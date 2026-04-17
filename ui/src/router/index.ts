import { createRouter, createWebHashHistory } from 'vue-router'
import DashboardView from '../views/DashboardView.vue'
import LogsView from '../views/LogsView.vue'
import TerminalView from '../views/TerminalView.vue'

export default createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', component: DashboardView },
    { path: '/logs', component: LogsView },
    { path: '/terminal', component: TerminalView },
  ],
})