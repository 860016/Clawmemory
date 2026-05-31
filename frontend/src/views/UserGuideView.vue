<template>
  <div class="doc-page">
    <div class="doc-hero">
      <div class="hero-bg"></div>
      <div class="hero-content">
        <div class="hero-badge">Open Source</div>
        <h1>ClawMemory</h1>
        <p class="hero-subtitle">{{ $t('guide.heroSubtitle') }}</p>
        <div class="hero-actions">
          <el-button type="primary" size="large" @click="scrollTo('quick-start')">
            <el-icon><CaretRight /></el-icon> {{ $t('guide.quickStart') }}
          </el-button>
          <el-button size="large" @click="scrollTo('memory')">
            <el-icon><Grid /></el-icon> {{ $t('guide.allFeatures') }}
          </el-button>
        </div>
      </div>
    </div>

    <div class="doc-tabs-bar">
      <div class="doc-tabs-scroll">
        <button
          v-for="section in sections"
          :key="section.id"
          :class="['doc-tab', { active: activeSection === section.id }]"
          @click="scrollTo(section.id)"
        >
          <span class="tab-icon">{{ section.icon }}</span>
          <span class="tab-text">{{ $t(section.label) }}</span>
        </button>
      </div>
    </div>

    <div class="doc-body">
      <main class="doc-content">
        <section id="quick-start" class="doc-section">
          <div class="section-header">
            <span class="section-icon">🚀</span>
            <h2>{{ $t('guide.gettingStarted') }}</h2>
          </div>
          <div class="section-body">
            <div class="steps-timeline">
              <div class="timeline-item" v-for="(step, i) in startSteps" :key="i">
                <div class="timeline-dot">{{ i + 1 }}</div>
                <div class="timeline-content">
                  <h3>{{ $t(step.title) }}</h3>
                  <p>{{ $t(step.desc) }}</p>
                  <div v-if="step.tip" class="timeline-tip">
                    <el-icon><InfoFilled /></el-icon>
                    <span>{{ $t(step.tip) }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section id="memory" class="doc-section">
          <div class="section-header">
            <span class="section-icon">🧠</span>
            <h2>{{ $t('guide.memorySystem') }}</h2>
          </div>
          <div class="section-body">
            <p class="section-intro">{{ $t('guide.memoryIntro') }}</p>
            <div class="layer-cards">
              <div class="layer-card" v-for="layer in memoryLayers" :key="layer.key" :class="layer.key">
                <div class="layer-header">
                  <span class="layer-emoji">{{ layer.icon }}</span>
                  <span class="layer-name">{{ $t(layer.label) }}</span>
                </div>
                <p class="layer-desc">{{ $t(layer.desc) }}</p>
                <div class="layer-tags">
                  <span class="layer-tag" v-for="tag in layer.tags" :key="tag">{{ tag }}</span>
                </div>
              </div>
            </div>
            <div class="feature-grid">
              <div class="mini-feature">
                <span class="mf-icon">🔍</span>
                <div>
                  <strong>{{ $t('guide.searchTypes') }}</strong>
                  <p>{{ $t('guide.searchTypesDesc') }}</p>
                </div>
              </div>
              <div class="mini-feature">
                <span class="mf-icon">📌</span>
                <div>
                  <strong>{{ $t('guide.reinforceTitle') }}</strong>
                  <p>{{ $t('guide.reinforceDesc') }}</p>
                </div>
              </div>
              <div class="mini-feature">
                <span class="mf-icon">🎯</span>
                <div>
                  <strong>{{ $t('guide.smartLoadTitle') }}</strong>
                  <p>{{ $t('guide.smartLoadDesc') }}</p>
                </div>
              </div>
              <div class="mini-feature">
                <span class="mf-icon">🔒</span>
                <div>
                  <strong>{{ $t('guide.encryptTitle') }}</strong>
                  <p>{{ $t('guide.encryptDesc') }}</p>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section id="knowledge" class="doc-section">
          <div class="section-header">
            <span class="section-icon">🕸️</span>
            <h2>{{ $t('guide.knowledgeGraph') }}</h2>
          </div>
          <div class="section-body">
            <p class="section-intro">{{ $t('guide.graphIntroNew') }}</p>
            <div class="graph-showcase">
              <div class="graph-feature">
                <div class="gf-icon">🔵</div>
                <h4>{{ $t('guide.entityMgmt') }}</h4>
                <p>{{ $t('guide.entityMgmtDesc') }}</p>
              </div>
              <div class="graph-feature">
                <div class="gf-icon">🔗</div>
                <h4>{{ $t('guide.relationMgmt') }}</h4>
                <p>{{ $t('guide.relationMgmtDesc') }}</p>
              </div>
              <div class="graph-feature">
                <div class="gf-icon">🌐</div>
                <h4>{{ $t('guide.graphVisual') }}</h4>
                <p>{{ $t('guide.graphVisualDesc') }}</p>
              </div>
              <div class="graph-feature">
                <div class="gf-icon">🤖</div>
                <h4>{{ $t('guide.autoGraph') }}</h4>
                <p>{{ $t('guide.autoGraphDesc') }}</p>
              </div>
            </div>
          </div>
        </section>

        <section id="wiki" class="doc-section">
          <div class="section-header">
            <span class="section-icon">📖</span>
            <h2>{{ $t('guide.wikiGuide') }}</h2>
          </div>
          <div class="section-body">
            <p class="section-intro">{{ $t('guide.wikiIntroNew') }}</p>
            <div class="wiki-features">
              <div class="wf-item" v-for="f in wikiFeatures" :key="f.icon">
                <span class="wf-icon">{{ f.icon }}</span>
                <div>
                  <strong>{{ $t(f.title) }}</strong>
                  <p>{{ $t(f.desc) }}</p>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section id="projects" class="doc-section">
          <div class="section-header">
            <span class="section-icon">📋</span>
            <h2>{{ $t('guide.projectMgmt') }}</h2>
          </div>
          <div class="section-body">
            <p class="section-intro">{{ $t('guide.projectIntro') }}</p>
            <div class="project-features">
              <div class="pf-card" v-for="f in projectFeatures" :key="f.icon">
                <span class="pf-icon">{{ f.icon }}</span>
                <h4>{{ $t(f.title) }}</h4>
                <p>{{ $t(f.desc) }}</p>
              </div>
            </div>
          </div>
        </section>

        <section id="sharing" class="doc-section">
          <div class="section-header">
            <span class="section-icon">🤝</span>
            <h2>{{ $t('guide.sharingTitle') }}</h2>
          </div>
          <div class="section-body">
            <p class="section-intro">{{ $t('guide.sharingIntro') }}</p>
            <div class="sharing-flow">
              <div class="flow-step">
                <div class="flow-num">1</div>
                <div class="flow-text">{{ $t('guide.sharingStep1') }}</div>
              </div>
              <div class="flow-arrow">→</div>
              <div class="flow-step">
                <div class="flow-num">2</div>
                <div class="flow-text">{{ $t('guide.sharingStep2') }}</div>
              </div>
              <div class="flow-arrow">→</div>
              <div class="flow-step">
                <div class="flow-num">3</div>
                <div class="flow-text">{{ $t('guide.sharingStep3') }}</div>
              </div>
            </div>
            <div class="share-types">
              <div class="share-type">
                <strong>{{ $t('guide.manualShare') }}</strong>
                <p>{{ $t('guide.manualShareDesc') }}</p>
              </div>
              <div class="share-type">
                <strong>{{ $t('guide.autoShare') }}</strong>
                <p>{{ $t('guide.autoShareDesc') }}</p>
              </div>
            </div>
          </div>
        </section>

        <section id="session" class="doc-section">
          <div class="section-header">
            <span class="section-icon">💬</span>
            <h2>{{ $t('guide.sessionMemory') }}</h2>
          </div>
          <div class="section-body">
            <p class="section-intro">{{ $t('guide.sessionIntro') }}</p>
            <div class="session-use-cases">
              <div class="use-case" v-for="uc in sessionUseCases" :key="uc.icon">
                <span>{{ uc.icon }}</span>
                <span>{{ $t(uc.text) }}</span>
              </div>
            </div>
          </div>
        </section>

        <section id="ai" class="doc-section">
          <div class="section-header">
            <span class="section-icon">🤖</span>
            <h2>{{ $t('guide.aiFeatures') }}</h2>
          </div>
          <div class="section-body">
            <p class="section-intro">{{ $t('guide.aiIntro') }}</p>
            <div class="ai-grid">
              <div class="ai-card" v-for="f in aiFeatures" :key="f.icon" :class="{ local: f.local }">
                <div class="ai-card-header">
                  <span class="ai-icon">{{ f.icon }}</span>
                  <span class="ai-badge" v-if="f.local">{{ $t('guide.local') }}</span>
                  <span class="ai-badge cloud" v-else>{{ $t('guide.cloud') }}</span>
                </div>
                <h4>{{ $t(f.title) }}</h4>
                <p>{{ $t(f.desc) }}</p>
              </div>
            </div>
            <div class="models-section">
              <h3>{{ $t('guide.supportedModels') }}</h3>
              <div class="model-grid">
                <div class="model-card" v-for="m in aiModels" :key="m.name">
                  <div class="model-logo">{{ m.logo }}</div>
                  <div class="model-info">
                    <strong>{{ m.name }}</strong>
                    <span>{{ m.models }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section id="chromadb" class="doc-section">
          <div class="section-header">
            <span class="section-icon">🔮</span>
            <h2>{{ $t('guide.chromadbGuide') }}</h2>
          </div>
          <div class="section-body">
            <p class="section-intro">{{ $t('guide.chromadbIntro') }}</p>
            <div class="install-methods">
              <div class="method-card recommended">
                <div class="method-badge">{{ $t('guide.recommended') }}</div>
                <h4>{{ $t('guide.oneClickInstall') }}</h4>
                <p>{{ $t('guide.oneClickDescNew') }}</p>
                <el-button type="primary" @click="$router.push('/')">{{ $t('guide.goDashboard') }}</el-button>
              </div>
              <div class="method-card">
                <h4>{{ $t('guide.manualInstall') }}</h4>
                <p>{{ $t('guide.manualInstallDesc') }}</p>
                <div class="code-block">
                  <code>pip install chromadb</code>
                  <button class="copy-btn" @click="copyCode('pip install chromadb')">
                    <el-icon><CopyDocument /></el-icon>
                  </button>
                </div>
                <div class="code-block">
                  <code>chroma run --host 0.0.0.0 --port 8000</code>
                  <button class="copy-btn" @click="copyCode('chroma run --host 0.0.0.0 --port 8000')">
                    <el-icon><CopyDocument /></el-icon>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section id="data" class="doc-section">
          <div class="section-header">
            <span class="section-icon">💾</span>
            <h2>{{ $t('guide.dataMgmt') }}</h2>
          </div>
          <div class="section-body">
            <div class="data-grid">
              <div class="data-card" v-for="f in dataFeatures" :key="f.icon">
                <span class="data-icon">{{ f.icon }}</span>
                <h4>{{ $t(f.title) }}</h4>
                <p>{{ $t(f.desc) }}</p>
              </div>
            </div>
          </div>
        </section>

        <section id="faq" class="doc-section">
          <div class="section-header">
            <span class="section-icon">❓</span>
            <h2>{{ $t('guide.faqTitle') }}</h2>
          </div>
          <div class="section-body">
            <div class="faq-list">
              <div
                v-for="(faq, i) in faqs"
                :key="i"
                :class="['faq-item', { open: openFaq === i }]"
                @click="openFaq = openFaq === i ? -1 : i"
              >
                <div class="faq-q">
                  <span>{{ $t(faq.q) }}</span>
                  <el-icon class="faq-arrow"><ArrowRight /></el-icon>
                </div>
                <div class="faq-a" v-if="openFaq === i">{{ $t(faq.a) }}</div>
              </div>
            </div>
          </div>
        </section>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage } from 'element-plus'

const { t } = useI18n()
const activeSection = ref('quick-start')
const openFaq = ref(-1)
const isMobile = ref(window.innerWidth < 768)

const sections = [
  { id: 'quick-start', icon: '🚀', label: 'guide.gettingStarted' },
  { id: 'memory', icon: '🧠', label: 'guide.memorySystem' },
  { id: 'knowledge', icon: '🕸️', label: 'guide.knowledgeGraph' },
  { id: 'wiki', icon: '📖', label: 'guide.wikiGuide' },
  { id: 'projects', icon: '📋', label: 'guide.projectMgmt' },
  { id: 'sharing', icon: '🤝', label: 'guide.sharingTitle' },
  { id: 'session', icon: '💬', label: 'guide.sessionMemory' },
  { id: 'ai', icon: '🤖', label: 'guide.aiFeatures' },
  { id: 'chromadb', icon: '🔮', label: 'guide.chromadbGuide' },
  { id: 'data', icon: '💾', label: 'guide.dataMgmt' },
  { id: 'faq', icon: '❓', label: 'guide.faqTitle' },
]

const startSteps = [
  { title: 'guide.step1Title', desc: 'guide.step1Desc', tip: 'guide.step1Tip' },
  { title: 'guide.step2Title', desc: 'guide.step2Desc', tip: '' },
  { title: 'guide.step3Title', desc: 'guide.step3Desc', tip: 'guide.step3Tip' },
  { title: 'guide.step4Title', desc: 'guide.step4Desc', tip: '' },
]

const memoryLayers = [
  { key: 'preference', icon: '⭐', label: 'memories.preference', desc: 'guide.layerPreferenceNew', tags: ['持久', '高优先'] },
  { key: 'knowledge', icon: '📚', label: 'memories.knowledge', desc: 'guide.layerKnowledgeNew', tags: ['持久', '可搜索'] },
  { key: 'short_term', icon: '⚡', label: 'memories.shortTerm', desc: 'guide.layerShortTermNew', tags: ['临时', '自动衰减'] },
  { key: 'private', icon: '🔐', label: 'memories.private', desc: 'guide.layerPrivateNew', tags: ['加密', '仅自己可见'] },
]

const wikiFeatures = [
  { icon: '📝', title: 'guide.wikiEdit', desc: 'guide.wikiEditDesc' },
  { icon: '📂', title: 'guide.wikiCategory', desc: 'guide.wikiCategoryDesc' },
  { icon: '🌳', title: 'guide.wikiTree', desc: 'guide.wikiTreeDesc' },
  { icon: '🤖', title: 'guide.wikiAI', desc: 'guide.wikiAIDesc' },
  { icon: '🔍', title: 'guide.wikiSearch', desc: 'guide.wikiSearchDesc' },
  { icon: '✅', title: 'guide.wikiStatus', desc: 'guide.wikiStatusDesc' },
]

const projectFeatures = [
  { icon: '📁', title: 'guide.projCreate', desc: 'guide.projCreateDesc' },
  { icon: '📝', title: 'guide.projNotes', desc: 'guide.projNotesDesc' },
  { icon: '🔗', title: 'guide.projContext', desc: 'guide.projContextDesc' },
  { icon: '🤖', title: 'guide.projExtract', desc: 'guide.projExtractDesc' },
]

const sessionUseCases = [
  { icon: '🔄', text: 'guide.sessionUC1' },
  { icon: '📋', text: 'guide.sessionUC2' },
  { icon: '🎯', text: 'guide.sessionUC3' },
  { icon: '📊', text: 'guide.sessionUC4' },
]

const aiFeatures = [
  { icon: '📉', title: 'guide.aiDecay', desc: 'guide.aiDecayDesc', local: true },
  { icon: '⚔️', title: 'guide.aiConflict', desc: 'guide.aiConflictDesc', local: true },
  { icon: '🧭', title: 'guide.aiRouter', desc: 'guide.aiRouterDesc', local: true },
  { icon: '🗜️', title: 'guide.aiCompress', desc: 'guide.aiCompressDesc', local: true },
  { icon: '✨', title: 'guide.aiExtract', desc: 'guide.aiExtractDesc', local: false },
  { icon: '📊', title: 'guide.aiReport', desc: 'guide.aiReportDesc', local: false },
  { icon: '🌐', title: 'guide.aiAutoGraph', desc: 'guide.aiAutoGraphDesc', local: false },
  { icon: '🧠', title: 'guide.aiReasoning', desc: 'guide.aiReasoningDesc', local: false },
  { icon: '💡', title: 'guide.aiEvolution', desc: 'guide.aiEvolutionDesc', local: false },
]

const aiModels = [
  { name: 'OpenAI', logo: '🟢', models: 'GPT-4o / GPT-4-turbo / GPT-3.5' },
  { name: 'Anthropic', logo: '🟠', models: 'Claude 3.5 Sonnet / Claude 3' },
  { name: 'DeepSeek', logo: '🔵', models: 'Chat / Reasoner' },
  { name: 'Moonshot', logo: '🌙', models: 'v1-8k / 32k / 128k' },
  { name: 'Zhipu', logo: '🟣', models: 'GLM-4 / Flash / Air' },
  { name: 'Qwen', logo: '🟡', models: 'Max / Plus / Turbo' },
]

const dataFeatures = [
  { icon: '📥', title: 'guide.dataImport', desc: 'guide.dataImportDesc' },
  { icon: '📤', title: 'guide.dataExport', desc: 'guide.dataExportDesc' },
  { icon: '🗄️', title: 'guide.dataBackup', desc: 'guide.dataBackupDesc' },
  { icon: '🧹', title: 'guide.dataDedup', desc: 'guide.dataDedupDesc' },
  { icon: '🏥', title: 'guide.dataHealth', desc: 'guide.dataHealthDesc' },
  { icon: '🔑', title: 'guide.dataAPI', desc: 'guide.dataAPIDesc' },
]

const faqs = [
  { q: 'guide.faq1Q', a: 'guide.faq1A' },
  { q: 'guide.faq2Q', a: 'guide.faq2A' },
  { q: 'guide.faq3Q', a: 'guide.faq3A' },
  { q: 'guide.faq4Q', a: 'guide.faq4A' },
  { q: 'guide.faq5Q', a: 'guide.faq5A' },
  { q: 'guide.faq6Q', a: 'guide.faq6A' },
  { q: 'guide.faq7Q', a: 'guide.faq7A' },
  { q: 'guide.faq8Q', a: 'guide.faq8A' },
]

function scrollTo(id: string) {
  const el = document.getElementById(id)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'start' })
    activeSection.value = id
  }
}

