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

const fs = require('fs');
const path = require('path');
const winston = require('winston');

// 判断平台
function getPlatform() {
  const platform = process.platform;
  if (platform === 'win32') return 'win32';
  if (platform === 'darwin') return 'darwin';
  return 'unsupported';
}

// 检查并创建目录，检查写权限
function ensureDirWritable(dirPath) {
  try {
    fs.mkdirSync(dirPath, { recursive: true });
    fs.accessSync(dirPath, fs.constants.W_OK);
    return true;
  } catch (e) {
    return false;
  }
}

function isHealthy(status){
  if ( status===200 ) return true;
  return false;
}

// 斐波那契数列生成器
function fibonacci(n, base) {
  let arr = [0, base];
  for (let i = 2; i < n + 2; i++) {
    arr[i] = arr[i - 1] + arr[i - 2];
  }
  return arr.slice(0, n);
}

// 检查端口
function checkPort(port, timeout = 3000) {
  return new Promise((resolvePort) => {
    const options = {
      hostname: 'localhost',
      port: port,
      path: '/',
      method: 'GET',
      timeout,
    };
    const req = require('http').request(options, (res) => {
      resolvePort(res.statusCode === 200);
    });
    req.on('error', () => resolvePort(false));
    req.on('timeout', () => {
      req.destroy();
      resolvePort(false);
    });
    req.end();
  });
}

// 日志工具，所有日志写入 AOG.log，带[info]/[warn]/[error]前缀
// TODO:写在用户目录下
const logFilePath = path.join(__dirname, 'aog.log');
const logFormat = winston.format.printf(({ level, message, timestamp }) => {
  let prefix = '[info]';
  if (level === 'warn') prefix = '[warn]';
  if (level === 'error') prefix = '[error]';
  return `${timestamp} ${prefix} ${message}`;
});
const logger = winston.createLogger({
  level: 'info',
  format: winston.format.combine(
    //TODO：开源用格林威治时间戳
    // 东八区时间戳
    winston.format.timestamp({
      format: () => new Date().toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' }),
    }),
    logFormat
  ),
  transports: [
    new winston.transports.File({ filename: logFilePath }),
    new winston.transports.Console(),
  ],
});

function logAndConsole(level, msg) {
  logger.log({ level, message: msg });
  if (level === 'info') console.log(msg);
  else if (level === 'warn') console.warn(msg);
  else if (level === 'error') console.error(msg);
}

// 下载文件（通用工具方法）
async function downloadFile(url, dest, options, retries = 3) {
  for (let attempt = 1; attempt <= retries; attempt++) {
    try {
      logger.info(`axios downloading... attempt ${attempt}`);
      const dirOk = ensureDirWritable(path.dirname(dest));
      if (!dirOk) throw new Error('目标目录不可写');
      const response = await require('axios').get(url, {
        ...options,
        responseType: 'stream',
        timeout: 15000,
        validateStatus: status => status === 200,
      });
      const writer = fs.createWriteStream(dest);
      await new Promise((resolve, reject) => {
        response.data.pipe(writer);
        writer.on('finish', resolve);
        writer.on('error', reject);
      });
      logger.info('axios download success');
      return true;
    } catch (err) {
      try { fs.unlinkSync(dest); } catch {}
      logger.warn(`下载失败（第${attempt}次）：${err.message}`);
      if (attempt === retries) {
        logger.error('多次下载失败，放弃');
        return false;
      }
    }
  }
  return false;
}

// 平台相关：获取aog可执行文件路径
function getAOGExecutablePath() {
  const userDir = require('os').homedir();
  const platform = getPlatform();
  if (platform === 'win32') {
    return path.join(userDir, 'AOG', 'aog.exe');
  } else if (platform === 'darwin') {
    return '/usr/local/bin/aog';
  }
  return null;
}

// 平台相关：运行安装包
function runInstallerByPlatform(installerPath) {
  const platform = getPlatform();
  if (platform === 'win32') {
    return new Promise((resolve, reject) => {
      const child = require('child_process').spawn(installerPath, ['/S'], { stdio: 'inherit' });
      child.on('error', reject);
      child.on('close', (code) => {
        code === 0 ? resolve() : reject(new Error(`Installer exited with code ${code}`));
      });
    });
  } else if (platform === 'darwin') {
    return new Promise((resolve, reject) => {
      const child = require('child_process').spawn('open', [installerPath], { stdio: 'ignore', detached: true });
      child.on('error', reject);
      // 可扩展轮询检测逻辑
      resolve();
    });
  }
  return Promise.reject(new Error('不支持的平台'));
}

module.exports = {
  getPlatform,
  ensureDirWritable,
  fibonacci,
  checkPort,
  logAndConsole,
  downloadFile,
  getAOGExecutablePath,
  runInstallerByPlatform,
  isHealthy
};
