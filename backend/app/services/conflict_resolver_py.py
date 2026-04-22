"""Conflict Resolver - Pure Python fallback"""
from typing import List, Dict, Any, Tuple

def detect_conflict(memory1: Dict[str, Any], memory2: Dict[str, Any]) -> Dict[str, Any]:
    """检测两个记忆是否冲突"""
    content1 = memory1.get("content", "")
    content2 = memory2.get("content", "")
    
    if not content1 or not content2:
        return {"has_conflict": False, "similarity": 0}
    
    # 简单的文本相似度
    words1 = set(content1.lower().split())
    words2 = set(content2.lower().split())
    
    if not words1 or not words2:
        return {"has_conflict": False, "similarity": 0}
    
    intersection = words1 & words2
    union = words1 | words2
    similarity = len(intersection) / len(union) if union else 0
    
    # 相似度高但内容不同可能是冲突
    has_conflict = similarity > 0.5 and content1 != content2
    
    return {
        "has_conflict": has_conflict,
        "similarity": similarity,
        "common_words": list(intersection),
    }

def resolve_conflict(memory1: Dict[str, Any], memory2: Dict[str, Any]) -> Dict[str, Any]:
    """尝试合并冲突的记忆"""
    conflict = detect_conflict(memory1, memory2)
    
    if not conflict["has_conflict"]:
        return {"resolved": True, "merged": memory1}
    
    # 简单的合并策略：保留更新的或更重要的
    importance1 = memory1.get("importance", 0.5)
    importance2 = memory2.get("importance", 0.5)
    
    if importance1 >= importance2:
        merged = dict(memory1)
        merged["conflict_resolved"] = True
        merged["merged_with"] = memory2.get("id")
    else:
        merged = dict(memory2)
        merged["conflict_resolved"] = True
        merged["merged_with"] = memory1.get("id")
    
    return {
        "resolved": True,
        "merged": merged,
        "strategy": "keep_higher_importance",
    }

def scan_for_conflicts(memories: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
    """扫描记忆列表中的冲突"""
    conflicts = []
    for i, m1 in enumerate(memories):
        for m2 in memories[i+1:]:
            result = detect_conflict(m1, m2)
            if result["has_conflict"]:
                conflicts.append({
                    "memory1_id": m1.get("id"),
                    "memory2_id": m2.get("id"),
                    "similarity": result["similarity"],
                })
    return conflicts

def get_conflict_summary(conflicts: List[Dict[str, Any]]) -> Dict[str, Any]:
    return {
        "total_conflicts": len(conflicts),
        "high_similarity": sum(1 for c in conflicts if c.get("similarity", 0) > 0.8),
        "medium_similarity": sum(1 for c in conflicts if 0.5 < c.get("similarity", 0) <= 0.8),
    }
