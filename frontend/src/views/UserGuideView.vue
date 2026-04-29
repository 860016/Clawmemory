<template>
  <div class="user-guide-page">
    <div class="page-header">
      <h1>📖 {{ $t('guide.title') }}</h1>
      <p class="page-desc">{{ $t('guide.subtitle') }}</p>
    </div>

    <div class="guide-grid">
      <div class="guide-card">
        <div class="card-icon">🚀</div>
        <h2>{{ $t('guide.gettingStarted') }}</h2>
        <div class="card-content">
          <ol class="steps-list">
            <li class="step-item">
              <span class="step-number">1</span>
              <div class="step-text">{{ $t('guide.step1') }}</div>
            </li>
            <li class="step-item">
              <span class="step-number">2</span>
              <div class="step-text">{{ $t('guide.step2') }}</div>
            </li>
            <li class="step-item">
              <span class="step-number">3</span>
              <div class="step-text">{{ $t('guide.step3') }}</div>
            </li>
            <li class="step-item">
              <span class="step-number">4</span>
              <div class="step-text">{{ $t('guide.step4') }}</div>
            </li>
          </ol>
        </div>
      </div>

      <div class="guide-card highlight">
        <div class="card-icon">🔮</div>
        <h2>{{ $t('guide.chromadbGuide') }}</h2>
        <div class="card-content">
          <p class="guide-intro">{{ $t('guide.chromadbIntro') }}</p>
          <div class="install-steps">
            <div class="install-step">
              <div class="step-header">
                <span class="step-badge">{{ $t('guide.method1') }}</span>
                <span class="step-label">{{ $t('guide.oneClickInstall') }}</span>
              </div>
              <p class="step-desc">{{ $t('guide.oneClickDesc') }}</p>
            </div>
            <div class="install-step">
              <div class="step-header">
                <span class="step-badge">{{ $t('guide.method2') }}</span>
                <span class="step-label">{{ $t('guide.manualInstall') }}</span>
              </div>
              <div class="code-block">
                <code>pip install chromadb</code>
                <el-button text size="small" @click="copyCode('pip install chromadb')">📋 {{ $t('common.copy') }}</el-button>
              </div>
              <div class="code-block">
                <code>chroma run --host 0.0.0.0 --port 8000</code>
                <el-button text size="small" @click="copyCode('chroma run --host 0.0.0.0 --port 8000')">📋 {{ $t('common.copy') }}</el-button>
              </div>
            </div>
          </div>
          <div class="chromadb-features">
            <h3>{{ $t('guide.features') }}</h3>
            <ul class="feature-list">
              <li><span class="check">✓</span> {{ $t('guide.feature1') }}</li>
              <li><span class="check">✓</span> {{ $t('guide.feature2') }}</li>
              <li><span class="check">✓</span> {{ $t('guide.feature3') }}</li>
              <li><span class="check">✓</span> {{ $t('guide.feature4') }}</li>
            </ul>
          </div>
        </div>
      </div>

      <div class="guide-card">
        <div class="card-icon">🧠</div>
        <h2>{{ $t('guide.memoryLayers') }}</h2>
        <div class="card-content">
          <div class="layer-explain">
            <div class="layer-item">
              <div class="layer-badge preference">{{ $t('memories.preference') }}</div>
              <p>{{ $t('guide.layerPreference') }}</p>
            </div>
            <div class="layer-item">
              <div class="layer-badge knowledge">{{ $t('memories.knowledge') }}</div>
              <p>{{ $t('guide.layerKnowledge') }}</p>
            </div>
            <div class="layer-item">
              <div class="layer-badge short_term">{{ $t('memories.shortTerm') }}</div>
              <p>{{ $t('guide.layerShortTerm') }}</p>
            </div>
            <div class="layer-item">
              <div class="layer-badge private">{{ $t('memories.private') }}</div>
              <p>{{ $t('guide.layerPrivate') }}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="guide-card">
        <div class="card-icon">🎯</div>
        <h2>{{ $t('guide.smartLoadTitle') }}</h2>
        <div class="card-content">
          <p class="guide-intro">{{ $t('guide.smartLoadIntro') }}</p>
          <div class="load-levels">
            <div class="level-item">
              <span class="level-badge summary">summary</span>
              <span class="level-desc">{{ $t('guide.smartLoadSummary') }}</span>
            </div>
            <div class="level-item">
              <span class="level-badge standard">standard</span>
              <span class="level-desc">{{ $t('guide.smartLoadStandard') }}</span>
            </div>
            <div class="level-item">
              <span class="level-badge full">full</span>
              <span class="level-desc">{{ $t('guide.smartLoadFull') }}</span>
            </div>
          </div>
          <ul class="feature-list">
            <li><span class="check">✓</span> {{ $t('guide.smartLoadFeature1') }}</li>
            <li><span class="check">✓</span> {{ $t('guide.smartLoadFeature2') }}</li>
            <li><span class="check">✓</span> {{ $t('guide.smartLoadFeature3') }}</li>
          </ul>
        </div>
      </div>

      <div class="guide-card">
        <div class="card-icon">📌</div>
        <h2>{{ $t('guide.reinforceTitle') }}</h2>
        <div class="card-content">
          <p class="guide-intro">{{ $t('guide.reinforceIntro') }}</p>
          <ul class="feature-list">
            <li><span class="check">✓</span> {{ $t('guide.reinforceFeature1') }}</li>
            <li><span class="check">✓</span> {{ $t('guide.reinforceFeature2') }}</li>
            <li><span class="check">✓</span> {{ $t('guide.reinforceFeature3') }}</li>
            <li><span class="check">✓</span> {{ $t('guide.reinforceFeature4') }}</li>
          </ul>
        </div>
      </div>

      <div class="guide-card">
        <div class="card-icon">🕸️</div>
        <h2>{{ $t('guide.knowledgeGraph') }}</h2>
        <div class="card-content">
          <p class="guide-intro">{{ $t('guide.graphIntro') }}</p>
          <ul class="feature-list">
            <li><span class="check">✓</span> {{ $t('guide.graphFeature1') }}</li>
            <li><span class="check">✓</span> {{ $t('guide.graphFeature2') }}</li>
            <li><span class="check">✓</span> {{ $t('guide.graphFeature3') }}</li>
          </ul>
        </div>
      </div>

      <div class="guide-card">
        <div class="card-icon">📖</div>
        <h2>{{ $t('guide.wikiGuide') }}</h2>
        <div class="card-content">
          <p class="guide-intro">{{ $t('guide.wikiIntro') }}</p>
          <ul class="feature-list">
            <li><span class="check">✓</span> {{ $t('guide.wikiFeature1') }}</li>
            <li><span class="check">✓</span> {{ $t('guide.wikiFeature2') }}</li>
            <li><span class="check">✓</span> {{ $t('guide.wikiFeature3') }}</li>
          </ul>
        </div>
      </div>

      <div class="guide-card full-width pro-card">
        <div class="card-icon">💎</div>
        <h2>{{ $t('guide.proFeatures') }}</h2>
        <div class="card-content">
          <div class="pro-sections">
            <div class="pro-section">
              <h3>{{ $t('guide.proLocalTitle') }}</h3>
              <p class="section-desc">{{ $t('guide.proLocalDesc') }}</p>
              <ul class="feature-list">
                <li><span class="check">✓</span> {{ $t('guide.proDecayTitle') }}: {{ $t('guide.proDecayDesc') }}</li>
                <li><span class="check">✓</span> {{ $t('guide.proConflictTitle') }}: {{ $t('guide.proConflictDesc') }}</li>
                <li><span class="check">✓</span> {{ $t('guide.proRouterTitle') }}: {{ $t('guide.proRouterDesc') }}</li>
              </ul>
            </div>
            <div class="pro-section">
              <h3>{{ $t('guide.proCloudTitle') }}</h3>
              <p class="section-desc">{{ $t('guide.proCloudDesc') }}</p>
              <ul class="feature-list">
                <li><span class="check">✓</span> {{ $t('guide.proExtractTitle') }}: {{ $t('guide.proExtractDesc') }}</li>
                <li><span class="check">✓</span> {{ $t('guide.proTrendTitle') }}: {{ $t('guide.proTrendDesc') }}</li>
                <li><span class="check">✓</span> {{ $t('guide.proReportTitle') }}: {{ $t('guide.proReportDesc') }}</li>
              </ul>
            </div>
          </div>
          <div class="pro-models">
            <h3>{{ $t('guide.proModelsTitle') }}</h3>
            <div class="model-grid">
              <div class="model-item">OpenAI: GPT-4o / GPT-4-turbo</div>
              <div class="model-item">Anthropic: Claude 3.5 Sonnet</div>
              <div class="model-item">DeepSeek: Chat / Reasoner</div>
              <div class="model-item">Moonshot: v1-8k/32k/128k</div>
              <div class="model-item">Zhipu: GLM-4 / Flash / Air</div>
              <div class="model-item">Qwen: Max / Plus / Turbo</div>
            </div>
          </div>
          <el-button type="primary" style="margin-top: 16px" @click="$router.push('/pro')">
            {{ $t('nav.pro') }} →
          </el-button>
        </div>
      </div>

      <div class="guide-card full-width">
        <div class="card-icon">📋</div>
        <h2>{{ $t('guide.dailyReportTitle') }}</h2>
        <div class="card-content">
          <p class="guide-intro">{{ $t('guide.dailyReportIntro') }}</p>
          <ul class="feature-list">
            <li><span class="check">✓</span> {{ $t('guide.dailyReportFeature1') }}</li>
            <li><span class="check">✓</span> {{ $t('guide.dailyReportFeature2') }}</li>
            <li><span class="check">✓</span> {{ $t('guide.dailyReportFeature3') }}</li>
            <li><span class="check">✓</span> {{ $t('guide.dailyReportFeature4') }}</li>
          </ul>
        </div>
      </div>

      <div class="guide-card full-width">
        <div class="card-icon">✨</div>
        <h2>{{ $t('guide.skillsTitle') }}</h2>
        <div class="card-content">
          <p class="guide-intro">{{ $t('guide.skillsIntro') }}</p>
          <div class="install-steps">
            <div class="install-step">
              <div class="step-header">
                <span class="step-badge">{{ $t('guide.skillsScanLabel') }}</span>
                <span class="step-label">{{ $t('guide.skillsScanTitle') }}</span>
              </div>
              <p class="step-desc">{{ $t('guide.skillsScanDesc') }}</p>
            </div>
            <div class="install-step">
              <div class="step-header">
                <span class="step-badge">{{ $t('guide.skillsInstallLabel') }}</span>
                <span class="step-label">{{ $t('guide.skillsInstallTitle') }}</span>
              </div>
              <p class="step-desc">{{ $t('guide.skillsInstallDesc') }}</p>
            </div>
          </div>
          <div class="chromadb-features">
            <h3>{{ $t('guide.skillsPathsTitle') }}</h3>
            <ul class="feature-list">
              <li><span class="check">📁</span> ~/.openclaw/skills/</li>
              <li><span class="check">📁</span> ~/.clawmemory/skills/</li>
              <li><span class="check">📁</span> {exe_dir}/skills/</li>
              <li><span class="check">📁</span> {work_dir}/skills/</li>
            </ul>
          </div>
        </div>
      </div>

      <div class="guide-card full-width">
        <div class="card-icon">📥</div>
        <h2>{{ $t('guide.memoryImportTitle') }}</h2>
        <div class="card-content">
          <p class="guide-intro">{{ $t('guide.memoryImportIntro') }}</p>
          <ul class="feature-list">
            <li><span class="check">✓</span> {{ $t('guide.memoryImportFeature1') }}</li>
            <li><span class="check">✓</span> {{ $t('guide.memoryImportFeature2') }}</li>
            <li><span class="check">✓</span> {{ $t('guide.memoryImportFeature3') }}</li>
          </ul>
        </div>
      </div>

      <div class="guide-card full-width">
        <div class="card-icon">⚙️</div>
        <h2>{{ $t('guide.proFallbackTitle') }}</h2>
        <div class="card-content">
          <p class="guide-intro">{{ $t('guide.proFallbackIntro') }}</p>
          <ul class="feature-list">
            <li><span class="check">✓</span> {{ $t('guide.proFallbackFeature1') }}</li>
            <li><span class="check">✓</span> {{ $t('guide.proFallbackFeature2') }}</li>
            <li><span class="check">✓</span> {{ $t('guide.proFallbackFeature3') }}</li>
            <li><span class="check">✓</span> {{ $t('guide.proFallbackFeature4') }}</li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'

