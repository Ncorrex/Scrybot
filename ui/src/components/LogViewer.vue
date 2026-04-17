<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted } from 'vue'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import { useLogsStore, type LogLine } from '../stores/logs'
import { useWebSocket } from '../composables/useWebSocket'

const props = withDefaults(defineProps<{ maxLines?: number; compact?: boolean }>(), {
  maxLines: 0,
  compact: false,
})

const store = useLogsStore()
const filter = ref('')
const autoScroll = ref(true)
const scrollEl = ref<HTMLElement | null>(null)

const wsUrl = computed(() => {
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${location.host}/ws/logs`
})

const { isConnected } = useWebSocket(wsUrl.value, (raw) => {
  try {
    const line: LogLine = JSON.parse(raw)
    store.add(line)
  } catch {
    // ignore malformed frames
  }
})

watch(isConnected, (v) => {
  store.connected = v
})

const displayed = computed(() => {
  let list = store.lines
  if (filter.value.trim()) {
    const q = filter.value.toLowerCase()
    list = list.filter((l) => l.msg.toLowerCase().includes(q))
  }
  if (props.maxLines > 0) {
    list = list.slice(-props.maxLines)
  }
  return list
})

async function scrollBottom() {
  await nextTick()
  if (scrollEl.value) {
    scrollEl.value.scrollTop = scrollEl.value.scrollHeight
  }
}

watch(
  () => store.lines.length,
  () => {
    if (autoScroll.value) scrollBottom()
  },
)

onMounted(scrollBottom)

function levelSeverity(level: string) {
  if (level === 'ERROR') return 'danger'
  if (level === 'WARN') return 'warn'
  return 'secondary'
}

function fmtTime(iso: string) {
  return new Date(iso).toLocaleTimeString()
}
</script>

<template>
  <div class="log-viewer flex flex-column gap-2 h-full">
    <div class="flex align-items-center gap-2 flex-shrink-0">
      <span class="flex align-items-center gap-1 text-xs">
        <span
          class="ws-dot"
          :class="isConnected ? 'ws-dot--connected' : 'ws-dot--disconnected'"
        />
        {{ isConnected ? 'Live' : 'Reconnecting…' }}
      </span>
      <InputText
        v-model="filter"
        placeholder="Filter logs…"
        size="small"
        class="flex-1"
      />
      <Button
        :icon="autoScroll ? 'pi pi-lock' : 'pi pi-lock-open'"
        v-tooltip="autoScroll ? 'Auto-scroll on' : 'Auto-scroll off'"
        :severity="autoScroll ? 'primary' : 'secondary'"
        size="small"
        text
        @click="autoScroll = !autoScroll"
      />
      <Button
        icon="pi pi-trash"
        v-tooltip="'Clear logs'"
        severity="secondary"
        size="small"
        text
        @click="store.clear()"
      />
    </div>

    <div
      ref="scrollEl"
      class="log-list flex-1 overflow-y-auto font-mono"
      :class="{ 'log-list--compact': compact }"
    >
      <div
        v-for="(line, i) in displayed"
        :key="i"
        class="log-line flex gap-2 align-items-start"
        :class="`log-line--${line.level.toLowerCase()}`"
      >
        <span class="log-time text-color-secondary flex-shrink-0">{{ fmtTime(line.ts) }}</span>
        <Tag
          :value="line.level"
          :severity="levelSeverity(line.level)"
          class="flex-shrink-0 log-tag"
        />
        <span class="log-msg">{{ line.msg }}</span>
      </div>
      <div v-if="displayed.length === 0" class="text-center text-color-secondary py-4 text-sm">
        {{ filter ? 'No matching log lines.' : 'Waiting for log output…' }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.log-list {
  background: var(--p-surface-900);
  border: 1px solid var(--p-surface-700);
  border-radius: var(--p-border-radius-md);
  padding: 0.5rem;
  min-height: 200px;
}

.log-list--compact {
  min-height: 120px;
}

.log-line {
  padding: 2px 4px;
  border-radius: 4px;
  font-size: 0.75rem;
  line-height: 1.5;
}

.log-line:hover {
  background: var(--p-surface-800);
}

.log-line--error .log-msg {
  color: var(--p-red-400);
}

.log-line--warn .log-msg {
  color: var(--p-yellow-400);
}

.log-time {
  font-size: 0.7rem;
  min-width: 6rem;
}

.log-tag {
  font-size: 0.6rem !important;
  padding: 0 4px !important;
}

.ws-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.ws-dot--connected {
  background: var(--p-green-500);
  box-shadow: 0 0 4px var(--p-green-500);
}

.ws-dot--disconnected {
  background: var(--p-red-500);
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}
</style>