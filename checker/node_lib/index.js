const express = require('express');
const http = require('http');
const fs = require('fs');
const path = require('path');
const os = require('os');
const { spawn } = require('child_process');

const webServerPort = 5000;
let userResponseResolve;
let userResponsePromise;

function waitForUserResponse() {
  return new Promise((resolve) => {
    userResponseResolve = resolve;
  });
}

function startWebServer() {
  const app = express();
  app.use(express.json());

  app.get('/install-prompt', (req, res) => {
    res.type('html').send(installPromptHTML);
  });

  app.get('/user-response', (req, res) => {
    const choice = req.query.choice === 'true';
    if (userResponseResolve) {
      userResponseResolve(choice);
      userResponseResolve = null;
    }
    res.sendStatus(200);
  });

  app.listen(webServerPort, () => {
    console.log(`Web server running on http://localhost:${webServerPort}`);
  });
}

function openBrowser(url) {
  let cmd, args;
  const platform = os.platform();
  if (platform === 'win32') {
    cmd = 'cmd';
    args = ['/c', 'start', url];
  } else if (platform === 'darwin') {
    cmd = 'open';
    args = [url];
  } else {
    cmd = 'xdg-open';
    args = [url];
  }
  try {
    spawn(cmd, args, { detached: true, stdio: 'ignore' });
    return true;
  } catch (err) {
    return false;
  }
}

function isByzeAvailable() {
  return new Promise((resolve) => {
    const options = {
      hostname: 'localhost',
      port: 16688,
      path: '/',
      method: 'GET',
      timeout: 3000,
    };
    const req = http.request(options, (res) => {
      resolve(res.statusCode === 200);
    });
    req.on('error', () => resolve(false));
    req.on('timeout', () => {
      req.destroy();
      resolve(false);
    });
    req.end();
  });
}

function downloadByze() {
  return new Promise((resolve) => {
    const url = 'http://120.232.136.73:31619/byzedev/byze.exe';
    const userDir = os.homedir();
    const destDir = path.join(userDir, 'Byze');
    const dest = path.join(destDir, 'byze.exe');
    fs.mkdir(destDir, { recursive: true }, (err) => {
      if (err) return resolve(false);
      const file = fs.createWriteStream(dest);
      http.get(url, (res) => {
        res.pipe(file);
        file.on('finish', () => {
          file.close(() => resolve(true));
        });
      }).on('error', () => resolve(false));
    });
  });
}

function installByze() {
  return new Promise((resolve) => {
    const userDir = os.homedir();
    const byzePath = path.join(userDir, 'Byze', 'byze.exe');
    const child = spawn(byzePath, ['server', 'start', '-d'], { detached: true, stdio: 'ignore' });

    child.on('error', () => resolve(false));
    setTimeout(() => resolve(true), 3000); // wait 3s for Byze server to start
    child.unref();
  });
}

// 执行 `byze import --file path/to/.byze`
function importByzeFile(filePath) {
  return new Promise((resolve) => {
    const userDir = os.homedir();
    const byzePath = path.join(userDir, 'Byze', 'byze.exe');
    const child = spawn(byzePath, ['import', '--file', filePath], { stdio: 'inherit' });

    child.on('exit', (code) => {
      resolve(code === 0);
    });

    child.on('error', () => resolve(false));
  });
}

const installPromptHTML = `
<html>
<body style="padding:20px;font-family:Arial">
    <h2>安装确认</h2>
    <p>需要安装Byze组件才能继续，是否允许？</p>
    <button onclick="respond(true)">同意安装</button>
    <button onclick="respond(false)">取消</button>
    <script>
        function respond(choice) {
            fetch('/user-response?choice=' + choice)
                .then(() => window.close());
        }
    </script>
</body>
</html>
`;

async function ByzeInit(byzeFilePath = path.join(process.cwd(), '.byze')) {
  const available = await isByzeAvailable();
  if (!available) {
    startWebServer();
    openBrowser(`http://localhost:${webServerPort}/install-prompt`);
    try {
      const choice = await Promise.race([
        waitForUserResponse(),
        new Promise((_, reject) => setTimeout(() => reject(new Error('超时')), 5 * 60 * 1000))
      ]);
      if (!choice) return;
      const downloaded = await downloadByze();
      if (!downloaded) return;
      const installed = await installByze();
      if (!installed) return;
    } catch (err) {
      console.error(err.message);
      return;
    }
  }

  console.log('Byze 已启动，导入 .byze 文件...');
  const imported = await importByzeFile(byzeFilePath);
  if (imported) {
    console.log(`成功导入: ${byzeFilePath}`);
  } else {
    console.error(`导入失败: ${byzeFilePath}`);
  }
}

module.exports = { ByzeInit };
