import { createApp } from 'vue'
import { VueQueryPlugin } from '@tanstack/vue-query'

import '@/styles/index.css'
import App from '@/App.vue'
import router from '@/router/index'
import { createPinia } from 'pinia'

const pinia = createPinia()
const app = createApp(App)

app.use(pinia)
app.use(router)
app.use(VueQueryPlugin, {
  enableDevtoolsV6Plugin: true,
})

app.mount('#app')
