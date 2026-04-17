<script setup lang="ts">
import Card from 'primevue/card'
import { useStatusStore } from '../stores/status'

const store = useStatusStore()
</script>

<template>
  <Card class="stats-card h-full">
    <template #title>
      <div class="flex align-items-center gap-2">
        <i class="pi pi-chart-bar text-primary" />
        Statistics
      </div>
    </template>
    <template #content>
      <div class="flex flex-column gap-3">
        <div class="text-center py-2">
          <div class="text-6xl font-bold text-primary mb-1">
            {{ store.status?.seen_count?.toLocaleString() ?? '—' }}
          </div>
          <div class="text-color-secondary text-sm">Cards Tracked</div>
        </div>
      </div>
    </template>
  </Card>
</template>