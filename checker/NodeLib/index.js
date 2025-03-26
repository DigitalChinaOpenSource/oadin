// const express = require('express');
const http = require('http');
const fs = require('fs');
const path = require('path');
const os = require('os');
const { spawn } = require('child_process');

class Byze {
  // 检查 Byze 服务是否启动
  IsByzeAvailiable(){
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

  // 检查用户目录是否存在 Byze.exe
  IsByzeExisted() {
    return new Promise((resolve) => {
        const userDir = os.homedir();
        const destDir = path.join(userDir, 'Byze');
        const dest = path.join(destDir, 'Byze.exe');
        resolve(fs.existsSync(dest));
    });
  }

  // 从服务器下载 Byze.exe
  DownloadByze() {
    return new Promise((resolve) => {
      const url = 'http://120.232.136.73:31619/Byzedev/Byze.exe';
      const userDir = os.homedir();
      const destDir = path.join(userDir, 'Byze');
      const dest = path.join(destDir, 'Byze.exe');

      fs.mkdir(destDir, { recursive: true }, (err) => {
        if (err) return resolve(false);

        const file = fs.createWriteStream(dest);
        http.get(url, (res) => {
          res.pipe(file);
          file.on('finish', () => {
            file.close(() => {
            // 下载完成后添加到环境变量
              const { exec } = require('child_process');
              const command = `setx PATH "%PATH%;${destDir}"`;

              exec('reg query "HKCU\\Environment" /v PATH', (error, stdout) => {
                  if (error) {
                      console.error(`查询 PATH 失败: ${error.message}`);
                      return resolve(false);
                  }
              
                  const match = stdout.match(/PATH\s+REG_SZ\s+(.+)/);
                  const currentPath = match ? match[1].trim() : "";
              
                  // 避免重复添加
                  if (!currentPath.includes(destDir)) {
                      const updatedPath = `${currentPath};${destDir}`;
              
                      // 安全追加 PATH
                      exec(`reg add "HKCU\\Environment" /v PATH /t REG_SZ /d "${updatedPath}" /f`, (error) => {
                          if (error) {
                              console.error(`更新 PATH 失败: ${error.message}`);
                              return resolve(false);
                          }
                          console.log("✅ 环境变量已成功添加！");
                          return resolve(true);
                      });
                  } else {
                      console.log("✅ 该路径已存在于 PATH 中，无需重复添加。");
                      return resolve(true);
                  }
              });
            });
          });
        }).on('error', () => resolve(false));
      });
    });
  }

  // 启动 Byze 服务
  InstallByze() {
    return new Promise((resolve) => {
      const userDir = os.homedir();
      const byzePath = path.join(userDir, 'Byze', 'byze.exe');

      const child = spawn(byzePath, ['server', 'start', '-d'], { detached: true, stdio: 'ignore' });

      child.on('error', (err) => {
        console.error(`启动 Byze 服务失败: ${err.message}`);
        return resolve(false);
      });

      const checkServer = () => {
        const options = {
            hostname: 'localhost',
            port: 16688,
            path: '/',
            method: 'GET',
            timeout: 3000,
        };

        const req = http.request(options, (res) => {
            if (res.statusCode === 200) {
                console.log('Byze 服务启动成功，端口正常');
                resolve(true);
            } else {
                console.error(`Byze 服务启动失败，返回状态码: ${res.statusCode}`);
                resolve(false);
            }
        });

        req.on('error', () => {
            console.error('Byze 服务未启动，端口不可用');
            resolve(false);
        });

        req.on('timeout', () => {
            console.error('检查 Byze 服务超时');
            req.destroy();
            resolve(false);
        });

        req.end();
      };
      setTimeout(checkServer, 3000);
      child.unref();
    });
  }

  ByzeInit(){

  }
}

module.exports = Byze;