function copyCode(code: string) {
  navigator.clipboard.writeText(code)
  ElMessage.success(t('common.copied'))
}

function onScroll() {
  const sectionEls = document.querySelectorAll('.doc-section')
  let current = 'quick-start'
  sectionEls.forEach(el => {
    const rect = el.getBoundingClientRect()
    if (rect.top <= 120) current = el.id
  })
  activeSection.value = current
}

function onResize() {
  isMobile.value = window.innerWidth < 768
}

onMounted(() => {
  window.addEventListener('scroll', onScroll, { passive: true })
  window.addEventListener('resize', onResize, { passive: true })
})

onUnmounted(() => {
  window.removeEventListener('scroll', onScroll)
  window.removeEventListener('resize', onResize)
})
</script>

<style scoped>
.doc-page {
  min-height: 100vh;
}

.doc-hero {
  position: relative;
  padding: 48px 28px 40px;
  text-align: center;
  overflow: hidden;
}

.hero-bg {
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.08) 0%, rgba(6, 182, 212, 0.06) 50%, rgba(16, 185, 129, 0.04) 100%);
  border-bottom: 1px solid var(--cm-border);
}

.hero-bg::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle at 30% 50%, rgba(16, 185, 129, 0.06) 0%, transparent 50%),
              radial-gradient(circle at 70% 50%, rgba(6, 182, 212, 0.04) 0%, transparent 50%);
  animation: hero-glow 8s ease-in-out infinite alternate;
}

