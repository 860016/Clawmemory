<template>
  <div class="settings-page">
    <div class="page-header">
      <h1>⚙️ {{ $t('settings.title') }}</h1>
    </div>

    <div class="settings-layout">
      <div class="settings-grid">
      <!-- 授权管理 -->
      <div class="settings-card" :class="{ 'section-highlight': activeSection === 'license' }" id="settings-license">
        <div class="card-title">🔑 {{ $t('settings.license') }}</div>
        <div class="license-status" v-if="license.active">
          <div class="status-row">
            <span class="status-label">{{ $t('settings.version') }}</span>
            <span class="status-value pro">{{ license.tier === 'enterprise' ? 'Enterprise' : 'Pro' }}</span>
          </div>
          <div class="status-row" v-if="license.key_hash">
            <span class="status-label">{{ $t('settings.licenseKey') }}</span>
            <span class="status-value">{{ license.key_hash }}••••</span>
          </div>
          <div class="status-row" v-if="license.expires_at">
            <span class="status-label">{{ $t('settings.expiresAt') }}</span>
            <span class="status-value" :class="{ 'text-warning': isExpiringSoon }">{{ formatDate(license.expires_at) }}</span>
          </div>
          <div class="status-row" v-if="license.device_slot">
            <span class="status-label">{{ $t('settings.deviceSlot') }}</span>
            <span class="status-value">{{ license.device_slot }}</span>
          </div>
          <div class="status-row" v-if="license.fingerprint">
            <span class="status-label">{{ $t('settings.deviceFingerprint') }}</span>
            <span class="status-value">{{ license.fingerprint }}•••</span>
          </div>
          <div class="status-row">
            <span class="status-label">{{ $t('settings.features') }}</span>
            <div class="feature-tags">
              <span class="ftag" v-for="f in license.features" :key="f">{{ featureLabels[f] || f }}</span>
            </div>
          </div>
          <div class="status-row" v-if="!license.is_valid && license.active">
            <span class="status-label">{{ $t('settings.licenseStatus') }}</span>
            <span class="status-value text-warning">{{ $t('settings.gracePeriodWarning') }}</span>
          </div>
          <el-button type="danger" plain size="small" @click="deactivateLicense" style="margin-top: 12px">{{ $t('settings.cancelLicense') }}</el-button>
        </div>
        <div v-else class="license-free">
          <div class="free-badge">{{ $t('settings.freeBadge') }}</div>
          <p class="free-desc">{{ $t('settings.freeDesc') }}</p>
          <div class="pricing">
            <div class="price-card">
              <div class="price-name">Pro {{ $t('settings.proAnnual') }}</div>
              <div class="price-amount">{{ $t('settings.proAnnualPrice') }}</div>
              <ul class="price-features">
                <li v-for="(f, i) in $tm('settings.proFeatures')" :key="i">{{ f }}</li>
              </ul>
            </div>
            <div class="price-card featured">
              <div class="price-badge">{{ $t('settings.recommended') }}</div>
              <div class="price-name">Pro {{ $t('settings.proLifetime') }}</div>
              <div class="price-amount">{{ $t('settings.proLifetimePrice') }}</div>
              <ul class="price-features">
                <li v-for="(f, i) in $tm('settings.lifetimeExtra')" :key="i">{{ f }}</li>
              </ul>
            </div>
          </div>
          <div class="activate-section">
            <el-input v-model="licenseKey" :placeholder="$t('settings.licensePlaceholder')" class="license-input" />
            <el-button type="primary" @click="activateLicense" :loading="activating">{{ $t('settings.activateLicense') }}</el-button>
          </div>
        </div>
      </div>

      <!-- 语言设置 -->
      <div class="settings-card">
        <div class="card-title">◇ {{ currentLocale === 'zh' ? 'Language' : '界面语言' }}</div>
        <div class="setting-item">
          <span>{{ currentLocale === 'zh' ? 'Language' : '界面语言' }}</span>
          <el-select v-model="currentLocale" @change="changeLocale" style="width: 140px">
            <el-option v-if="currentLocale === 'zh'" :label="$t('settings.english')" value="en" />
            <el-option v-else :label="$t('settings.chinese')" value="zh" />
          </el-select>
        </div>
      </div>

      <!-- AI 配置 -->
      <div class="settings-card" id="settings-ai">
        <div class="card-title">🧠 {{ $t('settings.aiConfig') }}</div>
        <div class="setting-item" v-if="aiConfig">
          <span>{{ $t('settings.aiProvider') }}</span>
          <span class="setting-desc">
            <el-tag :type="aiConfig.provider_source === 'reasoning' ? 'success' : 'primary'" size="small">
              {{ aiConfig.provider_name || aiConfig.provider_id || $t('settings.aiNotConfigured') }}
            </el-tag>
            <el-tag v-if="aiConfig.is_pro" type="warning" size="small" style="margin-left: 4px">Pro</el-tag>
          </span>
        </div>
        <div class="setting-item" v-if="aiConfig">
          <span>{{ $t('settings.aiModel') }}</span>
          <span class="setting-desc">{{ aiConfig.model || '-' }}</span>
        </div>
        <div class="setting-item" v-if="aiConfig && aiConfig.is_pro">
          <span>{{ $t('settings.aiCustomProvider') }}</span>
          <el-button size="small" @click="showAIConfigDialog = true">{{ $t('settings.aiConfigure') }}</el-button>
        </div>
        <div class="setting-item" v-if="!aiConfig">
          <span>{{ $t('settings.aiProvider') }}</span>
          <el-button size="small" @click="loadAIConfig" :loading="aiLoading">{{ $t('common.retry') }}</el-button>
        </div>
        <div class="setting-item">
          <span>{{ $t('settings.aiTestConnection') }}</span>
          <el-button size="small" type="primary" @click="testAIConnection" :loading="aiTesting">{{ $t('settings.aiTest') }}</el-button>
        </div>
        <div v-if="aiTestResult" class="ai-test-result" style="margin-top: 8px; padding: 10px; background: var(--cm-bg-secondary); border-radius: 6px; font-size: 12px">
          <span :style="{ color: aiTestResult.success ? 'var(--cm-success)' : 'var(--cm-danger)' }">
            {{ aiTestResult.success ? '✓ ' + $t('settings.aiTestSuccess') : '✗ ' + (aiTestResult.error || $t('settings.aiTestFailed')) }}
          </span>
        </div>
        <div class="setting-item" v-if="aiUsage">
          <span>{{ $t('settings.aiUsage') }}</span>
          <span class="setting-desc">{{ aiUsage.total_calls || 0 }} {{ $t('settings.aiCalls') }}</span>
        </div>
        <div class="ai-config-hint" v-if="aiConfig && !aiConfig.provider_id" style="margin-top: 8px; padding: 10px; background: var(--cm-bg-secondary); border-radius: 6px; font-size: 12px; color: var(--cm-text-muted)">
          <div style="font-weight: 600; margin-bottom: 4px">{{ $t('settings.aiConfigHint') }}</div>
          <div>{{ $t('settings.aiConfigHintDesc') }}</div>
        </div>
      </div>

      <!-- 安全设置 -->
      <div class="settings-card" :class="{ 'section-highlight': activeSection === 'security' }" id="settings-security">
        <div class="card-title">◇ {{ $t('settings.security') }}</div>
        <div class="setting-item">
          <span>{{ $t('settings.password') }}</span>
          <el-button size="small" @click="showPasswordDialog = true">
            {{ passwordSet ? $t('settings.changePassword') : $t('settings.setPassword') }}
          </el-button>
        </div>
      </div>

      <!-- 风险开关 -->
      <div class="settings-card" id="settings-risk-switches">
        <div class="card-title">⚠️ {{ $t('riskSwitch.title') }}</div>
        <p class="setting-desc" style="margin-bottom: 12px">{{ $t('riskSwitch.desc') }}</p>
        <div class="risk-category">
          <div class="risk-category-title">{{ $t('riskSwitch.categoryMemory') }}</div>
          <div class="setting-item">
            <span>{{ $t('riskSwitch.allowCrossAgentRead') }}</span>
            <el-switch v-model="riskSwitches.allow_cross_agent_read" @change="saveRiskSwitches" />
          </div>
          <div class="setting-item">
            <span>{{ $t('riskSwitch.allowCrossAgentWrite') }}</span>
            <el-switch v-model="riskSwitches.allow_cross_agent_write" @change="saveRiskSwitches" />
          </div>
          <div class="setting-item">
            <span>{{ $t('riskSwitch.allowMemoryVisibilityChange') }}</span>
            <el-switch v-model="riskSwitches.allow_visibility_change" @change="saveRiskSwitches" />
          </div>
        </div>
        <div class="risk-category">
          <div class="risk-category-title">{{ $t('riskSwitch.categorySharing') }}</div>
          <div class="setting-item">
            <span>{{ $t('riskSwitch.allowAutoSharing') }}</span>
            <el-switch v-model="riskSwitches.allow_auto_sharing" @change="saveRiskSwitches" />
          </div>
          <div class="setting-item">
            <span>{{ $t('riskSwitch.allowManualSharing') }}</span>
            <el-switch v-model="riskSwitches.allow_manual_sharing" @change="saveRiskSwitches" />
          </div>
          <div class="setting-item">
            <span>{{ $t('riskSwitch.allowBatchShare') }}</span>
            <el-switch v-model="riskSwitches.allow_batch_share" @change="saveRiskSwitches" />
          </div>
        </div>
        <div class="risk-category">
          <div class="risk-category-title">{{ $t('riskSwitch.categoryWriteback') }}</div>
          <div class="setting-item">
            <span>{{ $t('riskSwitch.allowWriteback') }}</span>
            <el-switch v-model="riskSwitches.allow_writeback" @change="saveRiskSwitches" />
          </div>
        </div>
      </div>

      <!-- Dialectic Reasoning -->
      <div class="settings-card" id="settings-reasoning">
        <div class="card-title">🔮 {{ $t('settings.reasoningConfig') }}</div>
        <p class="setting-desc" style="margin-bottom: 8px">{{ $t('settings.reasoningDesc') }}</p>
        <div class="setting-item" v-if="reasoningConfig">
          <span>{{ $t('settings.reasoningEnabled') }}</span>
          <el-switch v-model="reasoningEnabled" @change="updateReasoningEnabled" />
        </div>
        <div class="setting-item" v-if="reasoningConfig">
          <span>{{ $t('settings.reasoningProvider') }}</span>
          <el-select v-model="reasoningForm.provider" style="width: 200px" @change="onReasoningProviderChange">
            <el-option label="OpenAI" value="openai" />
            <el-option label="Ollama" value="ollama" />
            <el-option label="DeepSeek" value="deepseek" />
            <el-option label="Custom" value="custom" />
          </el-select>
        </div>
        <div class="setting-item" v-if="reasoningConfig">
          <span>{{ $t('settings.reasoningModel') }}</span>
          <el-input v-model="reasoningForm.model" style="width: 200px" :placeholder="$t('settings.reasoningModelPlaceholder')" />
        </div>
        <div class="setting-item" v-if="reasoningConfig">
          <span>{{ $t('settings.reasoningApiKey') }}</span>
          <el-input v-model="reasoningForm.api_key" type="password" show-password style="width: 200px" :placeholder="reasoningHasKey ? $t('settings.reasoningApiKeySet') : $t('settings.reasoningApiKeyPlaceholder')" />
        </div>
        <div class="setting-item" v-if="reasoningConfig && reasoningForm.provider === 'custom'">
          <span>{{ $t('settings.reasoningBaseUrl') }}</span>
          <el-input v-model="reasoningForm.base_url" style="width: 200px" placeholder="http://localhost:11434/v1" />
        </div>
        <div class="setting-item" v-if="reasoningConfig">
          <span>{{ $t('settings.reasoningDepth') }}</span>
          <el-slider v-model="reasoningForm.dialectic_depth" :min="1" :max="3" :step="1" :marks="{ 1: '1', 2: '2', 3: '3' }" style="width: 200px" />
        </div>
        <div class="setting-item" v-if="reasoningConfig">
          <span>{{ $t('settings.reasoningLevel') }}</span>
          <el-select v-model="reasoningForm.reasoning_level" style="width: 200px">
            <el-option label="Minimal" value="minimal" />
            <el-option label="Low" value="low" />
            <el-option label="Medium" value="medium" />
            <el-option label="High" value="high" />
            <el-option label="Max" value="max" />
          </el-select>
        </div>
        <div class="setting-item" v-if="reasoningConfig">
          <span></span>
          <div>
            <el-button size="small" type="primary" @click="saveReasoningConfig" :loading="reasoningSaving">{{ $t('common.save') }}</el-button>
            <el-button size="small" @click="testReasoningConnection" :loading="reasoningTesting">{{ $t('settings.reasoningTest') }}</el-button>
          </div>
        </div>
        <div v-if="reasoningTestResult" style="margin-top: 8px; padding: 10px; background: var(--cm-bg-secondary); border-radius: 6px; font-size: 12px">
          <span :style="{ color: reasoningTestResult.success ? 'var(--cm-success)' : 'var(--cm-danger)' }">
            {{ reasoningTestResult.success ? '✓ ' + $t('settings.reasoningTestSuccess') : '✗ ' + (reasoningTestResult.error || $t('settings.reasoningTestFailed')) }}
          </span>
        </div>
        <div v-if="!reasoningConfig" class="setting-item">
          <span>{{ $t('settings.reasoningProvider') }}</span>
          <el-button size="small" @click="loadReasoningConfig" :loading="reasoningLoading">{{ $t('common.retry') }}</el-button>
        </div>
        <div class="reasoning-hint" v-if="reasoningConfig" style="margin-top: 8px; padding: 10px; background: var(--cm-bg-secondary); border-radius: 6px; font-size: 12px; color: var(--cm-text-muted)">
          <div style="font-weight: 600; margin-bottom: 4px">{{ $t('settings.reasoningHint') }}</div>
          <div>{{ $t('settings.reasoningHintDesc') }}</div>
        </div>
      </div>

      <!-- API 密钥 -->
      <div class="settings-card">
        <div class="card-title">🔑 {{ $t('settings.apiKeys') }}</div>
        <p class="setting-desc" style="margin-bottom: 8px">{{ $t('settings.apiKeysDesc') }}</p>
        <el-alert type="info" :closable="false" style="margin-bottom: 12px; font-size: 12px">
          <template #title>{{ $t('settings.apiKeySecurityNote') }}</template>
        </el-alert>
        <div class="api-key-list" v-if="apiKeys.length > 0">
          <div class="api-key-item" v-for="key in apiKeys" :key="key.id">
            <div class="api-key-info">
              <span class="api-key-name">{{ key.name }}</span>
              <span class="api-key-prefix">{{ key.key_prefix }}••••••••</span>
              <span class="api-key-perms">{{ formatPermissions(key.permissions) }}</span>
              <span class="api-key-time">{{ key.last_used_at ? t('settings.apiKeyLastUsed') + ': ' + key.last_used_at.substring(0, 10) : t('settings.apiKeyNeverUsed') }}</span>
            </div>
            <el-button type="danger" plain size="small" @click="deleteApiKey(key.id)">{{ t('settings.apiKeyDelete') }}</el-button>
          </div>
        </div>
        <div v-else class="setting-item">
          <span class="setting-desc">{{ $t('settings.apiKeyNoKeys') }}</span>
        </div>
        <div class="setting-item" style="margin-top: 12px">
          <span class="setting-desc">{{ t('settings.apiKeyRemaining') }}: {{ 5 - apiKeys.length }}/5</span>
          <el-button type="primary" size="small" @click="openApiKeyDialog" :disabled="apiKeys.length >= 5">{{ $t('settings.createApiKey') }}</el-button>
        </div>
        <div class="api-usage-hint" style="margin-top: 12px; padding: 10px; background: var(--cm-bg-secondary); border-radius: 6px; font-size: 12px">
          <div style="font-weight: 600; margin-bottom: 6px">{{ $t('settings.apiKeyUsage') }}</div>
          <div style="color: var(--cm-text-muted); margin-bottom: 6px">{{ $t('settings.apiKeyUsageDesc') }}</div>
          <code style="display: block; padding: 8px; background: var(--cm-bg); border-radius: 4px; font-size: 11px; word-break: break-all; color: var(--cm-accent)">
            curl -X POST http://localhost:8765/api/v1/external/memories \<br>
            &nbsp;&nbsp;-H "X-API-Key: YOUR_KEY" \<br>
            &nbsp;&nbsp;-H "Content-Type: application/json" \<br>
            &nbsp;&nbsp;-d '{"key":"topic","value":"content"}'
          </code>
        </div>
      </div>

      <!-- 敏感内容设置 -->
      <div class="settings-card">
        <div class="card-title">🛡️ {{ $t('settings.recordSensitive') }}</div>
        <p class="setting-desc" style="margin-bottom: 8px">{{ $t('settings.recordSensitiveDesc') }}</p>
        <div class="setting-item">
          <span>{{ $t('settings.recordSensitive') }}</span>
          <el-switch v-model="recordSensitive" @change="updateRecordSensitive" />
        </div>
        <el-alert v-if="recordSensitive" type="warning" :closable="false" style="margin-top: 8px; font-size: 12px">
          <template #title>{{ $t('settings.recordSensitiveWarning') }}</template>
        </el-alert>
      </div>

      <!-- 客户端连接配置 -->
      <div class="settings-card" id="settings-openclaw">
        <div class="card-title">🔗 {{ $t('settings.clientConnection') }}</div>
        <div class="setting-item">
          <span>{{ $t('settings.connectedAgents') }}</span>
          <el-button size="small" @click="loadConnectedAgents" :loading="agentScanLoading">{{ $t('settings.refresh') }}</el-button>
        </div>
        <div v-if="connectedAgents.length > 0" class="agent-list">
          <div v-for="agent in connectedAgents" :key="agent.name" class="agent-item">
            <div class="agent-item-header">
              <span class="agent-name">{{ agent.display_name || agent.name }}</span>
              <el-tag type="success" size="small">{{ $t('settings.agentConnected') }}</el-tag>
            </div>
            <div v-if="agent.found_dirs && agent.found_dirs.length > 0" class="openclaw-paths">
              <div v-for="p in agent.found_dirs" :key="p" class="openclaw-path-item">
                <code>{{ p }}</code>
              </div>
            </div>
          </div>
        </div>
        <div v-if="connectedAgents.length === 0 && !agentScanLoading" class="setting-item">
          <span class="setting-desc">{{ $t('settings.noAgentsDetected') }}</span>
        </div>
        <div class="setting-item" v-if="agentSyncStatus">
          <span>{{ $t('settings.autoRecord') }}</span>
          <el-switch v-model="agentAutoSync" @change="toggleAgentSync" :loading="agentSyncLoading" />
        </div>
        <div v-if="agentSyncStatus" class="setting-item">
          <span>{{ $t('settings.forceSync') }}</span>
          <el-button size="small" type="primary" @click="forceAgentSync" :loading="agentSyncLoading">{{ $t('settings.syncNow') }}</el-button>
        </div>
        <div v-if="agentSyncStatus && agentSyncStatus.synced_count > 0" class="setting-item">
          <span>{{ $t('settings.syncedCount') }}</span>
          <span class="setting-desc">{{ agentSyncStatus.synced_count }}</span>
        </div>
        <div v-if="agentSyncStatus && agentSyncStatus.skipped_count > 0" class="setting-item">
          <span>{{ $t('settings.skippedCount') }}</span>
          <span class="setting-desc">{{ agentSyncStatus.skipped_count }}</span>
        </div>
        <div style="margin-top: 12px; padding: 10px; background: var(--cm-bg-secondary); border-radius: 6px; font-size: 12px">
          <div style="font-weight: 600; margin-bottom: 6px">🔌 {{ $t('settings.pluginInstallTitle') }}</div>
          <div style="color: var(--cm-text-muted); margin-bottom: 8px">{{ $t('settings.pluginInstallDesc') }}</div>
          <code style="display: block; padding: 8px; background: var(--cm-bg); border-radius: 4px; font-size: 11px; word-break: break-all; color: var(--cm-accent)">
            openclaw plugins install -l ./openclaw-plugin
          </code>
          <div style="color: var(--cm-text-muted); margin-top: 8px; font-size: 11px">{{ $t('settings.pluginConfigHint') }}</div>
          <code style="display: block; padding: 8px; background: var(--cm-bg); border-radius: 4px; font-size: 11px; word-break: break-all; color: var(--cm-accent); margin-top: 4px">
            { "plugins": { "slots": { "contextEngine": "clawmemory" }, "entries": { "clawmemory": { "enabled": true, "config": { "baseUrl": "http://localhost:8765", "apiKey": "cm_..." } } } } }
          </code>
          <el-alert type="info" :closable="false" style="margin-top: 8px; font-size: 11px">
            <template #title>{{ $t('settings.apiKeyHowToGet') }}</template>
          </el-alert>
          <el-alert type="warning" :closable="false" style="margin-top: 4px; font-size: 11px">
            <template #title>{{ $t('settings.apiKeyEnvFallback') }}</template>
          </el-alert>
        </div>
        <div style="margin-top: 12px; padding: 10px; background: var(--cm-bg-secondary); border-radius: 6px; font-size: 12px">
          <div style="font-weight: 600; margin-bottom: 6px">{{ $t('settings.agentsMdTitle') }}</div>
          <div style="color: var(--cm-text-muted); margin-bottom: 8px">{{ $t('settings.agentsMdDesc') }}</div>
          <el-button size="small" type="primary" @click="copyAgentsMD" :loading="agentsMdLoading">{{ $t('settings.agentsMdCopy') }}</el-button>
          <el-button size="small" @click="showAgentsMdPreview = true">{{ $t('settings.agentsMdPreview') }}</el-button>
        </div>
      </div>

      <!-- 数据管理 -->
      <div class="settings-card" :class="{ 'section-highlight': activeSection === 'data' }" id="settings-data">
        <div class="card-title">💾 {{ $t('settings.data') }}</div>
        <div class="setting-item">
          <span>{{ $t('settings.dbLocation') }}</span>
          <span class="setting-desc">data/clawmemory.db</span>
        </div>
        <div class="setting-item">
          <span>{{ $t('settings.exportData') }}</span>
          <el-input v-model="exportPassword" type="password" :placeholder="$t('settings.enterPassword')" size="small" style="width:160px;margin-right:8px" show-password />
          <el-button size="small" type="primary" @click="exportData" :disabled="!exportPassword">{{ $t('settings.export') }}</el-button>
        </div>
        <div class="setting-item">
          <span>{{ $t('settings.importData') }}</span>
          <el-upload :show-file-list="false" :before-upload="importData" accept=".json" action="" :auto-upload="false">
            <el-button size="small" type="warning">{{ $t('settings.chooseFile') }}</el-button>
          </el-upload>
        </div>
      </div>

      <!-- 记忆衰减设置 -->
      <div class="settings-card" :class="{ 'section-highlight': activeSection === 'decay' }" id="settings-decay">
        <div class="card-title">🧠 {{ $t('settings.memoryDecay') }}</div>
        <div class="decay-info" v-if="decayInfo">
          <div class="decay-stage-info">
            <div class="stage-item">
              <span class="stage-label">{{ $t('settings.stage15d') }}</span>
              <span class="stage-desc">{{ $t('settings.stage15dDesc') }}</span>
            </div>
            <div class="stage-item">
              <span class="stage-label">{{ $t('settings.stage30d') }}</span>
              <span class="stage-desc">{{ $t('settings.stage30dDesc') }}</span>
            </div>
            <div class="stage-item">
              <span class="stage-label">{{ $t('settings.stage60d') }}</span>
              <span class="stage-desc">{{ $t('settings.stage60dDesc') }}</span>
            </div>
            <div class="stage-item">
              <span class="stage-label">{{ $t('settings.stageTrash') }}</span>
              <span class="stage-desc">{{ $t('settings.stageTrashDesc') }}</span>
            </div>
          </div>
        </div>
        <div class="setting-item">
          <span>{{ $t('settings.autoDecay') }}</span>
          <el-switch v-model="decayEnabled" @change="updateDecaySettings" :loading="decayLoading" />
        </div>
        <div class="decay-stats" v-if="decayStats">
          <div class="stats-row">
            <span class="stats-label">{{ $t('settings.totalMemories') }}</span>
            <span class="stats-value">{{ decayStats.total }}</span>
          </div>
          <div class="stats-row">
            <span class="stats-label">{{ $t('settings.activeMemories') }}</span>
            <span class="stats-value">{{ decayStats.active }}</span>
          </div>
          <div class="stats-row">
            <span class="stats-label">{{ $t('settings.archivedMemories') }}</span>
            <span class="stats-value warning">{{ decayStats.archived }}</span>
          </div>
          <div class="stats-row">
            <span class="stats-label">{{ $t('settings.trashedMemories') }}</span>
            <span class="stats-value danger">{{ decayStats.trashed }}</span>
          </div>
        </div>
        <div class="decay-actions" v-if="decayStats && decayStats.trashed > 0">
          <el-button size="small" type="warning" @click="viewTrash">{{ $t('settings.viewTrash') }}</el-button>
          <el-button size="small" type="danger" @click="emptyTrash">{{ $t('settings.emptyTrash') }}</el-button>
        </div>
      </div>

      <!-- 记忆健康度 -->
      <div class="settings-card">
        <div class="card-title">💊 {{ $t('settings.memoryHealth') }}</div>
        <div v-if="healthScore" class="health-display">
          <div class="health-score" :class="healthGrade">{{ healthScore.overall_score }}</div>
          <div class="health-grade">{{ healthScore.grade }}</div>
          <div class="health-dimensions">
            <div class="dim-item" v-for="(dim, key) in healthScore.dimensions" :key="key">
              <span class="dim-label">{{ dim.label }}</span>
              <div class="dim-bar"><div class="dim-fill" :style="{ width: dim.score + '%' }"></div></div>
              <span class="dim-value">{{ dim.score }}</span>
            </div>
          </div>
          <div class="health-suggestions" v-if="healthScore.suggestions?.length">
            <div class="suggestion" v-for="(s, i) in healthScore.suggestions" :key="i">{{ s }}</div>
          </div>
        </div>
        <div v-else class="setting-item">
          <span>{{ $t('settings.clickToCheck') }}</span>
          <el-button size="small" type="primary" @click="checkHealth" :loading="healthLoading">{{ $t('settings.checkHealth') }}</el-button>
        </div>
      </div>

      <!-- 记忆去重 -->
      <div class="settings-card">
        <div class="card-title">🔍 {{ $t('settings.memoryDedup') }}</div>
        <div v-if="dedupResult" class="dedup-display">
          <div class="dedup-stats">
            <span>{{ $t('settings.totalMemories') }}: {{ dedupResult.total_memories }}</span>
            <span>{{ $t('settings.duplicates') }}: {{ dedupResult.total_duplicates }}</span>
            <span>{{ $t('settings.dedupRate') }}: {{ Math.round((dedupResult.dedup_rate || 0) * 100) }}%</span>
          </div>
          <div class="dedup-groups" v-if="dedupResult.duplicate_groups?.length">
            <div class="dedup-group" v-for="(g, i) in dedupResult.duplicate_groups" :key="i">
              <div class="dedup-group-header">{{ g.key }} ({{ $t('settings.similarity') }}: {{ Math.round(g.similarity * 100) }}%)</div>
              <div class="dedup-item" v-for="m in g.memories" :key="m.id">
                <span class="dedup-id">#{{ m.id }}</span>
                <span class="dedup-value">{{ m.value }}</span>
              </div>
            </div>
          </div>
        </div>
        <div class="setting-item">
          <span>{{ $t('settings.scanDuplicates') }}</span>
          <el-button size="small" type="primary" @click="scanDedup" :loading="dedupLoading">{{ $t('settings.scan') }}</el-button>
        </div>
      </div>

      <!-- 系统信息 -->
      <div class="settings-card" :class="{ 'section-highlight': activeSection === 'system' }" id="settings-system">
        <div class="card-title">◇ {{ $t('settings.system') }}</div>
        <div class="setting-item">
          <span>{{ $t('settings.version') }}</span>
          <span class="setting-desc">
            v{{ appVersion }}
            <el-tag v-if="updateInfo.has_update" type="success" size="small" style="margin-left: 8px">
              🆕 v{{ updateInfo.latest_version }}
            </el-tag>
            <el-tag v-else-if="updateInfo.checked" type="info" size="small" style="margin-left: 8px">
              ✓ {{ $t('settings.upToDate') }}
            </el-tag>
          </span>
        </div>
        <div class="setting-item" v-if="updateInfo.has_update">
          <span>{{ $t('settings.newVersion') }}</span>
          <span class="setting-desc">
            v{{ updateInfo.latest_version }}
            <a :href="updateInfo.download_url" target="_blank" class="update-link">{{ $t('settings.downloadUpdate') }}</a>
          </span>
        </div>
        <div class="setting-item" v-if="updateInfo.release_notes">
          <span>{{ $t('settings.releaseNotes') }}</span>
          <span class="setting-desc release-notes">{{ updateInfo.release_notes }}</span>
        </div>
        <div class="setting-item">
          <span>{{ $t('settings.coreEngine') }}</span>
          <span class="setting-desc">{{ coreEngine }}</span>
        </div>
        <div class="setting-item">
          <span>{{ $t('settings.checkUpdate') }}</span>
          <el-button size="small" @click="checkForUpdate" :loading="updateChecking">{{ $t('settings.checkNow') }}</el-button>
        </div>
        <div class="setting-item">
          <span>{{ $t('settings.resetPasswordTip') }}</span>
          <span class="setting-desc code-hint">{{ cliResetCommand }}</span>
        </div>
      </div>

      <!-- 邀请码管理 (仅创始账号可见) -->
      <div class="settings-card" :class="{ 'section-highlight': activeSection === 'invitations' }" id="settings-invitations" v-if="isFounder">
        <div class="card-title">🎫 {{ $t('settings.invitationManage') }}</div>
        <p class="section-desc">{{ $t('settings.invitationManageDesc') }}</p>
        <div class="setting-item">
          <span>{{ $t('settings.createInvitation') }}</span>
          <div class="invitation-create-form">
            <el-input-number v-model="newInvitationMaxUses" :min="1" :max="100" size="small" style="width: 120px" />
            <span class="form-label">{{ $t('settings.invitationMaxUses') }}</span>
            <el-input-number v-model="newInvitationExpiresHours" :min="0" size="small" style="width: 120px" />
            <span class="form-label">{{ $t('settings.invitationExpires') }}</span>
            <span class="form-label" v-if="newInvitationExpiresHours === 0">{{ $t('settings.invitationNeverExpires') }}</span>
            <el-button type="primary" size="small" @click="handleCreateInvitation" :loading="creatingInvitation">{{ $t('settings.createInvitation') }}</el-button>
          </div>
        </div>
        <div class="invitation-list" v-if="invitations.length > 0">
          <div class="invitation-item" v-for="inv in invitations" :key="inv.id">
            <div class="invitation-code-row">
              <code class="invitation-code">{{ inv.code }}</code>
              <el-button size="small" text @click="copyInvitationCode(inv.code)">{{ $t('settings.invitationCopy') }}</el-button>
              <el-tag size="small" :type="getInvitationTagType(inv)">{{ getInvitationStatus(inv) }}</el-tag>
              <el-button size="small" text type="danger" @click="handleDeleteInvitation(inv.id)">{{ $t('settings.invitationDelete') }}</el-button>
            </div>
            <div class="invitation-meta">
              <span>{{ inv.used_count }}/{{ inv.max_uses }} {{ $t('settings.invitationUsed') }}</span>
              <span v-if="inv.expires_at">· {{ $t('settings.invitationExpires') }}: {{ formatDate(inv.expires_at) }}</span>
              <span v-else>· {{ $t('settings.invitationNeverExpires') }}</span>
            </div>
          </div>
        </div>
        <div class="empty-hint" v-else>{{ $t('settings.invitationNoCodes') }}</div>
      </div>

      <!-- 用户管理 (仅创始账号可见) -->
      <div class="settings-card" :class="{ 'section-highlight': activeSection === 'users' }" id="settings-users" v-if="isFounder">
        <div class="card-title">👥 {{ $t('settings.userManage') }}</div>
        <p class="section-desc">{{ $t('settings.userManageDesc') }}</p>
        <div class="user-list" v-if="users.length > 0">
          <div class="user-item" v-for="u in users" :key="u.id">
            <div class="user-info">
              <span class="user-name">{{ u.username }}</span>
              <el-tag size="small" :type="u.is_founder ? 'danger' : u.role === 'admin' ? 'warning' : 'info'">
                {{ u.is_founder ? $t('settings.userFounder') : u.role === 'admin' ? $t('settings.userAdmin') : $t('settings.userNormal') }}
              </el-tag>
              <span class="user-time">{{ $t('settings.userCreatedAt') }}: {{ formatDate(u.created_at) }}</span>
            </div>
            <div class="user-actions" v-if="!u.is_founder">
              <el-button size="small" type="warning" @click="openResetUserPasswordDialog(u)">{{ $t('settings.userResetPassword') }}</el-button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <el-dialog v-model="showPasswordDialog" :title="passwordSet ? $t('settings.changePassword') : $t('settings.setPassword')" width="400px" :fullscreen="isMobile">
      <el-form label-position="top">
        <el-form-item v-if="passwordSet" :label="$t('settings.oldPassword')">
          <el-input v-model="oldPassword" type="password" show-password :placeholder="$t('settings.oldPasswordPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('settings.newPassword')">
          <el-input v-model="newPassword" type="password" show-password :placeholder="$t('settings.passwordMinLen')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showPasswordDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleSetPassword" :loading="settingPassword">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showResetUserPasswordDialog" :title="$t('settings.userResetPassword')" width="400px" :fullscreen="isMobile">
      <p style="margin-bottom: 12px; color: var(--cm-text-muted)">{{ t('settings.userResetPasswordConfirm', [resetTargetUser?.username]) }}</p>
      <el-form label-position="top">
        <el-form-item :label="$t('settings.newPassword')">
          <el-input v-model="resetUserNewPassword" type="password" show-password :placeholder="$t('settings.passwordMinLen')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showResetUserPasswordDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="handleResetUserPassword" :loading="resettingUserPassword">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showApiKeyDialog" :title="$t('settings.createApiKey')" width="460px" :fullscreen="isMobile" :close-on-click-modal="false">
      <div v-if="!newApiKeyRaw">
        <el-form label-position="top">
          <el-form-item :label="$t('settings.apiKeyName')">
            <el-input v-model="newApiKeyName" :placeholder="$t('settings.apiKeyNamePlaceholder')" />
          </el-form-item>
          <el-form-item :label="$t('settings.apiKeyPermission')">
            <el-select v-model="newApiKeyPerm" style="width: 100%">
              <el-option :label="$t('settings.apiKeyPermFull')" value="full" />
              <el-option :label="$t('settings.apiKeyPermReadWrite')" value="readwrite" />
              <el-option :label="$t('settings.apiKeyPermReadOnly')" value="readonly" />
            </el-select>
            <div style="margin-top: 6px; font-size: 11px; color: var(--cm-text-muted)">
              <template v-if="newApiKeyPerm === 'full'">{{ $t('settings.apiKeyPermFullDesc') }}</template>
              <template v-else-if="newApiKeyPerm === 'readwrite'">{{ $t('settings.apiKeyPermReadWriteDesc') }}</template>
              <template v-else>{{ $t('settings.apiKeyPermReadOnlyDesc') }}</template>
            </div>
          </el-form-item>
        </el-form>
      </div>
      <div v-else>
        <el-alert type="warning" :closable="false" style="margin-bottom: 12px">
          <template #title>{{ $t('settings.apiKeyCreatedWarning') }}</template>
        </el-alert>
        <div style="padding: 10px; background: var(--cm-bg-secondary); border-radius: 6px; font-family: monospace; font-size: 13px; word-break: break-all; color: var(--cm-accent)">
          {{ newApiKeyRaw }}
        </div>
        <el-button type="primary" size="small" style="margin-top: 8px" @click="copyApiKey">{{ $t('settings.apiKeyCopy') }}</el-button>
      </div>
      <template #footer>
        <el-button @click="showApiKeyDialog = false; newApiKeyRaw = ''">{{ $t('common.cancel') }}</el-button>
        <el-button v-if="!newApiKeyRaw" type="primary" @click="createApiKey" :loading="creatingApiKey">{{ $t('settings.createApiKey') }}</el-button>
        <el-button v-else type="primary" @click="showApiKeyDialog = false; newApiKeyRaw = ''">{{ $t('common.confirm') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showAIConfigDialog" :title="$t('settings.aiConfigure')" width="520px" :fullscreen="isMobile" :close-on-click-modal="false">
      <el-form label-position="top">
        <el-form-item :label="$t('settings.aiProvider')">
          <el-select v-model="aiForm.provider_id" @change="onAIProviderChange" style="width: 100%">
            <el-option v-for="p in aiProviders" :key="p.ID" :label="p.Name" :value="p.ID">
              <span>{{ p.Name }}</span>
              <el-tag v-if="p.Free" type="success" size="small" style="margin-left: 8px">{{ $t('settings.aiFree') }}</el-tag>
              <el-tag v-if="p.ProOnly" type="warning" size="small" style="margin-left: 4px">Pro</el-tag>
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('settings.aiModel')">
          <el-select v-model="aiForm.model" style="width: 100%">
            <el-option v-for="m in currentProviderModels" :key="m" :label="m" :value="m" />
          </el-select>
        </el-form-item>
        <el-form-item :label="$t('settings.aiApiKey')">
          <el-input v-model="aiForm.api_key" type="password" show-password :placeholder="$t('settings.aiApiKeyPlaceholder')" />
        </el-form-item>
        <el-form-item :label="$t('settings.aiBaseUrl')" v-if="aiForm.provider_id === 'custom'">
          <el-input v-model="aiForm.base_url" :placeholder="$t('settings.aiBaseUrlPlaceholder')" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAIConfigDialog = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="saveAIConfig" :loading="aiSaving">{{ $t('common.save') }}</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="showAgentsMdPreview" :title="$t('settings.agentsMdTitle')" width="640px" :fullscreen="isMobile">
      <div style="max-height: 500px; overflow-y: auto; padding: 12px; background: var(--cm-bg-secondary); border-radius: 6px; font-family: monospace; font-size: 12px; white-space: pre-wrap; word-break: break-all; line-height: 1.6">
{{ agentsMdContent }}
      </div>
      <template #footer>
        <el-button @click="showAgentsMdPreview = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="copyAgentsMD">{{ $t('settings.agentsMdCopy') }}</el-button>
      </template>
    </el-dialog>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from '../api/go-client'
import { setLocale, getLocale, translateError } from '../i18n'
import { aiApi } from '../api/go-ai'
import { reasoningApi } from '../api/go-reasoning'
import { useAuthStore } from '../stores/auth'

const { t } = useI18n()
const route = useRoute()
const authStore = useAuthStore()
const isMobile = ref(window.innerWidth <= 768)
const activeSection = ref((route.query.section as string) || '')
const sections = [
  { id: 'license', icon: '🔑', label: computed(() => t('settings.license')) },
  { id: 'ai', icon: '🧠', label: computed(() => t('settings.aiConfig')) },
  { id: 'security', icon: '🔒', label: computed(() => t('settings.security')) },
  { id: 'risk-switches', icon: '🛡️', label: computed(() => t('settings.riskControl')) },
  { id: 'reasoning', icon: '🤔', label: computed(() => t('settings.reasoningConfig')) },
  { id: 'openclaw', icon: '🌐', label: 'OpenClaw' },
  { id: 'data', icon: '💾', label: computed(() => t('settings.dataManagement')) },
  { id: 'decay', icon: '⏳', label: computed(() => t('settings.memoryDecay')) },
  { id: 'system', icon: '🖥️', label: computed(() => t('settings.systemInfo')) },
]
const license = ref<any>({ active: false, tier: 'oss', features: [] })
const activating = ref(false)
const licenseKey = ref('')
const exporting = ref(false)
const exportPassword = ref('')
const passwordSet = ref(false)
const showPasswordDialog = ref(false)
const oldPassword = ref('')
const newPassword = ref('')
const settingPassword = ref(false)
const coreEngine = ref('python')
const currentLocale = ref(getLocale())
const appVersion = ref('2.14.0')
const updateInfo = ref<any>({ checked: false, has_update: false, latest_version: '', download_url: '', release_notes: '' })
const updateChecking = ref(false)
const cliResetCommand = ref(navigator.platform.toLowerCase().includes('win') ? 'clawmemory.exe --reset-password NEW_PASSWORD' : './clawmemory --reset-password NEW_PASSWORD')

const apiKeys = ref<any[]>([])
const showApiKeyDialog = ref(false)
const newApiKeyName = ref('')
const newApiKeyRaw = ref('')
const newApiKeyPerm = ref('full')
const creatingApiKey = ref(false)

const decayEnabled = ref(false)
const decayLoading = ref(false)
const decayStats = ref<any>(null)
const decayInfo = ref<any>(null)
const recordSensitive = ref(false)

const healthScore = ref<any>(null)
const healthLoading = ref(false)
const healthGrade = computed(() => {
  if (!healthScore.value) return ''
  const s = healthScore.value.overall_score
  if (s >= 80) return 'grade-a'
  if (s >= 60) return 'grade-b'
  return 'grade-c'
})

const dedupResult = ref<any>(null)

const isFounder = computed(() => authStore.isFounder)
const invitations = ref<any[]>([])
const newInvitationMaxUses = ref(1)
const newInvitationExpiresHours = ref(0)
const creatingInvitation = ref(false)
const users = ref<any[]>([])
const showResetUserPasswordDialog = ref(false)
const resetTargetUser = ref<any>(null)
const resetUserNewPassword = ref('')
const resettingUserPassword = ref(false)
const dedupLoading = ref(false)

const aiConfig = ref<any>(null)
const aiLoading = ref(false)
const aiTesting = ref(false)
const aiTestResult = ref<any>(null)
const aiUsage = ref<any>(null)
const showAIConfigDialog = ref(false)
const aiSaving = ref(false)
const aiProviders = ref<any[]>([])
const aiForm = ref<any>({
  provider_id: '',
  model: '',
  api_key: '',
  base_url: '',
})

const currentProviderModels = computed(() => {
  const p = aiProviders.value.find((x: any) => x.ID === aiForm.value.provider_id)
  return p?.Models || []
})

const showAgentsMdPreview = ref(false)
const agentsMdContent = ref('')
const agentsMdLoading = ref(false)

const reasoningConfig = ref<any>(null)
const reasoningLoading = ref(false)
const reasoningSaving = ref(false)
const reasoningTesting = ref(false)
const reasoningTestResult = ref<any>(null)
const reasoningEnabled = ref(false)
const reasoningHasKey = ref(false)
const reasoningForm = ref<any>({
  provider: 'openai',
  model: '',
  api_key: '',
  base_url: '',
  dialectic_depth: 1,
  reasoning_level: 'medium',
  max_tokens: 1000,
})

const featureLabels: Record<string, string> = {
  ai_extract: t('settings.featAiExtract'),
  auto_graph: t('settings.featAutoGraph'),
  unlimited_graph: t('settings.featUnlimitedGraph'),
  auto_decay: t('settings.featAutoDecay'),
  decay_report: t('settings.featDecayReport'),
  prune_suggest: t('settings.featPruneSuggest'),
  reinforce: t('settings.featReinforce'),
  conflict_scan: t('settings.featConflictScan'),
  conflict_merge: t('settings.featConflictMerge'),
  smart_router: t('settings.featSmartRouter'),
  token_stats: t('settings.featTokenStats'),
  wiki: t('settings.featWiki'),
  auto_backup: t('settings.featAutoBackup'),
  compress: t('settings.featCompress'),
  evolution: t('settings.featEvolution'),
  api_access: t('settings.featApiAccess'),
  sso: t('settings.featSso'),
  audit_log: t('settings.featAuditLog'),
  time_travel: t('settings.featTimeTravel'),
  offline_mode: t('settings.featOfflineMode'),
}

const isExpiringSoon = computed(() => {
  if (!license.value.expires_at) return false
  try {
    const exp = new Date(license.value.expires_at)
    const now = new Date()
    const daysLeft = (exp.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
    return daysLeft < 30 && daysLeft > 0
  } catch { return false }
})

function formatDate(dateStr: string): string {
  if (!dateStr) return ''
  try {
    const d = new Date(dateStr)
    if (isNaN(d.getTime())) return dateStr
    return d.toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
  } catch { return dateStr }
}

onMounted(async () => {
  await Promise.all([loadLicense(), loadInitStatus(), loadInstallStatus(), loadDecaySettings(), loadDecayStats(), loadApiKeys(), loadRecordSensitiveSetting(), loadConnectedAgents(), loadAgentSyncStatus(), loadRiskSwitches(), loadAIConfig(), loadAIUsage(), loadReasoningConfig()])
  if (authStore.isFounder) {
    loadInvitations()
    loadUsers()
  }
  if (activeSection.value) {
    nextTick(() => scrollToSection(activeSection.value))
  }
})

watch(() => route.query.section, (section) => {
  if (section && typeof section === 'string') {
    activeSection.value = section
    nextTick(() => scrollToSection(section))
  }
})

function scrollToSection(section: string) {
  const el = document.getElementById(`settings-${section}`)
  if (el) {
    activeSection.value = section
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    el.classList.add('section-highlight')
    setTimeout(() => el.classList.remove('section-highlight'), 2000)
  }
}

onMounted(() => {
  const observer = new IntersectionObserver((entries) => {
    for (const entry of entries) {
      if (entry.isIntersecting) {
        const id = entry.target.id?.replace('settings-', '')
        if (id) activeSection.value = id
      }
    }
  }, { threshold: 0.3 })
  document.querySelectorAll('.settings-card[id]').forEach((el) => observer.observe(el))
  onUnmounted(() => observer.disconnect())
})

function changeLocale(locale: 'zh' | 'en') {
  setLocale(locale)
}

async function loadLicense() {
  try { const { data } = await axios.get('/license/info'); license.value = data } catch { license.value = null }
}

async function loadApiKeys() {
  try {
    const { data } = await axios.get('/api-keys')
    apiKeys.value = data.items || []
  } catch { apiKeys.value = [] }
}

function openApiKeyDialog() {
  if (apiKeys.value.length >= 5) {
    ElMessage.warning(t('settings.apiKeyMaxReached'))
    return
  }
  newApiKeyName.value = ''
  newApiKeyRaw.value = ''
  newApiKeyPerm.value = 'full'
  showApiKeyDialog.value = true
}

async function createApiKey() {
  if (!newApiKeyName.value.trim()) {
    ElMessage.warning(t('settings.apiKeyName'))
    return
  }
  creatingApiKey.value = true
  try {
    const permMap: Record<string, string> = {
      full: 'memories:read,memories:write,conversations:write,sessions:write,reason:execute',
      readwrite: 'memories:read,memories:write',
      readonly: 'memories:read',
    }
    const payload: any = { name: newApiKeyName.value.trim() }
    const perm = permMap[newApiKeyPerm.value]
    if (perm) payload.permissions = perm
    const { data } = await axios.post('/api-keys', payload)
    newApiKeyRaw.value = data.key
    ElMessage.success(t('settings.apiKeyCreated'))
    await loadApiKeys()
  } catch (e: any) {
    const errMsg = e.response?.data?.error || ''
    if (e.response?.status === 429) {
      ElMessage.error(t('settings.apiKeyRateLimit'))
    } else if (errMsg.includes('maximum')) {
      ElMessage.error(t('settings.apiKeyMaxReached'))
    } else {
      ElMessage.error(translateError(errMsg, t('common.failed')))
    }
  } finally {
    creatingApiKey.value = false
  }
}

function formatPermissions(perms: string) {
  if (!perms) return ''
  const permLabels: Record<string, string> = {
    'memories:read': '📖',
    'memories:write': '✏️',
    'conversations:write': '💬',
    'sessions:write': '📋',
    'reason:execute': '🧠',
    'read': '📖',
    'write': '✏️',
    'admin': '👑',
  }
  return perms.split(',').map((p: string) => permLabels[p.trim()] || p.trim()).join(' ')
}

async function deleteApiKey(id: number) {
  try {
    await ElMessageBox.confirm(t('settings.apiKeyDeleteConfirm'), t('common.confirm'), { type: 'warning' })
  } catch { return }
  try {
    await axios.delete(`/api-keys/${id}`)
    ElMessage.success(t('common.success'))
    await loadApiKeys()
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  }
}

function copyApiKey() {
  navigator.clipboard.writeText(newApiKeyRaw.value)
  ElMessage.success(t('settings.apiKeyCopied'))
}

const connectedAgents = ref<any[]>([])
const agentScanLoading = ref(false)
const agentSyncStatus = ref<any>(null)
const agentAutoSync = ref(true)
const agentSyncLoading = ref(false)

async function loadConnectedAgents() {
  agentScanLoading.value = true
  try {
    const { data } = await axios.get('/agents/connected')
    connectedAgents.value = data.agents || []
  } catch {
    connectedAgents.value = []
  } finally {
    agentScanLoading.value = false
  }
}

async function loadAgentSyncStatus() {
  agentSyncLoading.value = true
  try {
    const { data } = await axios.get('/agent-sync/status')
    agentSyncStatus.value = data
    agentAutoSync.value = data.auto_sync_enabled
  } catch {
    agentSyncStatus.value = null
  } finally {
    agentSyncLoading.value = false
  }
}

async function toggleAgentSync() {
  agentSyncLoading.value = true
  try {
    await axios.post('/agent-sync/toggle', { enabled: agentAutoSync.value })
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    agentAutoSync.value = !agentAutoSync.value
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally {
    agentSyncLoading.value = false
  }
}

async function forceAgentSync() {
  agentSyncLoading.value = true
  try {
    const { data } = await axios.post('/agent-sync/force')
    ElMessage.success(t('settings.syncCompleted', { count: data.synced_count || 0 }))
    await loadAgentSyncStatus()
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally {
    agentSyncLoading.value = false
  }
}

async function loadRecordSensitiveSetting() {
  try {
    const { data } = await axios.get('/settings')
    recordSensitive.value = !!data.record_sensitive_content
  } catch { recordSensitive.value = false }
}

const riskSwitches = ref<Record<string, boolean>>({
  allow_cross_agent_read: false,
  allow_cross_agent_write: false,
  allow_visibility_change: true,
  allow_auto_sharing: false,
  allow_manual_sharing: true,
  allow_batch_share: false,
  allow_writeback: false,
})

async function loadRiskSwitches() {
  try {
    const { data } = await axios.get('/risk-switches')
    if (data.switches) {
      riskSwitches.value = { ...riskSwitches.value, ...data.switches }
    }
  } catch { /* use defaults */ }
}

async function saveRiskSwitches() {
  try {
    await ElMessageBox.confirm(
      t('riskSwitch.confirmDesc'),
      t('riskSwitch.confirmTitle'),
      { type: 'warning' }
    )
  } catch {
    await loadRiskSwitches()
    return
  }
  try {
    await axios.put('/risk-switches', { switches: riskSwitches.value })
    ElMessage.success(t('riskSwitch.updated'))
  } catch (e: any) {
    await loadRiskSwitches()
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  }
}

async function updateRecordSensitive() {
  try {
    await axios.put('/settings', { record_sensitive_content: recordSensitive.value })
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    recordSensitive.value = !recordSensitive.value
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  }
}

async function loadInitStatus() {
  try { const { data } = await axios.get('/auth/init-status'); passwordSet.value = data.password_set } catch { passwordSet.value = true }
}

async function loadInvitations() {
  try {
    const { data } = await axios.get('/invitations')
    invitations.value = data.items || []
  } catch { invitations.value = [] }
}

async function loadUsers() {
  try {
    const { data } = await axios.get('/users')
    users.value = data.items || []
  } catch { users.value = [] }
}

async function handleCreateInvitation() {
  creatingInvitation.value = true
  try {
    await axios.post('/invitations', {
      max_uses: newInvitationMaxUses.value,
      expires_in_hours: newInvitationExpiresHours.value || 0,
    })
    ElMessage.success(t('settings.createInvitation') + ' ✓')
    await loadInvitations()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.failed'))
  } finally {
    creatingInvitation.value = false
  }
}

async function handleDeleteInvitation(id: number) {
  try {
    await axios.delete(`/invitations/${id}`)
    await loadInvitations()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.failed'))
  }
}

function copyInvitationCode(code: string) {
  navigator.clipboard.writeText(code)
  ElMessage.success(t('settings.invitationCopied'))
}

function getInvitationStatus(inv: any) {
  if (inv.used_count >= inv.max_uses) return t('settings.invitationUsed')
  if (inv.expires_at && new Date(inv.expires_at) < new Date()) return t('settings.invitationExpired')
  return t('settings.invitationUnused')
}

function getInvitationTagType(inv: any) {
  if (inv.used_count >= inv.max_uses) return 'info'
  if (inv.expires_at && new Date(inv.expires_at) < new Date()) return 'danger'
  return 'success'
}

function openResetUserPasswordDialog(user: any) {
  resetTargetUser.value = user
  resetUserNewPassword.value = ''
  showResetUserPasswordDialog.value = true
}

async function handleResetUserPassword() {
  if (!resetTargetUser.value || resetUserNewPassword.value.length < 6) {
    ElMessage.warning(t('settings.passwordMinLen'))
    return
  }
  resettingUserPassword.value = true
  try {
    await axios.post('/users/reset-password', {
      user_id: resetTargetUser.value.id,
      new_password: resetUserNewPassword.value,
    })
    ElMessage.success(t('settings.userResetPasswordSuccess', [resetTargetUser.value.username]))
    showResetUserPasswordDialog.value = false
    resetTargetUser.value = null
    resetUserNewPassword.value = ''
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('common.failed'))
  } finally {
    resettingUserPassword.value = false
  }
}

async function loadInstallStatus() {
  try { const { data } = await axios.get('/install-status'); coreEngine.value = data.checks?.security_engine || 'python'; if (data.version) appVersion.value = data.version } catch { coreEngine.value = 'go' }
}

async function checkForUpdate() {
  updateChecking.value = true
  try {
    const { data } = await axios.get('/check-update')
    updateInfo.value = {
      checked: true,
      has_update: data.has_update || false,
      latest_version: data.latest_version || '',
      download_url: data.download_url || '',
      release_notes: data.release_notes || '',
    }
  } catch {
    updateInfo.value.checked = true
  } finally {
    updateChecking.value = false
  }
}

async function exportData() {
  exporting.value = true
  try {
    const { data } = await axios.post('/data/export', { password: exportPassword.value })
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url; a.download = `clawmemory-export-${new Date().toISOString().split('T')[0]}.json`; a.click()
    URL.revokeObjectURL(url)
    ElMessage.success(t('settings.exportSuccess'))
  } catch { ElMessage.error(t('settings.exportFailed')) }
  finally { exporting.value = false }
}

async function importData(file: File) {
  const reader = new FileReader()
  reader.onload = async (e) => {
    try {
      const jsonData = JSON.parse(e.target?.result as string)
      await axios.post('/data/import', jsonData)
      ElMessage.success(t('settings.importSuccess'))
    } catch (e: any) {
      ElMessage.error(translateError(e.response?.data?.error, t('settings.importFailed')))
    }
  }
  reader.readAsText(file)
  return false
}

async function activateLicense() {
  if (!licenseKey.value) return
  if (licenseKey.value.trim().length < 8) {
    ElMessage.warning(t('settings.licenseKeyTooShort'))
    return
  }
  activating.value = true
  try {
    const { data } = await axios.post('/license/activate', { license_key: licenseKey.value.trim() })
    if (data.valid || data.active) {
      ElMessage.success(t('settings.activated'))
      licenseKey.value = ''
      await loadLicense()
    } else {
      ElMessage.error(data.message || t('common.failed'))
    }
  } catch (e: any) {
    const detail = e.response?.data?.error || e.response?.data?.detail
    if (typeof detail === 'string') {
      if (detail.includes('connect license server') || detail.includes('connection refused')) {
        ElMessage.error(t('settings.licenseServerUnreachable'))
      } else if (detail.includes('already bound') || detail.includes('设备已绑定')) {
        ElMessage.error(t('settings.deviceAlreadyBound'))
      } else if (detail.includes('expired') || detail.includes('已过期')) {
        ElMessage.error(t('settings.licenseExpired'))
      } else if (detail.includes('invalid') || detail.includes('无效')) {
        ElMessage.error(t('settings.licenseInvalid'))
      } else if (detail.includes('rate limit') || detail.includes('too many')) {
        ElMessage.error(t('settings.licenseRateLimited'))
      } else {
        ElMessage.error(translateError(detail, t('common.failed')))
      }
    } else {
      ElMessage.error(t('common.failed'))
    }
  } finally { activating.value = false }
}

async function deactivateLicense() {
  try {
    await ElMessageBox.confirm(t('settings.cancelConfirm'), t('common.confirm'), { type: 'warning' })
  } catch { return }
  try {
    await axios.post('/license/deactivate')
    ElMessage.success(t('settings.canceled')); await loadLicense()
  } catch (e: any) {
    const detail = e.response?.data?.error || ''
    if (detail.includes('connect') || detail.includes('server')) {
      ElMessage.warning(t('settings.deactivateServerWarning'))
      await loadLicense()
    } else {
      ElMessage.error(translateError(detail, t('common.failed')))
    }
  }
}

async function handleSetPassword() {
  if (newPassword.value.length < 6) { ElMessage.warning(t('settings.passwordMinLen')); return }
  settingPassword.value = true
  try {
    if (passwordSet.value) {
      // 修改密码 — 验证旧密码
      await axios.post('/auth/change-password', { old_password: oldPassword.value, new_password: newPassword.value })
    } else {
      // 首次设置密码
      await axios.post('/auth/set-password', { password: newPassword.value })
    }
    ElMessage.success(t('settings.passwordSet')); showPasswordDialog.value = false; oldPassword.value = ''; newPassword.value = ''; passwordSet.value = true
  } catch (e: any) { ElMessage.error(translateError(e.response?.data?.error || e.response?.data?.detail, t('common.failed'))) }
  finally { settingPassword.value = false }
}

async function loadDecaySettings() {
  try {
    const { data } = await axios.get('/memories/decay/settings')
    decayEnabled.value = data.enabled
    decayInfo.value = data
  } catch { decayEnabled.value = false }
}

async function loadDecayStats() {
  try {
    const { data } = await axios.get('/memories/decay/stats')
    decayStats.value = data.stats || data
  } catch { decayStats.value = null }
}

async function updateDecaySettings() {
  decayLoading.value = true
  try {
    await axios.put('/memories/decay/settings', { enabled: decayEnabled.value })
    ElMessage.success(decayEnabled.value ? t('settings.decayEnabled') : t('settings.decayDisabled'))
  } catch {
    ElMessage.error(t('common.failed'))
  } finally {
    decayLoading.value = false
  }
}

async function viewTrash() {
  window.location.href = '/memories?status=trashed'
}

async function emptyTrash() {
  try {
    await ElMessageBox.confirm(t('settings.emptyTrashConfirm'), t('settings.confirm'), { type: 'warning' })
    await axios.delete('/memories/trash')
    ElMessage.success(t('settings.trashEmptied'))
    await loadDecayStats()
  } catch (e: any) {
    if (e !== 'cancel' && e?.message !== 'cancel') {
      ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
    }
  }
}

async function checkHealth() {
  healthLoading.value = true
  try {
    const { data } = await axios.get('/memories/health')
    healthScore.value = data
  } catch {
    ElMessage.error(t('common.failed'))
  } finally {
    healthLoading.value = false
  }
}

async function scanDedup() {
  dedupLoading.value = true
  try {
    const { data } = await axios.get('/memories/dedup/scan')
    dedupResult.value = data
    if (data.total_duplicates > 0) {
      ElMessage.warning(t('settings.foundDuplicates', { count: data.total_duplicates }))
    } else {
      ElMessage.success(t('settings.noDuplicates'))
    }
  } catch {
    ElMessage.error(t('common.failed'))
  } finally {
    dedupLoading.value = false
  }
}

async function loadAIConfig() {
  aiLoading.value = true
  try {
    const { data } = await aiApi.getConfig()
    aiConfig.value = data
  } catch {
    aiConfig.value = null
  } finally {
    aiLoading.value = false
  }
}

async function loadAIUsage() {
  try {
    const { data } = await aiApi.getUsage()
    aiUsage.value = data
  } catch { aiUsage.value = null }
}

async function testAIConnection() {
  aiTesting.value = true
  aiTestResult.value = null
  try {
    const { data } = await aiApi.testConnection()
    aiTestResult.value = { success: true, ...data }
  } catch (e: any) {
    aiTestResult.value = { success: false, error: e.response?.data?.error || t('settings.aiTestFailed') }
  } finally {
    aiTesting.value = false
  }
}

async function loadReasoningConfig() {
  reasoningLoading.value = true
  try {
    const { data } = await reasoningApi.getConfig()
    reasoningConfig.value = data
    reasoningEnabled.value = data.enabled || false
    reasoningHasKey.value = data.has_api_key || false
    reasoningForm.value = {
      provider: data.provider || 'openai',
      model: data.model || '',
      api_key: '',
      base_url: data.base_url || '',
      dialectic_depth: data.dialectic_depth || 1,
      reasoning_level: data.reasoning_level || 'medium',
      max_tokens: data.max_tokens || 1000,
    }
  } catch {
    reasoningConfig.value = null
  } finally {
    reasoningLoading.value = false
  }
}

async function updateReasoningEnabled() {
  try {
    await reasoningApi.updateConfig({ enabled: reasoningEnabled.value })
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    reasoningEnabled.value = !reasoningEnabled.value
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  }
}

function onReasoningProviderChange() {
  if (reasoningForm.value.provider === 'ollama') {
    reasoningForm.value.base_url = 'http://localhost:11434/v1'
    reasoningForm.value.model = reasoningForm.value.model || 'llama3'
  } else if (reasoningForm.value.provider === 'deepseek') {
    reasoningForm.value.base_url = 'https://api.deepseek.com/v1'
    reasoningForm.value.model = reasoningForm.value.model || 'deepseek-chat'
  } else if (reasoningForm.value.provider === 'openai') {
    reasoningForm.value.base_url = ''
    reasoningForm.value.model = reasoningForm.value.model || 'gpt-4o-mini'
  }
}

async function saveReasoningConfig() {
  reasoningSaving.value = true
  try {
    const payload: Record<string, any> = { ...reasoningForm.value }
    if (!payload.api_key) {
      delete payload.api_key
    }
    await reasoningApi.updateConfig(payload)
    ElMessage.success(t('common.success'))
    await loadReasoningConfig()
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally {
    reasoningSaving.value = false
  }
}

async function testReasoningConnection() {
  reasoningTesting.value = true
  reasoningTestResult.value = null
  try {
    await reasoningApi.testConnection()
    reasoningTestResult.value = { success: true }
  } catch (e: any) {
    reasoningTestResult.value = { success: false, error: e.response?.data?.error || t('settings.reasoningTestFailed') }
  } finally {
    reasoningTesting.value = false
  }
}

async function loadAIProviders() {
  try {
    const { data } = await aiApi.getProviders()
    aiProviders.value = data.providers || []
  } catch { aiProviders.value = [] }
}

function onAIProviderChange() {
  const p = aiProviders.value.find((x: any) => x.ID === aiForm.value.provider_id)
  if (p && p.Models && p.Models.length > 0) {
    aiForm.value.model = p.Models[0]
  } else {
    aiForm.value.model = ''
  }
}

async function saveAIConfig() {
  aiSaving.value = true
  try {
    const payload: Record<string, any> = {
      provider_id: aiForm.value.provider_id,
      model: aiForm.value.model,
    }
    if (aiForm.value.api_key) payload.api_key = aiForm.value.api_key
    if (aiForm.value.base_url) payload.base_url = aiForm.value.base_url
    await aiApi.updateConfig(payload)
    ElMessage.success(t('common.success'))
    showAIConfigDialog.value = false
    await loadAIConfig()
  } catch (e: any) {
    const errMsg = e.response?.data?.error || ''
    if (e.response?.status === 403) {
      ElMessage.error(t('settings.aiProRequired'))
    } else {
      ElMessage.error(translateError(errMsg, t('common.failed')))
    }
  } finally {
    aiSaving.value = false
  }
}

watch(showAIConfigDialog, async (v) => {
  if (v) {
    await loadAIProviders()
    if (aiConfig.value) {
      aiForm.value.provider_id = aiConfig.value.provider_id || ''
      aiForm.value.model = aiConfig.value.model || ''
      aiForm.value.api_key = ''
      aiForm.value.base_url = aiConfig.value.base_url || ''
    }
  }
})

async function loadAgentsMD() {
  agentsMdLoading.value = true
  try {
    const { data } = await axios.get('/agent/agents-md')
    agentsMdContent.value = data.content || ''
  } catch {
    agentsMdContent.value = ''
  } finally {
    agentsMdLoading.value = false
  }
}

async function copyAgentsMD() {
  if (!agentsMdContent.value) {
    await loadAgentsMD()
  }
  if (!agentsMdContent.value) {
    ElMessage.error(t('common.failed'))
    return
  }
  try {
    await navigator.clipboard.writeText(agentsMdContent.value)
    ElMessage.success(t('settings.agentsMdCopied'))
  } catch {
    ElMessage.error(t('common.copyFailed'))
  }
}

watch(showAgentsMdPreview, async (v) => {
  if (v && !agentsMdContent.value) {
    await loadAgentsMD()
  }
})
</script>

<style scoped>
.settings-page { padding: 28px; max-width: 1200px; margin: 0 auto; }
.settings-layout { display: flex; gap: 20px; }
.page-header { margin-bottom: 24px; }
.page-header h1 { font-size: 24px; font-weight: 700; color: var(--cm-text); margin: 0; }
.settings-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(440px, 1fr)); gap: 16px; }
.settings-card { background: var(--cm-bg-secondary); border: 1px solid var(--cm-border); border-radius: 12px; padding: 20px; transition: border-color 0.3s, box-shadow 0.3s; }
.settings-card.section-highlight { border-color: #10B981; box-shadow: 0 0 0 2px rgba(16,185,129,0.2); }
.card-title { font-size: 16px; font-weight: 600; color: var(--cm-text); margin-bottom: 16px; padding-bottom: 12px; border-bottom: 1px solid var(--cm-border); }
.license-status .status-row { display: flex; justify-content: space-between; align-items: flex-start; padding: 8px 0; }
.status-label { color: var(--cm-text-muted); font-size: 13px; }
.status-value { font-size: 13px; color: var(--cm-text); }
.status-value.pro { color: #10B981; font-weight: 600; }
.status-value.text-warning { color: #F59E0B; font-weight: 600; }
.feature-tags { display: flex; flex-wrap: wrap; gap: 4px; max-width: 280px; justify-content: flex-end; }
.ftag { padding: 1px 8px; background: rgba(16,185,129,0.12); color: #10B981; border-radius: 4px; font-size: 11px; }
.license-free { text-align: center; }
.free-badge { font-size: 18px; font-weight: 600; color: var(--cm-text-muted); margin-bottom: 8px; }
.free-desc { color: var(--cm-text-muted); font-size: 13px; margin-bottom: 16px; }
.pricing { display: flex; gap: 12px; margin-bottom: 16px; }
.price-card { flex: 1; background: var(--cm-bg); border: 1px solid var(--cm-border); border-radius: 10px; padding: 16px; position: relative; }
.price-card.featured { border-color: rgba(16,185,129,0.4); background: rgba(16,185,129,0.04); }
.price-badge { position: absolute; top: -8px; right: 12px; background: #10B981; color: var(--cm-bg); font-size: 10px; padding: 2px 8px; border-radius: 8px; font-weight: 600; }
.price-name { font-size: 14px; font-weight: 600; color: var(--cm-text); margin-bottom: 4px; }
.price-amount { font-size: 24px; font-weight: 700; color: #10B981; }
.price-features { list-style: none; padding: 0; margin: 8px 0 0; font-size: 12px; color: var(--cm-text-muted); line-height: 1.8; text-align: left; }
.activate-section { display: flex; gap: 8px; justify-content: center; }
.license-input { width: 260px; }
.pro-install-status { margin-top: 12px; padding: 12px; background: rgba(16,185,129,0.05); border-radius: 8px; }
.status-text { font-size: 13px; color: var(--cm-text); margin-bottom: 8px; }
.section-desc { font-size: 13px; color: var(--cm-text-muted); margin: 0 0 12px; }
.invitation-create-form { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.form-label { font-size: 12px; color: var(--cm-text-muted); white-space: nowrap; }
.invitation-list { display: flex; flex-direction: column; gap: 8px; margin-top: 8px; }
.invitation-item { padding: 10px; background: var(--cm-bg); border-radius: 8px; border: 1px solid var(--cm-border); }
.invitation-code-row { display: flex; align-items: center; gap: 8px; }
.invitation-code { font-family: monospace; font-size: 12px; padding: 4px 8px; background: rgba(16,185,129,0.06); border-radius: 4px; user-select: all; word-break: break-all; }
.invitation-meta { font-size: 12px; color: var(--cm-text-muted); margin-top: 6px; }
.empty-hint { font-size: 13px; color: var(--cm-text-muted); text-align: center; padding: 20px; }
.user-list { display: flex; flex-direction: column; gap: 8px; margin-top: 8px; }
.user-item { display: flex; justify-content: space-between; align-items: center; padding: 10px; background: var(--cm-bg); border-radius: 8px; border: 1px solid var(--cm-border); }
.user-info { display: flex; align-items: center; gap: 8px; }
.user-name { font-weight: 500; font-size: 14px; }
.user-time { font-size: 12px; color: var(--cm-text-muted); }
.setting-item { display: flex; justify-content: space-between; align-items: center; padding: 10px 0; border-bottom: 1px solid var(--cm-border); font-size: 14px; color: var(--cm-text); }
.setting-desc { color: var(--cm-text-muted); font-size: 13px; }
.update-link { color: var(--cm-primary, #6366f1); text-decoration: none; margin-left: 8px; font-weight: 500; }
.update-link:hover { text-decoration: underline; }
.release-notes { max-width: 400px; white-space: pre-wrap; word-break: break-word; font-size: 12px; line-height: 1.5; }
.code-hint { font-family: monospace; background: var(--cm-bg, #f5f5f5); padding: 4px 8px; border-radius: 4px; font-size: 12px; color: var(--cm-text); user-select: all; }
.backup-list { margin-top: 12px; border-top: 1px solid var(--cm-border); padding-top: 8px; max-height: 200px; overflow-y: auto; }
.backup-item { display: flex; justify-content: space-between; align-items: center; padding: 8px 0; border-bottom: 1px solid var(--cm-border); }
.backup-item:last-child { border-bottom: none; }
.backup-info { display: flex; flex-direction: column; gap: 2px; }
.backup-name { font-size: 13px; color: var(--cm-text); font-weight: 500; }
.backup-meta { font-size: 11px; color: var(--cm-text-muted); }
.backup-actions { display: flex; gap: 4px; }
@media (max-width: 768px) {
  .settings-grid {
    grid-template-columns: 1fr;
  }
  .pricing {
    flex-direction: column;
  }
  .settings-page {
    padding: 16px;
  }
  .settings-card {
    padding: 16px;
  }
  .settings-card h2 {
    font-size: 16px;
  }
  .setting-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }
  .license-input {
    width: 100%;
  }
  .activate-section {
    flex-direction: column;
  }
  .activate-section .el-button {
    width: 100%;
  }
  .price-card {
    padding: 16px;
  }
  .price-name {
    font-size: 16px;
  }
  .price-amount {
    font-size: 20px;
  }
  .price-features {
    font-size: 11px;
  }
  .decay-stage-info {
    grid-template-columns: 1fr;
  }
  .stage-item {
    padding: 6px 10px;
  }
}

@media (max-width: 480px) {
  .settings-page {
    padding: 12px;
  }
  .settings-card {
    padding: 14px;
    border-radius: 10px;
  }
  .settings-card h2 {
    font-size: 15px;
  }
  .setting-item {
    padding: 8px 0;
  }
  .setting-desc {
    font-size: 12px;
  }
  .price-card {
    padding: 14px;
  }
  .price-amount {
    font-size: 18px;
  }
  .pro-install-status {
    padding: 10px;
  }
  .status-text {
    font-size: 12px;
  }
  .backup-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
  }
  .backup-actions {
    width: 100%;
    justify-content: flex-end;
  }
}
.decay-info { margin-bottom: 16px; }
.decay-stage-info { display: grid; grid-template-columns: repeat(2, 1fr); gap: 8px; }
.stage-item { display: flex; justify-content: space-between; padding: 8px 12px; background: var(--cm-bg); border-radius: 8px; font-size: 12px; }
.stage-label { color: var(--cm-text-muted); }
.stage-desc { color: var(--cm-text); font-weight: 500; }
.decay-stats { margin-top: 12px; padding-top: 12px; border-top: 1px solid var(--cm-border); }
.stats-row { display: flex; justify-content: space-between; padding: 6px 0; font-size: 13px; }
.stats-label { color: var(--cm-text-muted); }
.stats-value { color: var(--cm-text); font-weight: 500; }
.stats-value.warning { color: #ffc107; }
.stats-value.danger { color: #e91e63; }
.decay-actions { margin-top: 12px; display: flex; gap: 8px; }
.health-display { text-align: center; }
.health-score { font-size: 48px; font-weight: 700; margin: 8px 0; }
.health-score.grade-a { color: #10B981; }
.health-score.grade-b { color: #F59E0B; }
.health-score.grade-c { color: #EF4444; }
.health-grade { font-size: 18px; font-weight: 600; color: var(--cm-text-muted); margin-bottom: 12px; }
.health-dimensions { text-align: left; }
.dim-item { display: flex; align-items: center; gap: 8px; padding: 4px 0; }
.dim-label { font-size: 12px; color: var(--cm-text-muted); width: 60px; }
.dim-bar { flex: 1; height: 6px; background: var(--cm-border); border-radius: 3px; overflow: hidden; }
.dim-fill { height: 100%; background: #10B981; border-radius: 3px; transition: width 0.5s; }
.dim-value { font-size: 12px; color: var(--cm-text); width: 36px; text-align: right; }
.health-suggestions { margin-top: 12px; text-align: left; }
.suggestion { font-size: 12px; color: var(--cm-text-muted); padding: 4px 0; border-top: 1px solid var(--cm-border); }
.dedup-display { margin-bottom: 12px; }
.dedup-stats { display: flex; gap: 16px; font-size: 13px; color: var(--cm-text); margin-bottom: 8px; }
.dedup-groups { max-height: 200px; overflow-y: auto; }
.dedup-group { margin-bottom: 8px; }
.dedup-group-header { font-size: 13px; font-weight: 600; color: var(--cm-text); margin-bottom: 4px; }
.dedup-item { display: flex; gap: 8px; font-size: 12px; padding: 2px 0; }
.dedup-id { color: var(--cm-text-muted); }
.dedup-value { color: var(--cm-text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 280px; }
.api-key-list { display: flex; flex-direction: column; gap: 8px; }
.api-key-item { display: flex; justify-content: space-between; align-items: center; padding: 10px 12px; background: var(--cm-bg); border: 1px solid var(--cm-border); border-radius: 8px; }
.api-key-info { display: flex; flex-direction: column; gap: 2px; }
.api-key-name { font-size: 14px; font-weight: 500; color: var(--cm-text); }
.api-key-prefix { font-family: monospace; font-size: 12px; color: var(--cm-text-muted); }
.api-key-perms { font-size: 12px; color: var(--cm-text-secondary); letter-spacing: 2px; }
.api-key-time { font-size: 11px; color: var(--cm-text-muted); }
.openclaw-paths { margin: 8px 0; padding: 8px; background: var(--cm-bg); border-radius: 6px; }
.openclaw-path-item { font-size: 11px; color: var(--cm-text-muted); margin: 4px 0; word-break: break-all; }
.openclaw-path-item code { font-size: 11px; }
.agent-list { margin: 8px 0; }
.agent-item { padding: 10px; margin-bottom: 8px; background: var(--cm-bg); border-radius: 6px; }
.agent-item-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px; }
.agent-name { font-weight: 600; font-size: 13px; }
.risk-category { margin-bottom: 12px; }
.risk-category-title { font-weight: 600; font-size: 12px; color: var(--cm-text-muted); margin-bottom: 6px; padding-bottom: 4px; border-bottom: 1px solid var(--cm-border); }
</style>
