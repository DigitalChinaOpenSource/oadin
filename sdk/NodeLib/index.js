// const express = require('express');
const http = require('http');
const fs = require('fs');
const path = require('path');
const os = require('os');
const axios = require('axios');
const Ajv = require('ajv');
const addFormats = require('ajv-formats');
const EventEmitter = require('events');
const AdmZip = require('adm-zip');
const { spawn } = require('child_process');
const { execSync } = require('child_process');
const { promises: fsPromises } = require("fs");

const schemas = require('./schema.js');

function AddToUserPath(destDir) {
  const isMacOS = process.platform === 'darwin';

  if (isMacOS) {
    try {
      const shell = process.env.SHELL || '';
      let shellConfigName = '.zshrc';
      if (shell.includes('bash')) shellConfigName = '.bash_profile';
      
      const shellConfigPath = path.join(os.homedir(), shellConfigName);
      const exportLine = `export PATH="$PATH:${destDir}"\n`;

      // ensure the config file exists
      if (!fs.existsSync(shellConfigPath)) {
        fs.writeFileSync(shellConfigPath, '');
      }

      // check if the line already exists
      const content = fs.readFileSync(shellConfigPath, 'utf8');
      if (content.includes(exportLine)) {
        console.log('✅ 环境变量已存在');
        return true;
      }

      // append the line to the config file
      fs.appendFileSync(shellConfigPath, `\n${exportLine}`);
      console.log(`✅ 已添加到 ${shellConfigName}，请执行以下命令生效：\nsource ${shellConfigPath}`);
      return true;
    } catch (err) {
      console.error('❌ 添加环境变量失败:', err.message);
      return false;
    }
  } else {
    try {
      const regKey = 'HKCU\\Environment';
      let currentPath = '';

      try {
        const output = execSync(`REG QUERY "${regKey}" /v Path`, { 
          encoding: 'utf-8',
          stdio: ['pipe', 'pipe', 'ignore'] 
        });
        const match = output.match(/REG_EXPAND_SZ\s+(.*)/);
        currentPath = match ? match[1].trim() : '';
      } catch {}

      const paths = currentPath.split(';').filter(p => p);
      if (paths.includes(destDir)) {
        console.log('✅ 环境变量已存在');
        return true;
      }

      const newPath = currentPath ? `${currentPath};${destDir}` : destDir;
      execSync(`REG ADD "${regKey}" /v Path /t REG_EXPAND_SZ /d "${newPath}" /f`, { 
        stdio: 'inherit' 
      });
      
      console.log('✅ 已添加到环境变量，请重新启动应用程序使更改生效');
      return true;
    } catch (error) {
      console.error('❌ 添加环境变量失败:', error.message);
      return false;
    }
  }
}

