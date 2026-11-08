use anyhow::Result;
use axum::{Json, http::StatusCode, response::{IntoResponse, Response}};
use serde_json::{Value, json};
use std::{ffi::OsStr, net::{IpAddr, Ipv4Addr}, path::Path, time::{SystemTime, UNIX_EPOCH}};
use chrono::Utc;

pub fn json_error(code: StatusCode, msg: impl ToString) -> Response {
    (code, Json(json!({"error": msg.to_string()}))).into_response()
}

pub fn ok_json(v: Value) -> Response {
    Json(v).into_response()
}

pub fn now_string() -> String {
    Utc::now().to_rfc3339()
}

pub fn unix_now() -> usize {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_secs() as usize
}

/// 对齐 Go `normalizePath`：`filepath.Clean("/" + strings.TrimLeft(p, "/"))`
/// 结果恒为 "/" 或 "/xxx"（正确解析 ".."，不会逃出根目录）。
pub fn normalize_path(p: &str) -> String {
    let trimmed = p.trim_start_matches('/');
    let mut out: Vec<&str> = Vec::new();
    for part in trimmed.split('/') {
        match part {
            "" | "." => {}
            ".." => { out.pop(); }
            v => out.push(v),
        }
    }
    if out.is_empty() { "/".into() } else { format!("/{}", out.join("/")) }
}

/// 对齐 Go `filepath.Clean`（POSIX 语义），用于校验用户传入的绝对/相对路径。
pub fn path_clean(p: &str) -> String {
    let rooted = p.starts_with('/');
    let mut out: Vec<&str> = Vec::new();
    for part in p.split('/') {
        match part {
            "" | "." => {}
            ".." => {
                if !out.is_empty() && *out.last().unwrap() != ".." {
                    out.pop();
                } else if !rooted {
                    out.push("..");
                }
                // rooted 且 out 为空：忽略，保持在根
            }
            v => out.push(v),
        }
    }
    let mut s = out.join("/");
    if rooted { s.insert(0, '/'); }
    if s.is_empty() { ".".into() } else { s }
}

pub fn is_danger_path(p: &Path) -> bool {
    let s = p.to_string_lossy();
    ["/proc", "/sys", "/dev"]
        .iter()
        .any(|x| s == *x || s.starts_with(&format!("{x}/")))
}

pub fn safe_name(name: &str) -> String {
    Path::new(name)
        .file_name()
        .unwrap_or_else(|| OsStr::new("download"))
        .to_string_lossy()
        .to_string()
}

/// 对齐 Go isPrivateIP：回环/链路本地单播与多播/未指定 + 常见私有网段
/// （10/8、172.16/12、192.168/16、169.254/16、100.64/10 CGNAT、0/8、
/// IPv6 ULA fc00::/7、IPv6 链路本地 fe80::/10）。
pub fn is_private_ip(ip: IpAddr) -> bool {
    match ip {
        IpAddr::V4(v) => {
            if v.is_loopback() || v.is_link_local() || v.is_unspecified() || v.is_multicast() {
                return true;
            }
            is_in_v4(v, Ipv4Addr::new(10, 0, 0, 0), 8)
                || is_in_v4(v, Ipv4Addr::new(172, 16, 0, 0), 12)
                || is_in_v4(v, Ipv4Addr::new(192, 168, 0, 0), 16)
                || is_in_v4(v, Ipv4Addr::new(169, 254, 0, 0), 16)
                || is_in_v4(v, Ipv4Addr::new(100, 64, 0, 0), 10)
                || is_in_v4(v, Ipv4Addr::new(0, 0, 0, 0), 8)
        }
        IpAddr::V6(v) => {
            v.is_loopback()
                || v.is_unspecified()
                || v.is_multicast()
                // ULA fc00::/7
                || (v.segments()[0] & 0xfe00) == 0xfc00
                // link-local fe80::/10
                || (v.segments()[0] & 0xffc0) == 0xfe80
        }
    }
}

