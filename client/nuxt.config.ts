// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  modules: [
    '@nuxtjs/tailwindcss'
  ],
  app: {
    head: {
      title: 'GMWE',
      meta: [
        { name: 'theme-color', content: '#166345' },
        { name: 'apple-mobile-web-app-capable', content: 'yes' },
        { name: 'apple-mobile-web-app-title', content: 'GMWE' },
        { name: 'apple-mobile-web-app-status-bar-style', content: 'default' }
      ],
      link: [
        { rel: 'manifest', href: '/manifest.webmanifest' },
        { rel: 'apple-touch-icon', href: '/pwa-180.png' }
      ],
      script: [{ src: '/pwa.js', defer: true }]
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
