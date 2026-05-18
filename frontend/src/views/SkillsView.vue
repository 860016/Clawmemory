<template>
  <div class="skills-page">
    <div class="page-header">
      <h2>✨ {{ $t('skills.title') }}</h2>
      <div class="header-actions">
        <el-button-group>
          <el-button :type="activeTab === 'learned' ? 'primary' : 'default'" @click="activeTab = 'learned'">
            🧠 {{ $t('skills.learnedSkills') }}
          </el-button>
          <el-button :type="activeTab === 'scanned' ? 'primary' : 'default'" @click="activeTab = 'scanned'">
            📁 {{ $t('skills.scannedSkills') }}
          </el-button>
        </el-button-group>
      </div>
    </div>

    <!-- Learned Skills Tab -->
    <div v-if="activeTab === 'learned'">
      <div class="learned-actions">
        <el-button type="primary" @click="detectPatterns" :loading="detecting">
          {{ detecting ? $t('skills.detecting') : $t('skills.detectPatterns') }}
        </el-button>
        <el-button @click="autoCreateSkill" :loading="creating">
          {{ $t('skills.autoCreate') }}
        </el-button>
        <el-button @click="generateSuggestions" :loading="suggesting">
          {{ $t('skills.suggestions') }}
        </el-button>
        <el-select v-model="skillStatusFilter" size="small" style="width: 150px; margin-left: 8px" @change="loadLearnedSkills">
          <el-option :label="$t('skills.statusActive')" value="active" />
          <el-option :label="$t('skills.statusNeedsImprovement')" value="needs_improvement" />
          <el-option :label="$t('skills.statusInactive')" value="inactive" />
        </el-select>
      </div>

      <div v-if="detecting" class="loading-state">
        <el-icon class="spin" :size="32"><Loading /></el-icon>
        <p>{{ $t('skills.detecting') }}</p>
      </div>

      <div v-if="patterns.length && !detecting" class="pattern-section">
        <h3>{{ $t('skills.detectedPatterns') }} ({{ patterns.length }})</h3>
        <div class="pattern-grid">
          <div v-for="(p, idx) in patterns" :key="idx" class="pattern-card">
            <div class="pattern-icon">🔄</div>
            <div class="pattern-info">
              <div class="pattern-seq">{{ (p.action_sequence || []).join(' → ') }}</div>
              <div class="pattern-meta">
                {{ $t('skills.occurrence') }}: {{ p.occurrence_count }} · {{ $t('skills.sessions') }}: {{ p.session_count }}
              </div>
            </div>
            <el-tag v-if="p.worth_skill" type="success" size="small">{{ $t('skills.worthSkill') }}</el-tag>
          </div>
        </div>
      </div>

      <div v-if="learnedSkills.length" class="skill-section">
        <h3>{{ $t('skills.learnedSkills') }} ({{ learnedSkills.length }})</h3>
        <div class="skill-grid">
          <div v-for="skill in learnedSkills" :key="skill.id" class="skill-card" @click="showLearnedDetail(skill)">
            <div class="skill-icon">{{ skill.auto_created ? '🤖' : '✍️' }}</div>
            <div class="skill-info">
              <div class="skill-name">{{ skill.name }}</div>
              <div class="skill-desc">{{ skill.description || '—' }}</div>
            </div>
            <div class="skill-meta">
              <span class="badge">v{{ skill.version }}</span>
              <span class="success-rate" v-if="skill.usage_count > 0">
                {{ Math.round((skill.success_count / skill.usage_count) * 100) }}%
              </span>
              <el-tag v-if="skill.status === 'needs_improvement'" type="warning" size="small">⚠ {{ $t('skills.needsImprovement') }}</el-tag>
            </div>
          </div>
        </div>
      </div>

      <div v-if="pendingSuggestions.length" class="suggestion-section">
        <h3>{{ $t('skills.agentSuggestions') }} ({{ pendingSuggestions.length }})</h3>
        <div class="suggestion-grid">
          <div v-for="sug in pendingSuggestions" :key="sug.id" class="suggestion-card">
            <div class="sug-icon">💡</div>
            <div class="sug-info">
              <div class="sug-title">{{ sug.title }}</div>
              <div class="sug-desc">{{ sug.description }}</div>
            </div>
            <div class="sug-actions">
              <el-button size="small" type="primary" v-if="sug.import_url" @click="openImport(sug.import_url)">
                {{ $t('skills.import') }}
              </el-button>
              <el-button size="small" @click="dismissSuggestion(sug.id)">
                {{ $t('skills.dismiss') }}
              </el-button>
            </div>
          </div>
        </div>
      </div>

      <div v-if="!learnedSkills.length && !patterns.length && !pendingSuggestions.length && !detecting" class="empty-state">
        <el-icon :size="48" color="var(--cm-text-muted)"><MagicStick /></el-icon>
        <p>{{ $t('skills.noLearnedSkills') }}</p>
        <p class="hint">{{ $t('skills.learnedHint') }}</p>
      </div>
    </div>

    <!-- Scanned Skills Tab -->
    <div v-if="activeTab === 'scanned'">
      <div class="scanned-actions">
        <el-button type="primary" @click="scanSkills" :loading="scanning">
          {{ scanning ? $t('skills.scanning') : $t('skills.scanSkills') }}
        </el-button>
      </div>

      <div v-if="scanning" class="loading-state">
        <el-icon class="spin" :size="32"><Loading /></el-icon>
        <p>{{ $t('skills.scanning') }}</p>
      </div>

      <div v-if="scanned && !scanning">
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
    </div>

    <!-- Detail Dialog for Scanned Skills -->
    <el-dialog v-model="detailVisible" :title="detailSkill?.name" width="600px" :fullscreen="isMobile">
      <div v-if="detailSkill" class="skill-detail">
        <div class="detail-row"><strong>{{ $t('skills.name') }}:</strong> {{ detailSkill.name }}</div>
        <div class="detail-row"><strong>{{ $t('skills.description') }}:</strong> {{ detailSkill.description || '—' }}</div>
        <div class="detail-row"><strong>{{ $t('skills.version') }}:</strong> {{ detailSkill.version }}</div>
        <div class="detail-row"><strong>{{ $t('skills.author') }}:</strong> {{ detailSkill.author }}</div>
        <div class="detail-row"><strong>{{ $t('skills.scope') }}:</strong> {{ detailSkill.scope }}</div>
        <div class="detail-row" v-if="detailSkill.tags?.length">
          <strong>{{ $t('advanced.tags') }}:</strong>
          <el-tag v-for="tag in detailSkill.tags" :key="tag" size="small" style="margin: 2px">{{ tag }}</el-tag>
        </div>
        <div class="detail-row" v-if="detailSkill.files?.length">
          <strong>{{ $t('skills.files') }}:</strong>
          <div class="file-list">{{ detailSkill.files.join(', ') }}</div>
        </div>
        <div class="detail-body" v-if="detailSkill.body_full">
          <strong>{{ $t('advanced.content') }}:</strong>
          <pre>{{ detailSkill.body_full }}</pre>
        </div>
      </div>
    </el-dialog>

    <!-- Detail Dialog for Learned Skills -->
    <el-dialog v-model="learnedDetailVisible" :title="learnedDetailSkill?.name" width="650px" :fullscreen="isMobile">
      <div v-if="learnedDetailSkill" class="skill-detail">
        <div class="detail-row"><strong>{{ $t('skills.name') }}:</strong> {{ learnedDetailSkill.name }}</div>
        <div class="detail-row"><strong>{{ $t('skills.description') }}:</strong> {{ learnedDetailSkill.description || '—' }}</div>
        <div class="detail-row"><strong>{{ $t('skills.version') }}:</strong> v{{ learnedDetailSkill.version }}</div>
        <div class="detail-row"><strong>{{ $t('skills.category') }}:</strong> {{ learnedDetailSkill.category }}</div>
        <div class="detail-row"><strong>{{ $t('skills.source') }}:</strong> {{ learnedDetailSkill.source_agent }}</div>
        <div class="detail-row">
          <strong>{{ $t('skills.usageStats') }}:</strong>
          {{ learnedDetailSkill.usage_count }} {{ $t('skills.times') }} ·
          {{ $t('skills.success') }} {{ learnedDetailSkill.success_count }} ·
          {{ $t('skills.fail') }} {{ learnedDetailSkill.fail_count }}
          <span v-if="learnedDetailSkill.usage_count > 0">
            ({{ Math.round((learnedDetailSkill.success_count / learnedDetailSkill.usage_count) * 100) }}%)
          </span>
        </div>
        <div class="detail-row" v-if="learnedDetailSkill.trigger_keywords">
          <strong>{{ $t('skills.triggerKeywords') }}:</strong>
          <template v-if="parseJSON(learnedDetailSkill.trigger_keywords)">
            <el-tag v-for="kw in parseJSON(learnedDetailSkill.trigger_keywords)" :key="kw" size="small" style="margin: 2px">{{ kw }}</el-tag>
          </template>
          <span v-else>{{ learnedDetailSkill.trigger_keywords }}</span>
        </div>
        <div class="detail-row" v-if="learnedDetailSkill.steps">
          <strong>{{ $t('skills.steps') }}:</strong>
          <div class="steps-list">
            <template v-if="parseJSON(learnedDetailSkill.steps)">
              <div v-for="(step, idx) in parseJSON(learnedDetailSkill.steps)" :key="idx" class="step-item">
                <span class="step-num">{{ idx + 1 }}</span>
                <span>{{ typeof step === 'string' ? step : step.action || step.detail || JSON.stringify(step) }}</span>
              </div>
            </template>
            <pre v-else>{{ learnedDetailSkill.steps }}</pre>
          </div>
        </div>
        <div class="detail-row" v-if="learnedDetailSkill.known_pitfalls">
          <strong>{{ $t('skills.pitfalls') }}:</strong>
          <div class="pitfalls-list">
            <template v-if="parseJSON(learnedDetailSkill.known_pitfalls)">
              <el-alert v-for="(pf, idx) in parseJSON(learnedDetailSkill.known_pitfalls)" :key="idx"
                :title="typeof pf === 'string' ? pf : pf.pitfall"
                :description="typeof pf === 'object' ? pf.solution : ''"
                type="warning" :closable="false" style="margin-bottom: 6px" />
            </template>
            <pre v-else>{{ learnedDetailSkill.known_pitfalls }}</pre>
          </div>
        </div>
        <div class="detail-row" v-if="learnedDetailSkill.verification">
          <strong>{{ $t('skills.verification') }}:</strong> {{ learnedDetailSkill.verification }}
        </div>
        <div class="detail-actions">
          <el-button size="small" @click="improveSkill(learnedDetailSkill.id)" :loading="improving">
            {{ $t('skills.improve') }}
          </el-button>
          <el-button v-if="learnedDetailSkill.status === 'needs_improvement'" size="small" type="success" @click="reactivateSkill(learnedDetailSkill.id)">
            {{ $t('skills.reactivate') }}
          </el-button>
          <el-button size="small" type="danger" @click="deactivateSkill(learnedDetailSkill.id)">
            {{ $t('skills.deactivate') }}
          </el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useIsMobile } from '../composables/useIsMobile'
