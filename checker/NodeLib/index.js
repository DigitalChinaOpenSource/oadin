// const express = require('express');
const http = require('http');
const https = require('https');
const fs = require('fs');
const path = require('path');
const os = require('os');
const axios = require('axios');
const Ajv = require('ajv');
const addFormats = require('ajv-formats');
const EventEmitter = require('events');
const { execFile, spawn } = require('child_process');
const { promises: fsPromises } = require("fs");

const schemas = require('./schema.js');

function runInstaller(installerPath, isMacOS) {
  return new Promise((resolve, reject) => {
    if (isMacOS) {
      // 打开 GUI 安装器
      const child = spawn('open', [installerPath], { stdio: 'ignore', detached: true });

      child.on('error', reject);

      // 等待安装目录生成（轮询）
      const expectedPath = '/Applications/Byze.app'; // 你需要根据实际情况修改
      const maxRetries = 30; // 最多等 60 次（大约 60 秒）
      let retries = 0;

      const interval = setInterval(() => {
        if (fs.existsSync(expectedPath)) {
          clearInterval(interval);
          resolve();
        } else if (++retries >= maxRetries) {
          clearInterval(interval);
          reject(new Error('安装器未在超时前完成安装'));
        }
      }, 1000);
    } else {
      // Windows 安装器
      const child = spawn(installerPath, ['/S'], { stdio: 'inherit' });

      child.on('error', reject);
      child.on('close', (code) => {
        code === 0 ? resolve() : reject(new Error(`Installer exited with code ${code}`));
      });
    }
  });
}

class Byze {
  version = "byze/v0.2";

  constructor(version) {
    this.client = axios.create({
      baseURL: `http://localhost:16688/${this.version}`,
      headers: {"Content-Type": "application/json" },
    })
    this.ajv = new Ajv();
    addFormats(this.ajv);
  }

  async validateSchema(schema, data) {
    if (!data || Object.keys(data).length === 0) {
      return data;
    }
  
    const validate = this.ajv.compile(schema);
    if (!validate(data)) {
      throw new Error(`Schema validation failed: ${JSON.stringify(validate.errors)}`);
    }
    return data;
  }

  // 检查 Byze 服务是否启动
  IsByzeAvailiable() {
    return new Promise((resolve) => {
      const checkPort = (port) => {
        return new Promise((resolvePort) => {
          const options = {
            hostname: 'localhost',
            port: port,
            path: '/',
            method: 'GET',
            timeout: 3000,
          };
          const req = http.request(options, (res) => {
            resolvePort(res.statusCode === 200);
          });
          req.on('error', () => resolvePort(false));
          req.on('timeout', () => {
            req.destroy();
            resolvePort(false);
          });
          req.end();
        });
      };

      // 同时检查 16688 和 16677 两个端口
      Promise.all([checkPort(16688), checkPort(16677)]).then((results) => {
        resolve(results.every((status) => status)); // 两个端口都可用时返回 true
      });
    });
  }

  // 检查用户目录是否存在 Byze.exe
  IsByzeExisted() {
    return new Promise((resolve) => {
        const userDir = os.homedir();
        const platform = process.platform; // 获取当前平台

        let destDir;
        let dest;

        if (platform === 'win32') {
            // Windows 平台路径
            destDir = path.join(userDir, 'Byze');
            dest = path.join(destDir, 'byze.exe');
        } else if (platform === 'darwin') {
            // macOS 平台路径
            dest = '/usr/local/bin/byze';
        } else {
            console.error('❌ 不支持的操作系统');
            return resolve(false);
        }

        resolve(fs.existsSync(dest));
    });
  }

