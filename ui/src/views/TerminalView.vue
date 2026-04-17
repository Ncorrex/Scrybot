<script setup lang="ts">
// Terminal view — wraps the Terminal component full-page.
</script>

<template>
  <div class="terminal-view flex flex-column" style="height: calc(100vh - 120px)">
    <h2 class="mt-0 mb-3 flex align-items-center gap-2">
      <i class="pi pi-terminal text-primary" />
      Terminal
    </h2>
    <div class="flex-1 overflow-hidden">
      <Terminal />
    </div>
  </div>
</template>

<script lang="ts">
import Terminal from '../components/Terminal.vue'
export default { components: { Terminal } }
</script>