class Aog {
  version = "aog/v0.3";

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
      return new Error(`Schema validation failed: ${JSON.stringify(validate.errors)}`);
    }
    return data;
  }

  // check if aog.exe is running
  IsAogAvailiable(){
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

  // check if aog.exe is existed
  IsAogExisted() {
    return new Promise((resolve) => {
        const userDir = os.homedir();
        const platform = process.platform;

        let destDir;
        let dest;

        if (platform === 'win32') {
            // Windows PATH
            destDir = path.join(userDir, 'AOG');
            dest = path.join(destDir, 'aog.exe');
        } else if (platform === 'darwin') {
            // macOS PATH
            destDir = path.join(userDir, 'AOG');
            dest = path.join(destDir, 'aog');
        } else {
            console.error('❌ 不支持的操作系统');
            return resolve(false);
        }

        resolve(fs.existsSync(dest));
    });
}

  // download aog.exe from server
  DownloadAog() {
    return new Promise((resolve) => {
      const isMacOS = process.platform === 'darwin';
      const url = isMacOS 
        ? 'http://120.232.136.73:31619/aogdev/aog.zip'
        : 'http://120.232.136.73:31619/aogdev/aog.exe';
      
      const userDir = os.homedir();
      const destDir = path.join(userDir, 'AOG');
      const destFileName = isMacOS ? 'aog.zip' : 'aog.exe';
      const dest = path.join(destDir, destFileName);
  
      fs.mkdir(destDir, { recursive: true }, async (err) => {
        if (err) {
          console.error('❌ 创建目录失败:', err.message);
          return resolve(false);
        }
  
        console.log('🔍 正在下载文件:', url);
        const file = fs.createWriteStream(dest);
        
        const request = http.get(url, (res) => {
          if (res.statusCode !== 200) {
            console.error(`❌ 下载失败，HTTP 状态码: ${res.statusCode}`);
            file.close();
            fs.unlink(dest, () => {});
            return resolve(false);
          }
  
          res.pipe(file);
          file.on('finish', async () => {
            file.close();
            console.log('✅ 下载完成:', dest);
  
            // macOS
            if (isMacOS) {
              try {
                const zip = new AdmZip(dest);
                zip.extractAllTo(destDir, true);
                console.log('✅ 解压完成');
                fs.unlinkSync(dest);
                
                const execPath = path.join(destDir, 'aog');
                if (fs.existsSync(execPath)) {
                  fs.chmodSync(execPath, 0o755);
                }
              } catch (e) {
                console.error('❌ 解压失败:', e.message);
                return resolve(false);
              }
            }
  
            // add to PATH
            const done = await AddToUserPath(destDir);
            resolve(done);
          });
        });
  
        request.on('error', (err) => {
          console.error('❌ 下载失败:', err.message);
          file.close();
          fs.unlink(dest, () => {});
          resolve(false);
        });
      });
    });
  }

  // run aog
  InstallAog() {
    return new Promise((resolve) => {
      const isMacOS = process.platform === 'darwin';
      const userDir = os.homedir();
      const aogDir = path.join(userDir, 'Aog');
  
      // ensure aog.exe is in PATH
      if (!process.env.PATH.includes(aogDir)) {
        process.env.PATH = `${process.env.PATH}${path.delimiter}${aogDir}`;
      }
  
      const child = spawn('aog', ['server', 'start', '-d'], {
        stdio: 'ignore',
        windowsHide: true
      });
      child.unref();
  
      child.on('error', (err) => {
        console.error(`❌ 启动失败: ${err.message}`);
        if (err.code === 'ENOENT') {
          console.log([
            '💡 可能原因:',
            `1. 未找到aog可执行文件，请检查下载是否成功`,
            `2. 环境变量未生效，请尝试重启终端`
          ].filter(Boolean).join('\n'));
        }
        resolve(false);
      });

      const checkServer = (attempt = 1) => {
        const req = http.request({
          hostname: 'localhost',
          port: 16688,
          method: 'GET',
          timeout: 5000
        }, (res) => {
          if (res.statusCode === 200) {
            console.log('✅ 服务已就绪');
            resolve(true);
          } else {
            console.log(`⚠️ 服务响应异常: HTTP ${res.statusCode}`);
            if (attempt < 3) setTimeout(() => checkServer(attempt + 1), 2000);
            else resolve(false);
          }
        });
  
        req.on('error', () => {
          console.log(`⌛ 检测尝试 ${attempt}/3`);
          if (attempt < 3) setTimeout(() => checkServer(attempt + 1), 2000);
          else resolve(false);
        });
  
        req.on('timeout', () => {
          console.log(`⏳ 检测超时 ${attempt}/3`);
          req.destroy();
          if (attempt < 3) setTimeout(() => checkServer(attempt + 1), 2000);
          else resolve(false);
        });
  
        req.end();
      };
  
      setTimeout(() => checkServer(1), 5000);
      child.unref();
    });
  }

  // get services
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

  // install service
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

  // update service
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

  // get models
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

  // stream install model
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

        return eventEmitter;
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message || '请求失败',
        data: null,
      }
    }
}

  // cancel install model
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

  // delete model
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

  // Install service provider
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

  // update service provider
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

  // delete service provider
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

  // import .aog config
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

  // export .aog config
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
      const userDir = os.homedir();
      const destDir = path.join(userDir, 'AOG');
      const dest = path.join(destDir, '.aog');

      fs.mkdir(destDir, { recursive: true }, (err) => {
          if (err) {
              console.error(`创建目录失败: ${err.message}`);
              return;
          }

          const fileContent = JSON.stringify(res.data, null, 2);

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

  // get models from engine （deprecated）
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

  // get models recommended
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

  // get models supported
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

  // get models supported from smartvision
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

  // chat
  async Chat(data) {
    try {
      this.validateSchema(schemas.chatRequest, data);

      // wheather to use stream
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

        return eventEmitter;
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


  // Generate
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
          eventEmitter.emit('end');
        });
  
        return eventEmitter;
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
  
  // text to image
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

  // embed
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

  // 用于一键安装 AOG 和 导入配置
  async AogInit(path){
    const isAogAvailable = await this.IsAogAvailiable();
    if (isAogAvailable) {
      console.log('✅ AOG 服务已启动，跳过安装。');
      return true;
    }
    
    const isAogExisted = await this.IsAogExisted();
    if (!isAogExisted) {
      const downloadSuccess = await this.DownloadAog();
      if (!downloadSuccess) {
        console.error('❌ 下载 AOG 失败，请检查网络连接或手动下载。');
        return false;
      }
    } else {
      console.log('✅ AOG 已存在，跳过下载。');
    }

    const installSuccess = await this.InstallAog();
    if (!installSuccess) {
      console.error('❌ 启动 AOG 服务失败，请检查配置或手动启动。');
      return false;
    } else {
      console.log('✅ AOG 服务已启动。');
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

module.exports = Aog;