use anyhow::anyhow;
use axum::{
    Json,
    extract::{ConnectInfo, Multipart, Path as AxPath, Query, State},
    http::StatusCode,
    response::{IntoResponse, Response},
};
use chrono::Utc;
use flate2::{Compression, read::GzDecoder, write::GzEncoder};
use futures_util::StreamExt;
use rand::{Rng, distributions::Alphanumeric};
use serde::Deserialize;
use serde_json::{Value, json};
use std::{
    io::{Read, Write},
    net::ToSocketAddrs,
    path::{Path, PathBuf},
    time::{Duration, Instant},
};
use tar::{Archive as TarArchive, Builder as TarBuilder};
use tokio::fs;
use tokio_util::io::ReaderStream;
use walkdir::WalkDir;
use axum::body::Bytes;
use axum::http::{HeaderValue, header};
use zip::{ZipArchive, ZipWriter, write::SimpleFileOptions};

use crate::auth::req_user;
use crate::db::{
    copy_visibility, delete_visibility_tree, file_info, get_settings, get_share, get_share_by_id,
    insert_share, is_public, list_dir, list_public_paths, list_shares_by_user, migrate_visibility,
    save_settings, upsert_visibility,
};
use crate::types::{AppState, Settings};
use crate::util::{
    copy_path, is_danger_path, is_private_ip, json_error, normalize_path, ok_json, path_clean,
    real_ip, remove_all, safe_name,
};

// ── Request types ─────────────────────────────────────────────────────────────

#[derive(Deserialize)]
pub struct PathQuery {
    pub path: Option<String>,
    pub name: Option<String>,
    pub dir: Option<String>,
    pub subpath: Option<String>,
}

#[derive(Deserialize)]
pub struct OnePath {
    pub path: String,
    pub content: Option<String>,
    pub mode: Option<u32>,
}

#[derive(Deserialize)]
pub struct MoveReq {
    pub src: String,
    pub dst: String,
}

#[derive(Deserialize)]
pub struct BatchPaths {
    pub paths: Vec<String>,
    pub target: Option<String>,
    pub filename: Option<String>,
}

#[derive(Deserialize)]
pub struct VisReq {
    pub path: String,
    pub is_public: bool,
}

#[derive(Deserialize)]
pub struct ShareReq {
    pub path: String,
    #[serde(default)]
    pub is_dir: bool,
    pub expire_in: Option<i64>,
    pub max_views: Option<i64>,
}

#[derive(Deserialize)]
pub struct CompressReq {
    pub paths: Vec<String>,
    pub format: Option<String>,
    pub output: Option<String>,
    pub dir: Option<String>,
}

#[derive(Deserialize)]
pub struct DecompressReq {
    pub path: String,
    pub dir: Option<String>,
}

#[derive(Deserialize)]
pub struct FetchReq {
    pub url: String,
    pub filename: Option<String>,
    pub dir: Option<String>,
}

/// 对齐 Go UpdateSettings 请求字段
#[derive(Deserialize)]
pub struct SettingsReq {
    pub storage_dir: Option<String>,
    pub lang: Option<String>,
    pub ui_theme: Option<String>,
    pub ui_font: Option<String>,
    pub editor_font: Option<String>,
    pub show_hidden: Option<bool>,
    pub github_proxy_enabled: Option<bool>,
    pub file_view_mode: Option<String>,
    pub file_sort_by: Option<String>,
    pub file_sort_order: Option<String>,
}

// ── Settings handlers ─────────────────────────────────────────────────────────

/// 对齐 Go safeSettings：不含 id 与加密字段
pub fn safe_settings_json(s: &Settings) -> Value {
    json!({
        "storage_dir": s.storage_dir,
        "lang": s.lang,
        "ui_theme": s.ui_theme,
        "ui_font": s.ui_font,
        "editor_font": s.editor_font,
        "webdav_enabled": s.webdav_enabled,
        "webdav_sub_path": s.webdav_sub_path,
        "webdav_username": s.webdav_username,
        "show_hidden": s.show_hidden,
        "file_view_mode": s.file_view_mode,
        "file_sort_by": s.file_sort_by,
        "file_sort_order": s.file_sort_order,
        "github_proxy_enabled": s.github_proxy_enabled,
    })
}

pub async fn get_settings_api(State(state): State<AppState>) -> Response {
    match get_settings(&state.db) {
        Ok(s) => ok_json(safe_settings_json(&s)),
        Err(e) => json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    }
}

pub async fn update_settings(
    State(state): State<AppState>,
    Json(req): Json<SettingsReq>,
) -> Response {
    let mut s = match get_settings(&state.db) {
        Ok(v) => v,
        Err(e) => return json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    };
    if let Some(d) = req.storage_dir.filter(|v| !v.is_empty()) {
        // 安全检查：禁止指向系统关键目录（对齐 Go：先 filepath.Clean 再比对）
        let clean = path_clean(&d);
        let forbidden = [
            "/etc", "/bin", "/sbin", "/usr/bin", "/usr/sbin",
            "/boot", "/sys", "/proc", "/dev", "/run", "/var/run",
        ];
        if forbidden.iter().any(|f| clean == *f || clean.starts_with(&format!("{f}/"))) {
            return json_error(StatusCode::BAD_REQUEST, "storage_dir points to a system directory");
        }
        if let Err(e) = fs::create_dir_all(&clean).await {
            return json_error(
                StatusCode::INTERNAL_SERVER_ERROR,
                format!("cannot create storage dir: {e}"),
            );
        }
        s.storage_dir = d.clone();
        state.files.lock().await.set_root(PathBuf::from(&clean));
    }
    if let Some(v) = req.lang.filter(|v| !v.is_empty()) { s.lang = v }
    if let Some(v) = req.ui_theme.filter(|v| !v.is_empty()) { s.ui_theme = v }
    if let Some(v) = req.ui_font.filter(|v| !v.is_empty()) { s.ui_font = v }
    if let Some(v) = req.editor_font.filter(|v| !v.is_empty()) { s.editor_font = v }
    if let Some(v) = req.show_hidden { s.show_hidden = v }
    if let Some(v) = req.github_proxy_enabled { s.github_proxy_enabled = v }
    if let Some(v) = req.file_view_mode.filter(|v| {
        ["list", "detail", "icon-large", "icon-small"].contains(&v.as_str())
    }) {
        s.file_view_mode = v;
    }
    if let Some(v) = req.file_sort_by.filter(|v| {
        ["name", "date", "type", "size"].contains(&v.as_str())
    }) {
        s.file_sort_by = v;
    }
    if let Some(v) = req.file_sort_order.filter(|v| v == "asc" || v == "desc") {
        s.file_sort_order = v;
    }
    if let Err(e) = save_settings(&state.db, &s) {
        return json_error(StatusCode::INTERNAL_SERVER_ERROR, e);
    }
    ok_json(safe_settings_json(&s))
}

// ── File handlers ─────────────────────────────────────────────────────────────

pub async fn list_files(State(state): State<AppState>, Query(q): Query<PathQuery>) -> Response {
    let p = q.path.unwrap_or("/".into());
    match list_dir(&state, &p).await {
        Ok(v) => ok_json(json!({"files": v, "path": p})),
        Err(e) => json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    }
}

