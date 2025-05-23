import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  build: {
    outDir: 'dist',
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  plugins: [react()],
  server: {
    proxy: {
      '/byze/v0.2': {
        target: 'http://127.0.0.1:16688',
        // target: 'http://10.3.73.109:16688', // 雨浩的动态ip
        changeOrigin: true,
        configure: (proxy, options) => {
          proxy.on('proxyReq', (proxyReq, req, res) => {
            console.log('代理请求:', req.url, '到', options.target);
          });
          proxy.on('proxyRes', (proxyRes, req, res) => {
            console.log('收到代理响应:', proxyRes.statusCode);
          });
        },
      },
      '/health': {
        target: 'http://127.0.0.1:16688',
        changeOrigin: true,
        configure: (proxy, options) => {
          proxy.on('proxyReq', (proxyReq, req, res) => {
            console.log('代理请求:', req.url, '到', options.target);
          });
          proxy.on('proxyRes', (proxyRes, req, res) => {
            console.log('收到代理响应:', proxyRes.statusCode);
          });
        },
      },
    },
  },
});
