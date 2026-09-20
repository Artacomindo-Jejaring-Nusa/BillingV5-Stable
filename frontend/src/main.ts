// frontend/src/main.ts

import './assets/main.css'
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import 'mapbox-gl/dist/mapbox-gl.css';

// Optimize font loading for better LCP
const fontLink = document.createElement('link');
fontLink.rel = 'stylesheet';
fontLink.href = 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap';
fontLink.media = 'print';
fontLink.onload = () => {
    fontLink.media = 'all';
};
document.head.appendChild(fontLink);

// 1. Impor semua yang dibutuhkan
import App from './App.vue'
import router from './router'
// @ts-ignore - Vuetify CSS import
import 'vuetify/dist/vuetify.min.css'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
// Optimasi: import selective components untuk tree shaking
import * as vuetifyComponents from './plugins/vuetify'
import '@mdi/font/css/materialdesignicons.css'
import 'leaflet/dist/leaflet.css';

import { useAuthStore } from './stores/auth'
import NotificationsPlugin from './plugins/notifications'

// 2. Buat instance Vuetify dengan tema B2B SaaS (Slate / Zinc) & Component Defaults
const vuetify = createVuetify({
  components: {
    ...components,
    ...vuetifyComponents
  },
  directives,
  defaults: {
    VCard: {
      elevation: 0,
      rounded: 'lg',
    },
    VBtn: {
      elevation: 0,
      rounded: 'md',
    },
    VTextField: {
      variant: 'outlined',
      density: 'compact',
    },
    VSelect: {
      variant: 'outlined',
      density: 'compact',
    },
    VAutocomplete: {
      variant: 'outlined',
      density: 'compact',
    },
    VCombobox: {
      variant: 'outlined',
      density: 'compact',
    },
    VTextarea: {
      variant: 'outlined',
      density: 'compact',
    },
    VChip: {
      elevation: 0,
      rounded: 'md',
    },
  },
  theme: {
    defaultTheme: 'light',
    themes: {
      light: {
        dark: false,
        colors: {
          primary: '#0f172a',
          secondary: '#475569',
          accent: '#2563eb',
          error: '#ef4444',
          warning: '#f59e0b',
          info: '#0284c7',
          success: '#10b981',
          background: '#f8fafc',
          surface: '#ffffff',
        },
      },
      dark: {
        dark: true,
        colors: {
          primary: '#f8fafc',
          secondary: '#94a3b8',
          accent: '#3b82f6',
          error: '#f87171',
          warning: '#fbbf24',
          info: '#38bdf8',
          success: '#34d399',
          background: '#09090b',
          surface: '#18181b',
        },
      },
    },
  },
})


async function startup() {
  const app = createApp(App)
  const pinia = createPinia()
  
  app.use(pinia)

  // 3. Panggil aksi untuk memulihkan sesi
  // Ini akan berjalan SEBELUM aplikasi di-mount
  const authStore = useAuthStore()
  await authStore.initializeAuth()

  app.use(router)
  app.use(vuetify)
  app.use(NotificationsPlugin)

  // 4. Mount aplikasi setelah semua siap
  app.mount('#app')
}

startup()
