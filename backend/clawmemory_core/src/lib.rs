use pyo3::prelude::*;
use std::collections::HashSet;
use std::sync::Mutex;

mod memory_decay;
mod conflict_resolver;
mod token_router;

// ========== 全局状态 ==========
static FEATURES: Mutex<Option<HashSet<String>>> = Mutex::new(None);
static TIER: Mutex<Option<String>> = Mutex::new(None);
static LICENSE_HASH: Mutex<Option<String>> = Mutex::new(None);

/// 计算授权数据哈希（防篡改）
fn compute_hash(tier: &str, features: &HashSet<String>) -> String {
    use sha2::{Sha256, Digest};
    let mut sorted_features: Vec<&String> = features.iter().collect();
    sorted_features.sort();
    let raw = format!("{}|{}", tier, sorted_features.iter().map(|s| s.as_str()).collect::<Vec<_>>().join(","));
    let mut hasher = Sha256::new();
    hasher.update(raw.as_bytes());
    hex::encode(hasher.finalize())[..16].to_string()
}

/// 验证授权数据完整性
#[pyfunction]
fn verify_integrity() -> PyResult<bool> {
    let features = FEATURES.lock().unwrap();
    let tier = TIER.lock().unwrap();
    let hash = LICENSE_HASH.lock().unwrap();

    match (&*features, &*tier, &*hash) {
        (None, _, _) | (_, None, _) => {
            match &*tier {
                Some(t) => Ok(t == "oss"),
                None => Ok(true),
            }
        }
        (Some(f), Some(t), Some(h)) => {
            let expected = compute_hash(t, f);
            Ok(h == &expected)
        }
        (Some(_), Some(t), None) => Ok(t == "oss"),
        _ => Ok(false),
    }
}

/// 设置授权信息
#[pyfunction]
fn set_license(tier: &str, features: Vec<String>) -> PyResult<()> {
    let feature_set: HashSet<String> = features.into_iter().collect();
    let hash = compute_hash(tier, &feature_set);

    *FEATURES.lock().unwrap() = Some(feature_set);
    *TIER.lock().unwrap() = Some(tier.to_string());
    *LICENSE_HASH.lock().unwrap() = Some(hash);

    Ok(())
}

/// 检查功能是否启用
#[pyfunction]
fn check_feature(feature: &str) -> PyResult<bool> {
    let features = FEATURES.lock().unwrap();
    match &*features {
        Some(f) => Ok(f.contains(feature)),
        None => Ok(false),
    }
}

/// 获取当前授权等级
#[pyfunction]
fn get_tier() -> PyResult<String> {
    let tier = TIER.lock().unwrap();
    match &*tier {
        Some(t) => Ok(t.clone()),
        None => Ok("oss".to_string()),
    }
}

/// 重置授权状态
#[pyfunction]
fn reset() -> PyResult<()> {
    *FEATURES.lock().unwrap() = None;
    *TIER.lock().unwrap() = None;
    *LICENSE_HASH.lock().unwrap() = None;
    Ok(())
}

/// 验证 RSA 签名的授权数据
#[pyfunction]
fn verify_license(license_data_b64: &str, public_key_pem: &str) -> PyResult<bool> {
    use rsa::pkcs1v15::VerifyingKey;
    use rsa::signature::Verifier;
    use rsa::RsaPublicKey;
    use sha2::Sha256;

    let raw = match base64::engine::general_purpose::STANDARD.decode(license_data_b64) {
        Ok(r) => r,
        Err(_) => return Ok(false),
    };

    let payload: serde_json::Value = match serde_json::from_slice(&raw) {
        Ok(p) => p,
        Err(_) => return Ok(false),
    };

    let data_str = match payload.get("data").and_then(|v| v.as_str()) {
        Some(d) => d,
        None => return Ok(false),
    };

    let signature_b64 = match payload.get("signature").and_then(|v| v.as_str()) {
        Some(s) => s,
        None => return Ok(false),
    };

    let signature_bytes = match base64::engine::general_purpose::STANDARD.decode(signature_b64) {
        Ok(s) => s,
        Err(_) => return Ok(false),
    };

    let public_key = match RsaPublicKey::from_pem(public_key_pem) {
        Ok(k) => k,
        Err(_) => return Ok(false),
    };

    let verifying_key = VerifyingKey::<Sha256>::new(public_key);
    match verifying_key.verify(data_str.as_bytes(), &signature_bytes.into()) {
        Ok(()) => Ok(true),
        Err(_) => Ok(false),
    }
}

/// 获取编译信息（用于防篡改检测）
#[pyfunction]
fn get_build_info() -> PyResult<String> {
    Ok(format!(
        "clawmemory-core-2.0.0 rust-pyo3 {}",
        env!("CARGO_PKG_VERSION")
    ))
}

#[pymodule]
fn clawmemory_core(m: &Bound<'_, PyModule>) -> PyResult<()> {
    // Feature gate
    m.add_function(wrap_pyfunction!(set_license, m)?)?;
    m.add_function(wrap_pyfunction!(check_feature, m)?)?;
    m.add_function(wrap_pyfunction!(get_tier, m)?)?;
    m.add_function(wrap_pyfunction!(reset, m)?)?;
    m.add_function(wrap_pyfunction!(verify_integrity, m)?)?;
    m.add_function(wrap_pyfunction!(verify_license, m)?)?;
    m.add_function(wrap_pyfunction!(get_build_info, m)?)?;

    // Memory decay
    m.add_function(wrap_pyfunction!(memory_decay::calculate_decay, m)?)?;
    m.add_function(wrap_pyfunction!(memory_decay::should_prune, m)?)?;
    m.add_function(wrap_pyfunction!(memory_decay::reinforce, m)?)?;
    m.add_function(wrap_pyfunction!(memory_decay::decay_memory, m)?)?;
    m.add_function(wrap_pyfunction!(memory_decay::decay_batch, m)?)?;
    m.add_function(wrap_pyfunction!(memory_decay::get_decay_stats, m)?)?;

    // Conflict resolver
    m.add_function(wrap_pyfunction!(conflict_resolver::detect_conflict, m)?)?;
    m.add_function(wrap_pyfunction!(conflict_resolver::resolve_conflict, m)?)?;
    m.add_function(wrap_pyfunction!(conflict_resolver::scan_for_conflicts, m)?)?;
    m.add_function(wrap_pyfunction!(conflict_resolver::get_conflict_summary, m)?)?;

    // Token router
    m.add_function(wrap_pyfunction!(token_router::estimate_complexity, m)?)?;
    m.add_function(wrap_pyfunction!(token_router::route_model, m)?)?;
    m.add_function(wrap_pyfunction!(token_router::get_routing_stats, m)?)?;

    Ok(())
}
