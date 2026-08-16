import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

// 渲染进程构建：产物输出到 dist/，由 Electron 主进程加载
export default defineConfig({
  plugins: [vue()],
  base: './',
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    chunkSizeWarningLimit: 1500,
  },
  server: {
    port: 5174,
    strictPort: false,
  },
});