import { MagicStick, Loading } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { skillsApi } from '../api/go-skills'
import { openClawSkillsApi } from '../api/go-openclaw-skills'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const activeTab = ref<'learned' | 'scanned'>('learned')
const { isMobile } = useIsMobile()

const scanning = ref(false)
const scanned = ref(false)
const globalSkills = ref<any[]>([])
const workspaceSkills = ref<any[]>([])
const detailVisible = ref(false)
const detailSkill = ref<any>(null)

const detecting = ref(false)
const creating = ref(false)
const suggesting = ref(false)
const improving = ref(false)
const patterns = ref<any[]>([])
const learnedSkills = ref<any[]>([])
const pendingSuggestions = ref<any[]>([])
const learnedDetailVisible = ref(false)
const learnedDetailSkill = ref<any>(null)
const skillStatusFilter = ref('active')

onMounted(() => {
  loadLearnedSkills()
  loadSuggestions()
})

function parseJSON(str: string) {
  try {
    const parsed = JSON.parse(str)
    if (Array.isArray(parsed)) return parsed
    return null
  } catch {
    return null
  }
}

async function scanSkills() {
  scanning.value = true
  try {
    const { data } = await openClawSkillsApi.scan()
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
    const { data } = await openClawSkillsApi.getDetail({
      skill_dir: skill.skill_dir,
      scope: skill.scope,
    })
    detailSkill.value = data
  } catch {
    detailSkill.value = skill
  }
  detailVisible.value = true
}

