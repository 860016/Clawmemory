#!/usr/bin/env python3
"""
Pro 模块打包脚本

用法:
    python package_pro.py

输出:
    dist/pro_module.zip - 可直接部署到服务器的 Pro 模块包
"""

import os
import sys
import json
import zipfile
from pathlib import Path
from datetime import datetime

PRO_DIR = Path(__file__).parent / "backend" / "app" / "pro"
DIST_DIR = Path(__file__).parent / "dist"
VERSION = "1.0.0"


def package_pro():
    """打包 Pro 模块"""
    print("=" * 60)
    print("Pro Module Packager")
    print("=" * 60)

    if not PRO_DIR.exists():
        print(f"Error: Pro directory not found: {PRO_DIR}")
        sys.exit(1)

    if DIST_DIR.exists():
        import shutil
        shutil.rmtree(DIST_DIR)
    DIST_DIR.mkdir(parents=True)

    files_to_include = []
    for item in PRO_DIR.iterdir():
        if item.name.startswith("__"):
            continue
        if item.is_file() and item.suffix in ('.py', '.pyd', '.so'):
            files_to_include.append(item)

    if not files_to_include:
        print("No Pro module files found!")
        print("Please ensure Pro modules are in:", PRO_DIR)
        sys.exit(1)

    print(f"\nFound {len(files_to_include)} files:")
    for f in files_to_include:
        print(f"  - {f.name}")

    zip_path = DIST_DIR / "pro_module.zip"

    with zipfile.ZipFile(zip_path, 'w', zipfile.ZIP_DEFLATED) as zf:
        for f in files_to_include:
            zf.write(f, f.name)

        manifest = {
            "version": VERSION,
            "build_time": datetime.now().isoformat(),
            "files": [f.name for f in files_to_include],
        }
        zf.writestr(".pro_manifest.json", json.dumps(manifest, indent=2))
        zf.writestr(".pro_installed", "")

    zip_size = zip_path.stat().st_size
    print(f"\nPackage created: {zip_path}")
    print(f"Package size: {zip_size:,} bytes ({zip_size/1024:.1f} KB)")
    print(f"\nUpload this file to your server and set the download URL in:")
    print(f"  backend/app/config.py -> PRO_DOWNLOAD_URL")


if __name__ == "__main__":
    package_pro()
