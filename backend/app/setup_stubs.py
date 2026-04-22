"""
ClawMemory 安装脚本 - 自动创建缺失模块的占位文件

这个脚本会在首次安装时运行，确保所有必需的模块都存在。
对于 Pro 功能模块，会创建返回"需要 Pro 授权"的占位实现。
"""
import os
import sys
from pathlib import Path

def create_stub_files():
    """创建缺失模块的占位文件"""
    
    # 获取 app 目录
    app_dir = Path(__file__).parent
    print(f"ClawMemory 安装目录: {app_dir}")
    
    # 1. 确保 core/__init__.py 存在
    core_dir = app_dir / "core"
    core_dir.mkdir(exist_ok=True)
    
    core_init = core_dir / "__init__.py"
    if not core_init.exists():
        core_init.write_text("# Core module\n")
        print("✓ 创建 core/__init__.py")
    
    # 2. 确保 pro_loader.py 存在（如果不存在则创建 OSS 版本）
    pro_loader = core_dir / "pro_loader.py"
    if not pro_loader.exists():
        pro_loader.write_text('''"""
Pro Loader - OSS 版本（占位实现）

当 Pro 模块未安装时，使用此占位实现。
所有 Pro 功能将返回"需要 Pro 授权"状态。
"""
import logging
from pathlib import Path

logger = logging.getLogger(__name__)

PRO_DIR = Path(__file__).parent.parent / "pro"


def is_pro_installed() -> bool:
    """检查 Pro 模块是否已安装"""
    return False


def get_pro_info() -> dict:
    """获取 Pro 模块信息"""
    return {"installed": False}


def get_pro_module(name: str):
    """动态加载 Pro 模块（OSS 版本返回 None）"""
    return None


def get_pro_class(module_name: str, class_name: str):
    """从 Pro 模块获取类（OSS 版本返回 None）"""
    return None


async def download_pro_package(
    download_url: str,
    fallback_urls: list = None,
    timeout: int = 120,
) -> bool:
    """下载 Pro 模块包（OSS 版本）"""
    logger.warning("Pro 模块下载功能需要安装完整版 Pro 模块")
    return False


def uninstall_pro() -> bool:
    """卸载 Pro 模块（OSS 版本）"""
    return True
''')
        print("✓ 创建 core/pro_loader.py (OSS 占位版本)")
    
    # 3. 确保 pro/ 目录存在
    pro_dir = app_dir / "pro"
    pro_dir.mkdir(exist_ok=True)
    
    pro_init = pro_dir / "__init__.py"
    if not pro_init.exists():
        pro_init.write_text("# Pro module\n")
        print("✓ 创建 pro/__init__.py")
    
    # 4. 确保 pro/services/ 目录存在
    pro_services = pro_dir / "services"
    pro_services.mkdir(exist_ok=True)
    
    services_init = pro_services / "__init__.py"
    if not services_init.exists():
        services_init.write_text("# Pro services\n")
        print("✓ 创建 pro/services/__init__.py")
    
    # 5. 确保 pro/routers/ 目录存在
    pro_routers = pro_dir / "routers"
    pro_routers.mkdir(exist_ok=True)
    
    routers_init = pro_routers / "__init__.py"
    if not routers_init.exists():
        routers_init.write_text("# Pro routers\n")
        print("✓ 创建 pro/routers/__init__.py")
    
    # 6. 创建 pro_features.py 占位路由
    pro_features = pro_routers / "pro_features.py"
    if not pro_features.exists():
        pro_features.write_text('''"""Pro Features API - OSS 版本（占位实现）"""
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from app.database import get_db
from app.middleware.auth import get_current_user

router = APIRouter(prefix="/api/v1/pro", tags=["pro-features"])


@router.get("/decay/stats")
def get_decay_stats(_=Depends(get_current_user)):
    """Memory decay stats (requires memory_decay)"""
    raise HTTPException(status_code=403, detail="Pro feature: memory decay")


@router.post("/decay/apply")
def apply_decay(_=Depends(get_current_user)):
    """Apply memory decay (requires memory_decay)"""
    raise HTTPException(status_code=403, detail="Pro feature: memory decay")


@router.get("/conflict/stats")
def get_conflict_stats(_=Depends(get_current_user)):
    """Conflict resolution stats (requires conflict_resolution)"""
    raise HTTPException(status_code=403, detail="Pro feature: conflict resolution")


@router.post("/conflict/resolve")
def resolve_conflict(_=Depends(get_current_user)):
    """Resolve memory conflicts (requires conflict_resolution)"""
    raise HTTPException(status_code=403, detail="Pro feature: conflict resolution")


@router.get("/token/stats")
def get_token_stats(_=Depends(get_current_user)):
    """Token routing stats (requires token_routing)"""
    raise HTTPException(status_code=403, detail="Pro feature: token routing")


@router.post("/token/route")
def route_token(_=Depends(get_current_user)):
    """Route token (requires token_routing)"""
    raise HTTPException(status_code=403, detail="Pro feature: token routing")


@router.get("/features")
def get_pro_features(_=Depends(get_current_user)):
    """List all Pro features"""
    return {
        "features": [
            {"name": "memory_decay", "enabled": False, "description": "Memory decay"},
            {"name": "conflict_resolution", "enabled": False, "description": "Conflict resolution"},
            {"name": "token_routing", "enabled": False, "description": "Token routing"},
            {"name": "ai_extract", "enabled": False, "description": "AI extract"},
            {"name": "auto_graph", "enabled": False, "description": "Auto graph"},
            {"name": "auto_backup", "enabled": False, "description": "Auto backup"},
        ]
    }
''')
        print("✓ 创建 pro/routers/pro_features.py (OSS 占位版本)")
    
    # 7. 创建 wiki_ai_service.py 占位服务
    wiki_ai = pro_services / "wiki_ai_service.py"
    if not wiki_ai.exists():
        wiki_ai.write_text('''"""Wiki AI Service - OSS 版本（占位实现）"""
import logging

logger = logging.getLogger(__name__)


def get_wiki_ai_service_class():
    """获取 WikiAIService 类（OSS 版本返回 None）"""
    logger.warning("Wiki AI service requires Pro module")
    return None
''')
        print("✓ 创建 pro/services/wiki_ai_service.py (OSS 占位版本)")
    
    # 8. 创建 llm_service.py 占位服务
    llm_service = pro_services / "llm_service.py"
    if not llm_service.exists():
        llm_service.write_text('''"""LLM Service - OSS 版本（占位实现）"""
import logging

logger = logging.getLogger(__name__)


class _LLMServiceStub:
    """LLM 服务占位实现"""
    
    def extract_entities(self, text: str):
        logger.warning("LLM extract_entities requires Pro module")
        return []
    
    def discover_relations(self, text: str, entities: list):
        logger.warning("LLM discover_relations requires Pro module")
        return []


def _get_llm_service():
    """获取 LLM 服务（OSS 版本返回占位对象）"""
    return _LLMServiceStub()


llm_service = _get_llm_service()
''')
        print("✓ 创建 pro/services/llm_service.py (OSS 占位版本)")
    
    # 9. 创建 entity_extractor.py 占位服务
    entity_extractor = pro_services / "entity_extractor.py"
    if not entity_extractor.exists():
        entity_extractor.write_text('''"""Entity Extractor - OSS 版本（占位实现）"""
import logging
from sqlalchemy.orm import Session

logger = logging.getLogger(__name__)


class EntityExtractor:
    """实体提取器（OSS 占位实现）"""
    
    def __init__(self, db: Session):
        self.db = db
    
    def extract_from_memory(self, memory_id: int):
        """从记忆中提取实体（OSS 版本返回空列表）"""
        logger.warning("Entity extraction requires Pro module")
        return []
    
    def extract_batch(self, memory_ids: list):
        """批量提取实体（OSS 版本返回空列表）"""
        logger.warning("Batch entity extraction requires Pro module")
        return []
''')
        print("✓ 创建 pro/services/entity_extractor.py (OSS 占位版本)")
    
    # 10. 创建 relation_discovery.py 占位服务
    relation_discovery = pro_services / "relation_discovery.py"
    if not relation_discovery.exists():
        relation_discovery.write_text('''"""Relation Discovery - OSS 版本（占位实现）"""
import logging
from sqlalchemy.orm import Session

logger = logging.getLogger(__name__)


class RelationDiscovery:
    """关系发现（OSS 占位实现）"""
    
    def __init__(self, db: Session):
        self.db = db
    
    def discover_from_memory(self, memory_id: int):
        """从记忆中发现关系（OSS 版本返回空列表）"""
        logger.warning("Relation discovery requires Pro module")
        return []
    
    def discover_batch(self, memory_ids: list):
        """批量发现关系（OSS 版本返回空列表）"""
        logger.warning("Batch relation discovery requires Pro module")
        return []
''')
        print("✓ 创建 pro/services/relation_discovery.py (OSS 占位版本)")
    
    print("\n✓ 所有占位文件创建完成！")
    print("  OSS 版本可以正常启动，Pro 功能将显示'需要 Pro 授权'")


if __name__ == "__main__":
    print("=" * 50)
    print("ClawMemory 安装脚本")
    print("=" * 50)
    create_stub_files()
    print("\n安装完成！")
