<template>
  <div class="agent-filter-bar">
    <el-select v-model="layer" @change="emitChange" :placeholder="$t('memories.layer')" clearable style="width: 130px">
      <el-option :label="$t('memories.all')" value="" />
      <el-option v-for="(label, key) in layerLabels" :key="key" :label="label" :value="key" />
    </el-select>
    <el-select v-model="memoryType" @change="emitChange" :placeholder="$t('memories.memoryType')" clearable style="width: 140px">
      <el-option :label="$t('memories.all')" value="" />
      <el-option :label="$t('memories.knowledge')" value="knowledge" />
      <el-option :label="$t('memories.feedback')" value="feedback" />
      <el-option :label="$t('memories.project')" value="project" />
      <el-option :label="$t('memories.reference')" value="reference" />
      <el-option :label="$t('memories.user')" value="user" />
    </el-select>
    <el-select v-model="sourceAgent" @change="emitChange" :placeholder="$t('memories.sourceAgent')" clearable style="width: 140px">
      <el-option :label="$t('memories.all')" value="" />
      <el-option v-for="a in agents" :key="a.name" :label="a.display_name || a.name" :value="a.name" />
    </el-select>
    <el-select v-model="visibility" @change="emitChange" :placeholder="$t('memories.visibility')" clearable style="width: 130px">
      <el-option :label="$t('memories.all')" value="" />
      <el-option :label="$t('memories.visibilityPrivate')" value="private" />
      <el-option :label="$t('memories.visibilityShared')" value="shared" />
      <el-option :label="$t('memories.visibilityPublic')" value="public" />
    </el-select>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{
  agents: any[]
  initialLayer?: string
  initialMemoryType?: string
  initialSourceAgent?: string
  initialVisibility?: string
}>()

const emit = defineEmits<{
  change: [filters: { layer: string; memoryType: string; sourceAgent: string; visibility: string }]
}>()

const layer = ref(props.initialLayer || '')
const memoryType = ref(props.initialMemoryType || '')
const sourceAgent = ref(props.initialSourceAgent || '')
const visibility = ref(props.initialVisibility || '')

const layerLabels: Record<string, string> = {
  preference: t('memories.preference'),
  knowledge: t('memories.knowledge'),
  short_term: t('memories.shortTerm'),
  private: t('memories.private'),
}

function emitChange() {
  emit('change', {
    layer: layer.value,
    memoryType: memoryType.value,
    sourceAgent: sourceAgent.value,
    visibility: visibility.value,
  })
}

watch(() => props.initialLayer, (v) => { layer.value = v || '' })
watch(() => props.initialSourceAgent, (v) => { sourceAgent.value = v || '' })
watch(() => props.initialVisibility, (v) => { visibility.value = v || '' })
</script>

<style scoped>
.agent-filter-bar {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}
</style>