@keyframes hero-glow {
  0% { transform: translate(0, 0); }
  100% { transform: translate(-5%, 3%); }
}

.hero-content {
  position: relative;
  z-index: 1;
}

.hero-badge {
  display: inline-block;
  padding: 4px 14px;
  border-radius: 20px;
  background: rgba(16, 185, 129, 0.15);
  color: #10B981;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 1px;
  text-transform: uppercase;
  margin-bottom: 16px;
}

.doc-hero h1 {
  font-size: 36px;
  font-weight: 800;
  background: linear-gradient(135deg, #10B981, #06b6d4);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0 0 12px;
}

.hero-subtitle {
  font-size: 16px;
  color: var(--cm-text-secondary);
  margin: 0 0 24px;
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
  line-height: 1.6;
}

.hero-actions {
  display: flex;
  gap: 12px;
  justify-content: center;
}

.doc-body {
  max-width: 1100px;
  margin: 0 auto;
  padding: 0 28px 28px;
}

.doc-tabs-bar {
  position: sticky;
  top: 0;
  z-index: 10;
  background: var(--cm-bg);
  border-bottom: 1px solid var(--cm-border);
  padding: 0 28px;
}

.doc-tabs-scroll {
  max-width: 1100px;
  margin: 0 auto;
  display: flex;
  gap: 2px;
  overflow-x: auto;
  padding: 8px 0;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
}

.doc-tabs-scroll::-webkit-scrollbar {
  display: none;
}

.doc-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  border: none;
  background: none;
  cursor: pointer;
  border-radius: 8px;
  font-size: 13px;
  color: var(--cm-text-secondary);
  white-space: nowrap;
  transition: all 0.15s;
}

