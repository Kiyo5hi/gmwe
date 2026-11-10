import { createApp } from 'vue'
import App from './app.vue'
import { router } from './router'
import './styles.css'

const app = createApp(App)
app.use(router)
await router.isReady()
app.mount('#app')
