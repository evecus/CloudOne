use anyhow::Result;
use chrono::{DateTime, Utc};
use redb::{Database, ReadableDatabase, ReadableTable, TableDefinition};
use serde::{Deserialize, Serialize};
use std::{
    collections::HashSet,
    path::Path,
    sync::Arc,
    time::SystemTime,
};

use crate::types::{AppState, FileInfo, Settings, ShareLink, User};
use crate::util::{is_danger_path, normalize_path};

// ── 存储层：redb 单文件嵌入式 KV（替代 SQLite）────────────────────────────────
//
// 表结构：
//   users     key=id(u64)        value=User JSON
//   user_idx  key=username(&str) value=id(u64)
//   settings  key=0(u64)         value=SettingsRow JSON（含加密字段）
//   shares    key=id(u64)        value=ShareLink JSON
//   share_idx key=code(&str)     value=id(u64)
//   vis       key=path(&str)     value=1(u8)（仅存公开项）
//   meta      key=0/1(u64)       value=下一个自增 id

pub type Db = Arc<Database>;

const T_USERS: TableDefinition<'static, u64, &[u8]> = TableDefinition::new("users");
const T_USER_IDX: TableDefinition<'static, &str, u64> = TableDefinition::new("user_idx");
const T_SETTINGS: TableDefinition<'static, u64, &[u8]> = TableDefinition::new("settings");
const T_SHARES: TableDefinition<'static, u64, &[u8]> = TableDefinition::new("shares");
const T_SHARE_IDX: TableDefinition<'static, &str, u64> = TableDefinition::new("share_idx");
const T_VIS: TableDefinition<'static, &str, u8> = TableDefinition::new("vis");
const T_META: TableDefinition<'static, u64, u64> = TableDefinition::new("meta");

const META_NEXT_USER_ID: u64 = 0;
const META_NEXT_SHARE_ID: u64 = 1;

/// 打开（或创建）数据目录下的 redb 数据库文件
pub fn open_db(path: &Path) -> Result<Db> {
    Ok(Arc::new(Database::create(path)?))
}

/// settings 的完整存储行（含 skip_serializing 的加密字段）
#[derive(Serialize, Deserialize, Clone, Debug)]
struct SettingsRow {
    id: i64,
    storage_dir: String,
    lang: String,
    ui_theme: String,
    ui_font: String,
    editor_font: String,
    webdav_enabled: bool,
    webdav_sub_path: String,
    webdav_username: String,
    webdav_password_enc: String,
    jwt_secret_enc: String,
    show_hidden: bool,
    file_view_mode: String,
    file_sort_by: String,
    file_sort_order: String,
    github_proxy_enabled: bool,
}

impl From<&Settings> for SettingsRow {
    fn from(s: &Settings) -> Self {
        Self {
            id: s.id,
            storage_dir: s.storage_dir.clone(),
            lang: s.lang.clone(),
            ui_theme: s.ui_theme.clone(),
            ui_font: s.ui_font.clone(),
            editor_font: s.editor_font.clone(),
            webdav_enabled: s.webdav_enabled,
            webdav_sub_path: s.webdav_sub_path.clone(),
            webdav_username: s.webdav_username.clone(),
            webdav_password_enc: s.webdav_password_enc.clone(),
            jwt_secret_enc: s.jwt_secret_enc.clone(),
            show_hidden: s.show_hidden,
            file_view_mode: s.file_view_mode.clone(),
            file_sort_by: s.file_sort_by.clone(),
            file_sort_order: s.file_sort_order.clone(),
            github_proxy_enabled: s.github_proxy_enabled,
        }
    }
}

impl From<SettingsRow> for Settings {
    fn from(r: SettingsRow) -> Self {
        Settings {
            id: r.id,
            storage_dir: r.storage_dir,
            lang: r.lang,
            ui_theme: r.ui_theme,
            ui_font: r.ui_font,
            editor_font: r.editor_font,
            webdav_enabled: r.webdav_enabled,
            webdav_sub_path: r.webdav_sub_path,
            webdav_username: r.webdav_username,
            webdav_password_enc: r.webdav_password_enc,
            jwt_secret_enc: r.jwt_secret_enc,
            show_hidden: r.show_hidden,
            file_view_mode: r.file_view_mode,
            file_sort_by: r.file_sort_by,
            file_sort_order: r.file_sort_order,
            github_proxy_enabled: r.github_proxy_enabled,
        }
    }
}

