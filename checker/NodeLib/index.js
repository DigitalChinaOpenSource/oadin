// const express = require('express');
const http = require('http');
const fs = require('fs');
const path = require('path');
const os = require('os');
const axios = require('axios');
const Ajv = require('ajv');
const { spawn } = require('child_process');

const schemas = require('./schemas.json');

class Byze {
  version = "byze/v0.2";

  constructor(version) {
    this.client = axios.create({
      baseURL: `http://localhost:16688/${this.version}`,
      headers: {"Content-Type": "application/json" },
    })
    this.ajv = new Ajv();
  }

  async validateSchema(schema, data) {
    const validate = this.ajv.compile(schema);
    if(!validate(data)) {
      throw new Error(`Schema validation failed: ${JSON.stringify(validate.errors)}`);
    }
    return data;
  }

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
        const dest = path.join(destDir, 'byze.exe');
        resolve(fs.existsSync(dest));
    });
  }

  // 从服务器下载 Byze.exe
  DownloadByze() {
    return new Promise((resolve) => {
      const url = 'http://120.232.136.73:31619/Byzedev/byze.exe';
      const userDir = os.homedir();
      const destDir = path.join(userDir, 'Byze');
      const dest = path.join(destDir, 'byze.exe');

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
      process.env.PATH = `${process.env.PATH};${byzePath}`;
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
            console.error('Byze 服务未启动');
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

  // 执行 byze install chat
  InstallChat() {
    return new Promise((resolve) => {
      const userDir = os.homedir();
      const byzePath = path.join(userDir, 'Byze', 'byze.exe');
      process.env.PATH = `${process.env.PATH};${byzePath}`;

      const child = spawn(byzePath, ['install', 'chat'], { detached: true });

      child.stdout.on('data', (data) => {
        console.log(`stdout: ${data}`);
      });

      child.on('close', (code) => {
        if (code === 0) {
          console.log('安装 Byze 聊天插件成功');
          resolve(true);
        } else {
          console.error(`安装 Byze 聊天插件失败，退出码: ${code}`);
          resolve(false);
        }
      });

      child.on('error', (err) => {
        console.error(`启动 Byze 安装命令失败: ${err.message}`);
        resolve(false);
      });

      child.unref();
    });
  }

  // 获取当前服务
  async GetServices() {
    const res = await this.client.get('/service');
    return this.validateSchema(schemas.getServicesSchema, res.data);
  }

  // 创建新服务
  async CreateService(data) {
    this.validateSchema(schemas.installServiceRequestSchema, data);
    const res = await this.client.post('/service', data);
    return this.validateSchema(schemas.ResponseSchema, res.data);
  }
  
  // 更新服务
  async UpdateService(data) {
    this.validateSchema(schemas.updateServiceRequestSchema, data);
    const res = await this.client.put('/service', data);
    return this.validateSchema(schemas.ResponseSchema, res.data);
  }

  ByzeInit(){

  }
}

module.exports = Byze;