pub async fn delete_file(State(state): State<AppState>, Query(q): Query<PathQuery>) -> Response {
    let p = match q.path {
        Some(p) if !p.is_empty() => p,
        _ => return json_error(StatusCode::BAD_REQUEST, "path is required"),
    };
    if p == "/" {
        return json_error(StatusCode::INTERNAL_SERVER_ERROR, "cannot delete root directory");
    }
    let fm = state.files.lock().await.clone();
    match fm.safe_abs_path(&p).and_then(|abs| remove_all(&abs).map_err(Into::into)) {
        Ok(_) => {
            delete_visibility_tree(&state.db, &p);
            ok_json(json!({"ok": true}))
        }
        Err(e) => json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    }
}

pub async fn write_rel(state: &AppState, rel: &str, data: impl AsRef<[u8]>) -> anyhow::Result<()> {
    let fm = state.files.lock().await.clone();
    let abs = fm.safe_abs_path(rel)?;
    if let Some(p) = abs.parent() {
        fs::create_dir_all(p).await?;
    }
    fs::write(abs, data).await?;
    Ok(())
}

pub async fn upload_file(
    State(state): State<AppState>,
    Query(q): Query<PathQuery>,
    mut mp: Multipart,
) -> Response {
    let dir = q.path.unwrap_or("/".into());
    let mut failed = Vec::new();
    while let Ok(Some(field)) = mp.next_field().await {
        if field.name() != Some("files") { continue; }
        let name = field.file_name().map(|s| s.to_string()).unwrap_or_default();
        let safe = safe_name(&name);
        // 对齐 Go：文件名净化后为 "." / ".." / "" 时计入 failed
        if safe.is_empty() || safe == "." || safe == ".." {
            failed.push(name);
            continue;
        }
        match field.bytes().await {
            Ok(b) => {
                if write_rel(&state, &format!("{dir}/{safe}"), b).await.is_err() {
                    failed.push(name)
                }
            }
            Err(_) => failed.push(name),
        }
    }
    if failed.is_empty() {
        ok_json(json!({"ok": true}))
    } else {
        (StatusCode::MULTI_STATUS, Json(json!({"ok": true, "failed": failed}))).into_response()
    }
}

pub async fn mkdir(State(state): State<AppState>, Json(req): Json<OnePath>) -> Response {
    let fm = state.files.lock().await.clone();
    match fm.safe_abs_path(&req.path).and_then(|p| std::fs::create_dir_all(p).map_err(Into::into)) {
        Ok(_) => ok_json(json!({"ok": true})),
        Err(e) => json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    }
}

pub async fn move_file(State(state): State<AppState>, Json(req): Json<MoveReq>) -> Response {
    move_one(&state, &req.src, &req.dst).await
}

pub async fn copy_file(State(state): State<AppState>, Json(req): Json<MoveReq>) -> Response {
    copy_one(&state, &req.src, &req.dst).await
}

async fn move_one(state: &AppState, src: &str, dst: &str) -> Response {
    let fm = state.files.lock().await.clone();
    match (fm.safe_abs_path(src), fm.safe_abs_path(dst)) {
        (Ok(s), Ok(d)) => match std::fs::rename(&s, &d) {
            Ok(_) => {
                migrate_visibility(&state.db, src, dst);
                ok_json(json!({"ok": true}))
            }
            Err(e) => json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
        },
        (Err(e), _) | (_, Err(e)) => json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    }
}

async fn copy_one(state: &AppState, src: &str, dst: &str) -> Response {
    let fm = state.files.lock().await.clone();
    match (fm.safe_abs_path(src), fm.safe_abs_path(dst)) {
        (Ok(s), Ok(d)) => match copy_path(&s, &d) {
            Ok(_) => {
                copy_visibility(&state.db, src, dst);
                ok_json(json!({"ok": true}))
            }
            Err(e) => json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
        },
        (Err(e), _) | (_, Err(e)) => json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    }
}

pub async fn download_file(State(state): State<AppState>, Query(q): Query<PathQuery>) -> Response {
    let p = q.path.unwrap_or_default();
    let fm = state.files.lock().await.clone();
    // 对齐 Go DownloadFile：使用 AbsPath（无 danger 前缀检查）
    serve_file_path(fm.abs_path(&p), true).await
}

pub async fn create_file(State(state): State<AppState>, Json(req): Json<OnePath>) -> Response {
    match write_rel(&state, &req.path, req.content.unwrap_or_default()).await {
        Ok(_) => ok_json(json!({"ok": true})),
        Err(e) => json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    }
}

pub async fn get_file_content(State(state): State<AppState>, Query(q): Query<PathQuery>) -> Response {
    let p = q.path.unwrap_or_default();
    let fm = state.files.lock().await.clone();
    match fm.safe_abs_path(&p).and_then(|abs| {
        let mut f = std::fs::File::open(abs)?;
        let mut buf = Vec::new();
        Read::by_ref(&mut f).take(2 * 1024 * 1024).read_to_end(&mut buf)?;
        Ok(String::from_utf8_lossy(&buf).to_string())
    }) {
        Ok(c) => ok_json(json!({"content": c})),
        Err(e) => json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    }
}

pub async fn update_file_content(State(state): State<AppState>, Json(req): Json<OnePath>) -> Response {
    create_file(State(state), Json(req)).await
}

pub async fn get_permission(State(state): State<AppState>, Query(q): Query<PathQuery>) -> Response {
    let fm = state.files.lock().await.clone();
    let p = q.path.unwrap_or_default();
    #[cfg(unix)]
    use std::os::unix::fs::PermissionsExt;
    match fm.safe_abs_path(&p).and_then(|abs| Ok(std::fs::metadata(abs)?.permissions())) {
        Ok(perm) => {
            #[cfg(unix)]
            let m = perm.mode() & 0o777;
            #[cfg(not(unix))]
            let m = 0;
            ok_json(json!({"mode": m}))
        }
        Err(e) => json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    }
}

pub async fn set_permission(State(state): State<AppState>, Json(req): Json<OnePath>) -> Response {
    #[cfg(unix)]
    use std::os::unix::fs::PermissionsExt;
    let fm = state.files.lock().await.clone();
    let mode = req.mode.unwrap_or(0o644);
    match fm.safe_abs_path(&req.path).and_then(|abs| {
        #[cfg(unix)]
        std::fs::set_permissions(abs, std::fs::Permissions::from_mode(mode))?;
        #[cfg(not(unix))]
        let _ = mode;
        Ok(())
    }) {
        Ok(_) => ok_json(json!({"ok": true})),
        Err(e) => json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    }
}

pub async fn set_visibility(State(state): State<AppState>, Json(req): Json<VisReq>) -> Response {
    if req.is_public {
        if let Err(e) = mark_public_recursive(&state, &req.path).await {
            return json_error(StatusCode::INTERNAL_SERVER_ERROR, e);
        }
    } else {
        delete_visibility_tree(&state.db, &req.path);
    }
    ok_json(json!({"ok": true}))
}

async fn mark_public_recursive(state: &AppState, rel: &str) -> anyhow::Result<()> {
    let fm = state.files.lock().await.clone();
    let abs = fm.safe_abs_path(rel)?;
    for e in WalkDir::new(&abs).into_iter().filter_map(Result::ok) {
        let r = e.path().strip_prefix(&fm.root).unwrap();
        let rp = normalize_path(&format!("/{}", r.to_string_lossy()));
        upsert_visibility(&state.db, &rp, true)?;
    }
    Ok(())
}

// ── Batch handlers ────────────────────────────────────────────────────────────