.doc-tab:hover {
  background: var(--cm-bg-secondary);
  color: var(--cm-text);
}

.doc-tab.active {
  background: rgba(16, 185, 129, 0.1);
  color: #10B981;
  font-weight: 600;
}

.tab-icon {
  font-size: 14px;
}

.tab-text {
  white-space: nowrap;
}

.doc-content {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

.doc-section {
  scroll-margin-top: 80px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--cm-border);
}

.section-icon {
  font-size: 28px;
}

.section-header h2 {
  font-size: 22px;
  font-weight: 700;
  color: var(--cm-text);
  margin: 0;
}

.section-intro {
  font-size: 15px;
  color: var(--cm-text-secondary);
  line-height: 1.7;
  margin: 0 0 20px;
}

.steps-timeline {
  position: relative;
  padding-left: 32px;
}

.steps-timeline::before {
  content: '';
  position: absolute;
  left: 14px;
  top: 16px;
  bottom: 16px;
  width: 2px;
  background: linear-gradient(to bottom, #10B981, rgba(16, 185, 129, 0.2));
  border-radius: 1px;
}

.timeline-item {
  position: relative;
  margin-bottom: 24px;
}

.timeline-item:last-child {
  margin-bottom: 0;
}

.timeline-dot {
  position: absolute;
  left: -32px;
  top: 4px;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #10B981;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
  box-shadow: 0 0 0 4px rgba(16, 185, 129, 0.15);
}

.timeline-content h3 {
  font-size: 15px;
  font-weight: 600;
  color: var(--cm-text);
  margin: 0 0 6px;
}

.timeline-content p {
  font-size: 14px;
  color: var(--cm-text-secondary);
  line-height: 1.6;
  margin: 0;
}

.timeline-tip {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  margin-top: 8px;
  padding: 8px 12px;
  background: rgba(16, 185, 129, 0.06);
  border: 1px solid rgba(16, 185, 129, 0.15);
  border-radius: 8px;
  font-size: 12px;
  color: var(--cm-text-secondary);
  line-height: 1.5;
}

.timeline-tip .el-icon {
  color: #10B981;
  margin-top: 2px;
  flex-shrink: 0;
}

.layer-cards {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14px;
  margin-bottom: 24px;
}

.layer-card {
  padding: 18px;
  border-radius: 12px;
  border: 1px solid var(--cm-border);
  background: var(--cm-bg-secondary);
  transition: all 0.2s;
}

.layer-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.layer-card.preference { border-left: 3px solid #10B981; }
.layer-card.knowledge { border-left: 3px solid #06b6d4; }
.layer-card.short_term { border-left: 3px solid #ffc107; }
.layer-card.private { border-left: 3px solid #e91e63; }

.layer-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.layer-emoji {
  font-size: 20px;
}

.layer-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--cm-text);
}

.layer-desc {
  font-size: 13px;
  color: var(--cm-text-secondary);
  line-height: 1.5;
  margin: 0 0 10px;
}

.layer-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.layer-tag {
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--cm-bg);
  border: 1px solid var(--cm-border);
  font-size: 11px;
  color: var(--cm-text-muted);
}

.feature-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14px;
}

.mini-feature {
  display: flex;
  gap: 12px;
  padding: 14px;
  border-radius: 10px;
  background: var(--cm-bg-secondary);
  border: 1px solid var(--cm-border);
}

.mf-icon {
  font-size: 24px;
  flex-shrink: 0;
}

.mini-feature strong {
  font-size: 14px;
  color: var(--cm-text);
  display: block;
  margin-bottom: 4px;
}

.mini-feature p {
  font-size: 12px;
  color: var(--cm-text-secondary);
  margin: 0;
  line-height: 1.5;
}

.graph-showcase {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
}

.graph-feature {
  padding: 18px;
  border-radius: 12px;
  background: var(--cm-bg-secondary);
  border: 1px solid var(--cm-border);
  text-align: center;
  transition: all 0.2s;
}

.graph-feature:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.gf-icon {
  font-size: 32px;
  margin-bottom: 8px;
}

.graph-feature h4 {
  font-size: 14px;
  font-weight: 600;
  color: var(--cm-text);
  margin: 0 0 6px;
}

.graph-feature p {
  font-size: 12px;
  color: var(--cm-text-secondary);
  line-height: 1.5;
  margin: 0;
}

.wiki-features {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
}

.wf-item {
  display: flex;
  gap: 10px;
  padding: 14px;
  border-radius: 10px;
  background: var(--cm-bg-secondary);
  border: 1px solid var(--cm-border);
}

.wf-icon {
  font-size: 22px;
  flex-shrink: 0;
}

.wf-item strong {
  font-size: 14px;
  color: var(--cm-text);
  display: block;
  margin-bottom: 4px;
}

.wf-item p {
  font-size: 12px;
  color: var(--cm-text-secondary);
  margin: 0;
  line-height: 1.5;
}

.project-features {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
}

.pf-card {
  padding: 18px;
  border-radius: 12px;
  background: var(--cm-bg-secondary);
  border: 1px solid var(--cm-border);
  text-align: center;
  transition: all 0.2s;
}

.pf-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.pf-icon {
  font-size: 28px;
  display: block;
  margin-bottom: 8px;
}

.pf-card h4 {
  font-size: 14px;
  font-weight: 600;
  color: var(--cm-text);
  margin: 0 0 6px;
}

.pf-card p {
  font-size: 12px;
  color: var(--cm-text-secondary);
  line-height: 1.5;
  margin: 0;
}

.sharing-flow {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 16px;
  margin-bottom: 24px;
  padding: 24px;
  background: var(--cm-bg-secondary);
  border-radius: 12px;
  border: 1px solid var(--cm-border);
}

.flow-step {
  text-align: center;
}

.flow-num {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #10B981;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 700;
  margin: 0 auto 8px;
}

.flow-text {
  font-size: 13px;
  color: var(--cm-text);
  font-weight: 500;
}

.flow-arrow {
  font-size: 20px;
  color: #10B981;
  font-weight: 700;
}

.share-types {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 14px;
}

.share-type {
  padding: 18px;
  border-radius: 12px;
  background: var(--cm-bg-secondary);
  border: 1px solid var(--cm-border);
}

.share-type strong {
  font-size: 15px;
  color: var(--cm-text);
  display: block;
  margin-bottom: 6px;
}

.share-type p {
  font-size: 13px;
  color: var(--cm-text-secondary);
  line-height: 1.5;
  margin: 0;
}

.session-use-cases {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.use-case {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 18px;
  border-radius: 10px;
  background: var(--cm-bg-secondary);
  border: 1px solid var(--cm-border);
  font-size: 14px;
  color: var(--cm-text-secondary);
}

.use-case span:first-child {
  font-size: 20px;
}

.ai-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
  margin-bottom: 24px;
}

.ai-card {
  padding: 18px;
  border-radius: 12px;
  background: var(--cm-bg-secondary);
  border: 1px solid var(--cm-border);
  transition: all 0.2s;
}

.ai-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.ai-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}

