// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: [
    '@nuxtjs/tailwindcss'
  ],
  ssr: false,
  nitro: {
    devProxy: {
      '/api': {
        target: 'http://acey.k1yoshi.com/api'
      }
    }
  }
})
