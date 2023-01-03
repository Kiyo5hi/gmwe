// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: [
    '@nuxtjs/tailwindcss'
  ],
  app: {
    head: {
      title: 'For Acey'
    }
  },
  nitro: {
    devProxy: {
      '/api': {
        target: 'http://localhost:5500/api'
      }
    }
  },
  ssr: false
})