.ai-icon {
  font-size: 24px;
}

.ai-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 600;
  background: rgba(16, 185, 129, 0.15);
  color: #10B981;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.ai-badge.cloud {
  background: rgba(6, 182, 212, 0.15);
  color: #06b6d4;
}

.ai-card h4 {
  font-size: 14px;
  font-weight: 600;
  color: var(--cm-text);
  margin: 0 0 6px;
}

.ai-card p {
  font-size: 12px;
  color: var(--cm-text-secondary);
  line-height: 1.5;
  margin: 0;
}

.models-section {
  padding: 20px;
  border-radius: 12px;
  background: var(--cm-bg-secondary);
  border: 1px solid var(--cm-border);
}

.models-section h3 {
  font-size: 16px;
  font-weight: 600;
  color: var(--cm-text);
  margin: 0 0 16px;
}

.model-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 10px;
}

.model-card {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-radius: 8px;
  background: var(--cm-bg);
  border: 1px solid var(--cm-border);
}

.model-logo {
  font-size: 20px;
  flex-shrink: 0;
}

.model-info {
  display: flex;
  flex-direction: column;
}

.model-info strong {
  font-size: 13px;
  color: var(--cm-text);
}

.model-info span {
  font-size: 11px;
  color: var(--cm-text-muted);
}

