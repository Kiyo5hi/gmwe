import { createApp } from 'vue'
import App from './app.vue'
import { router } from './router'
import './styles.css'

const app = createApp(App)
app.use(router)
// Render the shared shell immediately; RouterView loads the first page separately.
app.mount('#app')
