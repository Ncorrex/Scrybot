<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'

const route = useRoute()
const router = useRouter()

const navItems = [
  { path: '/', icon: 'pi pi-home', label: 'Dashboard' },
  { path: '/logs', icon: 'pi pi-list', label: 'Logs' },
  { path: '/terminal', icon: 'pi pi-terminal', label: 'Terminal' },
]

function isActive(path: string) {
  return route.path === path
}
</script>

<template>
  <div class="app-shell">
    <!-- Sidebar -->
    <aside class="app-sidebar">
      <div class="sidebar-brand">
        <i class="pi pi-bolt text-primary text-2xl" />
        <span class="font-bold text-lg">Scrybot</span>
      </div>

      <nav class="sidebar-nav">
        <Button
          v-for="item in navItems"
          :key="item.path"
          :icon="item.icon"
          :label="item.label"
          :severity="isActive(item.path) ? 'primary' : 'secondary'"
          :text="!isActive(item.path)"
          class="w-full justify-content-start nav-btn"
          @click="router.push(item.path)"
        />
      </nav>

      <div class="sidebar-footer text-xs text-color-secondary">
        Scryfall Alert Bot
      </div>
    </aside>

    <!-- Main content -->
    <main class="app-main">
      <RouterView />
    </main>
  </div>
</template>

<style>
/* Global resets */
*,
*::before,
*::after {
  box-sizing: border-box;
}

html,
body,
#app {
  margin: 0;
  padding: 0;
  height: 100%;
  background: var(--p-surface-950, #020617);
  color: var(--p-surface-0, #ffffff);
  font-family: var(--p-font-family);
}
</style>

<style scoped>
.app-shell {
  display: grid;
  grid-template-columns: 220px 1fr;
  height: 100vh;
}

.app-sidebar {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding: 1rem;
  background: var(--p-surface-900);
  border-right: 1px solid var(--p-surface-700);
  overflow: hidden;
}

.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0.5rem 1rem;
  border-bottom: 1px solid var(--p-surface-700);
  margin-bottom: 0.5rem;
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1;
}

.nav-btn {
  text-align: left !important;
}

.sidebar-footer {
  padding-top: 0.75rem;
  border-top: 1px solid var(--p-surface-700);
  text-align: center;
}

.app-main {
  padding: 1.5rem;
  overflow-y: auto;
}
</style>