pub async fn batch_delete(State(state): State<AppState>, Json(req): Json<BatchPaths>) -> Response {
    let mut failed = Vec::new();
    for p in req.paths {
        if p.is_empty() || p == "/" { continue; }
        let fm = state.files.lock().await.clone();
        // 对齐 Go：Delete 成功才清理可见性；remove_all 对不存在的路径视为成功
        if let Ok(abs) = fm.safe_abs_path(&p) {
            match remove_all(&abs) {
                Ok(_) => delete_visibility_tree(&state.db, &p),
                Err(_) => failed.push(p.clone()),
            }
        } else {
            failed.push(p.clone());
        }
    }
    if failed.is_empty() {
        ok_json(json!({"ok": true}))
    } else {
        (StatusCode::MULTI_STATUS, Json(json!({"ok": false, "failed": failed}))).into_response()
    }
}

pub async fn batch_move(State(state): State<AppState>, Json(req): Json<BatchPaths>) -> Response {
    batch_move_copy(state, req, true).await
}

pub async fn batch_copy(State(state): State<AppState>, Json(req): Json<BatchPaths>) -> Response {
    batch_move_copy(state, req, false).await
}

async fn batch_move_copy(state: AppState, req: BatchPaths, mv: bool) -> Response {
    let target = match req.target {
        Some(t) if !t.is_empty() => t,
        _ => return json_error(StatusCode::BAD_REQUEST, "target is required"),
    };
    let verb = if mv { "move" } else { "copy" };
    let mut failed = Vec::new();
    let mut blocked = Vec::new();
    let tn = normalize_path(&target);
    for src in req.paths {
        let sn = normalize_path(&src);
        // 防止移动/复制到自身或子目录（会导致无限递归）
        if sn == tn || format!("{tn}/").starts_with(&format!("{sn}/")) {
            blocked.push(src);
            continue;
        }
        let dst = format!(
            "{}/{}",
            tn.trim_end_matches('/'),
            Path::new(&sn).file_name().unwrap_or_default().to_string_lossy()
        );
        let resp = if mv {
            move_one(&state, &src, &dst).await
        } else {
            copy_one(&state, &src, &dst).await
        };
        if !resp.status().is_success() { failed.push(src); }
    }
    if !blocked.is_empty() {
        return (StatusCode::BAD_REQUEST, Json(json!({
            "error": format!("cannot {verb} a folder into itself or its subdirectory"),
            "blocked": blocked,
        }))).into_response();
    }
    if failed.is_empty() {
        ok_json(json!({"ok": true}))
    } else {
        (StatusCode::MULTI_STATUS, Json(json!({"ok": false, "failed": failed}))).into_response()
    }
}

pub async fn batch_download(State(state): State<AppState>, Json(req): Json<BatchPaths>) -> Response {
    let fm = state.files.lock().await.clone();
    let fname = req.filename.unwrap_or("download".into());
    let fname = if fname.to_lowercase().ends_with(".zip") { fname } else { format!("{fname}.zip") };

    let entries: Vec<(PathBuf, String)> = req.paths.iter()
        .filter_map(|p| {
            fm.safe_abs_path(p).ok().map(|abs| {
                // 只使用最后一段名称作为 zip 内的根路径，避免打包出多层父目录
                let name = abs.file_name().unwrap_or_default().to_string_lossy().to_string();
                (abs, name)
            })
        })
        .collect();

    if entries.is_empty() {
        return json_error(StatusCode::BAD_REQUEST, "no valid paths");
    }

    // 流式写入响应（对齐 Go：无需临时文件、无大内存分配）
    let (tx, rx) = tokio::sync::mpsc::channel::<Result<Bytes, std::io::Error>>(32);
    tokio::task::spawn_blocking(move || {
        let writer = ChannelWriter { tx };
        let mut zw = ZipWriter::new_stream(writer);
        for (abs, name) in entries {
            let _ = zip_add_lstat(&mut zw, &abs, &name);
        }
        let _ = zw.finish();
    });

    let body = axum::body::Body::from_stream(tokio_stream::wrappers::ReceiverStream::new(rx));
    let mut resp = Response::new(body);
    *resp.status_mut() = StatusCode::OK;
    resp.headers_mut().insert(header::CONTENT_TYPE, HeaderValue::from_static("application/zip"));
    resp.headers_mut().insert(
        header::CONTENT_DISPOSITION,
        HeaderValue::from_str(&format!("attachment; filename=\"{fname}\"")).unwrap(),
    );
    resp
}

struct ChannelWriter {
    tx: tokio::sync::mpsc::Sender<Result<Bytes, std::io::Error>>,
}

impl Write for ChannelWriter {
    fn write(&mut self, buf: &[u8]) -> std::io::Result<usize> {
        self.tx
            .blocking_send(Ok(Bytes::copy_from_slice(buf)))
            .map_err(|_| std::io::Error::new(std::io::ErrorKind::BrokenPipe, "response channel closed"))?;
        Ok(buf.len())
    }
    fn flush(&mut self) -> std::io::Result<()> { Ok(()) }
}

// ── Public file handlers ──────────────────────────────────────────────────────

pub async fn list_public_files(State(state): State<AppState>, Query(q): Query<PathQuery>) -> Response {
    let p = q.path.unwrap_or("/".into());
    if p == "/" {
        // 对齐 Go GetAllPublicFlat：跳过父目录已公开的记录；文件不存在时清理脏记录
        let paths: Vec<String> = list_public_paths(&state.db).into_iter().collect();
        let public_set: std::collections::HashSet<String> = paths.iter().cloned().collect();
        let fm = state.files.lock().await.clone();
        let mut list = Vec::new();
        for rel in &paths {
            let parent = normalize_path(
                &Path::new(rel)
                    .parent()
                    .map(|p| p.to_string_lossy().to_string())
                    .unwrap_or_default(),
            );
            if public_set.contains(&parent) {
                continue;
            }
            if let Ok(abs) = fm.safe_abs_path(rel) {
                match file_info(&state.db, &abs, rel) {
                    Ok(fi) => list.push(fi),
                    Err(_) => delete_visibility_tree(&state.db, rel),
                }
            }
        }
        ok_json(json!({"files": list, "path": p}))
    } else {
        match list_dir(&state, &p).await {
            Ok(v) => ok_json(json!({"files": v.into_iter().filter(|f| f.is_public).collect::<Vec<_>>(), "path": p})),
            Err(_) => ok_json(json!({"files": [], "path": p})),
        }
    }
}

pub async fn serve_public_file(State(state): State<AppState>, AxPath(path): AxPath<String>) -> Response {
    let rel = format!("/{path}");
    if !is_public(&state.db, &rel) {
        return json_error(StatusCode::FORBIDDEN, "not public");
    }
    let fm = state.files.lock().await.clone();
    let abs = match fm.abs_path(&rel) {
        Ok(a) => a,
        Err(e) => return json_error(StatusCode::BAD_REQUEST, e),
    };
    if abs.is_dir() {
        match list_dir(&state, &rel).await {
            Ok(v) => ok_json(json!({"files": v.into_iter().filter(|f| f.is_public).collect::<Vec<_>>(), "path": rel})),
            Err(e) => json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
        }
    } else {
        serve_file_path(Ok(abs), false).await
    }
}

