use std::{
    env,
    fs::{self, File},
    io::{self, Write},
    path::{Path, PathBuf},
};

/// 前端静态文件在编译期内嵌进二进制（对齐 Go //go:embed all:frontend/dist）。
/// dist 缺失时直接编译失败，避免产出"不内嵌前端"的二进制。
fn main() -> io::Result<()> {
    let manifest_dir = PathBuf::from(env::var_os("CARGO_MANIFEST_DIR").expect("CARGO_MANIFEST_DIR is set by cargo"));
    let dist = manifest_dir.join("frontend").join("dist");
    println!("cargo:rerun-if-changed={}", dist.display());

    if !dist.is_dir() {
        panic!(
            "frontend/dist is missing: run `cd frontend && npm install && npm run build` before compiling the Rust binary"
        );
    }

    let out_dir = PathBuf::from(env::var_os("OUT_DIR").expect("OUT_DIR is set by cargo"));
    let mut generated = File::create(out_dir.join("embedded_frontend.rs"))?;

    let mut assets = Vec::new();
    collect_assets(&dist, &dist, &mut assets)?;
    assets.sort_by(|a, b| a.0.cmp(&b.0));

    writeln!(
        generated,
        "pub fn get(path: &str) -> Option<(&'static [u8], &'static str)> {{"
    )?;
    writeln!(generated, "    match path {{")?;
    for (relative, mime) in assets {
        writeln!(
            generated,
            "        {:?} => Some((include_bytes!(concat!(env!(\"CARGO_MANIFEST_DIR\"), \"/frontend/dist/{relative}\")).as_slice(), {mime:?})),",
            relative
        )?;
    }
    writeln!(generated, "        _ => None,")?;
    writeln!(generated, "    }}")?;
    writeln!(generated, "}}")?;

    Ok(())
}

fn collect_assets(
    root: &Path,
    dir: &Path,
    assets: &mut Vec<(String, &'static str)>,
) -> io::Result<()> {
    for entry in fs::read_dir(dir)? {
        let entry = entry?;
        let path = entry.path();
        let file_type = entry.file_type()?;
        if file_type.is_dir() {
            collect_assets(root, &path, assets)?;
        } else if file_type.is_file() {
            let relative = path
                .strip_prefix(root)
                .expect("asset path is under frontend dist")
                .to_string_lossy()
                .replace('\\', "/");
            let mime = mime_for_path(&path);
            assets.push((relative, mime));
        }
    }
    Ok(())
}

/// 对齐 Go main.go mime.AddExtensionType 的静态资源 MIME 表
fn mime_for_path(path: &Path) -> &'static str {
    match path
        .extension()
        .and_then(|ext| ext.to_str())
        .unwrap_or_default()
        .to_ascii_lowercase()
        .as_str()
    {
        "js" | "mjs" => "application/javascript; charset=utf-8",
        "css" => "text/css; charset=utf-8",
        "html" => "text/html; charset=utf-8",
        "json" | "map" => "application/json",
        "svg" => "image/svg+xml",
        "png" => "image/png",
        "jpg" | "jpeg" => "image/jpeg",
        "gif" => "image/gif",
        "webp" => "image/webp",
        "ico" => "image/x-icon",
        "woff" => "font/woff",
        "woff2" => "font/woff2",
        "ttf" => "font/ttf",
        _ => "application/octet-stream",
    }
}
