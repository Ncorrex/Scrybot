import { ref } from 'vue'
import { defineStore } from 'pinia'

export interface LogLine {
  ts: string
  level: 'INFO' | 'WARN' | 'ERROR'
  msg: string
}

const MAX_LOGS = 500

export const useLogsStore = defineStore('logs', () => {
  const lines = ref<LogLine[]>([])
  const connected = ref(false)

  function add(line: LogLine) {
    lines.value.push(line)
    if (lines.value.length > MAX_LOGS) {
      lines.value.splice(0, lines.value.length - MAX_LOGS)
    }
  }

  function clear() {
    lines.value = []
  }

  return { lines, connected, add, clear }
})