use anyhow::Result;
use axum::{
    Json,
    body::Body,
    extract::{ConnectInfo, State},
    http::{HeaderMap, Request, StatusCode, header},
    middleware::Next,
    response::Response,
};
use bcrypt::{DEFAULT_COST, hash, verify};
use jsonwebtoken::{Algorithm, DecodingKey, EncodingKey, Header, Validation, decode, encode};
use serde::Deserialize;
use serde_json::json;

use crate::db::{count_users, get_user_by_id, get_user_by_username, insert_user, update_user_db};
use crate::types::{AppState, Claims, User};
use crate::util::{json_error, now_string, ok_json, real_ip, unix_now};

pub fn gen_token(state: &AppState, user: &User) -> Result<String> {
    let now = unix_now();
    let claims = Claims {
        sub: user.id,
        version: user.token_version,
        iat: now,
        exp: now + 7 * 24 * 3600,
    };
    Ok(encode(
        &Header::default(),
        &claims,
        &EncodingKey::from_secret(&state.jwt_secret),
    )?)
}

pub async fn auth_middleware(
    State(state): State<AppState>,
    mut req: Request<Body>,
    next: Next,
) -> Response {
    let token = req
        .headers()
        .get(header::AUTHORIZATION)
        .and_then(|v| v.to_str().ok())
        .and_then(|s| s.strip_prefix("Bearer ").map(str::to_string))
        .or_else(|| {
            req.uri().query().and_then(|q| {
                url::form_urlencoded::parse(q.as_bytes())
                    .find(|(k, _)| k == "token")
                    .map(|(_, v)| v.into_owned())
            })
        });
    let Some(token) = token else {
        return json_error(StatusCode::UNAUTHORIZED, "unauthorized");
    };
    let data = match decode::<Claims>(
        &token,
        &DecodingKey::from_secret(&state.jwt_secret),
        &Validation::new(Algorithm::HS256),
    ) {
        Ok(v) => v,
        Err(_) => return json_error(StatusCode::UNAUTHORIZED, "unauthorized"),
    };
    let user = match get_user_by_id(&state.db, data.claims.sub) {
        Ok(u) => u,
        Err(_) => return json_error(StatusCode::UNAUTHORIZED, "user not found"),
    };
    if user.token_version != data.claims.version {
        return json_error(
            StatusCode::UNAUTHORIZED,
            "token has been revoked, please login again",
        );
    }
    req.extensions_mut().insert(user);
    next.run(req).await
}

pub fn req_user(req: &Request<Body>) -> User {
    req.extensions().get::<User>().unwrap().clone()
}

// ── Handlers ──────────────────────────────────────────────────────────────────

pub async fn auth_status(State(state): State<AppState>) -> Response {
    let c = count_users(&state.db).unwrap_or(0);
    ok_json(json!({"setup": c > 0}))
}

#[derive(Deserialize)]
pub struct AuthReq {
    pub username: String,
    pub password: String,
}

impl AuthReq {
    /// 对齐 Go binding:"required"：username / password 均不得为空
    fn validate(&self) -> Option<Response> {
        if self.username.is_empty() || self.password.is_empty() {
            return Some(json_error(
                StatusCode::BAD_REQUEST,
                "Key: 'AuthReq.username' Error: field is required",
            ));
        }
        None
    }
}