  DownloadByze() {
    return new Promise((resolve, reject) => {
      const isMacOS = process.platform === 'darwin';
      const url = isMacOS
        ? 'https://oss-aipc.dcclouds.com/byze/releases/macos/byze-installer-latest.pkg'
        : 'https://oss-aipc.dcclouds.com/byze/releases/windows/byze-installer-latest.exe';

      const userDir = os.homedir();
      const destDir = path.join(userDir, 'ByzeInstaller');
      const destFileName = isMacOS ? 'byze-installer-latest.pkg' : 'byze-installer-latest.exe';
      const dest = path.join(destDir, destFileName);

      // 创建目录
      fs.mkdirSync(destDir, { recursive: true });

      // 下载文件
      const file = fs.createWriteStream(dest);
      https.get(url, (response) => {
        if (response.statusCode !== 200) {
          return reject(new Error(`Download failed with status code: ${response.statusCode}`));
        }

        response.pipe(file);

        file.on('finish', () => {
          file.close(async () => {
            // 安装执行
            try {
              await runInstaller(dest, isMacOS);
              resolve(true);
            } catch (err) {
              console.error('Installer execution failed:', err);
              resolve(false);
            }
          });
        });
      }).on('error', (err) => {
        fs.unlinkSync(dest);
        reject(err);
      });
    });
  }

  // 启动 Byze 服务
  InstallByze() {
    return new Promise((resolve, reject) => {
      const currentPlatform = process.platform;

      if (currentPlatform === 'win32') {
        const batPath = path.resolve(__dirname, 'start-byze.bat');
        execFile('cmd.exe', ['/c', batPath], { windowsHide: true }, (error, stdout, stderr) => {
          const output = (stdout + stderr).toString().toLowerCase();
          if (error || output.includes('error')) return resolve(false);
          if (output.includes('byze server start successfully')) return resolve(true);
          return resolve(false);
        });
      } else if (currentPlatform === 'darwin') {
        // macOS: 使用 nohup + & 启动 byze，不依赖父进程
        const shellCommand = `nohup byze server start -d > /dev/null 2>&1 & echo "Byze launched"`;

        execFile('sh', ['-c', shellCommand], (error, stdout, stderr) => {
          const output = (stdout + stderr).toString().toLowerCase();
          if (error || output.includes('error')) return resolve(false);

          // 若需严格确认成功，可以考虑加延时检测
          return resolve(true);
        });
      } else {
        return reject(new Error(`Unsupported platform: ${currentPlatform}`));
      }
    });
  }


  // 执行 byze install chat
  InstallChat(remote = null) {
    return new Promise((resolve) => {
      const userDir = os.homedir();
      const byzePath = path.join(userDir, 'Byze', 'byze.exe');
      process.env.PATH = `${process.env.PATH};${byzePath}`;

      const child = spawn(byzePath, ['install', 'chat'], { detached: true, stdio: [ 'pipe', 'pipe', 'pipe'] });

      child.stdout.on('data', (data) => {
        console.log(`stdout: ${data}`);

        if (data.toString().includes('(y/n)')) {
          if (remote) {
            child.stdin.write('${autoAnswer}\n');
          } else {
            child.stdin.write('n\n');
          }
        }
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

  checkServerStatus(resolve) {
    const options = {
      hostname: 'localhost',
      port: 16688,
      path: '/',
      method: 'GET',
      timeout: 3000,
    };

    const req = http.request(options, (res) => {
      if (res.statusCode === 200) {
        console.log('✅ Byze 服务已启动');
        resolve(true);
      } else {
        console.error(`❌ Byze 服务未启动，HTTP 状态码: ${res.statusCode}`);
        resolve(false);
      }
    });
    req.on('error', (err) => {
      console.error(`❌ Byze 服务未启动: ${err.message}`);
      resolve(false);
    });
    req.end();
  }

  // 查看当前服务
  async GetServices() {
    try {
      const res = await this.client.get('/service');
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      }
      await this.validateSchema(schemas.getServicesSchema, res.data);
      return {
        code: 200,
        msg: res.data.message || null,
        data: res.data.data,
      };
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message,
        data: null,
      };
    }
  }

  // 创建新服务
  async InstallService(data) {
    try {
      this.validateSchema(schemas.installServiceRequestSchema, data);
      const res = await this.client.post('/service', data);
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null
        };
      }
      await this.validateSchema(schemas.ResponseSchema, res.data);
      return {
        code: 200,
        msg: res.data.message || null,
        data: null
      };
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message || '请求失败',
        data: null,
      };
    }
  }