pub async fn download_public_file(State(state): State<AppState>, Query(q): Query<PathQuery>) -> Response {
    let p = match q.path {
        Some(p) if !p.is_empty() => p,
        _ => return json_error(StatusCode::BAD_REQUEST, "path is required"),
    };
    if !is_public(&state.db, &p) {
        return json_error(StatusCode::FORBIDDEN, "not public");
    }
    let fm = state.files.lock().await.clone();
    // 对齐 Go DownloadPublicFile：AbsPath（无 danger 前缀检查）
    serve_file_path(fm.abs_path(&p), true).await
}

// ── Share handlers ────────────────────────────────────────────────────────────

pub async fn create_share(State(state): State<AppState>, req0: axum::extract::Request) -> Response {
    let user = req_user(&req0);
    let b = axum::body::to_bytes(req0.into_body(), usize::MAX).await.unwrap_or_default();
    // 对齐 Go ShouldBindJSON：无效请求体 400，不 panic
    let req: ShareReq = match serde_json::from_slice(&b) {
        Ok(v) => v,
        Err(e) => return json_error(StatusCode::BAD_REQUEST, e),
    };
    if req.path.is_empty() {
        return json_error(StatusCode::BAD_REQUEST, "path is required");
    }
    let code: String = rand::thread_rng()
        .sample_iter(&Alphanumeric)
        .take(8)
        .map(char::from)
        .collect();
    let exp = req.expire_in.filter(|v| *v > 0)
        .map(|s| (Utc::now() + chrono::Duration::seconds(s)).to_rfc3339());
    let now = crate::util::now_string();
    match insert_share(
        &state.db,
        &code,
        &req.path,
        req.is_dir,
        user.id,
        exp,
        req.max_views.unwrap_or(0),
        &now,
    ) {
        Ok(id) => {
            let link = get_share_by_id(&state.db, id).unwrap();
            Json(link).into_response()
        }
        Err(e) => json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    }
}

pub async fn list_shares(State(state): State<AppState>, req: axum::extract::Request) -> Response {
    let user = req_user(&req);
    Json(list_shares_by_user(&state.db, user.id)).into_response()
}

pub async fn delete_share(
    State(state): State<AppState>,
    AxPath(id): AxPath<i64>,
    req: axum::extract::Request,
) -> Response {
    let user = req_user(&req);
    crate::db::delete_share(&state.db, id, user.id);
    ok_json(json!({"ok": true}))
}

pub async fn access_share(
    State(state): State<AppState>,
    ConnectInfo(addr): ConnectInfo<std::net::SocketAddr>,
    headers: axum::http::HeaderMap,
    AxPath(code): AxPath<String>,
    Query(q): Query<PathQuery>,
) -> Response {
    // 防分享码爆破枚举：同一 IP 每分钟最多 30 次访问
    if !state.share_limiter.allow(real_ip(&headers, addr)).await {
        return json_error(
            StatusCode::TOO_MANY_REQUESTS,
            "too many requests, please try again later",
        );
    }
    let link = match get_share(&state.db, &code, false) {
        Ok(l) => l,
        Err(_) => return json_error(StatusCode::NOT_FOUND, "not found"),
    };
    if link.is_dir {
        let sp = q.subpath.unwrap_or_default();
        let lp = if sp.is_empty() {
            link.file_path.clone()
        } else {
            format!("{}/{}", link.file_path, normalize_path(&sp).trim_start_matches('/'))
        };
        let list = list_dir(&state, &lp).await.unwrap_or_default();
        ok_json(json!({"files": list, "is_dir": true, "code": link.code, "file_path": link.file_path}))
    } else {
        ok_json(json!({"is_dir": false, "code": link.code, "file_path": link.file_path}))
    }
}

pub async fn download_share(
    State(state): State<AppState>,
    ConnectInfo(addr): ConnectInfo<std::net::SocketAddr>,
    headers: axum::http::HeaderMap,
    AxPath(code): AxPath<String>,
    Query(q): Query<PathQuery>,
) -> Response {
    if !state.share_limiter.allow(real_ip(&headers, addr)).await {
        return json_error(
            StatusCode::TOO_MANY_REQUESTS,
            "too many requests, please try again later",
        );
    }
    let link = match get_share(&state.db, &code, true) {
        Ok(l) => l,
        Err(_) => return json_error(StatusCode::NOT_FOUND, "not found"),
    };
    let fm = state.files.lock().await.clone();
    let path = if link.is_dir {
        let subpath = match q.subpath {
            Some(s) if !s.is_empty() => s,
            _ => return json_error(StatusCode::BAD_REQUEST, "subpath required for directory share"),
        };
        // 净化 subpath，防路径穿越
        let subpath = normalize_path(&subpath).trim_start_matches('/').to_string();
        let file_path = format!("{}/{}", link.file_path, subpath);

        // 校验最终路径仍在分享目录内（对齐 Go DownloadShare）
        let share_abs = match fm.abs_path(&link.file_path) {
            Ok(p) => p,
            Err(_) => return json_error(StatusCode::BAD_REQUEST, "invalid share path"),
        };
        let target_abs = match fm.abs_path(&file_path) {
            Ok(p) => p,
            Err(_) => return json_error(StatusCode::BAD_REQUEST, "invalid path"),
        };
        if target_abs != share_abs && !target_abs.starts_with(&share_abs) {
            return json_error(StatusCode::FORBIDDEN, "access denied");
        }
        file_path
    } else {
        link.file_path.clone()
    };
    // 对齐 Go：AbsPath（无 danger 前缀检查）
    serve_file_path(fm.abs_path(&path), true).await
}

pub async fn serve_share_raw(
    State(state): State<AppState>,
    ConnectInfo(addr): ConnectInfo<std::net::SocketAddr>,
    headers: axum::http::HeaderMap,
    AxPath(code): AxPath<String>,
) -> Response {
    if !state.share_limiter.allow(real_ip(&headers, addr)).await {
        return json_error(
            StatusCode::TOO_MANY_REQUESTS,
            "too many requests, please try again later",
        );
    }
    let link = match get_share(&state.db, &code, true) {
        Ok(l) => l,
        Err(_) => return json_error(StatusCode::NOT_FOUND, "not found"),
    };
    if link.is_dir {
        return json_error(StatusCode::BAD_REQUEST, "cannot view directory as raw");
    }
    let fm = state.files.lock().await.clone();
    serve_file_path(fm.abs_path(&link.file_path), false).await
}

// ── File serving ──────────────────────────────────────────────────────────────

fn ext_mime_override(ext: &str) -> Option<&'static str> {
    match ext {
        "md" | "markdown" | "yaml" | "yml" | "toml" | "ini" | "conf" | "env"
        | "log" | "sh" | "bash" | "zsh" | "fish" | "dockerfile"
        | "go" | "py" | "rs" | "rb" | "java" | "c" | "cpp" | "h"
        | "ts" | "tsx" | "jsx" | "vue" | "swift" | "kt" | "lua"
        | "r" | "sql" | "graphql" => Some("text/plain; charset=utf-8"),
        "avif" => Some("image/avif"),
        "webp" => Some("image/webp"),
        _ => None,
    }
}

fn is_browser_previewable(mime: &str) -> bool {
    if mime.starts_with("text/") || mime.starts_with("image/")
        || mime.starts_with("audio/") || mime.starts_with("video/")
    {
        return true;
    }
    matches!(
        mime,
        "application/pdf" | "application/json" | "application/javascript"
        | "application/x-javascript" | "application/xml" | "application/xhtml+xml"
        | "application/atom+xml" | "application/rss+xml" | "application/svg+xml"
    ) || mime.contains("xml") || mime.contains("javascript")
}

