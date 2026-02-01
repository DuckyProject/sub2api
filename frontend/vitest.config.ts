import { resolve } from 'path'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vitest/config'

export default defineConfig({
  plugins: [vue()],
  // Keep test config self-contained: `vite.config.ts` is a callback config and
  // cannot be merged directly by Vite/Vitest.
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      // Match app build behavior: runtime-only vue-i18n to avoid CSP unsafe-eval.
      'vue-i18n': 'vue-i18n/dist/vue-i18n.runtime.esm-bundler.js'
    }
  },
  define: {
    __INTLIFY_JIT_COMPILATION__: true
  },
  test: {
    globals: true,
    environment: 'jsdom',
    include: ['src/**/*.{test,spec}.{js,ts,jsx,tsx}'],
    exclude: ['node_modules', 'dist'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      include: ['src/**/*.{js,ts,vue}'],
      exclude: ['node_modules', 'src/**/*.d.ts', 'src/**/*.spec.ts', 'src/**/*.test.ts', 'src/main.ts'],
      thresholds: {
        global: {
          statements: 80,
          branches: 80,
          functions: 80,
          lines: 80
        }
      }
    },
    setupFiles: ['./src/__tests__/setup.ts']
  }
})
