<script setup lang="ts">
import { ref } from 'vue'
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Card from 'primevue/card'
import Message from 'primevue/message'

interface OutputLine {
  text: string
  type: 'input' | 'output' | 'error'
}

const input = ref('')
const output = ref<OutputLine[]>([])
const loading = ref(false)

// Commands will be wired up when the backend supports them.
const COMMANDS_ENABLED = false

async function send() {
  const cmd = input.value.trim()
  if (!cmd) return

  output.value.push({ text: `> ${cmd}`, type: 'input' })
  input.value = ''

  if (!COMMANDS_ENABLED) {
    output.value.push({ text: 'No commands implemented yet.', type: 'error' })
    return
  }

  loading.value = true
  try {
    const res = await fetch('/api/command', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: cmd }),
    })
    const data = await res.json()
    output.value.push({ text: data.output ?? '', type: data.ok ? 'output' : 'error' })
  } catch (e) {
    output.value.push({ text: String(e), type: 'error' })
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Card class="h-full">
    <template #title>
      <div class="flex align-items-center gap-2">
        <i class="pi pi-terminal text-primary" />
        Terminal
      </div>
    </template>
    <template #content>
      <div class="flex flex-column gap-3 h-full">
        <Message severity="secondary" :closable="false" class="text-sm">
          <i class="pi pi-info-circle mr-2" />
          No commands are available yet. This terminal will be enabled in a future release.
        </Message>

        <div class="terminal-output font-mono flex-1 overflow-y-auto">
          <div
            v-for="(line, i) in output"
            :key="i"
            :class="[
              'terminal-line text-xs',
              `terminal-line--${line.type}`,
            ]"
          >
            {{ line.text }}
          </div>
          <div v-if="output.length === 0" class="text-color-secondary text-xs py-2">
            No output yet.
          </div>
        </div>

        <div class="flex gap-2">
          <InputText
            v-model="input"
            placeholder="Enter command…"
            :disabled="!COMMANDS_ENABLED || loading"
            class="flex-1 font-mono"
            size="small"
            @keyup.enter="send"
          />
          <Button
            icon="pi pi-send"
            :disabled="!COMMANDS_ENABLED || loading || !input.trim()"
            :loading="loading"
            size="small"
            @click="send"
          />
        </div>
      </div>
    </template>
  </Card>
</template>

<style scoped>
.terminal-output {
  background: var(--p-surface-900);
  border: 1px solid var(--p-surface-700);
  border-radius: var(--p-border-radius-md);
  padding: 0.5rem;
  min-height: 120px;
}

.terminal-line {
  line-height: 1.6;
}

.terminal-line--input {
  color: var(--p-primary-400);
}

.terminal-line--error {
  color: var(--p-red-400);
}

.terminal-line--output {
  color: var(--p-surface-200);
}
</style>