fn default_settings() -> Settings {
    Settings {
        id: 1,
        storage_dir: "./data/storage".into(),
        lang: "zh".into(),
        ui_theme: String::new(),
        ui_font: String::new(),
        editor_font: String::new(),
        webdav_enabled: false,
        webdav_sub_path: String::new(),
        webdav_username: String::new(),
        webdav_password_enc: String::new(),
        jwt_secret_enc: String::new(),
        show_hidden: false,
        file_view_mode: "detail".into(),
        file_sort_by: "name".into(),
        file_sort_order: "asc".into(),
        github_proxy_enabled: false,
    }
}

pub fn init_db(db: &Db) -> Result<()> {
    let txn = db.begin_write()?;
    {
        txn.open_table(T_USERS)?;
        txn.open_table(T_USER_IDX)?;
        txn.open_table(T_SHARES)?;
        txn.open_table(T_SHARE_IDX)?;
        txn.open_table(T_VIS)?;
        let mut meta = txn.open_table(T_META)?;
        if meta.get(META_NEXT_USER_ID)?.is_none() {
            meta.insert(META_NEXT_USER_ID, 1)?;
        }
        if meta.get(META_NEXT_SHARE_ID)?.is_none() {
            meta.insert(META_NEXT_SHARE_ID, 1)?;
        }
        let mut settings = txn.open_table(T_SETTINGS)?;
        if settings.get(0u64)?.is_none() {
            // 对齐 Go InitDB 默认值
            let row: SettingsRow = (&default_settings()).into();
            settings.insert(0u64, serde_json::to_vec(&row)?.as_slice())?;
        }
    }
    txn.commit()?;
    Ok(())
}

// ── Settings ─────────────────────────────────────────────────────────────────

pub fn get_settings(db: &Db) -> Result<Settings> {
    let txn = db.begin_read()?;
    let table = txn.open_table(T_SETTINGS)?;
    match table.get(0u64)? {
        Some(v) => Ok(serde_json::from_slice::<SettingsRow>(v.value())?.into()),
        None => Ok(default_settings()),
    }
}

pub fn save_settings(db: &Db, s: &Settings) -> Result<()> {
    let txn = db.begin_write()?;
    {
        let mut table = txn.open_table(T_SETTINGS)?;
        let row: SettingsRow = (s).into();
        table.insert(0u64, serde_json::to_vec(&row)?.as_slice())?;
    }
    txn.commit()?;
    Ok(())
}

pub fn set_jwt_secret_enc(db: &Db, enc: &str) -> Result<()> {
    let mut s = get_settings(db)?;
    s.jwt_secret_enc = enc.to_string();
    save_settings(db, &s)
}

// ── Users ────────────────────────────────────────────────────────────────────

/// User 的完整存储行（User 本体对 API 响应 skip 了 password/token_version）
#[derive(Serialize, Deserialize)]
struct UserRow {
    id: i64,
    username: String,
    password: String,
    token_version: i64,
    created_at: String,
    updated_at: String,
}

impl From<&User> for UserRow {
    fn from(u: &User) -> Self {
        Self {
            id: u.id,
            username: u.username.clone(),
            password: u.password.clone(),
            token_version: u.token_version,
            created_at: u.created_at.clone(),
            updated_at: u.updated_at.clone(),
        }
    }
}

fn parse_user(v: &[u8]) -> Result<User> {
    let r: UserRow = serde_json::from_slice(v)?;
    Ok(User {
        id: r.id,
        username: r.username,
        password: r.password,
        token_version: r.token_version,
        created_at: r.created_at,
        updated_at: r.updated_at,
    })
}

fn user_bytes(u: &User) -> Result<Vec<u8>> {
    Ok(serde_json::to_vec(&UserRow::from(u))?)
}

