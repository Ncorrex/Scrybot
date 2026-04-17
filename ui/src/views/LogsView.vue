<script setup lang="ts">
// Full-page log viewer — LogViewer manages its own WebSocket connection.
</script>

<template>
  <div class="logs-view flex flex-column" style="height: calc(100vh - 120px)">
    <h2 class="mt-0 mb-3 flex align-items-center gap-2">
      <i class="pi pi-list text-primary" />
      Live Log Stream
    </h2>
    <div class="flex-1 overflow-hidden">
      <LogViewer />
    </div>
  </div>
</template>

<script lang="ts">
import LogViewer from '../components/LogViewer.vue'
export default { components: { LogViewer } }
</script>