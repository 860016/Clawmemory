<template>
  <div class="settings-page">
    <div class="page-header">
      <h1>⚙️ {{ $t('settings.title') }}</h1>
    </div>

    <div class="settings-grid">
      <!-- 授权管理 -->
      <div class="settings-card" :class="{ 'section-highlight': activeSection === 'license' }" id="settings-license">
        <div class="card-title">🔑 {{ $t('settings.license') }}</div>
        <div class="license-status" v-if="license.active">
          <div class="status-row">
            <span class="status-label">{{ $t('settings.version') }}</span>
            <span class="status-value pro">{{ license.type === 'enterprise' ? 'Enterprise' : 'Pro' }}</span>
          </div>
          <div class="status-row" v-if="license.license_key">
            <span class="status-label">{{ $t('settings.licenseKey') }}</span>
            <span class="status-value">{{ license.license_key }}</span>
          </div>
          <div class="status-row" v-if="license.expires_at">
            <span class="status-label">{{ $t('settings.expiresAt') }}</span>
            <span class="status-value">{{ license.expires_at }}</span>
          </div>
          <div class="status-row" v-if="license.device_slot">
            <span class="status-label">{{ $t('settings.deviceSlot') }}</span>
            <span class="status-value">{{ license.device_slot }}</span>
          </div>
          <div class="status-row">
            <span class="status-label">{{ $t('settings.features') }}</span>
            <div class="feature-tags">
              <span class="ftag" v-for="f in license.features" :key="f">{{ featureLabels[f] || f }}</span>
            </div>
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
        <div class="card-title">◇ {{ $t('settings.language') }}</div>
        <div class="setting-item">
          <span>{{ $t('settings.language') }}</span>
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
            <el-tag :type="aiConfig.provider_id === 'nvidia-nim' ? 'success' : 'primary'" size="small">
              {{ aiConfig.provider_name || aiConfig.provider_id }}
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
        <div class="ai-free-hint" v-if="aiConfig && !aiConfig.is_pro" style="margin-top: 8px; padding: 10px; background: var(--cm-bg-secondary); border-radius: 6px; font-size: 12px; color: var(--cm-text-muted)">
          <div style="font-weight: 600; margin-bottom: 4px">{{ $t('settings.aiFreeHint') }}</div>
          <div>{{ $t('settings.aiFreeHintDesc') }}</div>
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

      <!-- OpenClaw 连接配置 -->
      <div class="settings-card" id="settings-openclaw">
        <div class="card-title">🔗 {{ $t('settings.openclawConnection') }}</div>
        <div class="setting-item" v-if="openclawStatus">
          <span>{{ $t('settings.connectionMode') }}</span>
          <el-tag :type="openclawStatus.mode === 'local' ? 'success' : 'warning'" size="small">
            {{ openclawStatus.mode === 'local' ? $t('settings.localMode') : $t('settings.remoteMode') }}
          </el-tag>
        </div>
        <div class="setting-item" v-if="openclawStatus">
          <span>{{ $t('settings.autoRecord') }}</span>
          <el-switch v-model="openclawAutoSync" @change="toggleOpenClawSync" :loading="openclawSyncLoading" />
        </div>
        <div v-if="openclawStatus && openclawStatus.mode === 'local'" class="openclaw-local-info">
          <div class="setting-item">
            <span>{{ $t('settings.localDetected') }}</span>
            <el-tag type="success" size="small">✓</el-tag>
          </div>
          <div v-if="openclawStatus.local_paths && openclawStatus.local_paths.length > 0" class="openclaw-paths">
            <div v-for="p in openclawStatus.local_paths" :key="p" class="openclaw-path-item">
              <code>{{ p }}</code>
            </div>
          </div>
          <div class="setting-item" v-if="openclawStatus.synced_count > 0">
            <span>{{ $t('settings.syncedCount') }}</span>
            <span class="setting-desc">{{ openclawStatus.synced_count }}</span>
          </div>
          <div class="setting-item" v-if="openclawStatus.skipped_count > 0">
            <span>{{ $t('settings.skippedCount') }}</span>
            <span class="setting-desc">{{ openclawStatus.skipped_count }}</span>
          </div>
          <div class="setting-item">
            <span>{{ $t('settings.forceSync') }}</span>
            <el-button size="small" type="primary" @click="forceOpenClawSync" :loading="openclawSyncLoading">{{ $t('settings.syncNow') }}</el-button>
          </div>
        </div>
        <div v-if="openclawStatus && openclawStatus.mode === 'remote'" class="openclaw-remote-info">
          <el-alert type="info" :closable="false" style="margin-bottom: 12px; font-size: 12px">
            <template #title>{{ $t('settings.remoteModeHint') }}</template>
          </el-alert>
          <div class="api-usage-hint" style="margin-top: 8px; padding: 10px; background: var(--cm-bg-secondary); border-radius: 6px; font-size: 12px">
            <div style="font-weight: 600; margin-bottom: 6px">{{ $t('settings.pushEndpoint') }}</div>
            <code style="display: block; padding: 8px; background: var(--cm-bg); border-radius: 4px; font-size: 11px; word-break: break-all; color: var(--cm-accent)">
              curl -X POST http://localhost:8765/api/v1/external/conversations \<br>
              &nbsp;&nbsp;-H "X-API-Key: YOUR_KEY" \<br>
              &nbsp;&nbsp;-H "Content-Type: application/json" \<br>
              &nbsp;&nbsp;-d '{"agent_name":"openclaw","session_id":"xxx","messages":[{"role":"user","content":"..."}]}'
            </code>
          </div>
          <div class="setting-item" style="margin-top: 12px" v-if="apiKeys.length === 0">
            <span class="setting-desc">{{ $t('settings.needApiKeyForRemote') }}</span>
            <el-button size="small" type="primary" @click="openApiKeyDialog">{{ $t('settings.createApiKey') }}</el-button>
          </div>
        </div>
        <div v-if="!openclawStatus" class="setting-item">
          <span>{{ $t('settings.checkingStatus') }}</span>
          <el-button size="small" @click="loadOpenClawStatus" :loading="openclawSyncLoading">{{ $t('settings.refresh') }}</el-button>
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
          <el-button size="small" type="primary" @click="exportData">{{ $t('settings.export') }}</el-button>
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
    </div>

    <el-dialog v-model="showPasswordDialog" :title="passwordSet ? $t('settings.changePassword') : $t('settings.setPassword')" width="400px">
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

    <el-dialog v-model="showApiKeyDialog" :title="$t('settings.createApiKey')" width="460px" :close-on-click-modal="false">
      <div v-if="!newApiKeyRaw">
        <el-form label-position="top">
          <el-form-item :label="$t('settings.apiKeyName')">
            <el-input v-model="newApiKeyName" :placeholder="$t('settings.apiKeyNamePlaceholder')" />
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

    <el-dialog v-model="showAIConfigDialog" :title="$t('settings.aiConfigure')" width="520px" :close-on-click-modal="false">
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

    <el-dialog v-model="showAgentsMdPreview" :title="$t('settings.agentsMdTitle')" width="640px">
      <div style="max-height: 500px; overflow-y: auto; padding: 12px; background: var(--cm-bg-secondary); border-radius: 6px; font-family: monospace; font-size: 12px; white-space: pre-wrap; word-break: break-all; line-height: 1.6">
{{ agentsMdContent }}
      </div>
      <template #footer>
        <el-button @click="showAgentsMdPreview = false">{{ $t('common.cancel') }}</el-button>
        <el-button type="primary" @click="copyAgentsMD">{{ $t('settings.agentsMdCopy') }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import axios from '../api/go-client'
import { setLocale, getLocale, translateError } from '../i18n'
import { aiApi } from '../api/go-ai'

const { t } = useI18n()
const route = useRoute()
const activeSection = ref((route.query.section as string) || '')
const license = ref<any>({ active: false, tier: 'oss', features: [] })
const activating = ref(false)
const licenseKey = ref('')
const exporting = ref(false)
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
  provider_id: 'nvidia-nim',
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
  // Enterprise only
  api_access: t('settings.featApiAccess'),
  sso: t('settings.featSso'),
  audit_log: t('settings.featAuditLog'),
  time_travel: t('settings.featTimeTravel'),
  offline_mode: t('settings.featOfflineMode'),
}

