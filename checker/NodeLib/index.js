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
const { exec } = require('child_process');
const { promises: fsPromises } = require("fs");

const schemas = require('./schema.js');

function AddToUserPath(destDir) {
  const isMacOS = process.platform === 'darwin';

  if (isMacOS) {
    try {
      // 确定shell配置文件
      const shell = process.env.SHELL || '';
      let shellConfigName = '.zshrc';
      if (shell.includes('bash')) shellConfigName = '.bash_profile';
      
      const shellConfigPath = path.join(os.homedir(), shellConfigName);
      const exportLine = `export PATH="$PATH:${destDir}"\n`;

      // 确保配置文件存在
      if (!fs.existsSync(shellConfigPath)) {
        fs.writeFileSync(shellConfigPath, '');
      }

      // 检查是否已存在路径
      const content = fs.readFileSync(shellConfigPath, 'utf8');
      if (content.includes(exportLine)) {
        console.log('✅ 环境变量已存在');
        return true;
      }

      // 追加路径到配置文件
      fs.appendFileSync(shellConfigPath, `\n${exportLine}`);
      console.log(`✅ 已添加到 ${shellConfigName}，请执行以下命令生效：\nsource ${shellConfigPath}`);
      return true;
    } catch (err) {
      console.error('❌ 添加环境变量失败:', err.message);
      return false;
    }
  } else {
    // Windows环境变量处理
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

      // 检查路径是否已存在
      const paths = currentPath.split(';').filter(p => p);
      if (paths.includes(destDir)) {
        console.log('✅ 环境变量已存在');
        return true;
      }

      // 更新Path值
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
        const platform = process.platform; // 获取当前平台

        let destDir;
        let dest;

        if (platform === 'win32') {
            // Windows 平台路径
            destDir = path.join(userDir, 'Byze');
            dest = path.join(destDir, 'byze.exe');
        } else if (platform === 'darwin') {
            // macOS 平台路径
            destDir = path.join(userDir, 'Byze');
            dest = path.join(destDir, 'byze'); // 假设 macOS 的可执行文件名为 'byze'
        } else {
            console.error('❌ 不支持的操作系统');
            return resolve(false);
        }

        resolve(fs.existsSync(dest));
    });
}

  // 从服务器下载 Byze.exe
  DownloadByze() {
    return new Promise((resolve) => {
      const isMacOS = process.platform === 'darwin';
      const url = isMacOS 
        ? 'http://120.232.136.73:31619/byzedev/byze.zip'
        : 'http://120.232.136.73:31619/byzedev/byze.exe';
      
      const userDir = os.homedir();
      const destDir = path.join(userDir, 'Byze');
      const destFileName = isMacOS ? 'byze.zip' : 'byze.exe';
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
  
            // macOS解压处理
            if (isMacOS) {
              try {
                const zip = new AdmZip(dest);
                zip.extractAllTo(destDir, true);
                console.log('✅ 解压完成');
                
                // 删除原始ZIP文件
                fs.unlinkSync(dest);
                
                // 设置可执行权限（根据需要）
                const execPath = path.join(destDir, 'byze'); // 假设解压后的可执行文件名
                if (fs.existsSync(execPath)) {
                  fs.chmodSync(execPath, 0o755);
                }
              } catch (e) {
                console.error('❌ 解压失败:', e.message);
                return resolve(false);
              }
            }
  
            // 添加环境变量
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

  // 启动 Byze 服务
  InstallByze() {
    return new Promise((resolve) => {
      const isMacOS = process.platform === 'darwin';
      const userDir = os.homedir();
      const byzeDir = path.join(userDir, 'Byze');
  
      // 确保PATH包含Byze目录（兼容跨平台）
      if (!process.env.PATH.includes(byzeDir)) {
        process.env.PATH = `${process.env.PATH}${path.delimiter}${byzeDir}`;
      }
  
      const child = spawn('byze', ['server', 'start', '-d'], {
        stdio: 'ignore',
        windowsHide: true
      });
      child.unref();
  
      child.on('error', (err) => {
        console.error(`❌ 启动失败: ${err.message}`);
        if (err.code === 'ENOENT') {
          console.log([
            '💡 可能原因:',
            `1. 未找到byze可执行文件，请检查下载是否成功`,
            `2. 环境变量未生效，请尝试重启终端`
          ].filter(Boolean).join('\n'));
        }
        resolve(false);
      });
  
      // 智能服务检测（带重试机制）
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
  
      // 动态调整首次检测时间
      setTimeout(() => checkServer(1), 5000);
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

  // 安装模型（流式）
  async InstallModelStream(data) {
    this.validateSchema(schemas.installModelRequestSchema, data);

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

          // 如果状态为 success，触发完成事件
          if (response.status === 'success') {
              eventEmitter.emit('end', response);
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
    } catch (error) {
      throw new Error(`流式安装模型失败: ${error.message}`);
    }
  }

  // 取消安装模型（流式）
  async CancelInstallModel(data) {
    this.validateSchema(schemas.cancelModelStreamRequestSchema, data);
    const res = await this.client.post('/model/stream/cancel', data);
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
    const data = await fsPromises.readFile(path, 'utf8');
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

  // 获取问学平台支持模型列表
  async GetSmartvisionModelsSupported(data){
    this.validateSchema(schemas.SmartvisionModelSupportRequest, data);
    // 添加请求头
    const res = await this.client.get('/model/support/smartvision', {params: data});
    return this.validateSchema(schemas.SmartvisionModelSupport, res.data);
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
  async Embed(data) {
    this.validateSchema(schemas.embeddingRequest, data);
    const res = await this.client.post('/services/embed', data);
    return this.validateSchema(schemas.embeddingResponse, res.data);
  }

  ByzeInit(){

  }
}

module.exports = Byze;