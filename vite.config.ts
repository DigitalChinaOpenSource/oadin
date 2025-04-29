import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { spawn } from 'child_process';
import path from 'path'

// 启动后端服务器的函数
function startBackendServer() {
  const server = spawn('node', ['src/server.js'], {
    stdio: 'inherit',
    shell: true
  });
  
  // 当 Vite 服务器关闭时，也关闭后端服务器
  process.on('exit', () => {
    server.kill();
  });
}

export default defineConfig({
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src')
    }
  },
  plugins: [
    react(),
    {
      name: 'start-backend',
      configureServer() {
        startBackendServer();
      }
    }
  ],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:33333',
        changeOrigin: true,
        rewrite: (path) => path
      }
    }
  }
});
