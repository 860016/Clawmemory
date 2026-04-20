#!/usr/bin/env python3
"""
Cython 编译脚本 - 将 Pro 模块编译为 .pyd/.so 二进制文件

用法:
    python setup_cython.py build      # 编译
    python setup_cython.py package    # 编译并打包为 zip
    python setup_cython.py clean      # 清理编译产物

输出:
    build/pro_package/    - 编译后的文件
    dist/pro_module.zip   - 分发包（用于 GitHub Releases）
"""

import os
import sys
import shutil
import zipfile
import subprocess
from pathlib import Path
from datetime import datetime

PRO_DIR = Path(__file__).parent / "backend" / "app" / "pro"
BUILD_DIR = Path(__file__).parent / "build" / "pro_package"
DIST_DIR = Path(__file__).parent / "dist"
VERSION = "1.0.0"


def run_cython_build():
    """使用 Cython 编译 Pro 模块"""
    print("=" * 60)
    print("Cython Pro Module Builder")
    print("=" * 60)

    if not PRO_DIR.exists():
        print(f"Error: Pro directory not found: {PRO_DIR}")
        sys.exit(1)

    print(f"\nPro directory: {PRO_DIR}")
    print(f"Build directory: {BUILD_DIR}")

    if BUILD_DIR.exists():
        shutil.rmtree(BUILD_DIR)
    BUILD_DIR.mkdir(parents=True, exist_ok=True)

    py_files = list(PRO_DIR.glob("*.py"))
    py_files = [f for f in py_files if not f.name.startswith("_")]

    if not py_files:
        print("No Python files found to compile")
        sys.exit(1)

    print(f"\nFound {len(py_files)} files to compile:")
    for f in py_files:
        print(f"  - {f.name}")

    print("\nInstalling Cython...")
    subprocess.check_call([sys.executable, "-m", "pip", "install", "cython"])

    print("\nCompiling with Cython...")
    for py_file in py_files:
        print(f"  Compiling {py_file.name}...")
        result = subprocess.run(
            [sys.executable, "-m", "cython", "-3", str(py_file)],
            capture_output=True,
            text=True,
        )
        if result.returncode != 0:
            print(f"    Warning: Cython compilation failed for {py_file.name}")
            print(f"    {result.stderr}")
        else:
            c_file = py_file.with_suffix(".c")
            if c_file.exists():
                print(f"    Generated {c_file.name}")

    print("\nBuilding binary extensions...")
    setup_py_content = '''
from setuptools import setup, Extension
from Cython.Build import cythonize
import os
import glob

extensions = []
for cython_file in glob.glob("*.pyx"):
    module_name = os.path.splitext(cython_file)[0]
    extensions.append(
        Extension(
            module_name,
            [cython_file],
            extra_compile_args=["-O3"],
        )
    )

setup(
    ext_modules=cythonize(
        extensions,
        compiler_directives={
            'language_level': '3',
            'boundscheck': False,
            'wraparound': False,
            'cdivision': True,
        },
        annotate=True,
    ),
)
'''

    for py_file in py_files:
        c_file = py_file.with_suffix(".c")
        if c_file.exists():
            pyx_file = BUILD_DIR / py_file.with_suffix(".pyx").name
            shutil.copy2(c_file, pyx_file)

    setup_py = BUILD_DIR / "setup.py"
    setup_py.write_text(setup_py_content)

    result = subprocess.run(
        [sys.executable, "setup.py", "build_ext", "--inplace"],
        cwd=BUILD_DIR,
        capture_output=True,
        text=True,
    )

    if result.returncode != 0:
        print(f"Warning: Binary compilation had issues:")
        print(result.stderr[:500])
    else:
        print("Binary compilation successful")

    print("\nCopying non-compiled files...")
    for item in PRO_DIR.iterdir():
        if item.name.startswith("__"):
            continue
        if item.is_file() and item.suffix == ".py":
            if not item.name.startswith("_"):
                shutil.copy2(item, BUILD_DIR / item.name)

    manifest = {
        "version": VERSION,
        "build_time": datetime.now().isoformat(),
        "python_version": f"{sys.version_info.major}.{sys.version_info.minor}",
        "platform": sys.platform,
        "files": [f.name for f in BUILD_DIR.iterdir() if f.is_file()],
    }

    import json
    (BUILD_DIR / ".pro_manifest.json").write_text(json.dumps(manifest, indent=2))
    (BUILD_DIR / ".pro_installed").touch()

    print(f"\nBuild complete! Files in: {BUILD_DIR}")
    for f in BUILD_DIR.iterdir():
        if f.is_file():
            size = f.stat().st_size
            print(f"  {f.name}: {size:,} bytes")


def create_package():
    """创建分发包"""
    print("\n" + "=" * 60)
    print("Creating distribution package...")
    print("=" * 60)

    if not BUILD_DIR.exists():
        print("Error: Build directory not found. Run 'build' first.")
        sys.exit(1)

    if DIST_DIR.exists():
        shutil.rmtree(DIST_DIR)
    DIST_DIR.mkdir(parents=True)

    zip_path = DIST_DIR / f"pro_module_v{VERSION}.zip"

    with zipfile.ZipFile(zip_path, 'w', zipfile.ZIP_DEFLATED) as zf:
        for item in BUILD_DIR.rglob("*"):
            if item.is_file():
                arcname = item.relative_to(BUILD_DIR)
                zf.write(item, arcname)

    zip_size = zip_path.stat().st_size
    print(f"\nPackage created: {zip_path}")
    print(f"Package size: {zip_size:,} bytes ({zip_size/1024:.1f} KB)")


def clean():
    """清理编译产物"""
    print("Cleaning build artifacts...")

    dirs_to_clean = [
        Path(__file__).parent / "build",
        DIST_DIR,
    ]

    for d in dirs_to_clean:
        if d.exists():
            shutil.rmtree(d)
            print(f"  Removed: {d}")

    c_files = list(PRO_DIR.glob("*.c"))
    for f in c_files:
        f.unlink()
        print(f"  Removed: {f}")

    print("Clean complete")


def main():
    if len(sys.argv) < 2:
        print("Usage: python setup_cython.py [build|package|clean]")
        sys.exit(1)

    command = sys.argv[1].lower()

    if command == "build":
        run_cython_build()
    elif command == "package":
        run_cython_build()
        create_package()
    elif command == "clean":
        clean()
    else:
        print(f"Unknown command: {command}")
        print("Usage: python setup_cython.py [build|package|clean]")
        sys.exit(1)


if __name__ == "__main__":
    main()
