use anyhow::Result;
use axum::{
    Router,
    body::Body,
    extract::DefaultBodyLimit,
    http::{Method, Request, StatusCode, header},
    middleware::{self, Next},
    response::{IntoResponse, Response},
    routing::{any, delete, get, post, put},
};
use clap::Parser;
use rand::RngCore;
use sha2::{Digest, Sha256};
use std::{path::PathBuf, sync::Arc, time::Duration};
use tokio::{fs, net::TcpListener};

mod auth;
mod db;
mod handler;
mod terminal;
mod types;
mod util;
mod webdav;

mod embedded_frontend {
    include!(concat!(env!("OUT_DIR"), "/embedded_frontend.rs"));
}

use auth::{
    auth_middleware, auth_status, get_user, login, setup, update_user,
};
use db::{get_settings, init_db, open_db, save_settings, set_jwt_secret_enc};
use handler::{
    batch_copy, batch_delete, batch_download, batch_move, compress_files,
    create_file, decompress_file, delete_file, detect_file_type, download_file,
    download_public_file, fetch_url, get_file_content, get_permission,
    get_settings_api, list_dir_tree, list_files, list_public_files, mkdir,
    move_file, copy_file, search_files, serve_public_file, set_permission,
    set_visibility, update_file_content, update_settings, upload_file,
    upload_folder,
};
use handler::{access_share, create_share, delete_share, download_share, list_shares, serve_share_raw};
use types::{AppState, FileManager, RateLimiter};
use util::decrypt_opt;
use webdav::webdav_handler;

/// 对齐 Go 版命令行参数：-config / -dir / -port / -host
#[derive(Parser, Debug)]
#[command(name = "cloudone", about = "CloudOne (Rust) - self-hosted file manager")]
struct Args {
    /// 数据配置目录 (存放 db、keys)
    #[arg(long)]
    config: Option<PathBuf>,
    /// 文件存储目录 (存放用户上传的文件)
    #[arg(long)]
    dir: Option<PathBuf>,
    /// 监听端口
    #[arg(long, default_value = "6677")]
    port: String,
    /// 监听地址
    #[arg(long, default_value = "0.0.0.0")]
    host: String,
}

#[derive(serde::Deserialize)]
struct WebdavReq {
    #[serde(default)]
    webdav_enabled: bool,
    #[serde(default)]
    webdav_sub_path: Option<String>,
    #[serde(default)]
    webdav_username: Option<String>,
    #[serde(default)]
    webdav_password: Option<String>,
}

#[tokio::main]
async fn main() -> Result<()> {
    let _ = rustls::crypto::ring::default_provider().install_default();

    let args = Args::parse();
    let (data_dir, mut storage_dir) = match (&args.config, &args.dir) {
        (Some(c), Some(d)) => (c.clone(), d.clone()),
        (Some(c), None) => (c.clone(), c.join("storage")),
        (None, Some(d)) => (PathBuf::from("./data"), d.clone()),
        (None, None) => (PathBuf::from("./data"), PathBuf::from("./data/storage")),
    };
    fs::create_dir_all(&data_dir).await?;
    fs::create_dir_all(&storage_dir).await?;

    let master_key_text = match std::env::var("CLOUDONE_MASTER_KEY") {
        Ok(v) => v,
        Err(_) => {
            let p = data_dir.join("master.key");
            match fs::read_to_string(&p).await {
                Ok(v) => v,
                Err(_) => {
                    let mut b = [0u8; 32];
                    rand::thread_rng().fill_bytes(&mut b);
                    let v = hex::encode(b);
                    fs::write(&p, &v).await?;
                    v
                }
            }
        }
    };
    let master_key: [u8; 32] = Sha256::digest(master_key_text.as_bytes()).into();

    let db = open_db(&data_dir.join("cloudone.redb"))?;
    init_db(&db)?;

    let mut settings = get_settings(&db)?;
    if let Some(dir) = args.dir {
        storage_dir = dir;
        settings.storage_dir = storage_dir.to_string_lossy().into_owned();
        save_settings(&db, &settings)?;
    } else if !settings.storage_dir.is_empty() {
        storage_dir = PathBuf::from(&settings.storage_dir);
    }
    fs::create_dir_all(&storage_dir).await?;
    settings = get_settings(&db)?;

    let jwt_secret = match std::env::var("CLOUDONE_JWT_SECRET") {
        Ok(v) => v,
        Err(_) => decrypt_opt(&master_key, &settings.jwt_secret_enc).unwrap_or_default(),
    };
    let jwt_secret = if jwt_secret.is_empty() {
        let mut b = [0u8; 32];
        rand::thread_rng().fill_bytes(&mut b);
        let s = hex::encode(b);
        let enc = util::encrypt(&master_key, &s)?;
        set_jwt_secret_enc(&db, &enc)?;
        s
    } else {
        jwt_secret
    };

    let state = AppState {
        db,
        files: Arc::new(tokio::sync::Mutex::new(FileManager::new(storage_dir))),
        jwt_secret: Arc::new(jwt_secret.into_bytes()),
        master_key: Arc::new(master_key),
        login_limiter: Arc::new(RateLimiter::new(5, Duration::from_secs(60))),
        webdav_limiter: Arc::new(RateLimiter::new(10, Duration::from_secs(60))),
        allowed_origin: std::env::var("CLOUDONE_ORIGIN").unwrap_or_default(),
    };
    state.login_limiter.spawn_cleanup();
    state.webdav_limiter.spawn_cleanup();

    let app = build_router(state);
    let addr = format!("{}:{}", args.host, args.port);
    eprintln!("CloudOne starting on {addr}");
    let listener = TcpListener::bind(&addr).await?;
    axum::serve(listener, app.into_make_service_with_connect_info::<std::net::SocketAddr>()).await?;
    Ok(())
}

