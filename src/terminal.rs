use axum::{
    extract::{Query, State, WebSocketUpgrade, ws::{Message, WebSocket}},
    http::StatusCode,
    response::Response,
};
use futures_util::{SinkExt, StreamExt};
use jsonwebtoken::{Algorithm, DecodingKey, Validation, decode};
use serde_json::{Value, json};
use std::{
    ffi::OsString,
    os::unix::io::{FromRawFd, RawFd},
    path::Path,
    process::Stdio,
};
use tokio::process::Command;

use crate::db::get_user_by_id;
use crate::types::{AppState, Claims};
use crate::util::json_error;

/// 对齐 Go TerminalWebSocket：JWT 经 ?token= 查询参数校验（authed 路由已由
/// 中间件校验，这里为双保险），支持 ?cwd= 指定初始工作目录。
pub async fn terminal_ws(
    State(state): State<AppState>,
    Query(q): Query<std::collections::HashMap<String, String>>,
    ws: WebSocketUpgrade,
) -> Response {
    let token_str = q.get("token").cloned().unwrap_or_default();
    if token_str.is_empty() {
        return json_error(StatusCode::UNAUTHORIZED, "unauthorized");
    }
    let data = match decode::<Claims>(
        &token_str,
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
    // 前端会带上当前浏览的虚拟目录（?cwd=/etc），换算为磁盘真实路径，
    // 让终端打开时就位于用户当前所在目录
    let cwd = {
        let fm = state.files.lock().await;
        q.get("cwd").filter(|v| !v.is_empty()).and_then(|v| fm.safe_abs_path(v).ok())
    };
    ws.on_upgrade(move |socket| handle_ws(socket, cwd))
}

async fn handle_ws(mut socket: WebSocket, cwd: Option<std::path::PathBuf>) {
    let master_fd = match tokio::task::spawn_blocking(open_pty).await {
        Ok(Ok(fd)) => fd,
        _ => {
            let _ = socket.send(Message::Text(
                json!({"type": "error", "data": "Failed to open PTY"}).to_string().into(),
            )).await;
            return;
        }
    };

    let slave_path = match get_slave_name(master_fd) {
        Ok(p) => p,
        Err(e) => {
            unsafe { libc::close(master_fd) };
            let _ = socket.send(Message::Text(
                json!({"type": "error", "data": format!("PTY slave error: {e}")}).to_string().into(),
            )).await;
            return;
        }
    };

    let slave_cpath = std::ffi::CString::new(slave_path.to_string_lossy().as_ref()).unwrap_or_default();
    let slave_fd = unsafe { libc::open(slave_cpath.as_ptr(), libc::O_RDWR | libc::O_NOCTTY) };
    if slave_fd < 0 {
        unsafe { libc::close(master_fd) };
        let _ = socket.send(Message::Text(
            json!({"type": "error", "data": "Failed to open PTY slave"}).to_string().into(),
        )).await;
        return;
    }

    let shell = if Path::new("/bin/bash").exists() { "/bin/bash" } else { "/bin/sh" };
    let mut cmd = Command::new(shell);
    // 以登录 shell 启动（argv[0] = "-bash"），读取 /etc/profile、~/.profile 等，
    // 与 SSH 登录会话一致地加载用户配置的 PATH 等环境变量
    cmd.arg0(format!("-{}", Path::new(shell).file_name().unwrap_or_default().to_string_lossy()));
    cmd.arg("-l");
    cmd.env("TERM", "xterm-256color").env("COLORTERM", "truecolor");
    // 若提供了合法存在的目录，则终端以该目录作为初始工作目录
    if let Some(dir) = cwd
        && let Ok(md) = tokio::fs::metadata(&dir).await
        && md.is_dir()
    {
        cmd.current_dir(dir);
    }
    unsafe {
        let sfd = slave_fd;
        cmd.pre_exec(move || {
            libc::setsid();
            libc::ioctl(sfd, libc::TIOCSCTTY, 0 as libc::c_int);
            libc::dup2(sfd, 0);
            libc::dup2(sfd, 1);
            libc::dup2(sfd, 2);
            if sfd > 2 { libc::close(sfd); }
            Ok(())
        });
    }
    cmd.stdin(Stdio::null()).stdout(Stdio::null()).stderr(Stdio::null());

    let mut child = match cmd.spawn() {
        Ok(c) => c,
        Err(e) => {
            unsafe { libc::close(slave_fd); libc::close(master_fd); }
            let _ = socket.send(Message::Text(
                json!({"type": "error", "data": format!("Failed to start shell: {e}")}).to_string().into(),
            )).await;
            return;
        }
    };

    unsafe { libc::close(slave_fd) };
    let _ = socket.send(Message::Text(json!({"type": "connected"}).to_string().into())).await;

    let master_file = unsafe { std::fs::File::from_raw_fd(master_fd) };
    let master_fd_clone = master_fd;
    let (mut tx, mut rx) = socket.split();
    let (out_tx, mut out_rx) = tokio::sync::mpsc::channel::<Result<Vec<u8>, String>>(64);

    // PTY → WebSocket；读取结束（含错误）时通知前端，与 Go 版一致
    std::thread::spawn(move || {
        use std::io::Read;
        let mut f = master_file;
        let mut buf = [0u8; 4096];
        loop {
            match f.read(&mut buf) {
                Ok(0) | Err(_) => break,
                Ok(n) => {
                    if out_tx.blocking_send(Ok(buf[..n].to_vec())).is_err() { break; }
                }
            }
        }
        let _ = out_tx.blocking_send(Err("closed".to_string()));
    });

    let ws_out = tokio::spawn(async move {
        while let Some(data) = out_rx.recv().await {
            let text = match data {
                Ok(bytes) => json!({"type": "output", "data": String::from_utf8_lossy(&bytes)}).to_string(),
                Err(_) => json!({"type": "closed"}).to_string(),
            };
            if tx.send(Message::Text(text.into())).await.is_err() { break; }
        }
    });

    while let Some(Ok(Message::Text(t))) = rx.next().await {
        if let Ok(v) = serde_json::from_str::<Value>(&t) {
            match v["type"].as_str().unwrap_or("") {
                "input" => {
                    let data = v["data"].as_str().unwrap_or("").as_bytes().to_vec();
                    let fd = master_fd_clone;
                    tokio::task::spawn_blocking(move || {
                        unsafe { libc::write(fd, data.as_ptr() as *const libc::c_void, data.len()); }
                    }).await.ok();
                }
                "resize" => {
                    let rows = v["rows"].as_u64().unwrap_or(24) as u16;
                    let cols = v["cols"].as_u64().unwrap_or(80) as u16;
                    resize_pty(master_fd_clone, if rows == 0 { 24 } else { rows }, if cols == 0 { 80 } else { cols });
                }
                _ => {}
            }
        }
    }

    let _ = child.kill().await;
    ws_out.abort();
}

fn open_pty() -> std::io::Result<RawFd> {
    let master_fd = unsafe {
        libc::open(c"/dev/ptmx".as_ptr(), libc::O_RDWR | libc::O_NOCTTY | libc::O_CLOEXEC)
    };
    if master_fd < 0 { return Err(std::io::Error::last_os_error()); }
    let unlock: libc::c_int = 0;
    if unsafe { libc::ioctl(master_fd, libc::TIOCSPTLCK, &unlock) } < 0 {
        unsafe { libc::close(master_fd) };
        return Err(std::io::Error::last_os_error());
    }
    Ok(master_fd)
}

fn get_slave_name(master_fd: RawFd) -> std::io::Result<OsString> {
    let mut ptn: libc::c_uint = 0;
    if unsafe { libc::ioctl(master_fd, libc::TIOCGPTN, &mut ptn) } < 0 {
        return Err(std::io::Error::last_os_error());
    }
    Ok(OsString::from(format!("/dev/pts/{}", ptn)))
}

fn resize_pty(master_fd: RawFd, rows: u16, cols: u16) {
    #[repr(C)]
    struct Winsize { ws_row: u16, ws_col: u16, ws_xpixel: u16, ws_ypixel: u16 }
    let ws = Winsize { ws_row: rows, ws_col: cols, ws_xpixel: 0, ws_ypixel: 0 };
    unsafe { libc::ioctl(master_fd, libc::TIOCSWINSZ, &ws); }
}
