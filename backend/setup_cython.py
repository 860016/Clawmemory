"""Cython 编译脚本 — 将 .pyx 编译为 .so/.pyd 二进制文件"""
from setuptools import setup, Extension
from Cython.Build import cythonize
import os

# 需要编译的核心模块
CORE_MODULES = [
    "app/core/feature_gate.pyx",
    "app/core/license_verifier.pyx",
    "app/core/token_router.pyx",
    "app/core/memory_decay.pyx",
    "app/core/conflict_resolver.pyx",
]

extensions = []
for mod in CORE_MODULES:
    if os.path.exists(mod):
        mod_name = mod.replace("/", ".").replace("\\", ".").replace(".pyx", "")
        extensions.append(Extension(mod_name, [mod]))

setup(
    name="clawmemory-core",
    ext_modules=cythonize(
        extensions,
        compiler_directives={
            "language_level": "3",
            "boundscheck": False,
            "wraparound": False,
            "cdivision": True,
        },
    ),
)
