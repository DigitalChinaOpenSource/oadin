// const express = require('express');
const http = require('http');
const fs = require('fs');
const path = require('path');
const os = require('os');
const axios = require('axios');
const Ajv = require('ajv');
const addFormats = require('ajv-formats');
const EventEmitter = require('events');
const { spawn } = require('child_process');
const { exec } = require('child_process');

const schemas = require('./schema.js');

function AddToUserPath(newPath) {
  return new Promise((resolve, reject) => {
    if (!path.isAbsolute(newPath)) {
      console.error('请使用绝对路径');
      return resolve(false);
    }

    // 查询当前用户 PATH
    exec('reg query "HKCU\\Environment" /v PATH', (err, stdout, stderr) => {
      let currentPath = '';

      if (!err) {
        const match = stdout.match(/PATH\s+REG_SZ\s+(.*)/);
        if (match) {
          currentPath = match[1].trim();
        }
      } else if (err.code !== 1) {
        console.error('读取注册表出错：', stderr || err.message);
        return resolve(false);
      }

      // 检查是否已包含
      const pathList = currentPath.split(';').map(p => p.trim());
      if (pathList.includes(newPath)) {
        console.log('该路径已存在于 PATH 中，无需添加。');
        return resolve(true);
      }

      // 拼接新的 PATH
      const newFullPath = currentPath
        ? `${currentPath};${newPath}`
        : newPath;

      // 写入注册表
      const command = `reg add "HKCU\\Environment" /v PATH /d "${newFullPath}" /f`;
      exec(command, (err, stdout, stderr) => {
        if (err) {
          console.error('写入注册表失败：', stderr || err.message);
          return resolve(false);
        }

        console.log(`✅ 成功将路径添加到用户 PATH：${newPath}`);
        resolve(true);
      });
    });
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
      // 如果 data 为空或是空对象，跳过验证
      return data;
    }
  
    const validate = this.ajv.compile(schema);
    if (!validate(data)) {
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
      const url = 'http://120.232.136.73:31397/browser/byzedev/byze.exe';
      const userDir = os.homedir();
      const destDir = path.join(userDir, 'Byze');
      const dest = path.join(destDir, 'byze.exe');
  
      fs.mkdir(destDir, { recursive: true }, async (err) => {
        if (err) {
          console.error('❌ 创建目录失败:', err.message);
          return resolve(false);
        }
  
        console.log('🔍 正在下载文件:', url);
        const file = fs.createWriteStream(dest);
        const request = http.get(url, (res) => {
          // 检查 HTTP 响应状态码
          if (res.statusCode !== 200) {
            console.error(`❌ 下载失败，HTTP 状态码: ${res.statusCode}`);
            file.close();
            fs.unlink(dest, () => {}); // 删除已创建的空文件
            return resolve(false);
          }
  
          res.pipe(file);
          file.on('finish', async () => {
            file.close();
            console.log('✅ 下载完成:', dest);
  
            // 下载完成后添加到环境变量
            const done = await AddToUserPath(destDir);
            resolve(done);
          });
        });
  
        request.on('error', (err) => {
          console.error('❌ 下载失败:', err.message);
          file.close();
          fs.unlink(dest, () => {}); // 删除已创建的空文件
          resolve(false);
        });
      });
    });
  }

  // 启动 Byze 服务
  InstallByze() {
    return new Promise((resolve) => {
      const userDir = os.homedir();
      const byzePath = path.join(userDir, 'Byze', 'byze.exe');
      process.env.PATH = `${process.env.PATH};${path.dirname(byzePath)}`;
      const child = spawn('byze', ['server', 'start', '-d'], {
        detached: true, 
        stdio: 'ignore',
        windowsHide: true, // 隐藏窗口  
      });

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
          timeout: 3000, // 超时时间为 3 秒
        };
      
        let isResolved = false; // 添加标志变量，防止重复触发
      
        const req = http.request(options, (res) => {
          if (!isResolved) {
            isResolved = true; // 标记请求已完成
            if (res.statusCode === 200) {
              console.log('Byze 服务启动成功，端口正常');
              resolve(true);
            } else {
              console.error(`Byze 服务启动失败，返回状态码: ${res.statusCode}`);
              resolve(false);
            }
          }
        });
      
        req.on('error', () => {
          if (!isResolved) {
            isResolved = true; // 标记请求已完成
            console.error('Byze 服务未启动');
            resolve(false);
          }
        });
      
        req.on('timeout', () => {
          if (!isResolved) {
            isResolved = true; // 标记请求已完成
            console.error('检查 Byze 服务超时');
            req.destroy();
            resolve(false);
          }
        });
      
        req.end();
      };
      setTimeout(checkServer, 3000);
      child.unref();
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

  // 查看当前服务
  async GetServices() {
    const res = await this.client.get('/service');
    return this.validateSchema(schemas.getServicesSchema, res.data);
  }

  // 创建新服务
  async InstallService(data) {
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

  // 查看模型
  async GetModels() {
    const res = await this.client.get('/model');
    return this.validateSchema(schemas.getModelsSchema, res.data);
  }

  // 安装模型
  async InstallModel(data) {
    this.validateSchema(schemas.installModelRequestSchema, data);
    const res = await this.client.post('/model', data);
    return this.validateSchema(schemas.ResponseSchema, res.data);
  }

  // 卸载模型
  async DeleteModel(data) {
    this.validateSchema(schemas.deleteModelRequestSchema, data);
    const res = await this.client.delete('/model', { data });
    return this.validateSchema(schemas.ResponseSchema, res.data);
  }

  // 查看服务提供商
  async GetServiceProviders() {
    const res = await this.client.get('/service_provider');
    return this.validateSchema(schemas.getServiceProvidersSchema, res.data);
  }

  // 新增服务提供商
  async InstallServiceProvider(data) {
    this.validateSchema(schemas.installServiceProviderRequestSchema, data);
    const res = await this.client.post('/service_provider', data);
    return this.validateSchema(schemas.ResponseSchema, res.data);
  }

  // 更新服务提供商
  async UpdateServiceProvider(data) {
    this.validateSchema(schemas.updateServiceProviderRequestSchema, data);
    const res = await this.client.put('/service_provider', data);
    return this.validateSchema(schemas.ResponseSchema, res.data);
  }

  // 删除服务提供商
  async DeleteServiceProvider(data) {
    this.validateSchema(schemas.deleteServiceProviderRequestSchema, data);
    const res = await this.client.delete('/service-provider', { data });
    return this.validateSchema(schemas.ResponseSchema, res.data);
  }

  // 导入配置文件
  async ImportConfig(path) {
    const data = fs.readFile(path, 'utf8', (err, data) => { 
      if (err) {
        console.error(err);
        return;
      }
      return data;
    });
    const res = await this.client.post('/service/import', data);
    return this.validateSchema(schemas.ResponseSchema, res.data);
  }

  // 导出配置文件
  async ExportConfig(data = {}) {
    this.validateSchema(schemas.exportRequestSchema, data);
    const res = await this.client.post('/service/export', data);

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

    return res.data;
  }

  // 获取模型列表
  async GetModelsAvailiable(){
    const res = await this.client.get('/services/models');
    return this.validateSchema(schemas.modelsResponse, res.data);
  }

  // 获取推荐模型列表
  async GetModelsRecommended(){
    const res = await this.client.get('/model/recommend');
    return this.validateSchema(schemas.recommendModelsResponse, res.data);
  }

  // 获取支持模型列表
  async GetModelsSupported(data){
    this.validateSchema(schemas.getModelsSupported, data);
    // 添加请求头
    const res = await this.client.get('/model/support', {params: data});
    return this.validateSchema(schemas.recommendModelsResponse, res.data);
  }

  // chat服务
  async Chat(data) {
    this.validateSchema(schemas.chatRequest, data);
  
    // 判断是否是流式
    const config = { responseType: data.stream ? 'stream' : 'json' };
    const res = await this.client.post('/services/chat', data, config);
  
    if (data.stream) {
      const eventEmitter = new EventEmitter();
  
      res.data.on('data', (chunk) => {
        try {
          const rawData = chunk.toString().trim();
          const jsonString = rawData.startsWith('data:') ? rawData.slice(5) : rawData;
          const response = JSON.parse(jsonString);
          eventEmitter.emit('data', response);  // 触发事件，实时传输数据
        } catch (err) {
          eventEmitter.emit('error', `解析流数据失败: ${err.message}`);
        }
      });
  
      res.data.on('error', (err) => {
        eventEmitter.emit('error', `流式响应错误: ${err.message}`);
      });

      res.data.on('end', () => {
        eventEmitter.emit('end');  // 触发结束事件
      });
  
      return eventEmitter;  // 返回 EventEmitter 实例
    } else {
      return this.validateSchema(schemas.chatResponse, res.data);
    }
  }


  // 生文服务
  async Generate(data) {
    this.validateSchema(schemas.generateRequest, data);

    const config = { responseType: data.stream ? 'stream' : 'json' };
    const res = await this.client.post('/services/generate', data, config);

    if (data.stream) {
      const eventEmitter = new EventEmitter();

      res.data.on('data', (chunk) => {
        try {
          const response = JSON.parse(chunk.toString());
          if (response) {
            this.validateSchema(schemas.generateResponse, response);
            eventEmitter.emit('data', response.response);  // 逐步传输响应内容
          }
        } catch (err) {
          eventEmitter.emit('error', `解析流数据失败: ${err.message}`);
        }
      });

      res.data.on('error', (err) => {
        eventEmitter.emit('error', `流式响应错误: ${err.message}`);
      });

      res.data.on('end', () => {
        eventEmitter.emit('end');  // 触发结束事件
      });

      return eventEmitter;  // 返回 EventEmitter 实例
    } else {
      return this.validateSchema(schemas.generateResponse, res.data);
    }
  }
  

  // 生图服务
  async TextToImage(data) {
    this.validateSchema(schemas.textToImageRequest, data);
    const res = await this.client.post('/services/text-to-image', data);
    return this.validateSchema(schemas.textToImageResponse, res.data);
  }

  // embed服务

  ByzeInit(){

  }
}

module.exports = Byze;