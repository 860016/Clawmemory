use pyo3::prelude::*;
use std::time::{SystemTime, UNIX_EPOCH};

const HALF_LIFE_DEFAULT: f64 = 30.0 * 24.0 * 3600.0; // 30 days
const REINFORCEMENT_FACTOR: f64 = 1.5;
const MIN_IMPORTANCE: f64 = 0.05;

/// Ebbinghaus-style exponential decay
#[pyfunction]
fn calculate_decay(importance: f64, age_seconds: f64, half_life: Option<f64>) -> PyResult<f64> {
    let hl = half_life.unwrap_or(HALF_LIFE_DEFAULT);
    if hl <= 0.0 {
        return Ok(importance);
    }
    let decay_factor = 2.0_f64.powf(-age_seconds / hl);
    Ok((importance * decay_factor).max(0.0))
}

/// Check if a memory should be pruned
#[pyfunction]
fn should_prune(importance: f64) -> PyResult<bool> {
    Ok(importance < MIN_IMPORTANCE)
}

/// Reinforce a memory's importance on access/update (logarithmic scaling)
#[pyfunction]
fn reinforce(importance: f64, factor: Option<f64>) -> PyResult<f64> {
    let f = factor.unwrap_or(REINFORCEMENT_FACTOR);
    let reinforced = importance + (1.0 - importance) * (1.0 - (-0.5 * f * importance).exp());
    Ok(reinforced.min(1.0))
}

/// Calculate decay for a single memory
#[pyfunction]
fn decay_memory(memory_id: i64, current_importance: f64, last_accessed_at: f64, now: Option<f64>, half_life: Option<f64>) -> PyResult<Py<PyAny>> {
    let n = now.unwrap_or_else(|| {
        SystemTime::now().duration_since(UNIX_EPOCH).unwrap_or_default().as_secs_f64()
    });
    let age_seconds = (n - last_accessed_at).max(0.0);
    let new_importance = calculate_decay(current_importance, age_seconds, half_life)?;

    Python::with_gil(|py| {
        let dict = pyo3::types::PyDict::new(py);
        dict.set_item("memory_id", memory_id)?;
        dict.set_item("new_importance", (new_importance * 1_000_000.0).round() / 1_000_000.0)?;
        dict.set_item("should_prune", new_importance < MIN_IMPORTANCE)?;
        dict.set_item("age_seconds", age_seconds)?;
        Ok(dict.into_any().unbind())
    })
}

/// Process a batch of memories for decay
#[pyfunction]
fn decay_batch(memories: Vec<(i64, f64, f64)>, now: Option<f64>, half_life: Option<f64>) -> PyResult<Vec<Py<PyAny>>> {
    let n = now.unwrap_or_else(|| {
        SystemTime::now().duration_since(UNIX_EPOCH).unwrap_or_default().as_secs_f64()
    });

    let results: Vec<Py<PyAny>> = memories.into_iter().map(|(id, importance, last_accessed)| {
        let age = (n - last_accessed).max(0.0);
        let new_imp = calculate_decay(importance, age, half_life).unwrap_or(0.0);

        Python::with_gil(|py| {
            let dict = pyo3::types::PyDict::new(py);
            dict.set_item("memory_id", id).unwrap();
            dict.set_item("new_importance", (new_imp * 1_000_000.0).round() / 1_000_000.0).unwrap();
            dict.set_item("should_prune", new_imp < MIN_IMPORTANCE).unwrap();
            dict.set_item("age_seconds", age).unwrap();
            dict.into_any().unbind()
        })
    }).collect();

    Ok(results)
}

/// Get decay statistics for a set of memories
#[pyfunction]
fn get_decay_stats(memories: Vec<(i64, f64, f64)>, now: Option<f64>) -> PyResult<Py<PyAny>> {
    let n = now.unwrap_or_else(|| {
        SystemTime::now().duration_since(UNIX_EPOCH).unwrap_or_default().as_secs_f64()
    });

    let total = memories.len();
    if total == 0 {
        return Python::with_gil(|py| {
            let dict = pyo3::types::PyDict::new(py);
            dict.set_item("total", 0)?;
            dict.set_item("prune_candidates", 0)?;
            dict.set_item("avg_importance", 0.0)?;
            dict.set_item("decayed_count", 0)?;
            Ok(dict.into_any().unbind())
        });
    }

    let mut prune_candidates = 0usize;
    let mut decayed_count = 0usize;
    let mut importance_sum = 0.0f64;

    for (_, importance, last_accessed) in &memories {
        let age = (n - last_accessed).max(0.0);
        let new_imp = calculate_decay(*importance, age, None).unwrap_or(0.0);
        importance_sum += new_imp;
        if new_imp < MIN_IMPORTANCE {
            prune_candidates += 1;
        }
        if new_imp < importance - 0.01 {
            decayed_count += 1;
        }
    }

    Python::with_gil(|py| {
        let dict = pyo3::types::PyDict::new(py);
        dict.set_item("total", total)?;
        dict.set_item("prune_candidates", prune_candidates)?;
        dict.set_item("avg_importance", (importance_sum / total as f64 * 10_000.0).round() / 10_000.0)?;
        dict.set_item("decayed_count", decayed_count)?;
        Ok(dict.into_any().unbind())
    })
}
