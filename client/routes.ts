import type { RouteRecordRaw } from 'vue-router'

// Also used by the build to emit each direct-entry HTML path for Caddy.
export const routes = [
  { path: '/', component: () => import('./pages/index.vue'), meta: { title: 'GMWE' } },
  { path: '/hitokoto', component: () => import('./pages/hitokoto.vue'), meta: { title: 'Hitokoto | GMWE' } },
  { path: '/our-story', component: () => import('./pages/our-story.vue'), meta: { title: 'Story | GMWE' } },
  { path: '/account', component: () => import('./pages/account.vue'), meta: { title: 'Account | GMWE' } }
] satisfies RouteRecordRaw[]