const { t } = useI18n()

function copyCode(code: string) {
  navigator.clipboard.writeText(code)
  ElMessage.success(t('common.copied'))
}
</script>

<style scoped>
.user-guide-page {
  padding: 28px;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;
}

.page-header h1 {
  font-size: 24px;
  font-weight: 700;
  color: var(--cm-text);
  margin: 0 0 8px;
}

.page-desc {
  font-size: 14px;
  color: var(--cm-text-muted);
  margin: 0;
}

.guide-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.guide-card {
  background: var(--cm-bg-secondary);
  border: 1px solid var(--cm-border);
  border-radius: 16px;
  padding: 24px;
  transition: all 0.2s ease;
}

.guide-card:hover {
  border-color: rgba(16, 185, 129, 0.3);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.guide-card.full-width {
  grid-column: 1 / -1;
}

.guide-card.highlight {
  background: linear-gradient(135deg, var(--cm-bg-secondary) 0%, rgba(139, 92, 246, 0.05) 100%);
  border-color: rgba(139, 92, 246, 0.2);
}

.guide-card.pro-card {
  background: linear-gradient(135deg, var(--cm-bg-secondary) 0%, rgba(16, 185, 129, 0.05) 100%);
  border-color: rgba(16, 185, 129, 0.2);
}

.card-icon {
  font-size: 32px;
  margin-bottom: 12px;
}

.guide-card h2 {
  font-size: 18px;
  font-weight: 600;
  color: var(--cm-text);
  margin: 0 0 16px;
}

.card-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.guide-intro {
  font-size: 14px;
  color: var(--cm-text-secondary);
  line-height: 1.6;
  margin: 0;
}

.steps-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.step-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.step-number {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: rgba(16, 185, 129, 0.15);
  color: #10B981;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
  flex-shrink: 0;
}

.step-text {
  font-size: 14px;
  color: var(--cm-text-secondary);
  line-height: 1.6;
  padding-top: 4px;
}

.install-steps {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.install-step {
  padding: 12px;
  background: var(--cm-bg);
  border-radius: 10px;
  border: 1px solid var(--cm-border);
}

.step-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}

.step-badge {
  padding: 2px 8px;
  border-radius: 4px;
  background: rgba(139, 92, 246, 0.15);
  color: #8B5CF6;
  font-size: 11px;
  font-weight: 600;
}

.step-label {
  font-size: 14px;
  font-weight: 600;
  color: var(--cm-text);
}

.step-desc {
  font-size: 13px;
  color: var(--cm-text-muted);
  margin: 0;
}

.code-block {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 10px 12px;
  background: var(--cm-bg-secondary);
  border-radius: 8px;
  border: 1px solid var(--cm-border);
  margin-top: 8px;
}

.code-block code {
  font-family: 'Cascadia Code', 'Fira Code', monospace;
  font-size: 13px;
  color: #10B981;
}

.chromadb-features h3 {
  font-size: 14px;
  font-weight: 600;
  color: var(--cm-text);
  margin: 0 0 10px;
}

.feature-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.feature-list li {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--cm-text-secondary);
}

