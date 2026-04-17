<script setup lang="ts">
import { computed } from 'vue'
import Card from 'primevue/card'
import Tag from 'primevue/tag'
import { useStatusStore } from '../stores/status'

const store = useStatusStore()

function fmt(iso: string | null | undefined): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleString()
}

const timeSince = computed(() => {
  if (!store.status?.last_check) return null
  const diff = Date.now() - new Date(store.status.last_check).getTime()
  const m = Math.floor(diff / 60_000)
  if (m < 1) return 'just now'
  if (m === 1) return '1 min ago'
  return `${m} min ago`
})
</script>

<template>
  <Card class="status-card h-full">
    <template #title>
      <div class="flex align-items-center gap-2">
        <i class="pi pi-server text-primary" />
        Bot Status
      </div>
    </template>
    <template #content>
      <div class="flex flex-column gap-3">
        <div class="flex align-items-center justify-content-between">
          <span class="text-color-secondary text-sm">Status</span>
          <Tag
            :value="store.status?.running ? 'Running' : 'Stopped'"
            :severity="store.status?.running ? 'success' : 'danger'"
            icon="pi pi-circle-fill"
          />
        </div>
        <div class="flex align-items-center justify-content-between">
          <span class="text-color-secondary text-sm">Last Poll</span>
          <span class="text-sm font-mono">
            {{ fmt(store.status?.last_check) }}
            <span v-if="timeSince" class="text-color-secondary ml-1">({{ timeSince }})</span>
          </span>
        </div>
        <div class="flex align-items-center justify-content-between">
          <span class="text-color-secondary text-sm">Next Poll</span>
          <span class="text-sm font-mono">{{ fmt(store.status?.next_check) }}</span>
        </div>
      </div>
    </template>
  </Card>
</template>