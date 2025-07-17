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

      // AppleScript 来请求密码
      const appleScript = `
        tell application "System Events"
          display dialog "Oadin 安装程序需要您的管理员权限来完成安装。\\n\\n请输入您的密码：" default answer "" with icon caution with text buttons {"取消", "安装"} default button "安装" with hidden answer
          set button_pressed to button returned of result
          if button_pressed is equal to "取消" then
            return "CANCELLED"
          else
            set admin_password to text returned of result
            return admin_password
          end if
        end tell
      `;

      // 执行 AppleScript
      const osascriptProcess = child_process.spawn('osascript', ['-e', appleScript], { stdio: ['pipe', 'pipe', 'pipe'] }); // stdio设置为pipe以捕获输出

      let password = '';
      let scriptError = '';

      osascriptProcess.stdout.on('data', (data) => {
        password += data.toString().trim();
      });

      osascriptProcess.stderr.on('data', (data) => {
        scriptError += data.toString();
        logAndConsole('error', `AppleScript 错误输出: ${data.toString()}`);
      });

      osascriptProcess.on('error', (err) => {
        logAndConsole('error', `执行 osascript 失败: ${err.message}`);
        reject(new Error(`无法执行 AppleScript 来获取密码: ${err.message}`));
      });

      osascriptProcess.on('close', (osascriptCode) => {
        if (osascriptCode !== 0) {
          logAndConsole('error', `AppleScript 退出码非零: ${osascriptCode}. 错误: ${scriptError || '未知'}`);
          return reject(new Error(`用户取消或 AppleScript 错误，无法获取密码。`));
        }

        if (password === 'CANCELLED') {
          logAndConsole('warn', '用户取消了密码输入。');
          return reject(new Error('用户取消了安装。'));
        }

        if (!password) {
          logAndConsole('error', '未获取到密码。');
          return reject(new Error('未获取到密码，安装无法继续。'));
        }

        logAndConsole('info', '密码已获取，正在尝试安装...');

        // 使用获取到的密码执行 sudo installer
        // 注意：将密码通过 stdin 传递给 sudo
        const installerProcess = child_process.spawn('sudo', ['-S', 'installer', '-pkg', installerPath, '-target', '/'], {
            // stdio: ['pipe', 'inherit', 'inherit'] 表示：
            // stdin: pipe (用于写入密码)
            // stdout: inherit (安装器的输出直接显示)
            // stderr: inherit (安装器的错误直接显示)
            // shell: true // 如果 sudo 不在 PATH 中，或者需要复杂管道，可能需要
        });

        // 将密码写入 sudo 进程的 stdin
        installerProcess.stdin.write(password + '\n');
        installerProcess.stdin.end(); // 关闭 stdin，表示密码已发送

        installerProcess.on('error', (err) => {
          logAndConsole('error', `安装程序启动失败: ${err.message}`);
          reject(err);
        });

        installerProcess.on('close', async (code) => {
          if (code === 0) {
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
                if (available) { // 如果服务可用
                  clearInterval(interval);
                  resolve();
                }
              } else if (++retries >= maxRetries) {
                clearInterval(interval);
                reject(new Error('安装器未在超时前完成安装或 Oadin 未成功安装。'));
              }
            }, 1000);

          } else {
            reject(new Error(`macOS 安装器退出，退出码: ${code}。安装可能失败，请检查密码或日志。`));
          }
        });
      });
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
  getOadinExecutablePath,
  runInstallerByPlatform,
  isHealthy
};