.install-methods {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.method-card {
  padding: 20px;
  border-radius: 12px;
  background: var(--cm-bg-secondary);
  border: 1px solid var(--cm-border);
  position: relative;
}

.method-card.recommended {
  border-color: rgba(16, 185, 129, 0.3);
  background: linear-gradient(135deg, var(--cm-bg-secondary), rgba(16, 185, 129, 0.04));
}

.method-badge {
  position: absolute;
  top: -8px;
  right: 16px;
  padding: 2px 10px;
  border-radius: 4px;
  background: #10B981;
  color: white;
  font-size: 11px;
  font-weight: 600;
}

.method-card h4 {
  font-size: 16px;
  font-weight: 600;
  color: var(--cm-text);
  margin: 0 0 8px;
}

.method-card p {
  font-size: 13px;
  color: var(--cm-text-secondary);
  line-height: 1.5;
  margin: 0 0 12px;
}

.code-block {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 10px 14px;
  background: var(--cm-bg);
  border-radius: 8px;
  border: 1px solid var(--cm-border);
  margin-bottom: 8px;
}

.code-block code {
  font-family: 'Cascadia Code', 'Fira Code', monospace;
  font-size: 13px;
  color: #10B981;
}

.copy-btn {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--cm-text-muted);
  padding: 4px;
  border-radius: 4px;
  transition: all 0.15s;
}