pub async fn serve_file_path(p: anyhow::Result<PathBuf>, attachment: bool) -> Response {
    let abs = match p {
        Ok(v) => v,
        Err(e) => return json_error(StatusCode::BAD_REQUEST, e),
    };
    if abs.is_dir() {
        return json_error(StatusCode::BAD_REQUEST, "cannot download a directory directly; use batch-download");
    }
    let file = match tokio::fs::File::open(&abs).await {
        Ok(f) => f,
        Err(_) => return json_error(StatusCode::NOT_FOUND, "not found"),
    };
    let meta = file.metadata().await.ok();
    let ext = abs.extension().and_then(|e| e.to_str()).unwrap_or("").to_lowercase();
    let mime = ext_mime_override(&ext)
        .map(|s| s.to_string())
        .unwrap_or_else(|| mime_guess::from_path(&abs).first_or_octet_stream().to_string());
    let name = abs.file_name().unwrap_or_default().to_string_lossy().into_owned();
    let base_mime = mime.split(';').next().unwrap_or("").trim();
    let disp = if attachment || !is_browser_previewable(base_mime) { "attachment" } else { "inline" };

    let stream = ReaderStream::new(file);
    let body = axum::body::Body::from_stream(stream);
    let mut resp = axum::response::Response::new(body);
    *resp.status_mut() = StatusCode::OK;
    if let Some(m) = meta {
        resp.headers_mut().insert(header::CONTENT_LENGTH, HeaderValue::from(m.len()));
    }
    resp.headers_mut().insert(header::CONTENT_TYPE, HeaderValue::from_str(&mime).unwrap());
    resp.headers_mut().insert(
        header::CONTENT_DISPOSITION,
        HeaderValue::from_str(&format!("{disp}; filename=\"{name}\"")).unwrap(),
    );
    resp.headers_mut().insert(
        header::HeaderName::from_static("x-content-type-options"),
        HeaderValue::from_static("nosniff"),
    );
    resp.headers_mut().insert(header::CACHE_CONTROL, HeaderValue::from_static("public, max-age=3600"));
    resp
}

// ── Misc file handlers ────────────────────────────────────────────────────────

/// 对齐 Go DetectFileType：读取前 512 字节做嗅探，
/// 含 NUL 字节视为二进制，否则视为文本；mime 返回嗅探结果。
pub async fn detect_file_type(State(state): State<AppState>, Query(q): Query<PathQuery>) -> Response {
    let fm = state.files.lock().await.clone();
    let p = q.path.unwrap_or_default();
    let abs = match fm.abs_path(&p) {
        Ok(a) => a,
        Err(e) => return json_error(StatusCode::BAD_REQUEST, e),
    };
    let mut f = match std::fs::File::open(&abs) {
        Ok(f) => f,
        Err(e) => return json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    };
    let mut buf = [0u8; 512];
    let n = match f.read(&mut buf) {
        Ok(n) => n,
        Err(e) => return json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    };
    let sample = &buf[..n];
    if n == 0 {
        // 空文件视为文本
        return ok_json(json!({"text": true, "mime": "text/plain"}));
    }
    let mime = detect_content_type(sample);
    let text = !sample.contains(&0);
    ok_json(json!({"text": text, "mime": mime}))
}

/// 对齐 Go http.DetectContentType 的精简实现（常见 magic signature + 文本判定）
fn detect_content_type(data: &[u8]) -> String {
    let s = data;
    let starts = |sig: &[u8]| s.starts_with(sig);
    let at = |off: usize, sig: &[u8]| s.len() >= off + sig.len() && &s[off..off + sig.len()] == sig;

    // BOM
    if starts(b"\xEF\xBB\xBF") { return "text/plain; charset=utf-8".into(); }
    if starts(b"\xFE\xFF") || starts(b"\xFF\xFE") { return "text/plain; charset=utf-16".into(); }

    // 精确签名
    if starts(b"%PDF-") { return "application/pdf".into(); }
    if starts(b"\x89PNG\r\n\x1a\n") { return "image/png".into(); }
    if starts(b"GIF87a") || starts(b"GIF89a") { return "image/gif".into(); }
    if starts(b"BM") { return "image/bmp".into(); }
    if starts(b"\x00\x00\x01\x00") { return "image/x-icon".into(); }
    if starts(b"PK\x03\x04") { return "application/zip".into(); }
    if starts(b"\x1f\x8b\x08") { return "application/x-gzip".into(); }
    if starts(b"Rar!\x1a\x07\x00") { return "application/x-rar-compressed".into(); }
    if starts(b"\x00asm") { return "application/wasm".into(); }
    if starts(b"ID3") { return "audio/mpeg".into(); }
    if starts(b"\x1a\x45\xdf\xa3") { return "video/webm".into(); }
    if starts(b"wOFF") { return "font/woff".into(); }
    if starts(b"wOF2") { return "font/woff2".into(); }
    if starts(b"OTTO") { return "font/otf".into(); }
    if starts(b"\x00\x01\x00\x00") { return "font/ttf".into(); }
    if starts(b"%!") { return "application/postscript".into(); }
    if starts(b"{\\rtf") { return "text/rtf".into(); }
    if starts(b"RIFF") {
        if at(8, b"WEBP") { return "image/webp".into(); }
        if at(8, b"AVI ") { return "video/avi".into(); }
        if at(8, b"WAVE") { return "audio/wave".into(); }
    }
    if s.len() >= 12 && &s[4..8] == b"ftyp" {
        return "video/mp4".into();
    }
    if starts(b"\xD0\xCF\x11\xE0\xA1\xB1\x1A\xE1") {
        return "application/vnd.ms-office".into();
    }

    // HTML（跳过空白后大小写不敏感前缀匹配）
    let trimmed: &[u8] = {
        let mut t = s;
        while let Some((f, rest)) = t.split_first() {
            if f.is_ascii_whitespace() { t = rest; } else { break; }
        }
        t
    };
    let lower: Vec<u8> = trimmed.iter().map(|b| b.to_ascii_lowercase()).collect();
    let lstarts = |sig: &str| lower.starts_with(sig.as_bytes());
    if lstarts("<!doctype html")
        || lstarts("<html") || lstarts("<head") || lstarts("<script")
        || lstarts("<iframe") || lstarts("<h1") || lstarts("<div") || lstarts("<font")
        || lstarts("<table") || lstarts("<a") || lstarts("<style") || lstarts("<title")
        || lstarts("<b") || lstarts("<body") || lstarts("<br") || lstarts("<p")
        || lstarts("<!--")
    {
        return "text/html; charset=utf-8".into();
    }
    if lstarts("<?xml") {
        return "text/xml; charset=utf-8".into();
    }

    if s.contains(&0) {
        "application/octet-stream".into()
    } else {
        "text/plain; charset=utf-8".into()
    }
}

