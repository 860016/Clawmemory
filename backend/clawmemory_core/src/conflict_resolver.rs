use pyo3::prelude::*;
use std::collections::HashMap;

const SEVERITY_LOW: &str = "low";
const SEVERITY_MEDIUM: &str = "medium";
const SEVERITY_HIGH: &str = "high";

const STRATEGY_KEEP_NEW: &str = "keep_new";
const STRATEGY_KEEP_OLD: &str = "keep_old";
const STRATEGY_MERGE: &str = "merge";
const STRATEGY_FLAG: &str = "flag_for_review";

const SAFETY_KEYS: &[&str] = &["password", "secret", "api_key", "token", "permission", "access"];

const OPPOSITES_TRUE: &[&str] = &["true", "yes", "是", "对", "correct", "enabled", "1"];
const OPPOSITES_FALSE: &[&str] = &["false", "no", "否", "错", "incorrect", "disabled", "0"];

fn keys_similar(a: &str, b: &str) -> bool {
    if a.is_empty() || b.is_empty() {
        return false;
    }
    let a_lower = a.to_lowercase().replace('_', " ").replace('-', " ");
    let b_lower = b.to_lowercase().replace('_', " ").replace('-', " ");
    a_lower.contains(&b_lower) || b_lower.contains(&a_lower)
}

fn extract_number(value: &str) -> Option<f64> {
    let cleaned: String = value.split_whitespace().next()
        .unwrap_or("")
        .chars()
        .filter(|c| c.is_ascii_digit() || *c == '.' || *c == '-')
        .collect();
    cleaned.parse().ok()
}

fn is_date_like(value: &str) -> bool {
    // Check common date patterns
    (value.contains('-') && value.chars().filter(|c| c.is_ascii_digit()).count() >= 6)
        || value.contains('/')
        || value.contains("年")
}

fn is_opposite_truth(val_a: &str, val_b: &str) -> bool {
    let a = val_a.trim().to_lowercase();
    let b = val_b.trim().to_lowercase();
    let a_in_true = OPPOSITES_TRUE.contains(&a.as_str());
    let a_in_false = OPPOSITES_FALSE.contains(&a.as_str());
    let b_in_true = OPPOSITES_TRUE.contains(&b.as_str());
    let b_in_false = OPPOSITES_FALSE.contains(&b.as_str());
    (a_in_true && b_in_false) || (a_in_false && b_in_true)
}

fn assess_severity(key: &str, val_a: &str, val_b: &str) -> &'static str {
    let key_lower = key.to_lowercase();
    if SAFETY_KEYS.iter().any(|sk| key_lower.contains(sk)) {
        return SEVERITY_HIGH;
    }
    if let (Some(na), Some(nb)) = (extract_number(val_a), extract_number(val_b)) {
        if (na - nb).abs() > f64::EPSILON {
            return SEVERITY_MEDIUM;
        }
    }
    if is_date_like(val_a) && is_date_like(val_b) {
        return SEVERITY_MEDIUM;
    }
    if is_opposite_truth(val_a, val_b) {
        return SEVERITY_HIGH;
    }
    SEVERITY_LOW
}

fn auto_select_strategy(conflict: &HashMap<String, String>) -> &'static str {
    let severity = conflict.get("severity").map(|s| s.as_str()).unwrap_or(SEVERITY_LOW);
    if severity == SEVERITY_HIGH {
        return STRATEGY_FLAG;
    }
    let source_a = conflict.get("source_a").map(|s| s.as_str()).unwrap_or("");
    let source_b = conflict.get("source_b").map(|s| s.as_str()).unwrap_or("");
    if (source_a == "user" || source_a == "manual") && source_b != "user" && source_b != "manual" {
        return STRATEGY_KEEP_OLD;
    }
    if (source_b == "user" || source_b == "manual") && source_a != "user" && source_a != "manual" {
        return STRATEGY_KEEP_NEW;
    }
    let layer_a = conflict.get("layer_a").map(|s| s.as_str()).unwrap_or("");
    let layer_b = conflict.get("layer_b").map(|s| s.as_str()).unwrap_or("");
    if layer_a == "preference" && layer_b != "preference" {
        return STRATEGY_KEEP_OLD;
    }
    if layer_b == "preference" && layer_a != "preference" {
        return STRATEGY_KEEP_NEW;
    }
    if severity == SEVERITY_LOW {
        STRATEGY_MERGE
    } else {
        STRATEGY_FLAG
    }
}

/// Detect conflict between two memories
#[pyfunction]
fn detect_conflict(memory_a: Py<PyAny>, memory_b: Py<PyAny>) -> PyResult<Option<Py<PyAny>>> {
    Python::with_gil(|py| {
        let a = memory_a.bind(py);
        let b = memory_b.bind(py);

        let key_a: String = a.getattr("key")?.extract()?;
        let key_b: String = b.getattr("key")?.extract()?;

        if key_a != key_b && !keys_similar(&key_a, &key_b) {
            return Ok(None);
        }

        let val_a: String = a.getattr("value")?.extract()?;
        let val_b: String = b.getattr("value")?.extract()?;

        if val_a.trim().to_lowercase() == val_b.trim().to_lowercase() {
            return Ok(None);
        }

        let severity = assess_severity(&key_a, &val_a, &val_b);

        let id_a: i64 = a.getattr("id")?.extract()?;
        let id_b: i64 = b.getattr("id")?.extract()?;
        let layer_a: String = a.getattr("layer")?.extract()?;
        let layer_b: String = b.getattr("layer")?.extract()?;
        let source_a: String = a.getattr("source")?.extract()?;
        let source_b: String = b.getattr("source")?.extract()?;

        let dict = pyo3::types::PyDict::new(py);
        dict.set_item("memory_a_id", id_a)?;
        dict.set_item("memory_b_id", id_b)?;
        dict.set_item("key", &key_a)?;
        dict.set_item("value_a", &val_a)?;
        dict.set_item("value_b", &val_b)?;
        dict.set_item("severity", severity)?;
        dict.set_item("layer_a", &layer_a)?;
        dict.set_item("layer_b", &layer_b)?;
        dict.set_item("source_a", &source_a)?;
        dict.set_item("source_b", &source_b)?;

        Ok(Some(dict.into_any().unbind()))
    })
}

