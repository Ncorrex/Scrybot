import { ref } from 'vue'
import { defineStore } from 'pinia'

export interface StatusData {
  last_check: string | null
  next_check: string | null
  seen_count: number
  running: boolean
}

export interface ConfigData {
  search_query: string
  poll_interval: string
  data_dir: string
  webhook_configured: boolean
}

export const useStatusStore = defineStore('status', () => {
  const status = ref<StatusData | null>(null)
  const config = ref<ConfigData | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  let pollTimer: ReturnType<typeof setInterval> | null = null

  async function fetchStatus() {
    try {
      const res = await fetch('/api/status')
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      status.value = await res.json()
    } catch (e) {
      error.value = String(e)
    }
  }

  async function fetchConfig() {
    try {
      const res = await fetch('/api/config')
      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      config.value = await res.json()
    } catch (e) {
      error.value = String(e)
    }
  }

  async function refresh() {
    loading.value = true
    await Promise.all([fetchStatus(), fetchConfig()])
    loading.value = false
  }

  function startPolling() {
    refresh()
    pollTimer = setInterval(refresh, 30_000)
  }

  function stopPolling() {
    if (pollTimer !== null) {
      clearInterval(pollTimer)
      pollTimer = null
    }
  }

  return { status, config, loading, error, refresh, startPolling, stopPolling }
})