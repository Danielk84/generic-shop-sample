import { createApp, defineAsyncComponent } from 'vue'
import { VueQueryPlugin } from '@tanstack/vue-query'

import '@/styles/index.css'
import App from '@/App.vue'
import router from '@/router/index'
import { createPinia } from 'pinia'
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate'

const pinia = createPinia()
pinia.use(piniaPluginPersistedstate)

const app = createApp(App)

app.use(pinia)
app.use(router)
app.use(VueQueryPlugin, {
  enableDevtoolsV6Plugin: true,
})

const globalComponents: Array<{
  name: string
  path: string
}> = [
  // used for recursion.
  {
    name: 'CommentsList',
    path: 'components/common/comments/CommentsList',
  },
] as const

globalComponents.forEach((v) => {
  app.component(
    v.name,
    defineAsyncComponent(() => import(/* @vite-ignore */ `./${v.path}.vue`)),
  )
})

app.mount('#app')
