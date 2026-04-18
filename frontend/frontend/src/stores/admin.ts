import { defineStore } from 'pinia'
import { ref } from 'vue'
import { adminApi } from '../api/admin'
import { modelsApi } from '../api/models'
import { skillsApi } from '../api/skills'
import { nodesApi } from '../api/nodes'
import { backupsApi } from '../api/backups'

export const useAdminStore = defineStore('admin', () => {
  // Users
  const users = ref<any[]>([])
  async function fetchUsers() {
    const resp = await adminApi.listUsers()
    users.value = resp.data
  }
  async function toggleUserActive(id: number, is_active: boolean) {
    await adminApi.toggleUserActive(id, is_active)
  }

  // Models
  const models = ref<any[]>([])
  async function fetchModels() {
    const resp = await modelsApi.list()
    models.value = resp.data
  }
  async function createModel(data: any) {
    const resp = await modelsApi.create(data)
    models.value.push(resp.data)
    return resp.data
  }
  async function deleteModel(id: number) {
    await modelsApi.delete(id)
    models.value = models.value.filter(m => m.id !== id)
  }

  // Skills
  const skills = ref<any[]>([])
  async function fetchSkills() {
    const resp = await skillsApi.list()
    skills.value = resp.data
  }
  async function createSkill(data: any) {
    const resp = await skillsApi.create(data)
    skills.value.push(resp.data)
    return resp.data
  }
  async function deleteSkill(id: number) {
    await skillsApi.delete(id)
    skills.value = skills.value.filter(s => s.id !== id)
  }

  // Nodes
  const nodes = ref<any[]>([])
  async function fetchNodes() {
    const resp = await nodesApi.list()
    nodes.value = resp.data
  }
  async function deleteNode(id: number) {
    await nodesApi.delete(id)
    nodes.value = nodes.value.filter(n => n.id !== id)
  }

  // Backups
  const backups = ref<any[]>([])
  async function fetchBackups() {
    const resp = await backupsApi.list()
    backups.value = resp.data
  }
  async function createBackup() {
    const resp = await backupsApi.create()
    backups.value.unshift(resp.data)
    return resp.data
  }
  async function deleteBackup(id: number) {
    await backupsApi.delete(id)
    backups.value = backups.value.filter(b => b.id !== id)
  }

  // License
  const licenseInfo = ref<any>(null)
  async function fetchLicenseInfo() {
    const resp = await adminApi.getLicenseInfo()
    licenseInfo.value = resp.data
  }
  async function activateLicense(key: string) {
    const resp = await adminApi.activateLicense(key)
    licenseInfo.value = resp.data
    return resp.data
  }

  return {
    users, fetchUsers, toggleUserActive,
    models, fetchModels, createModel, deleteModel,
    skills, fetchSkills, createSkill, deleteSkill,
    nodes, fetchNodes, deleteNode,
    backups, fetchBackups, createBackup, deleteBackup,
    licenseInfo, fetchLicenseInfo, activateLicense,
  }
})
