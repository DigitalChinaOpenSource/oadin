const fs = require('fs');
const path = require('path');
const winston = require('winston');
const { MAC_OADIN_PATH, OADIN_HEALTH, OADIN_ENGINE_PATH, } = require('./constants.js');
const axios = require('axios');
const child_process = require('child_process');


async function isOadinAvailable(retries = 5, interval = 1000) {
  logAndConsole('info', '检测Oadin服务可用性...');
  const fibArr = fibonacci(retries, interval);
  for (let attempt = 0; attempt < retries; attempt++) {
    try {
      const [healthRes, engineHealthRes] = await Promise.all([
        axios.get(OADIN_HEALTH),
        axios.get(OADIN_ENGINE_PATH)
      ]);
      const healthOk = isHealthy(healthRes.status);
      const engineOk = isHealthy(engineHealthRes.status);
      logAndConsole('info', `/health: ${healthOk ? '正常' : '异常'}, /engine/health: ${engineOk ? '正常' : '异常'}`);
      if (healthOk && engineOk) return true;
    } catch (err) {
      logAndConsole('warn', `健康检查失败: ${err.message}`);
    }
    if (attempt < retries - 1) {
      await new Promise(r => setTimeout(r, fibArr[attempt]));
    }
  }
  logAndConsole('warn', 'Oadin服务不可用');
  return false;
}

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

// 日志工具，所有日志写入 oadin.log，带[info]/[warn]/[error]前缀
// TODO:写在用户目录下
const logFilePath = path.join(__dirname, 'oadin.log');
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

// 平台相关：获取oadin可执行文件路径
function getOadinExecutablePath() {
  const userDir = require('os').homedir();
  const platform = getPlatform();
  if (platform === 'win32') {
    return path.join(userDir, 'Oadin', 'oadin.exe');
  } else if (platform === 'darwin') {
    return MAC_OADIN_PATH;
  }
  return null;
}

// 平台相关：运行安装包
function runInstallerByPlatform0(installerPath) {
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
      // 轮询检测安装目录生成
      const expectedPath = MAC_OADIN_PATH;
      const maxRetries = 100;
      let retries = 0;

      const interval = setInterval(async () => {
        if (fs.existsSync(expectedPath)) {
          console.log("oadin 已添加到 /usr/local/bin ");
          // 检查服务是否可用
          const available = await isOadinAvailable(2, 1000);
          if (available) {
            clearInterval(interval);
            resolve();
          }
        } else if (++retries >= maxRetries) {
          clearInterval(interval);
          reject(new Error('安装器未在超时前完成安装'));
        }
      }, 1000);

    });
  }
  return Promise.reject(new Error('不支持的平台'));
}