onMounted(async () => {
  await Promise.all([loadLicense(), loadInitStatus(), loadInstallStatus(), loadDecaySettings(), loadDecayStats(), loadApiKeys(), loadRecordSensitiveSetting(), loadOpenClawStatus(), loadAIConfig(), loadAIUsage()])
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
    el.scrollIntoView({ behavior: 'smooth', block: 'center' })
    el.classList.add('section-highlight')
    setTimeout(() => el.classList.remove('section-highlight'), 2000)
  }
}

function changeLocale(locale: 'zh' | 'en') {
  setLocale(locale)
}

async function loadLicense() {
  try { const { data } = await axios.get('/license/info'); license.value = data } catch {}
}

async function loadApiKeys() {
  try {
    const { data } = await axios.get('/api-keys')
    apiKeys.value = data.items || []
  } catch {}
}

function openApiKeyDialog() {
  if (apiKeys.value.length >= 5) {
    ElMessage.warning(t('settings.apiKeyMaxReached'))
    return
  }
  newApiKeyName.value = ''
  newApiKeyRaw.value = ''
  showApiKeyDialog.value = true
}

async function createApiKey() {
  if (!newApiKeyName.value.trim()) {
    ElMessage.warning(t('settings.apiKeyName'))
    return
  }
  creatingApiKey.value = true
  try {
    const { data } = await axios.post('/api-keys', { name: newApiKeyName.value.trim() })
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

const openclawStatus = ref<any>(null)
const openclawAutoSync = ref(true)
const openclawSyncLoading = ref(false)

async function loadOpenClawStatus() {
  openclawSyncLoading.value = true
  try {
    const { data } = await axios.get('/openclaw-sync/status')
    openclawStatus.value = data
    openclawAutoSync.value = data.auto_sync_enabled
  } catch {
    openclawStatus.value = null
  } finally {
    openclawSyncLoading.value = false
  }
}

async function toggleOpenClawSync() {
  openclawSyncLoading.value = true
  try {
    await axios.post('/openclaw-sync/toggle', { enabled: openclawAutoSync.value })
    ElMessage.success(t('common.success'))
  } catch (e: any) {
    openclawAutoSync.value = !openclawAutoSync.value
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally {
    openclawSyncLoading.value = false
  }
}

async function forceOpenClawSync() {
  openclawSyncLoading.value = true
  try {
    const { data } = await axios.post('/openclaw-sync/force')
    ElMessage.success(t('settings.syncCompleted', { count: data.synced_count || 0 }))
    await loadOpenClawStatus()
  } catch (e: any) {
    ElMessage.error(translateError(e.response?.data?.error, t('common.failed')))
  } finally {
    openclawSyncLoading.value = false
  }
}

async function loadRecordSensitiveSetting() {
  try {
    const { data } = await axios.get('/settings')
    recordSensitive.value = !!data.record_sensitive_content
  } catch {}
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
  try { const { data } = await axios.get('/auth/init-status'); passwordSet.value = data.password_set } catch {}
}

async function loadInstallStatus() {
  try { const { data } = await axios.get('/install-status'); coreEngine.value = data.checks?.security_engine || 'python'; if (data.version) appVersion.value = data.version } catch {}
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
    const { data } = await axios.get('/data/export')
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
  activating.value = true
  try {
    const { data } = await axios.post('/license/activate', { license_key: licenseKey.value })
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
      ElMessage.error(translateError(detail, t('common.failed')))
    } else {
      ElMessage.error(t('common.failed'))
    }
  } finally { activating.value = false }
}

async function deactivateLicense() {
  try {
    await ElMessageBox.confirm(t('settings.cancelConfirm'), t('common.confirm'), { type: 'warning' })
    await axios.post('/license/deactivate')
    ElMessage.success(t('settings.canceled')); await loadLicense()
  } catch {}
}

async function handleSetPassword() {
  if (newPassword.value.length < 4) { ElMessage.warning(t('settings.passwordMinLen')); return }
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
  } catch {}
}

async function loadDecayStats() {
  try {
    const { data } = await axios.get('/memories/decay/stats')
    decayStats.value = data.stats || data
  } catch {}
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
  } catch {}
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
  } catch {}
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

async function loadAIProviders() {
  try {
    const { data } = await aiApi.getProviders()
    aiProviders.value = data.providers || []
  } catch {}
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
      aiForm.value.provider_id = aiConfig.value.provider_id || 'nvidia-nim'
      aiForm.value.model = aiConfig.value.model || ''
      aiForm.value.api_key = ''
      aiForm.value.base_url = aiConfig.value.base_url || ''
    }
  }
})

async function loadAgentsMD() {
  agentsMdLoading.value = true
  try {
    const { data } = await axios.get('/openclaw/agents-md')
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
.settings-page { padding: 28px; max-width: 1000px; margin: 0 auto; }
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
.api-key-time { font-size: 11px; color: var(--cm-text-muted); }
.openclaw-paths { margin: 8px 0; padding: 8px; background: var(--cm-bg); border-radius: 6px; }
.openclaw-path-item { font-size: 11px; color: var(--cm-text-muted); margin: 4px 0; word-break: break-all; }
.openclaw-path-item code { font-size: 11px; }
</style>