pub fn count_users(db: &Db) -> Result<i64> {
    let txn = db.begin_read()?;
    let table = txn.open_table(T_USERS)?;
    let mut n: i64 = 0;
    for row in table.iter()? {
        let _ = row?;
        n += 1;
    }
    Ok(n)
}

pub fn get_user_by_id(db: &Db, id: i64) -> Result<User> {
    let txn = db.begin_read()?;
    let table = txn.open_table(T_USERS)?;
    table
        .get(id as u64)?
        .map(|v| parse_user(v.value()))
        .transpose()?
        .ok_or_else(|| anyhow::anyhow!("user not found"))
}

pub fn get_user_by_username(db: &Db, username: &str) -> Result<Option<User>> {
    let txn = db.begin_read()?;
    let idx = txn.open_table(T_USER_IDX)?;
    let Some(id) = idx.get(username)?.map(|g| g.value()) else {
        return Ok(None);
    };
    let users = txn.open_table(T_USERS)?;
    users
        .get(id)?
        .map(|v| parse_user(v.value()))
        .transpose()
}

pub fn insert_user(db: &Db, username: &str, password_hash: &str, now: &str) -> Result<i64> {
    let txn = db.begin_write()?;
    let id = {
        let mut meta = txn.open_table(T_META)?;
        let id = meta.get(META_NEXT_USER_ID)?.map(|g| g.value()).unwrap_or(1);
        meta.insert(META_NEXT_USER_ID, id + 1)?;
        id
    };
    {
        let mut users = txn.open_table(T_USERS)?;
        let mut idx = txn.open_table(T_USER_IDX)?;
        let u = User {
            id: id as i64,
            username: username.to_string(),
            password: password_hash.to_string(),
            token_version: 0,
            created_at: now.to_string(),
            updated_at: now.to_string(),
        };
        users.insert(id, user_bytes(&u)?.as_slice())?;
        idx.insert(username, id)?;
    }
    txn.commit()?;
    Ok(id as i64)
}

/// 更新用户名 / 密码 / token_version。
/// 安全约束：改名时校验新用户名未被其他账号占用（防止通过改名劫持
/// 已存在用户名的登录索引），并同步清理旧用户名索引（防止旧名残留可登录）。
pub fn update_user_db(
    db: &Db,
    id: i64,
    username: &str,
    password_hash: &str,
    token_version: i64,
    now: &str,
) -> Result<()> {
    let txn = db.begin_write()?;
    {
        let mut users = txn.open_table(T_USERS)?;
        let mut u = users
            .get(id as u64)?
            .map(|v| parse_user(v.value()))
            .transpose()?
            .ok_or_else(|| anyhow::anyhow!("user not found"))?;
        let old_username = u.username.clone();
        if username != old_username {
            let mut idx = txn.open_table(T_USER_IDX)?;
            if let Some(existing) = idx.get(username)?.map(|g| g.value())
                && existing != id as u64
            {
                anyhow::bail!("username already taken");
            }
            idx.remove(old_username.as_str())?;
            idx.insert(username, id as u64)?;
        }
        u.username = username.to_string();
        u.password = password_hash.to_string();
        u.token_version = token_version;
        u.updated_at = now.to_string();
        users.insert(id as u64, user_bytes(&u)?.as_slice())?;
    }
    txn.commit()?;
    Ok(())
}

/// WebDAV 用：未配置独立用户名时回落到第一个 CloudOne 账户
pub fn first_username(db: &Db) -> Option<String> {
    let txn = db.begin_read().ok()?;
    let table = txn.open_table(T_USERS).ok()?;
    let row = table.iter().ok()?.next()?.ok()?;
    let (_, v) = row;
    parse_user(v.value()).ok().map(|u| u.username)
}

pub fn get_password_by_username(db: &Db, username: &str) -> Option<String> {
    get_user_by_username(db, username)
        .ok()
        .flatten()
        .map(|u| u.password)
}

// ── ShareLinks ───────────────────────────────────────────────────────────────

fn parse_share(v: &[u8]) -> Result<ShareLink> {
    Ok(serde_json::from_slice(v)?)
}

