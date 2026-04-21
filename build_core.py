#!/usr/bin/env python3
"""
Cython 编译脚本 - 将 core 模块编译为 .pyd/.so 二进制文件

用法:
    python build_core.py build      # 编译
    python build_core.py clean      # 清理编译产物
"""

import os
import sys
import json
import shutil
import subprocess
from pathlib import Path
from datetime import datetime

CORE_DIR = Path(__file__).parent / "backend" / "app" / "core"
BUILD_DIR = Path(__file__).parent / "build" / "core_package"
DIST_DIR = Path(__file__).parent / "dist"
VERSION = "1.0.0"


def run_cython_build():
    """使用 Cython 编译 core 模块"""
    print("=" * 60)
    print("Cython Core Module Builder")
    print("=" * 60)

    if not CORE_DIR.exists():
        print(f"Error: Core directory not found: {CORE_DIR}")
        sys.exit(1)

    if BUILD_DIR.exists():
        shutil.rmtree(BUILD_DIR)
    BUILD_DIR.mkdir(parents=True)

    pyx_files = list(CORE_DIR.glob("*.pyx"))

    if not pyx_files:
        print("No .pyx files found to compile")
        sys.exit(1)

    print(f"\nFound {len(pyx_files)} files to compile:")
    for f in pyx_files:
        print(f"  - {f.name}")

    print("\nInstalling Cython...")
    subprocess.check_call([sys.executable, "-m", "pip", "install", "cython", "-q"])

    # 创建临时构建目录
    temp_build = BUILD_DIR / "temp"
    temp_build.mkdir()

    # 复制源文件
    for f in pyx_files:
        shutil.copy2(f, temp_build / f.name)

    # 创建 setup.py
    setup_content = '''
from setuptools import setup, Extension
from Cython.Build import cythonize
import glob
import os

ext_modules = []
for pyx_file in glob.glob("*.pyx"):
    module_name = os.path.splitext(pyx_file)[0]
    ext_modules.append(
        Extension(
            module_name,
            [pyx_file],
        )
    )

setup(
    ext_modules=cythonize(
        ext_modules,
        compiler_directives={
            'language_level': '3',
            'boundscheck': False,
            'wraparound': False,
        },
    ),
    script_args=['build_ext', '--inplace'],
)
'''
    (temp_build / "setup.py").write_text(setup_content)

    print("\nCompiling with Cython...")
    result = subprocess.run(
        [sys.executable, "setup.py"],
        cwd=temp_build,
        capture_output=True,
        text=True,
    )

    if result.returncode != 0:
        print(f"Compilation error:\n{result.stderr[-1000:]}")
        print("\nFalling back to .py distribution...")
        for f in pyx_files:
            shutil.copy2(f, BUILD_DIR / f.name)
    else:
        print("Compilation successful!")
        # 复制编译产物
        for item in temp_build.iterdir():
            if item.suffix in ('.pyd', '.so', '.py'):
                if not item.name.startswith("_"):
                    shutil.copy2(item, BUILD_DIR / item.name)

    # 清理临时目录
    shutil.rmtree(temp_build)

    # 创建 manifest
    manifest = {
        "version": VERSION,
        "build_time": datetime.now().isoformat(),
        "python_version": f"{sys.version_info.major}.{sys.version_info.minor}",
        "platform": sys.platform,
        "files": [f.name for f in BUILD_DIR.iterdir() if f.is_file() and not f.name.startswith("setup")],
    }
    (BUILD_DIR / ".core_manifest.json").write_text(json.dumps(manifest, indent=2))
    (BUILD_DIR / ".core_installed").touch()

    print(f"\nBuild complete! Files in: {BUILD_DIR}")
    total_size = 0
    for f in BUILD_DIR.iterdir():
        if f.is_file():
            size = f.stat().st_size
            total_size += size
            print(f"  {f.name}: {size:,} bytes")
    print(f"\nTotal size: {total_size:,} bytes ({total_size/1024:.1f} KB)")


def clean():
    """清理编译产物"""
    print("Cleaning build artifacts...")
    for d in [Path(__file__).parent / "build", DIST_DIR]:
        if d.exists():
            shutil.rmtree(d)
            print(f"  Removed: {d}")
    print("Clean complete")


if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python build_core.py [build|clean]")
        sys.exit(1)

    cmd = sys.argv[1].lower()
    if cmd == "build":
        run_cython_build()
    elif cmd == "clean":
        clean()
    else:
        print(f"Unknown command: {cmd}")