/// 对齐 Go SearchFiles：跳过危险目录，最多 200 条
pub async fn search_files(State(state): State<AppState>, Query(q): Query<PathQuery>) -> Response {
    let Some(name) = q.name.filter(|n| !n.is_empty()) else {
        return json_error(StatusCode::BAD_REQUEST, "name is required");
    };
    let dir = q.dir.unwrap_or("/".into());
    let fm = state.files.lock().await.clone();
    let abs = match fm.safe_abs_path(&dir) {
        Ok(a) => a,
        Err(e) => return json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    };
    let root = fm.root();
    let kw = name.to_lowercase();
    let mut res = Vec::new();
    for e in WalkDir::new(abs)
        .into_iter()
        .filter_entry(|e| !is_danger_path(e.path()))
        .filter_map(Result::ok)
        .skip(1)
    {
        if res.len() >= 200 { break; }
        if e.file_name().to_string_lossy().to_lowercase().contains(&kw) {
            let rel = e.path().strip_prefix(&root).unwrap();
            res.push(json!({
                "name": e.file_name().to_string_lossy(),
                "path": normalize_path(&format!("/{}", rel.to_string_lossy())),
                "is_dir": e.file_type().is_dir(),
            }));
        }
    }
    ok_json(json!({"results": res}))
}

pub async fn list_dir_tree(State(state): State<AppState>, Query(q): Query<PathQuery>) -> Response {
    let p = q.path.unwrap_or("/".into());
    match list_dir(&state, &p).await {
        Ok(v) => ok_json(json!({"dirs": v.into_iter().filter(|x| x.is_dir).map(|x| json!({"name": x.name, "path": x.path, "is_dir": true})).collect::<Vec<_>>(), "path": p})),
        Err(e) => json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    }
}

/// 对齐 Go UploadFolder：支持 files + paths 字段按顺序配对（保留文件夹结构）
pub async fn upload_folder(
    State(state): State<AppState>,
    Query(q): Query<PathQuery>,
    mut mp: Multipart,
) -> Response {
    let dir = q.path.unwrap_or("/".into());
    let mut files: Vec<(String, anyhow::Result<Bytes>)> = Vec::new();
    let mut paths: Vec<String> = Vec::new();
    while let Ok(Some(field)) = mp.next_field().await {
        match field.name() {
            Some("files") => {
                let name = field.file_name().map(|s| s.to_string()).unwrap_or_default();
                let data = field.bytes().await.map_err(|e| anyhow!("{e}"));
                files.push((name, data));
            }
            Some("paths") => {
                if let Ok(v) = field.text().await {
                    paths.push(v);
                }
            }
            _ => {}
        }
    }

    let mut failed = Vec::new();
    for (i, (name, data)) in files.into_iter().enumerate() {
        let mut rel = name;
        if let Some(p) = paths.get(i).filter(|p| !p.is_empty()) {
            rel = p.clone();
        }
        // 净化路径，防止路径穿越
        let clean = normalize_path(&rel).trim_start_matches('/').to_string();
        if clean.is_empty() {
            failed.push(rel);
            continue;
        }
        match data {
            Ok(b) => {
                if write_rel(&state, &format!("{dir}/{clean}"), b).await.is_err() {
                    failed.push(clean);
                }
            }
            Err(_) => failed.push(clean),
        }
    }
    if failed.is_empty() {
        ok_json(json!({"ok": true}))
    } else {
        (StatusCode::MULTI_STATUS, Json(json!({"ok": true, "failed": failed}))).into_response()
    }
}

// ── Compress / Decompress ─────────────────────────────────────────────────────

pub async fn compress_files(State(state): State<AppState>, Json(mut req): Json<CompressReq>) -> Response {
    let fmt_str = req.format.take().unwrap_or("zip".into());
    let dir = req.dir.unwrap_or("/".into());
    let mut out = req.output.unwrap_or("archive".into());
    let fmt = match fmt_str.as_str() {
        "tar" => { if !out.ends_with(".tar") { out.push_str(".tar") } "tar" }
        "tar.gz" => { if !out.ends_with(".tar.gz") { out.push_str(".tar.gz") } "tar.gz" }
        _ => { if !out.ends_with(".zip") { out.push_str(".zip") } "zip" }
    };
    let fm = state.files.lock().await.clone();

    struct SrcEntry { abs: PathBuf, arc_name: String }
    let root = fm.root();
    let mut entries: Vec<SrcEntry> = Vec::new();
    for p in &req.paths {
        // 对齐 Go：使用 AbsPath，压缩包内路径为相对 storageRoot 的完整路径
        if let Ok(abs) = fm.abs_path(p) {
            let arc_name = abs
                .strip_prefix(&root)
                .map(|r| r.to_string_lossy().replace('\\', "/"))
                .unwrap_or_else(|_| abs.file_name().unwrap_or_default().to_string_lossy().to_string());
            entries.push(SrcEntry { abs, arc_name });
        }
    }
    if entries.is_empty() {
        return json_error(StatusCode::BAD_REQUEST, "no valid paths");
    }
    let out_abs = match fm.abs_path(&format!("{dir}/{out}")) {
        Ok(p) => p,
        Err(e) => return json_error(StatusCode::BAD_REQUEST, e),
    };
    if let Some(p) = out_abs.parent() { let _ = std::fs::create_dir_all(p); }

    let out_clone = out.clone();
    let join_result = tokio::task::spawn_blocking(move || -> anyhow::Result<()> {
        let file = std::fs::File::create(&out_abs)?;
        match fmt {
            "zip" => {
                let mut zw = ZipWriter::new(file);
                for e in entries { zip_add_lstat(&mut zw, &e.abs, &e.arc_name)?; }
                zw.finish()?;
            }
            _ => {
                if fmt == "tar.gz" {
                    let enc = GzEncoder::new(file, Compression::default());
                    let mut tw = TarBuilder::new(enc);
                    tw.follow_symlinks(false);
                    for e in entries { tar_add(&mut tw, &e.abs, &e.arc_name)?; }
                    tw.into_inner()?.finish()?;
                } else {
                    let mut tw = TarBuilder::new(file);
                    tw.follow_symlinks(false);
                    for e in entries { tar_add(&mut tw, &e.abs, &e.arc_name)?; }
                    tw.finish()?;
                }
            }
        }
        Ok(())
    });
    let result = match join_result.await {
        Ok(r) => r,
        Err(e) => return json_error(StatusCode::INTERNAL_SERVER_ERROR, anyhow!(e)),
    };
    match result {
        Ok(_) => ok_json(json!({"ok": true, "output": out_clone})),
        Err(e) => json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    }
}

