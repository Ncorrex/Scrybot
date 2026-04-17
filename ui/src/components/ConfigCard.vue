<script setup lang="ts">
import Card from 'primevue/card'
import Tag from 'primevue/tag'
import { useStatusStore } from '../stores/status'

const store = useStatusStore()
</script>

<template>
  <Card class="config-card h-full">
    <template #title>
      <div class="flex align-items-center gap-2">
        <i class="pi pi-cog text-primary" />
        Configuration
      </div>
    </template>
    <template #content>
      <div class="flex flex-column gap-3">
        <div>
          <div class="text-color-secondary text-xs mb-1">Search Query</div>
          <code class="text-xs surface-100 p-1 border-round block overflow-auto">
            {{ store.config?.search_query ?? '—' }}
          </code>
        </div>
        <div class="flex align-items-center justify-content-between">
          <span class="text-color-secondary text-sm">Poll Interval</span>
          <Tag :value="store.config?.poll_interval ?? '—'" severity="secondary" />
        </div>
        <div class="flex align-items-center justify-content-between">
          <span class="text-color-secondary text-sm">Data Dir</span>
          <code class="text-xs">{{ store.config?.data_dir ?? '—' }}</code>
        </div>
        <div class="flex align-items-center justify-content-between">
          <span class="text-color-secondary text-sm">Webhook</span>
          <Tag
            :value="store.config?.webhook_configured ? 'Configured' : 'Missing'"
            :severity="store.config?.webhook_configured ? 'success' : 'danger'"
          />
        </div>
      </div>
    </template>
  </Card>
</template>