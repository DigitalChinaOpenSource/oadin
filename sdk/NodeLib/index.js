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

  // run `aog install chat`
  InstallChat(remote = null) {
    return new Promise((resolve) => {
      const userDir = os.homedir();
      const aogPath = path.join(userDir, 'AOG', 'aog.exe');
      process.env.PATH = `${process.env.PATH};${aogPath}`;

      const child = spawn(aogPath, ['install', 'chat'], { detached: true, stdio: [ 'pipe', 'pipe', 'pipe'] });

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
          console.log('安装 aog 聊天插件成功');
          resolve(true);
        } else {
          console.error(`安装 aog 聊天插件失败，退出码: ${code}`);
          resolve(false);
        }
      });

      child.on('error', (err) => {
        console.error(`启动 aog 安装命令失败: ${err.message}`);
        resolve(false);
      });

      child.unref();
    });
  }

  // get services
  async GetServices() {
    try {
      const res = await this.client.get('/service');
      return this.validateSchema(schemas.getServicesSchema, res.data);
    } catch (error) {
      throw new Error(`获取服务失败: ${error.message}`);
    }
  }

  // install service
  async InstallService(data) {
    try {
      this.validateSchema(schemas.installServiceRequestSchema, data);
      const res = await this.client.post('/service', data);
      return this.validateSchema(schemas.ResponseSchema, res.data);
    } catch (error) {
      throw new Error(`创建服务失败: ${error.message}`);
    }
  }
  
  // eidt service
  async UpdateService(data) {
    try {
      const res = await this.client.put('/service', data);
      return res.data;
    } catch (error) {
      throw new Error(`更新服务失败: ${error.message}`);
    }

  }

  // 查get models
  async GetModels() {
    try {
      const res = await this.client.get('/model');
      return this.validateSchema(schemas.getModelsSchema, res.data);
    } catch (error) { 
      throw new Error(`获取模型失败: ${error.message}`);
    }
  }

  // install model
  async InstallModel(data) {
    try {
      this.validateSchema(schemas.installModelRequestSchema, data);
      const res = await this.client.post('/model', data);
      return this.validateSchema(schemas.ResponseSchema, res.data);
    } catch (error) {
      throw new Error(`安装模型失败: ${error.message}`);
    }
  }

  // install model（stream）
  async InstallModelStream(data) {
    try {
      this.validateSchema(schemas.installModelRequestSchema, data);
    } catch (error) {
      throw new Error(`流式安装模型失败: ${error.message}`);
    }

    const config = { responseType: 'stream' };

    try {
      const res = await this.client.post('/model/stream', data, config);
      const eventEmitter = new EventEmitter();

      res.data.on('data', (chunk) => {
        try {
          const rawData = chunk.toString().trim();
          const jsonString = rawData.startsWith('data:') ? rawData.slice(5) : rawData;
          const response = JSON.parse(jsonString);

          eventEmitter.emit('data', response);

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
          eventEmitter.emit('end');
      });

      return eventEmitter; 
    } catch (error) {
      throw new Error(`流式安装模型失败: ${error.message}`);
    }
  }

  // cancel install model (stream)
  async CancelInstallModel(data) {
    try {
      this.validateSchema(schemas.cancelModelStreamRequestSchema, data);
      const res = await this.client.post('/model/stream/cancel', data);
      return this.validateSchema(schemas.ResponseSchema, res.data);
    } catch (error) {
      throw new Error(`取消安装模型失败: ${error.message}`);
    }
  }

  // delete model
  async DeleteModel(data) {
    try {
      this.validateSchema(schemas.deleteModelRequestSchema, data);
      const res = await this.client.delete('/model', { data });
      return this.validateSchema(schemas.ResponseSchema, res.data);
    } catch (error) {
      throw new Error(`卸载模型失败: ${error.message}`);
    }
  }

  // get service providers
  async GetServiceProviders() {
    try {
      const res = await this.client.get('/service_provider');
      return this.validateSchema(schemas.getServiceProvidersSchema, res.data);
    } catch (error) {
      throw new Error(`获取服务提供商失败: ${error.message}`);
    }
  }

  // install service provider
  async InstallServiceProvider(data) {
    try {
      this.validateSchema(schemas.installServiceProviderRequestSchema, data);
      const res = await this.client.post('/service_provider', data);
      return this.validateSchema(schemas.ResponseSchema, res.data);
    } catch (error) {
      throw new Error(`新增服务提供商失败: ${error.message}`);
    }
  }

  // edit service provider
  async UpdateServiceProvider(data) {
    try {
      this.validateSchema(schemas.updateServiceProviderRequestSchema, data);
      const res = await this.client.put('/service_provider', data);
      return this.validateSchema(schemas.ResponseSchema, res.data);
    } catch (error) {
      throw new Error(`更新服务提供商失败: ${error.message}`);
    }
  }

  // delete service provider
  async DeleteServiceProvider(data) {
    try {
      this.validateSchema(schemas.deleteServiceProviderRequestSchema, data);
      const res = await this.client.delete('/service-provider', { data });
      return this.validateSchema(schemas.ResponseSchema, res.data);
    } catch (error) {
      throw new Error(`删除服务提供商失败: ${error.message}`);
    }
  }

  // import config
  async ImportConfig(path) {
    try {
      const data = await fsPromises.readFile(path, 'utf8');
      const res = await this.client.post('/service/import', data);
      return this.validateSchema(schemas.ResponseSchema, res.data);
    } catch (error) {
      throw new Error(`导入配置文件失败: ${error.message}`);
    }
  }

  // export config
  async ExportConfig(data = {}) {
    try{
      this.validateSchema(schemas.exportRequestSchema, data);
      const res = await this.client.post('/service/export', data);

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

      return res.data;
    } catch (error) {
      throw new Error(`导出配置文件失败: ${error.message}`);
    }
  }

  // get models availiable
  async GetModelsAvailiable(){
    try {
      const res = await this.client.get('/services/models');
      return this.validateSchema(schemas.modelsResponse, res.data);
    } catch (error) {
      throw new Error(`获取模型列表失败: ${error.message}`);
    }
  }

  // get models recommended
  async GetModelsRecommended(){
    try {
      const res = await this.client.get('/model/recommend');
      return this.validateSchema(schemas.recommendModelsResponse, res.data);
    } catch (error) {
      throw new Error(`获取推荐模型列表失败: ${error.message}`);
    }
  }

  // get models supported
  async GetModelsSupported(data){
    try {
      this.validateSchema(schemas.getModelsSupported, data);
      const res = await this.client.get('/model/support', {params: data});
      return this.validateSchema(schemas.recommendModelsResponse, res.data);
    } catch (error) {
      throw new Error(`获取支持模型列表失败: ${error.message}`);
    }
  }

  // chat
  async Chat(data) {
    this.validateSchema(schemas.chatRequest, data);
  
    // wheather stream is true, set responseType to stream
    const config = { responseType: data.stream ? 'stream' : 'json' };
    const res = await this.client.post('/services/chat', data, config);
  
    if (data.stream) {
      const eventEmitter = new EventEmitter();
  
      res.data.on('data', (chunk) => {
        try {
          const rawData = chunk.toString().trim();
          const jsonString = rawData.startsWith('data:') ? rawData.slice(5) : rawData;
          const response = JSON.parse(jsonString);
          eventEmitter.emit('data', response);
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
      return this.validateSchema(schemas.chatResponse, res.data);
    }
  }


  // generate
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
            eventEmitter.emit('data', response.response);
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
      return this.validateSchema(schemas.generateResponse, res.data);
    }
  }
  
  // text to image
  async TextToImage(data) {
    try {
      this.validateSchema(schemas.textToImageRequest, data);
      const res = await this.client.post('/services/text_to_image', data);
      return this.validateSchema(schemas.textToImageResponse, res.data);
    } catch (error) {
      throw new Error(`生图服务请求失败: ${error.message}`);
    }
  }

  // embed
  async Embed(data) {
    try {
      this.validateSchema(schemas.embeddingRequest, data);
      const res = await this.client.post('/services/embed', data);
      return this.validateSchema(schemas.embeddingResponse, res.data);
    } catch (error) {
      throw new Error(`Embed服务请求失败: ${error.message}`);
    }
  }

  // 
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