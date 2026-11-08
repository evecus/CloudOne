use serde::{Deserialize, Serialize};
use std::{collections::HashMap, path::{Path, PathBuf}, sync::Arc, time::{Duration, SystemTime}};
use tokio::sync::Mutex;

use crate::db::Db;

// ── 数据模型 ──────────────────────────────────────────────────────────────────

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct User {
    pub id: i64,
    pub username: String,
    #[serde(skip_serializing)]
    pub password: String,
    #[serde(skip_serializing)]
    pub token_version: i64,
    pub created_at: String,
    pub updated_at: String,
}

#[derive(Debug, Serialize, Clone)]
pub struct Settings {
    pub id: i64,
    pub storage_dir: String,
    pub lang: String,
    pub ui_theme: String,
    pub ui_font: String,
    pub editor_font: String,
    pub webdav_enabled: bool,
    pub webdav_sub_path: String,
    pub webdav_username: String,
    #[serde(skip_serializing)]
    pub webdav_password_enc: String,
    #[serde(skip_serializing)]
    pub jwt_secret_enc: String,
    pub show_hidden: bool,
    /// list | detail | icon-large | icon-small
    pub file_view_mode: String,
    /// name | date | type | size
    pub file_sort_by: String,
    /// asc | desc
    pub file_sort_order: String,
    pub github_proxy_enabled: bool,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct ShareLink {
    pub id: i64,
    pub code: String,
    pub file_path: String,
    pub is_dir: bool,
    pub user_id: i64,
    pub expires_at: Option<String>,
    pub max_views: i64,
    pub view_count: i64,
    pub created_at: String,
}

#[derive(Debug, Serialize, Clone)]
pub struct FileInfo {
    pub name: String,
    pub path: String,
    pub is_dir: bool,
    pub size: u64,
    pub mod_time: String,
    pub is_public: bool,
    pub mode: u32,
}

#[derive(Debug, Serialize)]
#[allow(dead_code)]
pub struct SearchResult {
    pub name: String,
    pub path: String,
    pub is_dir: bool,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Claims {
    pub sub: i64,
    pub version: i64,
    pub exp: usize,
    pub iat: usize,
}

// ── AppState ──────────────────────────────────────────────────────────────────

#[derive(Clone)]
pub struct AppState {
    pub db: Db,
    pub files: Arc<Mutex<FileManager>>,
    pub jwt_secret: Arc<Vec<u8>>,
    pub master_key: Arc<[u8; 32]>,
    pub login_limiter: Arc<RateLimiter>,
    pub webdav_limiter: Arc<RateLimiter>,
    /// 公开分享访问限流（防分享码爆破枚举）
    pub share_limiter: Arc<RateLimiter>,
    /// CLOUDONE_ORIGIN：允许的跨域来源，空 = 仅允许无 Origin / "null"
    pub allowed_origin: String,
}

// ── FileManager ───────────────────────────────────────────────────────────────

#[derive(Clone)]
pub struct FileManager {
    pub root: PathBuf,
}

impl FileManager {
    pub fn new(root: PathBuf) -> Self {
        Self { root }
    }
    pub fn set_root(&mut self, root: PathBuf) {
        self.root = root;
    }
    pub fn root(&self) -> PathBuf {
        self.root.clone()
    }
    pub fn abs_path(&self, rel: &str) -> anyhow::Result<PathBuf> {
        use crate::util::normalize_path;
        let norm = normalize_path(rel);
        let sub = norm.trim_start_matches('/');
        let abs = self.root.join(sub);
        let root = self.root.clone();
        if root != Path::new("/") && !abs.starts_with(&root) {
            return Err(anyhow::anyhow!("invalid path: path traversal detected"));
        }
        Ok(abs)
    }
    pub fn safe_abs_path(&self, rel: &str) -> anyhow::Result<PathBuf> {
        use crate::util::is_danger_path;
        let abs = self.abs_path(rel)?;
        if is_danger_path(&abs) {
            return Err(anyhow::anyhow!("access denied: virtual filesystem path"));
        }
        Ok(abs)
    }
}

// ── RateLimiter ───────────────────────────────────────────────────────────────

#[derive(Default)]
pub struct RateLimiter {
    pub max: usize,
    pub window: Duration,
    pub hits: Mutex<HashMap<String, Vec<SystemTime>>>,
}

impl RateLimiter {
    pub fn new(max: usize, window: Duration) -> Self {
        Self {
            max,
            window,
            hits: Mutex::new(HashMap::new()),
        }
    }

    /// 启动后台清理任务，每 5 分钟清理超过 2 个窗口时长未活动的 IP 条目
    pub fn spawn_cleanup(self: &Arc<Self>) {
        let limiter = Arc::clone(self);
        tokio::spawn(async move {
            let mut interval =
                tokio::time::interval(tokio::time::Duration::from_secs(300));
            loop {
                interval.tick().await;
                let cutoff = SystemTime::now() - limiter.window * 2;
                let mut guard = limiter.hits.lock().await;
                guard.retain(|_, timestamps| timestamps.iter().any(|t| *t > cutoff));
            }
        });
    }

    pub async fn allow(&self, ip: String) -> bool {
        let now = SystemTime::now();
        let cutoff = now - self.window;
        let mut guard = self.hits.lock().await;
        let entry = guard.entry(ip).or_default();
        entry.retain(|t| *t > cutoff);
        if entry.len() >= self.max {
            return false;
        }
        entry.push(now);
        true
    }
}
