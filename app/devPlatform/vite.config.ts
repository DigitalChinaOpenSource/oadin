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
    host: '0.0.0.0',
    // host: '127.0.0.1', // 指定主机地址
    port: 1666, // 指定端口号
    strictPort: true, // 如果端口已被占用，则会直接退出而不是尝试下一个可用端口
    proxy: {
      ['/api']: {
        target: 'http://10.3.74.123:3000', // 向秋怡的后端服务代理请求
        // target: 'http://10.3.73.66:3000', // 叶鑫的后端服务代理请求
        changeOrigin: true,
        rewrite: (path) => path.replace(new RegExp(`^/api`), '/api'),
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
