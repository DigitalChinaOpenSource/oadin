//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';
import { API_PREFIX, API_HEALTH_ENDPOINT } from './src/constants';
import svgr from 'vite-plugin-svgr';

export default defineConfig({
  // base: './',
  build: {
    outDir: 'dist',
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  plugins: [react(), svgr()],
  server: {
    host: '0.0.0.0',
    // host: '127.0.0.1', // 指定主机地址
    port: 16698, // 指定端口号
    strictPort: true, // 如果端口已被占用，则会直接退出而不是尝试下一个可用端口
    proxy: {
      [API_PREFIX]: {
        target: 'http://127.0.0.1:16688',
        changeOrigin: true,
        rewrite: (path) => path.replace(new RegExp(`^${API_PREFIX}`), API_PREFIX),
        configure: (proxy, options) => {
          proxy.on('proxyReq', (proxyReq, req, res) => {
            console.log('代理请求:', req.url, '到', options.target);
          });
          proxy.on('proxyRes', (proxyRes, req, res) => {
            console.log('收到代理响应:', proxyRes.statusCode);
          });
        },
      },
      [API_HEALTH_ENDPOINT]: {
        target: 'http://127.0.0.1:16688',
        changeOrigin: true,
        rewrite: (path) => path.replace(new RegExp(`^${API_HEALTH_ENDPOINT}`), API_HEALTH_ENDPOINT),
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
