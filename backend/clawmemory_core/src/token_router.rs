use pyo3::prelude::*;
use std::collections::HashMap;

/// Estimate message complexity on a 0.0-1.0 scale
#[pyfunction]
fn estimate_complexity(message: &str, context_length: Option<u32>, has_code: Option<bool>, has_reasoning: Option<bool>) -> PyResult<f64> {
    let ctx_len = context_length.unwrap_or(0) as f64;
    let code = has_code.unwrap_or(false) || detect_code(message);
    let reasoning = has_reasoning.unwrap_or(false) || detect_reasoning(message);

    let mut score = 0.0f64;

    // Length factor (logarithmic)
    let msg_len = message.len() as f64;
    let length_score = (1.0 + msg_len).ln() / (1.0 + 4000.0f64).ln();
    score += length_score * 0.25;

    // Context factor
    let ctx_score = (ctx_len / 50.0).min(1.0);
    score += ctx_score * 0.2;

    // Code detection
    if code {
        score += 0.25;
    }

    // Reasoning detection
    if reasoning {
        score += 0.3;
    }

    Ok(score.min(1.0))
}

fn detect_code(message: &str) -> bool {
    let indicators = ["```", "def ", "class ", "import ", "function ", "return ", "=>", "->", "fn "];
    indicators.iter().any(|ind| message.contains(ind))
}

fn detect_reasoning(message: &str) -> bool {
    let keywords = [
        "analyze", "explain", "why", "how does", "compare", "evaluate",
        "reason", "think about", "consider", "implication", "consequence",
        "分析", "解释", "为什么", "如何", "比较", "评估", "推理",
    ];
    let msg_lower = message.to_lowercase();
    keywords.iter().any(|kw| msg_lower.contains(kw))
}

/// Route a message to the most appropriate model
#[pyfunction]
fn route_model(message: &str, available_models: Vec<Py<PyAny>>, context_length: Option<u32>, user_tier: Option<&str>, budget_remaining: Option<f64>) -> PyResult<Py<PyAny>> {
    let ctx_len = context_length.unwrap_or(0);
    let tier = user_tier.unwrap_or("oss");
    let budget = budget_remaining;

    let complexity = estimate_complexity(message, Some(ctx_len), None, None)?;

    Python::with_gil(|py| {
        // Categorize models by tier
        let mut categorized: HashMap<&str, Vec<Py<PyAny>>> = HashMap::new();
        categorized.insert("lightweight", Vec::new());
        categorized.insert("standard", Vec::new());
        categorized.insert("powerful", Vec::new());

        for m in &available_models {
            let model_tier: String = m.bind(py).getattr("tier")?.extract::<String>().unwrap_or_else(|_| "standard".to_string());
            let t = match model_tier.as_str() {
                "lightweight" => "lightweight",
                "powerful" => "powerful",
                _ => "standard",
            };
            categorized.get_mut(t).unwrap().push((*m).clone_ref(py));
        }

        // Determine target tier
        let target_tier = if complexity <= 0.3 {
            "lightweight"
        } else if complexity <= 0.7 {
            "standard"
        } else {
            "powerful"
        };

        let mut target = target_tier;
        if let Some(b) = budget {
            if b < 0.01 {
                target = "lightweight";
            }
        }
        if tier == "oss" && complexity < 0.7 {
            target = "lightweight";
        }

        // Select model with fallback
        let mut selected: Option<Py<PyAny>> = None;
        let mut routing_reason = "complexity_based".to_string();
        for tier_order in [target, "lightweight", "standard", "powerful"] {
            if let Some(candidates) = categorized.get(tier_order) {
                if !candidates.is_empty() {
                    selected = Some(candidates[0].clone_ref(py));
                    if tier_order != target {
                        routing_reason = format!("fallback_to_{}", tier_order);
                    }
                    break;
                }
            }
        }

        let selected_model = selected.unwrap_or_else(|| available_models[0].clone_ref(py));
        if routing_reason == "complexity_based" && available_models.is_empty() {
            routing_reason = "first_available".to_string();
        }

        // Estimate cost
        let model_tier: String = selected_model.bind(py).getattr("tier")?.extract::<String>().unwrap_or_else(|_| "standard".to_string());
        let cost_per_1k = match model_tier.as_str() {
            "lightweight" => 0.00015,
            "powerful" => 0.015,
            _ => 0.003,
        };
        let word_count = message.split_whitespace().count() as f64;
        let estimated_tokens = (word_count * 1.3 + ctx_len as f64 * 50.0) as u32;
        let estimated_cost = (estimated_tokens as f64 / 1000.0) * cost_per_1k;

        let dict = pyo3::types::PyDict::new(py);
        dict.set_item("selected_model", selected_model)?;
        dict.set_item("complexity", (complexity * 10_000.0).round() / 10_000.0)?;
        dict.set_item("routing_reason", &routing_reason)?;
        dict.set_item("estimated_cost", (estimated_cost * 1_000_000.0).round() / 1_000_000.0)?;
        dict.set_item("estimated_tokens", estimated_tokens)?;
        Ok(dict.into_any().unbind())
    })
}

/// Analyze routing history for optimization insights
#[pyfunction]
fn get_routing_stats(routing_history: Vec<Py<PyAny>>) -> PyResult<Py<PyAny>> {
    Python::with_gil(|py| {
        let total = routing_history.len();
        if total == 0 {
            let dict = pyo3::types::PyDict::new(py);
            dict.set_item("total", 0)?;
            dict.set_item("distribution", HashMap::<String, u32>::new())?;
            dict.set_item("avg_complexity", 0.0)?;
            dict.set_item("total_cost", 0.0)?;
            return Ok(dict.into_any().unbind());
        }

        let mut distribution = HashMap::new();
        distribution.insert("lightweight".to_string(), 0u32);
        distribution.insert("standard".to_string(), 0u32);
        distribution.insert("powerful".to_string(), 0u32);
        let mut complexity_sum = 0.0f64;
        let mut total_cost = 0.0f64;

        for entry in &routing_history {
            let model = entry.bind(py).getitem("selected_model")?;
            let tier: String = model.getattr("tier")?.extract::<String>().unwrap_or_else(|_| "standard".to_string());
            if let Some(count) = distribution.get_mut(&tier) {
                *count += 1;
            }
            let complexity: f64 = entry.bind(py).getitem("complexity")?.extract().unwrap_or(0.0);
            complexity_sum += complexity;
            let cost: f64 = entry.bind(py).getitem("estimated_cost")?.extract().unwrap_or(0.0);
            total_cost += cost;
        }

        let dict = pyo3::types::PyDict::new(py);
        dict.set_item("total", total)?;
        dict.set_item("distribution", distribution)?;
        dict.set_item("avg_complexity", (complexity_sum / total as f64 * 10_000.0).round() / 10_000.0)?;
        dict.set_item("total_cost", (total_cost * 1_000_000.0).round() / 1_000_000.0)?;
        Ok(dict.into_any().unbind())
    })
}