.copy-btn:hover {
  color: #10B981;
  background: rgba(16, 185, 129, 0.1);
}

.data-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
}

.data-card {
  padding: 18px;
  border-radius: 12px;
  background: var(--cm-bg-secondary);
  border: 1px solid var(--cm-border);
  text-align: center;
  transition: all 0.2s;
}

.data-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
}

.data-icon {
  font-size: 28px;
  display: block;
  margin-bottom: 8px;
}

.data-card h4 {
  font-size: 14px;
  font-weight: 600;
  color: var(--cm-text);
  margin: 0 0 6px;
}

.data-card p {
  font-size: 12px;
  color: var(--cm-text-secondary);
  line-height: 1.5;
  margin: 0;
}

.faq-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.faq-item {
  padding: 14px 18px;
  border-radius: 10px;
  background: var(--cm-bg-secondary);
  border: 1px solid var(--cm-border);
  cursor: pointer;
  transition: all 0.2s;
}

.faq-item:hover {
  border-color: rgba(16, 185, 129, 0.3);
}

.faq-item.open {
  border-color: rgba(16, 185, 129, 0.4);
  background: linear-gradient(135deg, var(--cm-bg-secondary), rgba(16, 185, 129, 0.03));
}

.faq-q {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  font-size: 14px;
  font-weight: 500;
  color: var(--cm-text);
}