fn is_in_v4(ip: Ipv4Addr, network: Ipv4Addr, prefix: u8) -> bool {
    let mask = if prefix == 0 {
        0
    } else {
        u32::MAX << (32 - prefix)
    };
    (u32::from(ip) & mask) == (u32::from(network) & mask)
}

/// 对齐 Go RealIP：直连 IP 非私有 → 完全不信任转发头；
/// 来自私有网段（可信反代）时才读取 X-Forwarded-For / X-Real-IP。
pub fn real_ip(headers: &axum::http::HeaderMap, addr: std::net::SocketAddr) -> String {
    let remote = addr.ip();
    if !is_private_ip(remote) {
        return remote.to_string();
    }
    if let Some(ip) = headers
        .get("x-forwarded-for")
        .and_then(|v| v.to_str().ok())
        .and_then(|s| s.split(',').next())
        .map(str::trim)
        .filter(|s| !s.is_empty())
    {
        return ip.to_string();
    }
    if let Some(ip) = headers
        .get("x-real-ip")
        .and_then(|v| v.to_str().ok())
        .map(str::trim)
        .filter(|s| !s.is_empty())
    {
        return ip.to_string();
    }
    remote.to_string()
}

/// AES-256-GCM 加密（nonce 前置 + base64，与 Go 版 Encrypt 完全一致）
pub fn encrypt(key: &[u8; 32], plaintext: &str) -> Result<String> {
    use aes_gcm::{Aes256Gcm, Nonce, aead::{Aead, KeyInit}};
    use base64::{Engine as _, engine::general_purpose};
    use rand::RngCore;
    let cipher = Aes256Gcm::new(key.into());
    let mut nonce = [0u8; 12];
    rand::thread_rng().fill_bytes(&mut nonce);
    let ct = cipher
        .encrypt(Nonce::from_slice(&nonce), plaintext.as_bytes())
        .map_err(|_| anyhow::anyhow!("encrypt failed"))?;
    let mut out = nonce.to_vec();
    out.extend(ct);
    Ok(general_purpose::STANDARD.encode(out))
}

/// AES-256-GCM 解密，失败返回 None
pub fn decrypt_opt(key: &[u8; 32], enc: &str) -> Option<String> {
    use aes_gcm::{Aes256Gcm, Nonce, aead::{Aead, KeyInit}};
    use base64::{Engine as _, engine::general_purpose};
    if enc.is_empty() {
        return Some(String::new());
    }
    let data = general_purpose::STANDARD.decode(enc).ok()?;
    if data.len() < 12 {
        return None;
    }
    let cipher = Aes256Gcm::new(key.into());
    let pt = cipher
        .decrypt(Nonce::from_slice(&data[..12]), &data[12..])
        .ok()?;
    String::from_utf8(pt).ok()
}

/// 对齐 Go copyPath（WebDAV COPY 使用）
pub fn copy_path(src: &Path, dst: &Path) -> std::io::Result<()> {
    let md = std::fs::metadata(src)?;
    if md.is_dir() {
        std::fs::create_dir_all(dst)?;
        for e in std::fs::read_dir(src)? {
            let e = e?;
            copy_path(&e.path(), &dst.join(e.file_name()))?;
        }
    } else {
        if let Some(p) = dst.parent() {
            std::fs::create_dir_all(p)?;
        }
        std::fs::copy(src, dst)?;
    }
    Ok(())
}

/// 对齐 Go os.RemoveAll：目标不存在视为成功
pub fn remove_all(p: &Path) -> std::io::Result<()> {
    match std::fs::symlink_metadata(p) {
        Ok(md) => {
            if md.is_dir() {
                std::fs::remove_dir_all(p)
            } else {
                std::fs::remove_file(p)
            }
        }
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => Ok(()),
        Err(e) => Err(e),
    }
}