pub fn get_share_by_id(db: &Db, id: i64) -> Result<ShareLink> {
    let txn = db.begin_read()?;
    let table = txn.open_table(T_SHARES)?;
    table
        .get(id as u64)?
        .map(|v| parse_share(v.value()))
        .transpose()?
        .ok_or_else(|| anyhow::anyhow!("share not found"))
}

#[allow(clippy::too_many_arguments)]
pub fn insert_share(
    db: &Db,
    code: &str,
    file_path: &str,
    is_dir: bool,
    user_id: i64,
    expires_at: Option<String>,
    max_views: i64,
    created_at: &str,
) -> Result<i64> {
    let txn = db.begin_write()?;
    let id = {
        let mut meta = txn.open_table(T_META)?;
        let id = meta.get(META_NEXT_SHARE_ID)?.map(|g| g.value()).unwrap_or(1);
        meta.insert(META_NEXT_SHARE_ID, id + 1)?;
        id
    };
    {
        let mut shares = txn.open_table(T_SHARES)?;
        let mut idx = txn.open_table(T_SHARE_IDX)?;
        let link = ShareLink {
            id: id as i64,
            code: code.to_string(),
            file_path: file_path.to_string(),
            is_dir,
            user_id,
            expires_at,
            max_views,
            view_count: 0,
            created_at: created_at.to_string(),
        };
        shares.insert(id, serde_json::to_vec(&link)?.as_slice())?;
        idx.insert(code, id)?;
    }
    txn.commit()?;
    Ok(id as i64)
}

pub fn list_shares_by_user(db: &Db, user_id: i64) -> Vec<ShareLink> {
    let Ok(txn) = db.begin_read() else { return Vec::new() };
    let Ok(table) = txn.open_table(T_SHARES) else { return Vec::new() };
    let mut out = Vec::new();
    if let Ok(rows) = table.iter() {
        for row in rows {
            let Ok((_, v)) = row else { continue };
            let Ok(l) = parse_share(v.value()) else { continue };
            if l.user_id == user_id {
                out.push(l);
            }
        }
    }
    out
}

/// 删除指定分享（仅当 user_id 匹配时生效，对齐原 SQL `WHERE id=? AND user_id=?`）
pub fn delete_share(db: &Db, id: i64, user_id: i64) {
    let Ok(txn) = db.begin_write() else { return };
    {
        let Ok(mut shares) = txn.open_table(T_SHARES) else { return };
        let link = shares
            .get(id as u64)
            .ok()
            .flatten()
            .map(|v| parse_share(v.value()))
            .transpose()
            .ok()
            .flatten();
        let Some(link) = link else { return };
        if link.user_id != user_id {
            return;
        }
        let _ = shares.remove(id as u64);
        let Ok(mut idx) = txn.open_table(T_SHARE_IDX) else { return };
        let _ = idx.remove(link.code.as_str());
    }
    let _ = txn.commit();
}

pub fn get_share(db: &Db, code: &str, count: bool) -> Result<ShareLink> {
    let mut l = {
        let txn = db.begin_read()?;
        let idx = txn.open_table(T_SHARE_IDX)?;
        let Some(id) = idx.get(code)?.map(|g| g.value()) else {
            return Err(anyhow::anyhow!("share not found"));
        };
        let shares = txn.open_table(T_SHARES)?;
        shares
            .get(id)?
            .map(|v| parse_share(v.value()))
            .transpose()?
            .ok_or_else(|| anyhow::anyhow!("share not found"))?
    };
    if let Some(e) = &l.expires_at
        && DateTime::parse_from_rfc3339(e)
            .map(|d| d.with_timezone(&Utc) < Utc::now())
            .unwrap_or(false)
    {
        return Err(anyhow::anyhow!("share link has expired"));
    }
    if l.max_views > 0 && l.view_count >= l.max_views {
        return Err(anyhow::anyhow!("share link has reached maximum views"));
    }
    if count {
        l.view_count += 1;
        let txn = db.begin_write()?;
        {
            let mut shares = txn.open_table(T_SHARES)?;
            shares.insert(l.id as u64, serde_json::to_vec(&l)?.as_slice())?;
        }
        txn.commit()?;
    }
    Ok(l)
}

