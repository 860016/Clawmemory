"""Cython 编译脚本 — 将 .pyx 编译为 .so/.pyd 二进制文件

注意：
- 此脚本仅编译计算引擎模块（decay/conflict/router）
- 授权引擎模块（feature_gate/license_verifier）不再编译，因为：
  1. 它们在 .pyx 版本中永远锁定为 OSS，无法激活 Pro
  2. 真正的授权功能必须由 clawmemory_core (C/Rust wheel) 提供
  3. 编译它们没有意义，只会增加逆向难度但无法阻止破解

架构说明：
- 计算引擎：可独立使用，提供 Pro 功能的计算逻辑
- 授权引擎：必须使用 clawmemory_core，.pyx 版本永远锁定 OSS
"""
from setuptools import setup, Extension
from Cython.Build import cythonize
import os

# 仅编译计算引擎模块（不编译授权引擎）
COMPUTE_MODULES = [
    "app/core/memory_decay.pyx",
    "app/core/conflict_resolver.pyx",
    "app/core/token_router.pyx",
]

# 授权引擎模块不再编译（它们在 .pyx 版本中永远锁定 OSS）
# LICENSE_MODULES = [
#     "app/core/feature_gate.pyx",    # 已移除：永远返回 OSS
#     "app/core/license_verifier.pyx", # 已移除：永远返回空 dict
# ]

extensions = []
for mod in COMPUTE_MODULES:
    if os.path.exists(mod):
        mod_name = mod.replace("/", ".").replace("\\", ".").replace(".pyx", "")
        extensions.append(Extension(mod_name, [mod]))

setup(
    name="clawmemory-compute",
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