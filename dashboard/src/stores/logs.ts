import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { TraceLog } from '@/types'

const TRACE_KEY = 'gemini-trace-logs'
const MAX_ITEMS = 500

function load<T>(key: string): T[] {
  try {
    const raw = localStorage.getItem(key)
    return raw ? (JSON.parse(raw) as T[]) : []
  } catch {
    return []
  }
}

function save(key: string, data: unknown[]) {
  try {
    localStorage.setItem(key, JSON.stringify(data))
  } catch {
    // ignore storage errors
  }
}

function generateId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`
}

export const useLogsStore = defineStore('logs', () => {
  const traceLogs = ref<TraceLog[]>(load<TraceLog>(TRACE_KEY))

  function addTraceLog(entry: Omit<TraceLog, 'id' | 'timestamp'>) {
    const item: TraceLog = {
      ...entry,
      id: generateId(),
      timestamp: new Date().toISOString(),
    }
    traceLogs.value.unshift(item)
    if (traceLogs.value.length > MAX_ITEMS) {
      traceLogs.value = traceLogs.value.slice(0, MAX_ITEMS)
    }
    save(TRACE_KEY, traceLogs.value)
  }

  function clearTraceLogs() {
    traceLogs.value = []
    localStorage.removeItem(TRACE_KEY)
  }

  return {
    traceLogs,
    addTraceLog,
    clearTraceLogs,
  }
})
