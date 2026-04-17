<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Card from 'primevue/card'
import StatusCard from '../components/StatusCard.vue'
import StatsCard from '../components/StatsCard.vue'
import ConfigCard from '../components/ConfigCard.vue'
import LogViewer from '../components/LogViewer.vue'
import { useStatusStore } from '../stores/status'

const store = useStatusStore()
const router = useRouter()

onMounted(() => store.startPolling())
onUnmounted(() => store.stopPolling())
</script>

<template>
  <div class="dashboard flex flex-column gap-4">
    <div class="grid">
      <div class="col-12 md:col-4">
        <StatusCard />
      </div>
      <div class="col-12 md:col-4">
        <StatsCard />
      </div>
      <div class="col-12 md:col-4">
        <ConfigCard />
      </div>
    </div>

    <Card>
      <template #title>
        <div class="flex align-items-center justify-content-between">
          <div class="flex align-items-center gap-2">
            <i class="pi pi-list text-primary" />
            Recent Logs
          </div>
          <Button
            label="Full Log Viewer"
            icon="pi pi-external-link"
            severity="secondary"
            size="small"
            text
            @click="router.push('/logs')"
          />
        </div>
      </template>
      <template #content>
        <div style="height: 320px">
          <LogViewer :max-lines="50" compact />
        </div>
      </template>
    </Card>
  </div>
</template>