  // 更新服务
  async UpdateService(data) {
    try {
      this.validateSchema(schemas.updateServiceRequestSchema, data);
      const res = await this.client.put('/service', data);
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      }
      await this.validateSchema(schemas.ResponseSchema, res.data);
      return {
        code: 200,
        msg: res.data.message || null,
        data: null
      };
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message || '请求失败',
        data: null,
      };
    }
  }

  // 查看模型
  async GetModels() {
    try {
      const res = await this.client.get('/model');
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      }
      await this.validateSchema(schemas.getModelsSchema, res.data);
      return {
        code: 200,
        msg: res.data.message || null,
        data: res.data.data,
      };
    } catch (error){    
      return {
        code: 400,
        msg: error.response?.data?.message || error.message,
        data: null,
      };
    }
  }

  // 安装模型
  async InstallModel(data) {
    try {
      this.validateSchema(schemas.installModelRequestSchema, data);
      const res = await this.client.post('/model', data);
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      }
      await this.validateSchema(schemas.ResponseSchema, res.data);
      return {
        code: 200,
        msg: res.data.message || null,
        data: null
      };
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message || '请求失败',
        data: null,
      };
    }
  }

  // 安装模型（流式）
  async InstallModelStream(data) {
    try {
      this.validateSchema(schemas.installModelRequestSchema, data);
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message || '请求失败',
        data: null,
      };
    }

    const config = { responseType: 'stream' };
    try {
        const res = await this.client.post('/model/stream', data, config);
        const eventEmitter = new EventEmitter();

        res.data.on('data', (chunk) => {
            try {
              // 解析流数据
              const rawData = chunk.toString().trim();
              const jsonString = rawData.startsWith('data:') ? rawData.slice(5) : rawData;
              const response = JSON.parse(jsonString);

              // 触发事件，传递解析后的数据
              eventEmitter.emit('data', response);

              // 如果状态为 "success"，触发完成事件
              if (response.status === 'success') {
                eventEmitter.emit('end', response);
              }

              if (response.status === 'canceled') {
                eventEmitter.emit('canceled', response);
              }

              if (response.status === 'error') {
                eventEmitter.emit('end', response);
              }

            } catch (err) {
              eventEmitter.emit('error', `解析流数据失败: ${err.message}`);
            }
        });

        res.data.on('error', (err) => {
          eventEmitter.emit('error', `流式响应错误: ${err.message}`);
        });

        // res.data.on('end', () => {
        //     eventEmitter.emit('end'); // 触发结束事件
        // });

        return eventEmitter; // 返回 EventEmitter 实例
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message || '请求失败',
        data: null,
      }
    }
}

  // 取消安装模型（流式）
  async CancelInstallModel(data) {
    try {
      this.validateSchema(schemas.cancelModelStreamRequestSchema, data);
      const res = await this.client.post('/model/stream/cancel', data);
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      }
      return {
        code: 200,
        msg: res.data.message || null,
        data: null
      };
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message || '请求失败',
        data: null,
      };
    }
  }

  // 卸载模型
  async DeleteModel(data) {
    try {
      this.validateSchema(schemas.deleteModelRequestSchema, data);
      const res = await this.client.delete('/model', { data });
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      }
      await this.validateSchema(schemas.ResponseSchema, res.data);
      return {
        code: 200,
        msg: res.data.message || null,
        data: null
      };
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message || '请求失败',
        data: null,
      };
    }
  }

  // 查看服务提供商
  async GetServiceProviders() {
    try {
      const res = await this.client.get('/service_provider');
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      }
      await this.validateSchema(schemas.getServiceProvidersSchema, res.data);
      return {
        code: 200,
        msg: res.data.message || null,
        data: res.data.data,
      };
    } catch (error){    
      return {
        code: 400,
        msg: error.response?.data?.message || error.message,
        data: null,
      };
    }
  }

  // 新增服务提供商
  async InstallServiceProvider(data) {
    try {
      this.validateSchema(schemas.installServiceProviderRequestSchema, data);
      const res = await this.client.post('/service_provider', data);
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      }
      await this.validateSchema(schemas.ResponseSchema, res.data);
      return {
        code: 200,
        msg: res.data.message || null,
        data: null,
      };
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message || '请求失败',
        data: null,
      };
    }
  }

  // 更新服务提供商
  async UpdateServiceProvider(data) {
    try {
      this.validateSchema(schemas.updateServiceProviderRequestSchema, data);
      const res = await this.client.put('/service_provider', data);
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      }
      await this.validateSchema(schemas.ResponseSchema, res.data);
      return {
        code: 200,
        msg: res.data.message || null,
        data: null,
      };
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message || '请求失败',
        data: null,
      };
    }
  }

  // 删除服务提供商
  async DeleteServiceProvider(data) {
    try {
      this.validateSchema(schemas.deleteServiceProviderRequestSchema, data);
      const res = await this.client.delete('/service-provider', { data });
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      }
      await this.validateSchema(schemas.ResponseSchema, res.data);
      return {
        code: 200,
        msg: res.data.message || null,
        data: null,
      };
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message || '请求失败',
        data: null,
      };
    }
  }

  // 导入配置文件
  async ImportConfig(path) {
    try {
      const data = await fsPromises.readFile(path, 'utf8');
      const res = await this.client.post('/service/import', data);
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      }
      await this.validateSchema(schemas.ResponseSchema, res.data);
      return {
        code: 200,
        msg: res.data.message || null,
        data: null,
      };
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message || '请求失败',
        data: null,
      };
    }
  }

  // 导出配置文件
  async ExportConfig(data = {}) {
    try{
      this.validateSchema(schemas.exportRequestSchema, data);
      const res = await this.client.post('/service/export', data);
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      }
      // 将响应数据存入 .byze 文件
      const userDir = os.homedir();
      const destDir = path.join(userDir, 'Byze');
      const dest = path.join(destDir, '.byze');

      // 确保目录存在并写入文件
      fs.mkdir(destDir, { recursive: true }, (err) => {
          if (err) {
              console.error(`创建目录失败: ${err.message}`);
              return;
          }

          // 将响应数据序列化为 JSON 字符串
          const fileContent = JSON.stringify(res.data, null, 2); // 格式化为易读的 JSON

          fs.writeFile(dest, fileContent, (err) => {
              if (err) {
                  console.error(`写入文件失败: ${err.message}`);
                  return;
              }
              console.log(`已将生成文件写入到 ${dest}`);
          });
      });

      return {
        code: 200,
        msg: res.data.message || null,
        data: res.data.data,
      };
    } catch (error){    
      return {
        code: 400,
        msg: error.response?.data?.message || error.message,
        data: null,
      };
    }
  }

  // 获取模型列表 （已弃用）
  async GetModelsAvailiable() {
    try {
      const res = await this.client.get('/services/models');
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || null,
        }
      }
      this.validateSchema(schemas.modelsResponse, res.data);
    } catch (error) {
      return { status: 0, err_msg: `获取模型列表失败: ${error.message}`, data: null };
    }
  }

  // 获取推荐模型列表
  async GetModelsRecommended() {
    try {
      const res = await this.client.get('/model/recommend');
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      }
      await this.validateSchema(schemas.recommendModelsResponse, res.data);
      return {
        code: 200,
        msg: res.data.message || null,
        data: res.data.data,
      };
    } catch (error){    
      return {
        code: 400,
        msg: error.response?.data?.message || error.message,
        data: null,
      };
    }
  }

  // 获取支持模型列表
  async GetModelsSupported(data) {
    try {
      this.validateSchema(schemas.getModelsSupported, data);
      const res = await this.client.get('/model/support', { params: data });
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      }
      await this.validateSchema(schemas.recommendModelsResponse, res.data);
      return {
        code: 200,
        msg: res.data.message || null,
        data: res.data.data,
      };
    } catch (error){    
      return {
        code: 400,
        msg: error.response?.data?.message || error.message,
        data: null,
      };
    }
  }

  // 获取问学平台支持模型列表
  async GetSmartvisionModelsSupported(data) {
    try {
      this.validateSchema(schemas.SmartvisionModelSupportRequest, data);
      const res = await this.client.get('/model/support/smartvision', { params: data });
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      }
      await this.validateSchema(schemas.SmartvisionModelSupport, res.data);
      return {
        code: 200,
        msg: res.data.message || null,
        data: res.data.data,
      };
    } catch (error){    
      return {
        code: 400,
        msg: error.response?.data?.message || error.message,
        data: null,
      };
    }
  }

  // chat服务
  async Chat(data) {
    try {
      this.validateSchema(schemas.chatRequest, data);

      // 判断是否是流式
      const config = { responseType: data.stream ? 'stream' : 'json' };
      const res = await this.client.post('/services/chat', data, config);
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      };

      if (data.stream) {
        const eventEmitter = new EventEmitter();

        res.data.on('data', (chunk) => {
          try {
            const rawData = chunk.toString().trim();
            const jsonString = rawData.startsWith('data:') ? rawData.slice(5) : rawData;
            const response = JSON.parse(jsonString);
            eventEmitter.emit('data', response); // 触发事件，实时传输数据
          } catch (err) {
            eventEmitter.emit('error', `解析流数据失败: ${err.message}`);
          }
        });

        res.data.on('error', (err) => {
          eventEmitter.emit('error', `流式响应错误: ${err.message}`);
        });

        res.data.on('end', () => {
          eventEmitter.emit('end'); // 触发结束事件
        });

        return eventEmitter; // 返回 EventEmitter 实例
      } else {
        // 非流式响应处理
        await this.validateSchema(schemas.chatResponse, res.data);
        return {
          code: 200,
          msg: res.data.message || null,
          data: res.data,
        };
      };
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message,
        data: null,
      };
    }
  }


  // 生文服务
  async Generate(data) {
    try {
      this.validateSchema(schemas.generateRequest, data);
  
      const config = { responseType: data.stream ? 'stream' : 'json' };
      const res = await this.client.post('/services/generate', data, config);
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      }
  
      if (data.stream) {
        const eventEmitter = new EventEmitter();
  
        res.data.on('data', (chunk) => {
          try {
            const response = JSON.parse(chunk.toString());
            if (response) {
              this.validateSchema(schemas.generateResponse, response);
              eventEmitter.emit('data', response.response); // 逐步传输响应内容
            }
          } catch (err) {
            eventEmitter.emit('error', `解析流数据失败: ${err.message}`);
          }
        });
  
        res.data.on('error', (err) => {
          eventEmitter.emit('error', `流式响应错误: ${err.message}`);
        });
  
        res.data.on('end', () => {
          eventEmitter.emit('end'); // 触发结束事件
        });
  
        return eventEmitter; // 返回 EventEmitter 实例
      } else {
        // 非流式响应处理
        await this.validateSchema(schemas.generateResponse, res.data);
        return {
          code: 200,
          msg: res.data.message || null,
          data: res.data,
        };
      }
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message,
        data: null,
      };
    }
  }
  
  // 生图服务
  async TextToImage(data) {
    try {
      this.validateSchema(schemas.textToImageRequest, data);
      const res = await this.client.post('/services/text-to-image', data);
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      };
      await this.validateSchema(schemas.textToImageResponse, res.data);
      return {
        code: 200,
        msg: res.data.message || null,
        data: res.data,
      };
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message,
        data: null,
      };
    }
  }

  // embed服务
  async Embed(data) {
    try {
      this.validateSchema(schemas.embeddingRequest, data);
      const res = await this.client.post('/services/embed', data);
      if (res.status !== 200) {
        return {
          code: 400,
          msg: res.data?.message || 'Bad Request',
          data: null,
        };
      };
      await this.validateSchema(schemas.embeddingResponse, res.data);
      return {
        code: 200,
        msg: res.data.message || null,
        data: res.data,
      };
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message,
        data: null,
      };
    }
  }

  // 用于一键安装 Byze 和 导入配置
  async ByzeInit(path){
    const isByzeAvailable = await this.IsByzeAvailiable();
    if (isByzeAvailable) {
      console.log('✅ Byze 服务已启动，跳过安装。');
      return true;
    }
    
    const isByzeExisted = await this.IsByzeExisted();
    if (!isByzeExisted) {
      const downloadSuccess = await this.DownloadByze();
      if (!downloadSuccess) {
        console.error('❌ 下载 Byze 失败，请检查网络连接或手动下载。');
        return false;
      }
    } else {
      console.log('✅ Byze 已存在，跳过下载。');
    }

    const installSuccess = await this.InstallByze();
    if (!installSuccess) {
      console.error('❌ 启动 Byze 服务失败，请检查配置或手动启动。');
      return false;
    } else {
      console.log('✅ Byze 服务已启动。');
    }

    const importSuccess = await this.ImportConfig(path);
    if (!importSuccess) {
      console.error('❌ 导入配置文件失败。');
      return false;
    } else {
      console.log('✅ 配置文件导入成功。');
    }
  }
}

module.exports = Byze;