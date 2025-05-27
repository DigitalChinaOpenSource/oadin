import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  base: './',
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
    host: '127.0.0.1', // 指定主机地址
    port: 16699, // 指定端口号
    strictPort: true, // 如果端口已被占用，则会直接退出而不是尝试下一个可用端口
    proxy: {
      '/byze/v0.2': {
        target: 'http://127.0.0.1:16688',
        // target: 'http://10.3.73.109:16688', // 雨浩的动态ip
        // target: 'http://10.3.74.111:16688', // 朱灿的本地ip
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
      // '/api': {
      //   target: 'http://10.3.74.111:16688', // 朱灿本地环境
      //   changeOrigin: true,
      // },
      '/health': {
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
    },
  },
});