/// Resolve a detected conflict
#[pyfunction]
fn resolve_conflict(conflict: Py<PyAny>, strategy: Option<String>) -> PyResult<Py<PyAny>> {
    Python::with_gil(|py| {
        let c = conflict.bind(py);

        let strat = strategy.unwrap_or_else(|| {
            let mut m = HashMap::new();
            let sev: String = c.getitem("severity")?.extract()?;
            let sa: String = c.getitem("source_a")?.extract()?;
            let sb: String = c.getitem("source_b")?.extract()?;
            let la: String = c.getitem("layer_a")?.extract()?;
            let lb: String = c.getitem("layer_b")?.extract()?;
            m.insert("severity".to_string(), sev);
            m.insert("source_a".to_string(), sa);
            m.insert("source_b".to_string(), sb);
            m.insert("layer_a".to_string(), la);
            m.insert("layer_b".to_string(), lb);
            auto_select_strategy(&m).to_string()
        });

        let id_a: i64 = c.getitem("memory_a_id")?.extract()?;
        let id_b: i64 = c.getitem("memory_b_id")?.extract()?;
        let val_a: String = c.getitem("value_a")?.extract()?;
        let val_b: String = c.getitem("value_b")?.extract()?;

        let (winner, action, merged) = match strat.as_str() {
            "keep_new" => {
                if id_b > id_a {
                    ("b", "update_a", None)
                } else {
                    ("a", "update_b", None)
                }
            }
            "keep_old" => {
                if id_a < id_b {
                    ("a", "update_b", None)
                } else {
                    ("b", "update_a", None)
                }
            }
            "merge" => {
                ("both", "merge", Some(format!("[merged: {} | {}]", val_a.trim(), val_b.trim())))
            }
            _ => {
                (":none", "flag_for_review", None)
            }
        };

        let dict = pyo3::types::PyDict::new(py);
        dict.set_item("conflict", conflict)?;
        dict.set_item("strategy", &strat)?;
        dict.set_item("winner", winner)?;
        dict.set_item("action", action)?;
        if let Some(mv) = merged {
            dict.set_item("merged_value", mv)?;
        }
        Ok(dict.into_any().unbind())
    })
}

/// Scan a list of memories for conflicts
#[pyfunction]
fn scan_for_conflicts(memories: Vec<Py<PyAny>>) -> PyResult<Vec<Py<PyAny>>> {
    Python::with_gil(|py| {
        // Extract key/id from each memory for grouping
        let mut by_key: HashMap<String, Vec<usize>> = HashMap::new();
        for (i, m) in memories.iter().enumerate() {
            let key: String = m.bind(py).getattr("key")?.extract()?;
            if !key.is_empty() {
                by_key.entry(key).or_default().push(i);
            }
        }

        let mut conflicts = Vec::new();
        for indices in by_key.values() {
            if indices.len() < 2 {
                continue;
            }
            for i in 0..indices.len() {
                for j in (i + 1)..indices.len() {
                    if let Some(conflict) = detect_conflict(
                        memories[indices[i]].clone_ref(py),
                        memories[indices[j]].clone_ref(py),
                    )? {
                        conflicts.push(conflict);
                    }
                }
            }
        }
        Ok(conflicts)
    })
}

/// Get summary statistics of detected conflicts
#[pyfunction]
fn get_conflict_summary(conflicts: Vec<Py<PyAny>>) -> PyResult<Py<PyAny>> {
    Python::with_gil(|py| {
        let mut by_severity = HashMap::new();
        by_severity.insert("low".to_string(), 0usize);
        by_severity.insert("medium".to_string(), 0usize);
        by_severity.insert("high".to_string(), 0usize);
        let mut auto_resolvable = 0usize;
        let mut needs_review = 0usize;

        for c in &conflicts {
            let sev: String = c.bind(py).getitem("severity")?.extract()?;
            *by_severity.entry(sev.clone()).or_insert(0) += 1;

            let mut m = HashMap::new();
            let sa: String = c.bind(py).getitem("source_a")?.extract()?;
            let sb: String = c.bind(py).getitem("source_b")?.extract()?;
            let la: String = c.bind(py).getitem("layer_a")?.extract()?;
            let lb: String = c.bind(py).getitem("layer_b")?.extract()?;
            m.insert("severity".to_string(), sev);
            m.insert("source_a".to_string(), sa);
            m.insert("source_b".to_string(), sb);
            m.insert("layer_a".to_string(), la);
            m.insert("layer_b".to_string(), lb);

            if auto_select_strategy(&m) == "flag_for_review" {
                needs_review += 1;
            } else {
                auto_resolvable += 1;
            }
        }

        let dict = pyo3::types::PyDict::new(py);
        dict.set_item("total", conflicts.len())?;
        dict.set_item("by_severity", by_severity)?;
        dict.set_item("auto_resolvable", auto_resolvable)?;
        dict.set_item("needs_review", needs_review)?;
        Ok(dict.into_any().unbind())
    })
}
