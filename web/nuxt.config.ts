export default defineNuxtConfig({
  compatibilityDate: '2026-01-01',
  devtools: { enabled: false },
  modules: ['@nuxt/eslint'],
  runtimeConfig: {
    backendBaseUrl: 'http://localhost:8080',
    public: {
      apiBaseUrl: '/api',
    },
  },
})