// ── FileVisibilities ─────────────────────────────────────────────────────────

pub fn is_public(db: &Db, rel: &str) -> bool {
    let Ok(txn) = db.begin_read() else { return false };
    let Ok(table) = txn.open_table(T_VIS) else { return false };
    table
        .get(normalize_path(rel).as_str())
        .ok()
        .flatten()
        .map(|v| v.value() == 1)
        .unwrap_or(false)
}

pub fn upsert_visibility(db: &Db, rel: &str, public: bool) -> Result<()> {
    let txn = db.begin_write()?;
    {
        let mut table = txn.open_table(T_VIS)?;
        let key = normalize_path(rel);
        if public {
            table.insert(key.as_str(), 1u8)?;
        } else {
            let _ = table.remove(key.as_str());
        }
    }
    txn.commit()?;
    Ok(())
}

/// 删除文件/目录及其所有子路径的可见性记录。
/// 对齐 Go DeleteVisibilityPrefix：精确字符串前缀匹配（key == prefix 或
/// key 以 prefix + "/" 开头）。
pub fn delete_visibility_tree(db: &Db, rel: &str) {
    let n = normalize_path(rel);
    let Ok(txn) = db.begin_write() else { return };
    {
        let Ok(mut table) = txn.open_table(T_VIS) else { return };
        let to_del: Vec<String> = table
            .iter()
            .map(|rows| {
                rows.filter_map(|r| {
                    let (k, _) = r.ok()?;
                    let key = k.value().to_string();
                    (key == n || key.starts_with(&format!("{n}/"))).then_some(key)
                })
                .collect()
            })
            .unwrap_or_default();
        for k in to_del {
            let _ = table.remove(k.as_str());
        }
    }
    let _ = txn.commit();
}

pub fn migrate_visibility(db: &Db, src: &str, dst: &str) {
    copy_visibility(db, src, dst);
    delete_visibility_tree(db, src);
}

/// 对齐 Go copyVisibility / RenameVisibilityPrefix：把 src 前缀的可见性
/// 记录迁移到 dst 前缀（含自身与所有子路径）。
pub fn copy_visibility(db: &Db, src: &str, dst: &str) {
    let sn = normalize_path(src);
    let dn = normalize_path(dst);
    let Ok(txn) = db.begin_write() else { return };
    {
        let Ok(mut table) = txn.open_table(T_VIS) else { return };
        let matched: Vec<(String, u8)> = table
            .iter()
            .map(|rows| {
                rows.filter_map(|r| {
                    let (k, v) = r.ok()?;
                    let key = k.value().to_string();
                    (key == sn || key.starts_with(&format!("{sn}/"))).then(|| (key.clone(), v.value()))
                })
                .collect()
            })
            .unwrap_or_default();
        for (old, public) in matched {
            let newp = if old == sn {
                dn.clone()
            } else {
                format!("{}{}", dn, &old[sn.len()..])
            };
            if public == 1 {
                let _ = table.insert(newp.as_str(), 1u8);
            } else {
                let _ = table.remove(newp.as_str());
            }
        }
    }
    let _ = txn.commit();
}

/// 对齐 Go ListPublicVisibility：返回所有公开路径
pub fn list_public_paths(db: &Db) -> Vec<String> {
    let Ok(txn) = db.begin_read() else { return Vec::new() };
    let Ok(table) = txn.open_table(T_VIS) else { return Vec::new() };
    table
        .iter()
        .map(|rows| {
            rows.filter_map(|r| {
                let (k, v) = r.ok()?;
                (v.value() == 1).then(|| k.value().to_string())
            })
            .collect()
        })
        .unwrap_or_default()
}

/// 批量查询用：全部公开路径集合（list_dir 避免逐项查询）
pub fn public_set(db: &Db) -> HashSet<String> {
    list_public_paths(db).into_iter().collect()
}

// ── 文件信息 ─────────────────────────────────────────────────────────────────