.check {
  color: #10B981;
  font-weight: 700;
}

.layer-explain {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.layer-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.layer-badge {
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
  flex-shrink: 0;
}

.layer-badge.preference {
  background: rgba(16, 185, 129, 0.15);
  color: #10B981;
}

.layer-badge.knowledge {
  background: rgba(6, 182, 212, 0.15);
  color: #06b6d4;
}

.layer-badge.short_term {
  background: rgba(255, 193, 7, 0.15);
  color: #ffc107;
}

.layer-badge.private {
  background: rgba(233, 30, 99, 0.15);
  color: #e91e63;
}

.layer-item p {
  font-size: 13px;
  color: var(--cm-text-secondary);
  line-height: 1.5;
  margin: 0;
}

.load-levels {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.level-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.level-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  min-width: 70px;
  text-align: center;
}

.level-badge.summary {
  background: rgba(230, 162, 60, 0.15);
  color: #e6a23c;
}

.level-badge.standard {
  background: rgba(6, 182, 212, 0.15);
  color: #06b6d4;
}

.level-badge.full {
  background: rgba(103, 194, 79, 0.15);
  color: #67c23a;
}

.level-desc {
  font-size: 13px;
  color: var(--cm-text-secondary);
}

.pro-sections {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.pro-section h3 {
  font-size: 14px;
  font-weight: 600;
  color: var(--cm-text);
  margin: 0 0 8px;
}

.section-desc {
  font-size: 12px;
  color: var(--cm-text-muted);
  margin: 0 0 12px;
}

.pro-models {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--cm-border);
}

.pro-models h3 {
  font-size: 14px;
  font-weight: 600;
  color: var(--cm-text);
  margin: 0 0 12px;
}

.model-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.model-item {
  padding: 6px 10px;
  background: var(--cm-bg);
  border-radius: 6px;
  font-size: 12px;
  color: var(--cm-text-secondary);
  border: 1px solid var(--cm-border);
}

@media (max-width: 768px) {
  .user-guide-page {
    padding: 16px;
  }
  .guide-grid {
    grid-template-columns: 1fr;
  }
  .page-header h1 {
    font-size: 20px;
  }
  .guide-card {
    padding: 18px;
  }
  .guide-card h2 {
    font-size: 16px;
  }
  .pro-sections {
    grid-template-columns: 1fr;
  }
  .model-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 480px) {
  .user-guide-page {
    padding: 12px;
  }
  .page-header h1 {
    font-size: 18px;
  }
  .guide-card {
    padding: 16px;
    border-radius: 12px;
  }
  .card-icon {
    font-size: 28px;
  }
  .guide-card h2 {
    font-size: 15px;
  }
  .step-number {
    width: 24px;
    height: 24px;
    font-size: 12px;
  }
  .step-text {
    font-size: 13px;
  }
  .code-block {
    flex-direction: column;
    align-items: stretch;
  }
  .code-block .el-button {
    justify-content: center;
  }
  .model-grid {
    grid-template-columns: 1fr;
  }
}
</style>