.faq-arrow {
  transition: transform 0.2s;
  color: var(--cm-text-muted);
  flex-shrink: 0;
}

.faq-item.open .faq-arrow {
  transform: rotate(90deg);
  color: #10B981;
}

.faq-a {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid var(--cm-border);
  font-size: 13px;
  color: var(--cm-text-secondary);
  line-height: 1.7;
}

@media (max-width: 768px) {
  .doc-hero {
    padding: 32px 16px 28px;
  }

  .doc-hero h1 {
    font-size: 28px;
  }

  .hero-subtitle {
    font-size: 14px;
  }

  .hero-actions {
    flex-direction: column;
    align-items: center;
  }

  .doc-body {
    padding: 0 16px 28px;
  }

  .doc-tabs-bar {
    padding: 0 12px;
  }

  .layer-cards {
    grid-template-columns: 1fr;
  }

  .feature-grid {
    grid-template-columns: 1fr;
  }

  .graph-showcase {
    grid-template-columns: repeat(2, 1fr);
  }

  .wiki-features {
    grid-template-columns: 1fr;
  }

  .project-features {
    grid-template-columns: repeat(2, 1fr);
  }

  .sharing-flow {
    flex-direction: column;
    gap: 8px;
    padding: 16px;
  }

  .flow-arrow {
    transform: rotate(90deg);
  }

  .share-types {
    grid-template-columns: 1fr;
  }

  .session-use-cases {
    grid-template-columns: 1fr;
  }

  .ai-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .model-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .install-methods {
    grid-template-columns: 1fr;
  }

  .data-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .section-header h2 {
    font-size: 18px;
  }
}

@media (max-width: 480px) {
  .doc-hero {
    padding: 24px 12px 20px;
  }

  .doc-hero h1 {
    font-size: 24px;
  }

  .doc-body {
    padding: 0 12px 28px;
  }

  .doc-tab {
    padding: 6px 10px;
    font-size: 12px;
  }

  .tab-icon {
    font-size: 12px;
  }

  .graph-showcase {
    grid-template-columns: 1fr;
  }

  .project-features {
    grid-template-columns: 1fr;
  }

  .ai-grid {
    grid-template-columns: 1fr;
  }

  .model-grid {
    grid-template-columns: 1fr;
  }

  .data-grid {
    grid-template-columns: 1fr;
  }

  .section-header h2 {
    font-size: 16px;
  }

  .section-icon {
    font-size: 22px;
  }
}
</style>