pub async fn decompress_file(State(state): State<AppState>, Json(req): Json<DecompressReq>) -> Response {
    let fm = state.files.lock().await.clone();
    // 对齐 Go：AbsPath（无 danger 前缀检查）
    let abs = match fm.abs_path(&req.path) {
        Ok(p) => p,
        Err(e) => return json_error(StatusCode::BAD_REQUEST, e),
    };
    let dest_rel = req.dir.unwrap_or_else(|| {
        Path::new(&req.path).parent()
            .map(|p| { let s = p.to_string_lossy(); if s.is_empty() || s == "." { "/".into() } else { s.into_owned() } })
            .unwrap_or_else(|| "/".into())
    });
    let dest = match fm.abs_path(&dest_rel) {
        Ok(p) => p,
        Err(e) => return json_error(StatusCode::BAD_REQUEST, e),
    };
    let name = abs.file_name().unwrap_or_default().to_string_lossy().to_lowercase();

    fn safe_extract(dest: &Path, entry_name: &str, reader: &mut dyn Read, mode: u32, is_dir: bool) -> std::io::Result<()> {
        use std::os::unix::fs::PermissionsExt;
        // 清理路径，防止 ../ 穿越（对齐 Go safeExtract）
        let clean = dest.join(Path::new(&format!("/{}", entry_name)).components()
            .filter(|c| matches!(c, std::path::Component::Normal(_)))
            .collect::<PathBuf>());
        if !clean.starts_with(dest) || clean == dest { return Ok(()); }
        if is_dir { return std::fs::create_dir_all(&clean); }
        if let Some(parent) = clean.parent() { std::fs::create_dir_all(parent)?; }
        let perm = std::fs::Permissions::from_mode((mode & 0o777) | 0o600);
        let mut f = std::fs::OpenOptions::new().create(true).write(true).truncate(true).open(&clean)?;
        f.set_permissions(perm)?;
        std::io::copy(reader, &mut f)?;
        Ok(())
    }

    let result_join = tokio::task::spawn_blocking(move || -> anyhow::Result<()> {
        std::fs::create_dir_all(&dest)?;
        if name.ends_with(".zip") {
            let f = std::fs::File::open(&abs)?;
            let mut za = ZipArchive::new(f)?;
            for i in 0..za.len() {
                let mut zf = za.by_index(i)?;
                let entry_name = zf.name().to_string();
                let is_dir = zf.is_dir();
                let mode = zf.unix_mode().unwrap_or(0o644);
                safe_extract(&dest, &entry_name, &mut zf, mode, is_dir)?;
            }
        } else if name.ends_with(".tar.gz") || name.ends_with(".tgz") {
            let f = std::fs::File::open(&abs)?;
            let mut tr = TarArchive::new(GzDecoder::new(f));
            for entry in tr.entries()? {
                let mut entry = entry?;
                let hdr = entry.header().clone();
                let entry_path = hdr.path()?.to_string_lossy().into_owned();
                let is_dir = hdr.entry_type().is_dir();
                let mode = hdr.mode().unwrap_or(0o644);
                safe_extract(&dest, &entry_path, &mut entry, mode, is_dir)?;
            }
        } else if name.ends_with(".tar") {
            let f = std::fs::File::open(&abs)?;
            let mut tr = TarArchive::new(f);
            for entry in tr.entries()? {
                let mut entry = entry?;
                let hdr = entry.header().clone();
                let entry_path = hdr.path()?.to_string_lossy().into_owned();
                let is_dir = hdr.entry_type().is_dir();
                let mode = hdr.mode().unwrap_or(0o644);
                safe_extract(&dest, &entry_path, &mut entry, mode, is_dir)?;
            }
        } else if name.ends_with(".gz") {
            // 单文件 .gz
            let f = std::fs::File::open(&abs)?;
            let mut gz = GzDecoder::new(f);
            let out_name = abs.file_stem().unwrap_or_default().to_string_lossy().into_owned();
            safe_extract(&dest, &out_name, &mut gz, 0o644, false)?;
        } else {
            return Err(anyhow!("unsupported archive format"));
        }
        Ok(())
    });
    let result: anyhow::Result<()> = match result_join.await {
        Ok(r) => r.map_err(|e| anyhow!(e)),
        Err(e) => Err(anyhow!(e)),
    };
    match result {
        Ok(_) => ok_json(json!({"ok": true})),
        Err(e) => json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
    }
}

// ── URL fetch（SSRF 防护 + GitHub 代理可选加速，对齐 Go FetchURL）──────────────

/// 对齐 Go githubProxyPrefixes（8 个，按探测延迟择优，开启加速时不直连）
const GITHUB_PROXY_PREFIXES: &[&str] = &[
    "https://ghfast.top/",
    "https://gh-proxy.com/",
    "https://gh.ddlc.top/",
    "https://ghproxy.it/",
    "https://ghproxy.net/",
    "https://gh.con.sh/",
    "https://gh-proxy.net/",
    "https://cdn.gh-proxy.com/",
];

fn is_github_url(raw: &str) -> bool {
    let github_hosts = [
        "github.com", "raw.githubusercontent.com", "objects.githubusercontent.com",
        "codeload.github.com", "releases.githubusercontent.com",
        "gist.githubusercontent.com", "gist.github.com",
    ];
    if let Ok(u) = url::Url::parse(raw)
        && let Some(host) = u.host_str()
    {
        let h = host.to_lowercase();
        for gh in &github_hosts {
            if h == *gh || h.ends_with(&format!(".{}", gh)) { return true; }
        }
    }
    false
}

/// 对齐 Go validatePublicURL：DNS 解析后校验所有 IP，防 DNS 重绑定
fn validate_public_url_sync(raw: &str) -> anyhow::Result<()> {
    let u = url::Url::parse(raw)?;
    if u.scheme() != "http" && u.scheme() != "https" {
        return Err(anyhow!("only http/https URLs are allowed"));
    }
    let host = u.host_str().ok_or_else(|| anyhow!("missing host"))?;
    if host.eq_ignore_ascii_case("localhost") {
        return Err(anyhow!("requests to private/internal addresses are not allowed"));
    }
    let port = u.port_or_known_default().unwrap_or(80);
    let addrs: Vec<std::net::SocketAddr> = (host, port)
        .to_socket_addrs()
        .map_err(|e| anyhow!("cannot resolve host: {e}"))?
        .collect();
    if addrs.is_empty() {
        return Err(anyhow!("cannot resolve host"));
    }
    for addr in addrs {
        if is_private_ip(addr.ip()) {
            return Err(anyhow!("requests to private/internal addresses are not allowed"));
        }
    }
    Ok(())
}

/// 对齐 Go rankGithubProxies：并发探测所有代理前缀（HEAD 失败退化为
/// Range GET，4s 超时），按延迟从低到高排序，全部失败返回空。
async fn rank_github_proxies(sample_url: &str) -> Vec<String> {
    let probe_client = match reqwest::Client::builder()
        .timeout(Duration::from_secs(4))
        .redirect(reqwest::redirect::Policy::none())
        .build()
    {
        Ok(c) => c,
        Err(_) => return Vec::new(),
    };
    let mut handles = Vec::new();
    for prefix in GITHUB_PROXY_PREFIXES {
        let client = probe_client.clone();
        let target = format!("{prefix}{sample_url}");
        handles.push(tokio::spawn(async move {
            let start = Instant::now();
            let ok = match client.head(&target).send().await {
                Ok(r) => r.status().as_u16() < 400,
                // 部分代理不支持 HEAD，退化为小范围 GET 探测
                Err(_) => match client.get(&target).header("Range", "bytes=0-0").send().await {
                    Ok(r) => r.status().as_u16() < 400,
                    Err(_) => false,
                },
            };
            (prefix.to_string(), start.elapsed(), ok)
        }));
    }
    let mut results = Vec::new();
    for h in handles {
        if let Ok((p, lat, true)) = h.await {
            results.push((p, lat));
        }
    }
    results.sort_by_key(|(_, lat)| *lat);
    results.into_iter().map(|(p, _)| p).collect()
}

