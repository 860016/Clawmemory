"""Memory Decay - Pure Python fallback"""
import math
from datetime import datetime, timezone
from typing import List, Dict, Any

def calculate_decay(age_hours: float, importance: float, access_count: int) -> float:
    """计算记忆衰减值"""
    if importance <= 0:
        return 1.0
    time_factor = math.log1p(age_hours) / 10.0
    access_factor = math.log1p(access_count) / 5.0
    decay = time_factor * (1.1 - min(importance, 1.0)) - access_factor
    return max(0.0, min(1.0, decay))

def should_prune(decay_value: float, threshold: float = 0.8) -> bool:
    return decay_value >= threshold

def reinforce(current_strength: float, amount: float = 0.1) -> float:
    return min(1.0, current_strength + amount)

def decay_memory(memory: Dict[str, Any]) -> Dict[str, Any]:
    """对单个记忆进行衰减计算"""
    created_at = memory.get("created_at")
    if isinstance(created_at, str):
        from datetime import datetime
        created_at = datetime.fromisoformat(created_at.replace('Z', '+00:00'))
    
    age_hours = 0
    if created_at:
        age_hours = (datetime.now(timezone.utc) - created_at).total_seconds() / 3600
    
    importance = memory.get("importance", 0.5)
    access_count = memory.get("access_count", 0)
    
    decay_value = calculate_decay(age_hours, importance, access_count)
    memory["decay_value"] = decay_value
    memory["should_prune"] = should_prune(decay_value)
    return memory

def decay_batch(memories: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
    return [decay_memory(m) for m in memories]

def get_decay_stats(memories: List[Dict[str, Any]]) -> Dict[str, Any]:
    if not memories:
        return {"total": 0, "avg_decay": 0, "prune_candidates": 0}
    
    decayed = decay_batch(memories)
    decay_values = [m.get("decay_value", 0) for m in decayed]
    prune_candidates = sum(1 for m in decayed if m.get("should_prune", False))
    
    return {
        "total": len(memories),
        "avg_decay": sum(decay_values) / len(decay_values),
        "prune_candidates": prune_candidates,
    }

def get_decay_stage(decay_value: float) -> str:
    if decay_value < 0.3:
        return "fresh"
    elif decay_value < 0.6:
        return "stable"
    elif decay_value < 0.8:
        return "fading"
    return "critical"

def should_archive(decay_value: float) -> bool:
    return decay_value >= 0.7

def should_trash(decay_value: float) -> bool:
    return decay_value >= 0.9

def should_prune_from_trash(age_days: float) -> bool:
    return age_days >= 30

def get_stage_info(stage: str) -> Dict[str, Any]:
    stages = {
        "fresh": {"color": "green", "label": "新鲜"},
        "stable": {"color": "blue", "label": "稳定"},
        "fading": {"color": "orange", "label": "衰减中"},
        "critical": {"color": "red", "label": "临界"},
    }
    return stages.get(stage, {"color": "gray", "label": "未知"})