async function loadLearnedSkills() {
  try {
    const { data } = await skillsApi.listSkills(skillStatusFilter.value)
    learnedSkills.value = data.skills || []
  } catch {
    learnedSkills.value = []
  }
}

async function detectPatterns() {
  detecting.value = true
  try {
    const { data } = await skillsApi.detectPatterns()
    patterns.value = data.patterns || []
    if (patterns.value.length === 0) {
      ElMessage.info(t('skills.noPatterns'))
    }
  } catch {
    patterns.value = []
  } finally {
    detecting.value = false
  }
}

async function autoCreateSkill() {
  creating.value = true
  try {
    const { data } = await skillsApi.createSkill({ use_ai: true })
    if (data.status === 'created') {
      ElMessage.success(t('skills.skillCreated'))
      loadLearnedSkills()
    } else if (data.status === 'duplicate') {
      ElMessage.info(data.message)
    } else if ((data.skills && data.skills.length > 0) || (data.new_skills && data.new_skills.length > 0)) {
      const count = data.skills?.length || data.new_skills?.length || data.skills_created || 0
      ElMessage.success(t('skills.skillsCreated', { count }))
      loadLearnedSkills()
    } else if (data.skills_created > 0) {
      ElMessage.success(t('skills.skillsCreated', { count: data.skills_created }))
      loadLearnedSkills()
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('skills.createFailed'))
  } finally {
    creating.value = false
  }
}