pub async fn fetch_url(State(state): State<AppState>, Json(req): Json<FetchReq>) -> Response {
    if let Err(e) = validate_public_url_sync(&req.url) {
        return json_error(StatusCode::BAD_REQUEST, e);
    }
    let fm = state.files.lock().await.clone();
    // 对齐 Go：AbsPath（无 danger 前缀检查）
    let dir = match fm.abs_path(&req.dir.clone().unwrap_or("/".into())) {
        Ok(p) => p,
        Err(e) => return json_error(StatusCode::BAD_REQUEST, e),
    };
    let _ = fs::create_dir_all(&dir).await;

    let client = reqwest::Client::builder()
        .timeout(Duration::from_secs(30 * 60))
        // 对齐 Go CheckRedirect：限制 10 次跳转且每个跳转 URL 都过 SSRF 校验
        .redirect(reqwest::redirect::Policy::custom(|attempt| {
            if attempt.previous().len() >= 10 {
                return attempt.error("too many redirects");
            }
            match validate_public_url_sync(attempt.url().as_str()) {
                Ok(()) => attempt.follow(),
                Err(e) => attempt.error(e),
            }
        }))
        .build()
        .unwrap();

    // 读取"GitHub 加速下载"开关：开启且是 GitHub 链接时才走代理（不直连）；
    // 关闭（默认）时始终直连，完全不使用代理。
    let github_proxy_enabled = get_settings(&state.db)
        .map(|s| s.github_proxy_enabled)
        .unwrap_or(false);
    let mut urls_to_try: Vec<String> = vec![req.url.clone()];
    if github_proxy_enabled && is_github_url(&req.url) {
        let ranked = rank_github_proxies(&req.url).await;
        if ranked.is_empty() {
            return json_error(
                StatusCode::INTERNAL_SERVER_ERROR,
                "no available github proxy, please try again later or disable acceleration",
            );
        }
        urls_to_try = ranked
            .into_iter()
            .map(|prefix| format!("{prefix}{}", req.url))
            .collect();
    }

    // 确定保存文件名（从原始 URL 推断，避免代理前缀影响）。
    // 对齐 Go：先判空串（前端"留空"提交的是 "" 而非 null），
    // 再 filepath.Base(url.Path)——url.Path 已百分号解码、Base 会去掉尾部斜杠。
    let fname = req
        .filename
        .clone()
        .filter(|s| !s.is_empty())
        .unwrap_or_else(|| infer_filename_from_url(&req.url));
    let out = dir.join(safe_name(&fname));

    let mut last_err = String::new();
    for try_url in &urls_to_try {
        match client.get(try_url).send().await {
            Ok(r) if r.status().as_u16() < 400 => {
                match tokio::fs::File::create(&out).await {
                    Err(e) => return json_error(StatusCode::INTERNAL_SERVER_ERROR, e),
                    Ok(mut file) => {
                        let mut stream = r.bytes_stream();
                        let mut ok = true;
                        while let Some(chunk) = stream.next().await {
                            match chunk {
                                Ok(bytes) => {
                                    if let Err(e) = tokio::io::AsyncWriteExt::write_all(&mut file, &bytes).await {
                                        last_err = format!("download interrupted [{try_url}]: {e}");
                                        ok = false;
                                        break;
                                    }
                                }
                                Err(e) => { last_err = format!("download interrupted [{try_url}]: {e}"); ok = false; break; }
                            }
                        }
                        if ok { return ok_json(json!({"ok": true})); }
                        let _ = tokio::fs::remove_file(&out).await;
                    }
                }
            }
            Ok(r) => { last_err = format!("remote returned {} [{}]", r.status(), try_url); }
            Err(e) => { last_err = format!("fetch failed [{}]: {}", try_url, e); }
        }
    }
    json_error(StatusCode::INTERNAL_SERVER_ERROR, last_err)
}

/// 对齐 Go FetchURL：filename 留空时从原始 URL 推断保存名。
/// 等价于 filepath.Base(u.Path)：u.Path 为百分号解码后的路径，
/// Base 先去掉全部尾部斜杠再取最后一段；取不到时回落 "download"。
/// 注：Go 对 ".." 会放行（Join 后越界写父目录），此处回落 "download" 更安全。
fn infer_filename_from_url(raw: &str) -> String {
    let Ok(u) = url::Url::parse(raw) else {
        return "download".into();
    };
    let path = percent_encoding::percent_decode_str(u.path()).decode_utf8_lossy();
    let base = path.trim_end_matches('/').rsplit('/').next().unwrap_or("");
    if base.is_empty() || base == "." || base == ".." {
        "download".into()
    } else {
        base.to_string()
    }
}

// ── Archive helpers ───────────────────────────────────────────────────────────

/// 对齐 Go addToZip：目录递归、无法访问的文件跳过、Deflate 压缩
pub fn zip_add_lstat<W: Write + std::io::Seek>(
    zw: &mut ZipWriter<W>,
    abs: &Path,
    name: &str,
) -> zip::result::ZipResult<()> {
    use std::os::unix::fs::PermissionsExt;
    let info = match abs.symlink_metadata() {
        Ok(m) => m,
        Err(_) => return Ok(()),
    };
    if info.is_dir() {
        for e in std::fs::read_dir(abs)? {
            let e = e?;
            zip_add_lstat(zw, &e.path(), &format!("{}/{}", name, e.file_name().to_string_lossy()))?;
        }
    } else if info.is_file() {
        let opts = SimpleFileOptions::default()
            .compression_method(zip::CompressionMethod::Deflated)
            .unix_permissions(info.permissions().mode());
        zw.start_file(name, opts)?;
        let mut f = std::fs::File::open(abs)?;
        std::io::copy(&mut f, zw)?;
    }
    Ok(())
}

fn tar_add<W: Write>(tw: &mut TarBuilder<W>, abs: &Path, name: &str) -> std::io::Result<()> {
    let info = abs.symlink_metadata()?;
    if info.is_dir() {
        let mut hdr = tar::Header::new_gnu();
        hdr.set_metadata(&info);
        hdr.set_path(format!("{}/", name))?;
        hdr.set_cksum();
        tw.append(&hdr, std::io::empty())?;
        for e in std::fs::read_dir(abs)? {
            let e = e?;
            tar_add(tw, &e.path(), &format!("{}/{}", name, e.file_name().to_string_lossy()))?;
        }
    } else if info.is_file() {
        let mut hdr = tar::Header::new_gnu();
        hdr.set_metadata(&info);
        hdr.set_path(name)?;
        hdr.set_cksum();
        let mut f = std::fs::File::open(abs)?;
        tw.append(&hdr, &mut f)?;
    }
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::infer_filename_from_url;

    #[test]
    fn infer_filename_matches_go() {
        // 常规：取最后一段
        assert_eq!(infer_filename_from_url("https://host/file.zip"), "file.zip");
        assert_eq!(
            infer_filename_from_url("https://github.com/git/git/archive/refs/tags/v2.0.zip"),
            "v2.0.zip"
        );
        // query 忽略
        assert_eq!(
            infer_filename_from_url("https://host/file.zip?token=x"),
            "file.zip"
        );
        // 尾部斜杠：对齐 Go filepath.Base，取目录名
        assert_eq!(infer_filename_from_url("https://host/path/"), "path");
        assert_eq!(infer_filename_from_url("https://host/a/b//"), "b");
        // 百分号解码：对齐 Go url.Parse 的 u.Path
        assert_eq!(
            infer_filename_from_url("https://host/a/b/my%20file.zip"),
            "my file.zip"
        );
        // 无路径 / 根路径 / 危险段 → download
        assert_eq!(infer_filename_from_url("https://host"), "download");
        assert_eq!(infer_filename_from_url("https://host/"), "download");
        assert_eq!(infer_filename_from_url("https://host/a/.."), "download");
        assert_eq!(infer_filename_from_url("not a url"), "download");
    }
}
