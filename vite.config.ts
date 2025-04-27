import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { spawn } from 'child_process';

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

// https://vitejs.dev/config/
export default defineConfig({
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
        target: 'http://localhost:3001',
        changeOrigin: true,
        rewrite: (path) => path
      }
    }
  }
});