async function loadSuggestions() {
  try {
    const { data } = await skillsApi.getSuggestions()
    pendingSuggestions.value = data.suggestions || []
  } catch {
    pendingSuggestions.value = []
  }
}

async function generateSuggestions() {
  suggesting.value = true
  try {
    const { data } = await skillsApi.generateSuggestions()
    pendingSuggestions.value = data.suggestions || []
    ElMessage.success(t('skills.suggestionsGenerated', { count: data.count || 0 }))
  } catch {
    ElMessage.error(t('skills.suggestionFailed'))
  } finally {
    suggesting.value = false
  }
}

async function dismissSuggestion(id: number) {
  try {
    await skillsApi.dismissSuggestion(id)
    pendingSuggestions.value = pendingSuggestions.value.filter(s => s.id !== id)
    ElMessage.success(t('skills.dismissed'))
  } catch {
    ElMessage.error(t('skills.dismissFailed'))
  }
}

function openImport(url: string) {
  window.open(url, '_blank')
}

async function showLearnedDetail(skill: any) {
  learnedDetailSkill.value = skill
  learnedDetailVisible.value = true
}

async function improveSkill(id: number) {
  improving.value = true
  try {
    const { data } = await skillsApi.improveSkill(id)
    if (data.status === 'kept') {
      ElMessage.info(data.message)
    } else if (data.status === 'improved') {
      ElMessage.success(t('skills.skillImproved') + ' v' + data.skill?.version)
      loadLearnedSkills()
    }
  } catch (e: any) {
    ElMessage.error(e.response?.data?.error || t('skills.improveFailed'))
  } finally {
    improving.value = false
  }
}

async function deactivateSkill(id: number) {
  try {
    const currentStatus = learnedDetailSkill.value?.status || 'active'
    await skillsApi.patchSkill(id, {
      field: 'status',
      old_value: currentStatus,
      new_value: 'inactive',
    })
    ElMessage.success(t('skills.deactivated'))
    learnedDetailVisible.value = false
    loadLearnedSkills()
  } catch {
    ElMessage.error(t('skills.deactivateFailed'))
  }
}

async function reactivateSkill(id: number) {
  try {
    await skillsApi.patchSkill(id, {
      field: 'status',
      old_value: 'needs_improvement',
      new_value: 'active',
    })
    ElMessage.success(t('skills.reactivated'))
    learnedDetailVisible.value = false
    loadLearnedSkills()
  } catch {
    ElMessage.error(t('skills.reactivateFailed'))
  }
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

.header-actions {
  display: flex;
  gap: 8px;
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

.learned-actions, .scanned-actions {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
  flex-wrap: wrap;
}

.skill-section, .pattern-section, .suggestion-section {
  margin-bottom: 32px;
}

.skill-section h3, .pattern-section h3, .suggestion-section h3 {
  font-size: 15px;
  font-weight: 600;
  color: var(--cm-text);
  margin-bottom: 12px;
  padding-bottom: 8px;
  border-bottom: 1px solid var(--cm-border);
}

.skill-grid, .pattern-grid, .suggestion-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 12px;
}

.skill-card, .pattern-card, .suggestion-card {
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

.skill-card:hover, .pattern-card:hover {
  border-color: var(--cm-primary);
  box-shadow: 0 2px 8px rgba(var(--cm-primary-rgb), 0.1);
}

.skill-icon, .pattern-icon, .sug-icon {
  font-size: 24px;
  flex-shrink: 0;
}

.skill-info, .pattern-info, .sug-info {
  flex: 1;
  min-width: 0;
}

.skill-name, .sug-title {
  font-weight: 600;
  font-size: 14px;
  color: var(--cm-text);
}

.skill-desc, .pattern-seq, .sug-desc {
  font-size: 12px;
  color: var(--cm-text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.pattern-meta {
  font-size: 11px;
  color: var(--cm-text-muted);
  margin-top: 2px;
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

.success-rate {
  font-size: 11px;
  padding: 2px 6px;
  border-radius: 4px;
  background: rgba(16, 185, 129, 0.1);
  color: #10b981;
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

.sug-actions {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
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

.steps-list {
  margin-top: 6px;
}

.step-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 6px;
  font-size: 13px;
}

.step-num {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: rgba(var(--cm-primary-rgb), 0.1);
  color: var(--cm-primary);
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
}

.pitfalls-list {
  margin-top: 6px;
}

.detail-actions {
  margin-top: 16px;
  display: flex;
  gap: 8px;
}

@media (max-width: 768px) {
  .skill-grid, .pattern-grid, .suggestion-grid { grid-template-columns: 1fr; }
  .page-header { flex-direction: column; gap: 12px; align-items: flex-start; }
}
</style>
