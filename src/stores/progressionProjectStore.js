import { defineStore } from 'pinia'
import { computed, ref } from 'vue'
import service from '../services/progressionProjectService'

export const useProgressionProjectStore = defineStore('progressionProjects', () => {
  const projects = ref([])
  const activeId = ref(null)
  const loading = ref(false)
  const loaded = ref(false)
  let loadPromise = null
  const error = ref('')
  const activeProject = computed(() => projects.value.find((p) => p.id === activeId.value) || null)
  async function load() {
    if (loaded.value) return
    if (loadPromise) return loadPromise
    loadPromise = loadOnce().finally(() => {
      loadPromise = null
    })
    return loadPromise
  }
  async function loadOnce() {
    loading.value = true
    try {
      const d = await service.list()
      projects.value = d.projects || []
      loaded.value = true
      activeId.value ||= projects.value[0]?.id || null
    } catch (e) {
      error.value = e.message
    } finally {
      loading.value = false
    }
  }
  async function persist() {
    await service.save({ version: 1, projects: projects.value })
  }
  async function create(name = 'Untitled progression') {
    if (!loaded.value) await load()
    if (!loaded.value) throw new Error('Load projects before creating one')
    const p = {
      id: `project-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
      name,
      orderMode: 'inferred',
      snapshots: [],
    }
    projects.value.push(p)
    activeId.value = p.id
    await persist()
    return p
  }
  async function update(project) {
    const i = projects.value.findIndex((p) => p.id === project.id)
    if (i >= 0) {
      projects.value[i] = { ...project }
      await persist()
    }
  }
  async function remove(id) {
    projects.value = projects.value.filter((p) => p.id !== id)
    if (activeId.value === id) activeId.value = projects.value[0]?.id || null
    await persist()
  }
  return {
    projects,
    activeId,
    activeProject,
    loading,
    loaded,
    error,
    load,
    create,
    update,
    remove,
    sync: service.state,
    retry: () => service.retry(),
    resolveConflict: async (choice) => {
      const d = await service.resolveConflict(choice)
      projects.value = d.projects
      activeId.value = projects.value[0]?.id || null
    },
    replaceDocument: async (value) => {
      await load()
      if (!loaded.value) throw new Error('Load projects first')
      projects.value = value.projects
      await persist()
      return true
    },
  }
})
