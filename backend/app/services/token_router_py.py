"""Token Router - Pure Python fallback"""
from typing import Dict, Any, List

def estimate_complexity(text: str) -> Dict[str, Any]:
    """估计文本复杂度"""
    if not text:
        return {"tokens": 0, "complexity": "low", "score": 0}
    
    # 简单的token估算（中文字符 + 英文单词）
    import re
    chinese_chars = len(re.findall(r'[\u4e00-\u9fff]', text))
    english_words = len(re.findall(r'[a-zA-Z]+', text))
    tokens = chinese_chars + english_words
    
    if tokens < 100:
        complexity = "low"
        score = 1
    elif tokens < 500:
        complexity = "medium"
        score = 2
    else:
        complexity = "high"
        score = 3
    
    return {
        "tokens": tokens,
        "complexity": complexity,
        "score": score,
    }

def route_model(text: str, available_models: List[str] = None) -> Dict[str, Any]:
    """根据文本复杂度路由到合适的模型"""
    complexity = estimate_complexity(text)
    
    if available_models is None:
        available_models = ["gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"]
    
    score = complexity["score"]
    if score == 1:
        model = available_models[0] if available_models else "gpt-3.5-turbo"
    elif score == 2:
        model = available_models[1] if len(available_models) > 1 else available_models[0]
    else:
        model = available_models[-1] if available_models else "gpt-4-turbo"
    
    return {
        "model": model,
        "complexity": complexity["complexity"],
        "tokens": complexity["tokens"],
        "reason": f"Complexity score: {score}",
    }

def get_routing_stats(history: List[Dict[str, Any]]) -> Dict[str, Any]:
    """获取路由统计"""
    if not history:
        return {"total": 0, "model_distribution": {}}
    
    model_counts = {}
    for h in history:
        model = h.get("model", "unknown")
        model_counts[model] = model_counts.get(model, 0) + 1
    
    return {
        "total": len(history),
        "model_distribution": model_counts,
    }
