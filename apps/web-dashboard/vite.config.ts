import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    port: Number(process.env.PORT) || 5173,
    proxy: {
      '/api': {
        target: process.env.VITE_PROXY_TARGET || 'http://api-gateway:8080',
        changeOrigin: true,
      },
    },
  },
})