fn build_router(state: AppState) -> Router {
    let authed = Router::new()
        .route("/user", get(get_user).put(update_user))
        .route("/settings", get(get_settings_api).put(update_settings))
        .route("/files", get(list_files).delete(delete_file))
        .route("/files/upload", post(upload_file))
        .route("/files/mkdir", post(mkdir))
        .route("/files/move", post(move_file))
        .route("/files/copy", post(copy_file))
        .route("/files/download", get(download_file))
        .route("/files/create", post(create_file))
        .route("/files/content", get(get_file_content).put(update_file_content))
        .route("/files/detect", get(detect_file_type))
        .route("/files/visibility", put(set_visibility))
        .route("/files/permission", get(get_permission).put(set_permission))
        .route("/files/batch-delete", post(batch_delete))
        .route("/files/batch-download", post(batch_download))
        .route("/files/batch-move", post(batch_move))
        .route("/files/batch-copy", post(batch_copy))
        .route("/files/compress", post(compress_files))
        .route("/files/decompress", post(decompress_file))
        .route("/files/fetch-url", post(fetch_url))
        .route("/files/search", get(search_files))
        .route("/files/upload-folder", post(upload_folder))
        .route("/files/dirtree", get(list_dir_tree))
        .route("/ws/terminal", get(terminal::terminal_ws))
        .route("/share", get(list_shares).post(create_share))
        .route("/share/{id}", delete(delete_share))
        .route("/webdav/settings", get(get_webdav_settings).put(update_webdav_settings))
        .route_layer(middleware::from_fn_with_state(state.clone(), auth_middleware));

    let api = Router::new()
        .route("/auth/status", get(auth_status))
        .route("/auth/setup", post(setup))
        .route("/auth/login", post(login))
        .route("/s/{code}", get(access_share))
        .route("/s/{code}/download", get(download_share))
        .merge(authed);

    Router::new()
        .nest("/api", api)
        .route("/public", get(list_public_files))
        .route("/pub/list", get(list_public_files))
        .route("/pub/dl", get(download_public_file))
        .route("/raw/{*path}", get(serve_public_file))
        .route("/s/{code}/raw", get(serve_share_raw))
        .route("/dav", any(webdav_handler))
        .route("/dav/", any(webdav_handler))
        .route("/dav/{*path}", any(webdav_handler))
        .route("/", get(serve_index))
        .route("/assets/{*path}", get(serve_assets))
        .fallback(static_fallback)
        .layer(middleware::from_fn_with_state(state.clone(), cors_middleware))
        .layer(DefaultBodyLimit::max(usize::MAX))
        .with_state(state)
}

// ── CORS（对齐 gin-contrib/cors + CLOUDONE_ORIGIN 语义）───────────────────────
//
// CLOUDONE_ORIGIN 未设置时仅允许无 Origin / "null" 的请求（即浏览器地址栏
// 直接访问，不允许任意站点跨域）；设置后只允许该 origin。
// 方法/头列表与 Go 版完全一致。

const CORS_ALLOW_METHODS: &str = "GET, POST, PUT, DELETE, OPTIONS, PROPFIND, PROPPATCH, MKCOL, COPY, MOVE, LOCK, UNLOCK, HEAD";
const CORS_ALLOW_HEADERS: &str = "Origin, Content-Type, Authorization, Depth, Destination, Lock-Token, Overwrite, Timeout, If";
const CORS_EXPOSE_HEADERS: &str = "Content-Disposition, DAV, Lock-Token";

fn origin_allowed(allowed: &str, origin: &str) -> bool {
    if allowed.is_empty() {
        origin.is_empty() || origin == "null"
    } else {
        origin == allowed
    }
}

