#!/usr/bin/env python3
"""
Windows Cython 编译辅助脚本
自动设置 MSVC 环境并编译 Pro 模块
"""

import os
import sys
import subprocess
from pathlib import Path

def find_msvc():
    """查找 MSVC 安装路径"""
    base = Path("C:/Program Files (x86)/Microsoft Visual Studio/2022/BuildTools")
    if not base.exists():
        base = Path("C:/Program Files/Microsoft Visual Studio/2022/BuildTools")
    if not base.exists():
        return None, None
    
    vc = base / "VC" / "Tools" / "MSVC"
    if not vc.exists():
        return None, None
    
    # 获取最新版本
    versions = sorted(vc.iterdir(), reverse=True)
    if not versions:
        return None, None
    
    version_dir = versions[0]
    bin_dir = version_dir / "bin" / "Hostx64" / "x64"
    return version_dir, bin_dir

def main():
    version_dir, bin_dir = find_msvc()
    if not version_dir:
        print("MSVC not found!")
        sys.exit(1)
    
    print(f"Found MSVC: {version_dir}")
    print(f"Bin dir: {bin_dir}")
    
    # 设置环境变量
    os.environ["PATH"] = str(bin_dir) + os.pathsep + os.environ["PATH"]
    os.environ["INCLUDE"] = str(version_dir / "include")
    os.environ["LIB"] = str(version_dir / "lib" / "x64")
    
    # 验证 cl.exe
    result = subprocess.run([str(bin_dir / "cl.exe")], capture_output=True)
    print(f"cl.exe found: {result.returncode == 0}")
    
    # 运行编译
    print("\nRunning Cython build...")
    subprocess.run([sys.executable, "setup_cython.py", "package"])

if __name__ == "__main__":
    main()