pub fn file_info(db: &Db, abs: &Path, rel: &str) -> Result<FileInfo> {
    let md = std::fs::metadata(abs)?;
    #[cfg(unix)]
    use std::os::unix::fs::PermissionsExt;
    #[cfg(unix)]
    let mode = md.permissions().mode();
    #[cfg(not(unix))]
    let mode = 0u32;
    Ok(FileInfo {
        name: Path::new(rel)
            .file_name()
            .unwrap_or_default()
            .to_string_lossy()
            .to_string(),
        path: normalize_path(rel),
        is_dir: md.is_dir(),
        size: if md.is_dir() { 0 } else { md.len() },
        mod_time: DateTime::<Utc>::from(md.modified().unwrap_or(SystemTime::now())).to_rfc3339(),
        is_public: is_public(db, rel),
        mode,
    })
}

/// 对齐 Go ListDir：不做隐藏文件过滤（由前端按 show_hidden 设置过滤）。
pub async fn list_dir(state: &AppState, rel: &str) -> Result<Vec<FileInfo>> {
    let fm = state.files.lock().await.clone();
    let abs = fm.safe_abs_path(rel)?;
    let mut rd = tokio::fs::read_dir(&abs).await?;
    let base = normalize_path(rel);

    let mut entries: Vec<(std::path::PathBuf, String)> = Vec::new();
    while let Some(e) = rd.next_entry().await? {
        let p = e.path();
        if is_danger_path(&p) {
            continue;
        }
        let name = e.file_name().to_string_lossy().into_owned();
        let child = normalize_path(&format!("{}/{}", base, name));
        entries.push((p, child));
    }

    // 批量取公开集合，避免逐项查询
    let vis_set = public_set(&state.db);

    let mut out = Vec::new();
    for (p, child) in entries {
        let abs_p = p.clone();
        let meta = tokio::task::spawn_blocking(move || std::fs::metadata(&abs_p)).await;
        if let Ok(Ok(md)) = meta {
            #[cfg(unix)]
            use std::os::unix::fs::PermissionsExt;
            #[cfg(unix)]
            let mode = md.permissions().mode();
            #[cfg(not(unix))]
            let mode = 0u32;
            let name = p.file_name().unwrap_or_default().to_string_lossy().into_owned();
            out.push(FileInfo {
                name,
                path: child.clone(),
                is_dir: md.is_dir(),
                size: if md.is_dir() { 0 } else { md.len() },
                mod_time: DateTime::<Utc>::from(
                    md.modified().unwrap_or(SystemTime::now()),
                )
                .to_rfc3339(),
                is_public: vis_set.contains(&child),
                mode,
            });
        }
    }
    Ok(out)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn rename_user_duplicate_rejected_and_old_idx_cleaned() {
        let dir = std::env::temp_dir().join(format!("c1db-test-{}", std::process::id()));
        let _ = std::fs::remove_dir_all(&dir);
        std::fs::create_dir_all(&dir).unwrap();
        let db = open_db(&dir.join("test.redb")).unwrap();
        init_db(&db).unwrap();
        let id1 = insert_user(&db, "alice", "h1", "now").unwrap();
        let id2 = insert_user(&db, "bob", "h2", "now").unwrap();

        // 改名到已存在的用户名 → 拒绝（防劫持 user_idx）
        assert!(update_user_db(&db, id2, "alice", "h2", 0, "now").is_err());
        // idx 未被破坏：alice 仍指向 id1，bob 仍指向 id2
        assert_eq!(get_user_by_username(&db, "alice").unwrap().unwrap().id, id1);
        assert_eq!(get_user_by_username(&db, "bob").unwrap().unwrap().id, id2);

        // 正常改名：旧索引清理、新索引生效
        update_user_db(&db, id2, "carol", "h2", 0, "now").unwrap();
        assert!(get_user_by_username(&db, "bob").unwrap().is_none());
        assert_eq!(get_user_by_username(&db, "carol").unwrap().unwrap().id, id2);
        // 改回自己的旧名也允许
        update_user_db(&db, id2, "bob", "h2", 0, "now").unwrap();
        assert_eq!(get_user_by_username(&db, "bob").unwrap().unwrap().id, id2);

        let _ = std::fs::remove_dir_all(&dir);
    }
}