async fn cors_middleware(
    axum::extract::State(state): axum::extract::State<AppState>,
    req: Request<Body>,
    next: Next,
) -> Response {
    // 对齐 gin-contrib/cors applyCors：
    // 1. 无 Origin → 非 CORS 请求，直接放行
    // 2. Origin 与 Host 同源（http(s)://{host}）→ 放行（浏览器同源 fetch 也会带 Origin）
    // 3. origin 校验失败 → 403（预检与实际请求均如此）
    // 4. OPTIONS → 预检 204；其余请求 → 加 CORS 头后继续
    let origin = req
        .headers()
        .get(header::ORIGIN)
        .and_then(|v| v.to_str().ok())
        .unwrap_or("")
        .to_string();
    if origin.is_empty() {
        return next.run(req).await;
    }
    let host = req
        .headers()
        .get(header::HOST)
        .and_then(|v| v.to_str().ok())
        .unwrap_or("")
        .to_string();
    if origin == format!("http://{host}") || origin == format!("https://{host}") {
        return next.run(req).await;
    }
    if !origin_allowed(&state.allowed_origin, &origin) {
        return StatusCode::FORBIDDEN.into_response();
    }

    if req.method() == Method::OPTIONS {
        let mut resp = StatusCode::NO_CONTENT.into_response();
        let h = resp.headers_mut();
        if let Ok(v) = header::HeaderValue::from_str(&origin) {
            h.insert(header::ACCESS_CONTROL_ALLOW_ORIGIN, v);
        }
        h.insert(header::ACCESS_CONTROL_ALLOW_METHODS, header::HeaderValue::from_static(CORS_ALLOW_METHODS));
        h.insert(header::ACCESS_CONTROL_ALLOW_HEADERS, header::HeaderValue::from_static(CORS_ALLOW_HEADERS));
        return resp;
    }

    let mut resp = next.run(req).await;
    if let Ok(v) = header::HeaderValue::from_str(&origin) {
        resp.headers_mut().insert(header::ACCESS_CONTROL_ALLOW_ORIGIN, v);
    }
    resp.headers_mut().insert(
        header::ACCESS_CONTROL_EXPOSE_HEADERS,
        header::HeaderValue::from_static(CORS_EXPOSE_HEADERS),
    );
    resp
}

// ── WebDAV settings handlers (kept in main to avoid circular deps) ────────────

async fn get_webdav_settings(axum::extract::State(state): axum::extract::State<AppState>) -> Response {
    let s = get_settings(&state.db).unwrap();
    util::ok_json(serde_json::json!({
        "webdav_enabled": s.webdav_enabled,
        "webdav_sub_path": s.webdav_sub_path,
        "webdav_username": s.webdav_username,
        "webdav_has_password": !s.webdav_password_enc.is_empty(),
    }))
}

async fn update_webdav_settings(
    axum::extract::State(state): axum::extract::State<AppState>,
    axum::Json(req): axum::Json<WebdavReq>,
) -> Response {
    let mut enc = None;
    if let Some(p) = req.webdav_password.filter(|s| !s.is_empty()) {
        let h = bcrypt::hash(p, bcrypt::DEFAULT_COST).unwrap();
        enc = Some(util::encrypt(&state.master_key, &h).unwrap());
    }
    let mut s = get_settings(&state.db).unwrap();
    s.webdav_enabled = req.webdav_enabled;
    s.webdav_sub_path = req.webdav_sub_path.unwrap_or_default();
    s.webdav_username = req.webdav_username.unwrap_or_default();
    if let Some(e) = enc {
        s.webdav_password_enc = e;
    }
    let _ = save_settings(&state.db, &s);
    get_webdav_settings(axum::extract::State(state)).await
}

// ── Static frontend（编译期内嵌，仅读内嵌资源，与 Go //go:embed 一致）──────────

async fn serve_index() -> Response {
    match embedded_frontend::get("index.html") {
        Some((data, mime)) => embedded_frontend_response(data, mime),
        None => StatusCode::INTERNAL_SERVER_ERROR.into_response(),
    }
}

async fn serve_assets(axum::extract::Path(path): axum::extract::Path<String>) -> Response {
    match embedded_frontend::get(&format!("assets/{path}")) {
        Some((data, mime)) => embedded_frontend_response(data, mime),
        None => StatusCode::NOT_FOUND.into_response(),
    }
}

async fn static_fallback(req: Request<Body>) -> Response {
    let path = req.uri().path().trim_start_matches('/').to_string();
    if !path.is_empty()
        && let Some((data, mime)) = embedded_frontend::get(&path)
    {
        return embedded_frontend_response(data, mime);
    }
    // SPA fallback：未命中的路由返回 index.html
    match embedded_frontend::get("index.html") {
        Some((data, mime)) => embedded_frontend_response(data, mime),
        None => StatusCode::INTERNAL_SERVER_ERROR.into_response(),
    }
}

fn embedded_frontend_response(data: &'static [u8], mime: &'static str) -> Response {
    use axum::http::HeaderValue;
    let mut resp = Response::new(Body::from(axum::body::Bytes::from_static(data)));
    resp.headers_mut().insert(header::CONTENT_TYPE, HeaderValue::from_static(mime));
    resp.headers_mut().insert(
        header::HeaderName::from_static("x-content-type-options"),
        HeaderValue::from_static("nosniff"),
    );
    resp
}