pub async fn setup(
    State(state): State<AppState>,
    ConnectInfo(addr): ConnectInfo<std::net::SocketAddr>,
    headers: HeaderMap,
    Json(req): Json<AuthReq>,
) -> Response {
    // 防抢注：setup 与登录共用限流（5/min/IP），防止首次部署窗口被公网扫描器
    // 抢先注册管理员账号
    if !state.login_limiter.allow(real_ip(&headers, addr)).await {
        return json_error(
            StatusCode::TOO_MANY_REQUESTS,
            "too many requests, please try again later",
        );
    }
    if let Some(resp) = req.validate() {
        return resp;
    }
    let c = count_users(&state.db).unwrap_or(0);
    if c > 0 {
        return json_error(StatusCode::BAD_REQUEST, "already setup");
    }
    let password = req.password.clone();
    let h = match tokio::task::spawn_blocking(move || hash(password, DEFAULT_COST)).await {
        Ok(Ok(v)) => v,
        Ok(Err(_)) => {
            return json_error(StatusCode::INTERNAL_SERVER_ERROR, "failed to hash password")
        }
        Err(_) => return json_error(StatusCode::INTERNAL_SERVER_ERROR, "internal error"),
    };
    let c2 = count_users(&state.db).unwrap_or(0);
    if c2 > 0 {
        return json_error(StatusCode::BAD_REQUEST, "already setup");
    }
    let now = now_string();
    let id = match insert_user(&state.db, &req.username, &h, &now) {
        Ok(id) => id,
        Err(e) => return json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    };
    let user = get_user_by_id(&state.db, id).unwrap();
    let token = gen_token(&state, &user).unwrap();
    ok_json(json!({"token": token, "user": user}))
}

pub async fn login(
    State(state): State<AppState>,
    ConnectInfo(addr): ConnectInfo<std::net::SocketAddr>,
    headers: HeaderMap,
    Json(req): Json<AuthReq>,
) -> Response {
    // 限流在参数校验之前：无效请求同样计数（对齐 Go Login 首行 Allow）
    if !state.login_limiter.allow(real_ip(&headers, addr)).await {
        return json_error(
            StatusCode::TOO_MANY_REQUESTS,
            "too many login attempts, please try again later",
        );
    }
    if let Some(resp) = req.validate() {
        return resp;
    }
    let row = get_user_by_username(&state.db, &req.username).ok().flatten();
    let Some(row) = row else {
        return json_error(StatusCode::UNAUTHORIZED, "invalid credentials");
    };
    let user = row;
    let password = req.password.clone();
    let stored_hash = user.password.clone();
    let valid = tokio::task::spawn_blocking(move || verify(password, &stored_hash))
        .await
        .ok()
        .and_then(|r| r.ok())
        .unwrap_or(false);
    if !valid {
        return json_error(StatusCode::UNAUTHORIZED, "invalid credentials");
    }
    let token = gen_token(&state, &user).unwrap();
    ok_json(json!({"token": token, "user": user}))
}

pub async fn get_user(req: Request<Body>) -> Response {
    ok_json(json!(req_user(&req)))
}

#[derive(Deserialize)]
pub struct UpdateUserReq {
    pub username: Option<String>,
    pub password: Option<String>,
}

pub async fn update_user(State(state): State<AppState>, req0: Request<Body>) -> Response {
    let user = req_user(&req0);
    let bytes = axum::body::to_bytes(req0.into_body(), usize::MAX)
        .await
        .unwrap_or_default();
    // 对齐 Go ShouldBindJSON：无效请求体直接 400，不静默降级
    let req: UpdateUserReq = match serde_json::from_slice(&bytes) {
        Ok(v) => v,
        Err(e) => return json_error(StatusCode::BAD_REQUEST, e),
    };
    let mut version = user.token_version;
    let mut password = user.password.clone();
    let mut username = user.username.clone();
    if let Some(u) = req.username.filter(|s| !s.is_empty()) {
        username = u;
    }
    let password_changed = if let Some(p) = req.password.filter(|s| !s.is_empty()) {
        password = tokio::task::spawn_blocking(move || hash(p, DEFAULT_COST))
            .await
            .unwrap()
            .unwrap();
        version += 1;
        true
    } else {
        false
    };
    let now = now_string();
    if let Err(e) = update_user_db(&state.db, user.id, &username, &password, version, &now) {
        let msg = e.to_string();
        let code = if msg.contains("already taken") {
            StatusCode::BAD_REQUEST
        } else {
            StatusCode::INTERNAL_SERVER_ERROR
        };
        return json_error(code, msg);
    }
    let user = get_user_by_id(&state.db, user.id).unwrap();
    let mut resp = json!({"user": user});
    if password_changed {
        resp["token"] = json!(gen_token(&state, &user).unwrap());
    }
    ok_json(resp)
}
