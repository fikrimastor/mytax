export default defineNuxtConfig({
  modules: ['@nuxt/ui'],
  css: ['~/assets/css/main.css'],
  devtools: { enabled: false },
  runtimeConfig: { public: { apiBase: process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8080' } },
  app: { head: { title: 'mytax | Malaysia Tax Claim Checker', meta: [{ name: 'description', content: 'Preliminary Malaysian personal tax claim checks.' }] } },
  compatibilityDate: '2025-05-01'
})