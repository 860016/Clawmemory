<template>
  <div class="skills-page">
    <div class="page-header">
      <h2>✨ {{ $t('skills.title') }}</h2>
      <el-button type="primary" @click="scanSkills" :loading="scanning">
        {{ scanning ? $t('skills.scanning') : $t('skills.scanSkills') }}
      </el-button>
    </div>

    <div v-if="!scanned && !scanning" class="empty-state">
      <el-icon :size="48" color="var(--cm-text-muted)"><MagicStick /></el-icon>
      <p>{{ $t('skills.noSkills') }}</p>
      <p class="hint">~/.openclaw/skills/ · .openclaw/skills/</p>
    </div>

    <div v-if="scanning" class="loading-state">
      <el-icon class="spin" :size="32"><Loading /></el-icon>
      <p>{{ $t('skills.scanning') }}</p>
    </div>

    <div v-if="scanned && !scanning">
      <!-- Global Skills -->
      <div v-if="globalSkills.length" class="skill-section">
        <h3>{{ $t('skills.globalSkills') }} ({{ globalSkills.length }})</h3>
        <div class="skill-grid">
          <div v-for="skill in globalSkills" :key="skill.skill_dir" class="skill-card" @click="showDetail(skill)">
            <div class="skill-icon">🌐</div>
            <div class="skill-info">
              <div class="skill-name">{{ skill.name }}</div>
              <div class="skill-desc">{{ skill.description || '—' }}</div>
            </div>
            <div class="skill-meta">
              <span class="badge">{{ skill.version }}</span>
              <span class="scope global">{{ skill.scope }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Workspace Skills -->
      <div v-if="workspaceSkills.length" class="skill-section">
        <h3>{{ $t('skills.workspaceSkills') }} ({{ workspaceSkills.length }})</h3>
        <div class="skill-grid">
          <div v-for="skill in workspaceSkills" :key="skill.skill_dir" class="skill-card" @click="showDetail(skill)">
            <div class="skill-icon">📁</div>
            <div class="skill-info">
              <div class="skill-name">{{ skill.name }}</div>
              <div class="skill-desc">{{ skill.description || '—' }}</div>
            </div>
            <div class="skill-meta">
              <span class="badge">{{ skill.version }}</span>
              <span class="scope workspace">{{ skill.scope }}</span>
            </div>
          </div>
        </div>
      </div>

      <div v-if="!globalSkills.length && !workspaceSkills.length" class="empty-state">
        <p>{{ $t('skills.noSkills') }}</p>
      </div>

      <div class="total-bar" v-if="globalSkills.length || workspaceSkills.length">
        {{ $t('skills.scanSkills') }}: {{ globalSkills.length + workspaceSkills.length }}
      </div>
    </div>

    <!-- Detail Dialog -->
    <el-dialog v-model="detailVisible" :title="detailSkill?.name" width="600px">
      <div v-if="detailSkill" class="skill-detail">
        <div class="detail-row"><strong>{{ $t('skills.name') }}:</strong> {{ detailSkill.name }}</div>
        <div class="detail-row"><strong>{{ $t('skills.description') }}:</strong> {{ detailSkill.description || '—' }}</div>
        <div class="detail-row"><strong>{{ $t('skills.version') }}:</strong> {{ detailSkill.version }}</div>
        <div class="detail-row"><strong>{{ $t('skills.author') }}:</strong> {{ detailSkill.author }}</div>
        <div class="detail-row"><strong>{{ $t('skills.scope') }}:</strong> {{ detailSkill.scope }}</div>
        <div class="detail-row" v-if="detailSkill.tags?.length">
          <strong>Tags:</strong>
          <el-tag v-for="tag in detailSkill.tags" :key="tag" size="small" style="margin: 2px">{{ tag }}</el-tag>
        </div>
        <div class="detail-row" v-if="detailSkill.files?.length">
          <strong>{{ $t('skills.files') }}:</strong>
          <div class="file-list">{{ detailSkill.files.join(', ') }}</div>
        </div>
        <div class="detail-body" v-if="detailSkill.body_full">
          <strong>Content:</strong>
          <pre>{{ detailSkill.body_full }}</pre>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { MagicStick, Loading } from '@element-plus/icons-vue'
import axios from '../api/client'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const scanning = ref(false)
const scanned = ref(false)
const globalSkills = ref<any[]>([])
const workspaceSkills = ref<any[]>([])
const detailVisible = ref(false)
const detailSkill = ref<any>(null)

async function scanSkills() {
  scanning.value = true
  try {
    const { data } = await axios.get('/openclaw-skills/scan')
    globalSkills.value = data.global_skills || []
    workspaceSkills.value = data.workspace_skills || []
    scanned.value = true
  } catch {
    globalSkills.value = []
    workspaceSkills.value = []
    scanned.value = true
  } finally {
    scanning.value = false
  }
}

async function showDetail(skill: any) {
  try {
    const { data } = await axios.get('/openclaw-skills/detail', {
      params: { skill_dir: skill.skill_dir, scope: skill.scope },
    })
    detailSkill.value = data
  } catch {
    detailSkill.value = skill
  }
  detailVisible.value = true
}
</script>

<style scoped>
.skills-page {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
}

.page-header h2 {
  margin: 0;
  font-size: 20px;
  color: var(--cm-text);
}

.empty-state, .loading-state {
  text-align: center;
  padding: 60px 20px;
  color: var(--cm-text-muted);
}

.hint {
  font-size: 12px;
  opacity: 0.6;
  margin-top: 8px;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.skill-section {
  margin-bottom: 32px;
}

.skill-section h3 {
  font-size: 15px;
  font-weight: 600;
  color: var(--cm-text);
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--cm-border);
}

.skill-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 12px;
}

.skill-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px;
  border-radius: 10px;
  border: 1px solid var(--cm-border);
  background: var(--cm-bg-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.skill-card:hover {
  border-color: var(--cm-primary);
  box-shadow: 0 2px 8px rgba(var(--cm-primary-rgb), 0.1);
}

.skill-icon {
  font-size: 24px;
  flex-shrink: 0;
}

.skill-info {
  flex: 1;
  min-width: 0;
}

.skill-name {
  font-weight: 600;
  font-size: 14px;
  color: var(--cm-text);
}

.skill-desc {
  font-size: 12px;
  color: var(--cm-text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.skill-meta {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
  flex-shrink: 0;
}

.badge {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 10px;
  background: rgba(var(--cm-primary-rgb), 0.1);
  color: var(--cm-primary);
}

.scope {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 4px;
  text-transform: uppercase;
}

.scope.global {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}

.scope.workspace {
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
}

.total-bar {
  text-align: center;
  color: var(--cm-text-muted);
  font-size: 13px;
  padding: 12px;
}

.skill-detail .detail-row {
  margin-bottom: 10px;
  font-size: 14px;
}

.detail-body {
  margin-top: 16px;
}

.detail-body pre {
  background: var(--cm-bg-secondary);
  border: 1px solid var(--cm-border);
  border-radius: 8px;
  padding: 12px;
  overflow-x: auto;
  font-size: 13px;
  max-height: 400px;
  overflow-y: auto;
  white-space: pre-wrap;
}

.file-list {
  font-size: 13px;
  color: var(--cm-text-muted);
  margin-top: 4px;
}

@media (max-width: 768px) {
  .skill-grid { grid-template-columns: 1fr; }
}
</style>