function runInstallerByPlatform(installerPath) {
  const platform = getPlatform();
  if (platform === 'win32') {
    return new Promise((resolve, reject) => {
      const child = child_process.spawn(installerPath, ['/S'], { stdio: 'inherit' });
      child.on('error', reject);
      child.on('close', (code) => {
        code === 0 ? resolve() : reject(new Error(`Installer exited with code ${code}`));
      });
    });
  } else if (platform === 'darwin') {
    return new Promise((resolve, reject) => {
      console.log(`正在启动 macOS 安装程序：${installerPath}`);
      logAndConsole('info', `正在启动 macOS 安装程序：${installerPath}`);

      // 构建 AppleScript 来执行 installer 命令并请求管理员权限
      // 这会触发 macOS 的图形界面密码认证弹窗
      const appleScript = `
        set pkgPath to "${installerPath}"
        set installCommand to "installer -pkg " & quoted form of pkgPath & " -target /"
        try
            -- 执行 shell 命令并请求管理员权限
            -- 这将触发 macOS 的图形界面密码认证弹窗
            do shell script installCommand with administrator privileges
            return "INSTALL_SUCCESS" -- 安装成功
        on error errMsg number errNum
            return "INSTALL_FAILED: " & errMsg & " (错误码: " & errNum & ")" -- 安装失败及错误信息
        end try
      `;

      // 使用 osascript 执行 AppleScript
      // stdio 设置为 pipe 以捕获 AppleScript 的返回结果
      const osascriptProcess = child_process.spawn('osascript', ['-e', appleScript], {
        stdio: ['pipe', 'pipe', 'pipe']
      });

      let scriptOutput = '';
      let scriptError = '';

      // 捕获 AppleScript 的标准输出（成功或失败消息）
      osascriptProcess.stdout.on('data', (data) => {
        scriptOutput += data.toString().trim();
      });

      // 捕获 osascript 自身的错误输出（例如 AppleScript 语法错误）
      osascriptProcess.stderr.on('data', (data) => {
        scriptError += data.toString();
        logAndConsole('error', `osascript stderr: ${data.toString()}`);
      });

      // 监听 osascript 进程的错误（例如 osascript 命令找不到）
      osascriptProcess.on('error', (err) => {
        logAndConsole('error', `执行 osascript 失败: ${err.message}`);
        reject(new Error(`无法执行 AppleScript 来运行安装程序: ${err.message}`));
      });

      // 监听 osascript 进程关闭
      osascriptProcess.on('close', async (osascriptCode) => {
        if (osascriptCode !== 0) {
          // osascript 自身执行失败 (非AppleScript内部错误)
          logAndConsole('error', `osascript 退出码非零: ${osascriptCode}. 错误: ${scriptError || '未知'}`);
          return reject(new Error(`AppleScript 执行失败，退出码: ${osascriptCode}. 错误: ${scriptError || '未知'}`));
        }

        if (scriptOutput === 'INSTALL_SUCCESS') {
          console.log('macOS 安装程序已成功完成。');
          logAndConsole('info', 'macOS 安装程序已成功完成。');
          
          // 轮询检测安装目录生成和 Oadin 服务可用性
          const expectedPath = MAC_OADIN_PATH;
          const maxRetries = 100;
          let retries = 0;

          const interval = setInterval(async () => {
            if (fs.existsSync(expectedPath)) {
              console.log("oadin 已添加到 /usr/local/bin ");
              logAndConsole('info', "oadin 已添加到 /usr/local/bin ");
              const available = await isOadinAvailable(2, 1000); // 检查服务是否可用
              if (available) {
                clearInterval(interval);
                // 成功后显示一个确认弹窗（可选，因为安装包本身可能会显示）
                // child_process.exec('osascript -e \'display dialog "oadin 安装成功！" buttons {"确定"} default button "确定" with icon note\'');
                resolve();
              }
            } else if (++retries >= maxRetries) {
              clearInterval(interval);
              const errorMessage = '安装器未在超时前完成安装或 Oadin 未成功安装。';
              logAndConsole('error', errorMessage);
              reject(new Error(errorMessage));
            }
          }, 1000);

        } else if (scriptOutput.startsWith('INSTALL_FAILED:')) {
          // AppleScript 内部捕获到安装失败
          const errorDetail = scriptOutput.substring('INSTALL_FAILED:'.length).trim();
          console.error(`安装失败: ${errorDetail}`);
          logAndConsole('error', `安装失败: ${errorDetail}`);
          // 失败后显示一个错误弹窗（可选，因为安装包本身可能会显示）
          // child_process.exec(`osascript -e 'display dialog "安装失败：${errorDetail}" buttons {"确定"} default button "确定" with icon stop'`);
          reject(new Error(`macOS 安装失败: ${errorDetail}`));
        } else {
          // 未知输出，可能是 AppleScript 逻辑问题
          const unknownError = `AppleScript 返回未知结果: ${scriptOutput}. 原始错误: ${scriptError}`;
          console.error(unknownError);
          logAndConsole('error', unknownError);
          reject(new Error(unknownError));
        }
      });
    });
  }
  return Promise.reject(new Error('不支持的平台')); // 不支持的平台则拒绝 Promise
}

module.exports = {
  getPlatform,
  ensureDirWritable,
  fibonacci,
  checkPort,
  logAndConsole,
  downloadFile,
  getOadinExecutablePath,
  runInstallerByPlatform,
  isHealthy
};
