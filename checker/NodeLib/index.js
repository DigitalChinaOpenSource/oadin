// const express = require('express');
const http = require('http');
const https = require('https');
const fs = require('fs');
const path = require('path');
const os = require('os');
const axios = require('axios');
const EventEmitter = require('events');
const _ = require('lodash');
const { execFile, spawn } = require('child_process');
const { promises: fsPromises } = require("fs");

const schemas = require('./schema.js');
const tools = require('./tools.js');
const { logAndConsole, downloadFile, getOadinExecutablePath, runInstallerByPlatform, isHealthy } = require('./tools.js');
const { instance, requestWithSchema } = require('./axiosInstance.js')
const { VERSION, MAC_OADIN_PATH, PLATFORM_CONFIG, OADIN_HEALTH, OADIN_ENGINE_PATH, } = require('./constants.js');

class Oadin {
  constructor(version) {
    this.version = version || "oadin/v0.2";
    this.client = instance;
    logAndConsole('info', `Oadin类初始化，版本: ${this.version}`);
  }

  async _requestWithSchema({ method, url, data, schema }) {
    logAndConsole('info', `请求API: ${method.toUpperCase()} ${url}`);
    return await requestWithSchema({ method, url, data, schema });
  }

  // 检查 Oadin 服务是否启动
  async isOadinAvailable(retries = 5, interval = 1000) {
    logAndConsole('info', '检测Oadin服务可用性...');
    const fibArr = tools.fibonacci(retries, interval);
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

  // 检查用户目录是否存在 Oadin.exe
  async isOadinExisted() {
    const dest = getOadinExecutablePath();
    const existed = fs.existsSync(dest);
    logAndConsole('info', `检测Oadin可执行文件是否存在: ${dest}，结果: ${existed}`);
    if (existed){
      const latest = await this.isOadinLatest();
      logAndConsole('info', `Oadin是否最新版: ${latest}`);
      return latest;
    }
    return existed;
  }

  async isOadinLatest() { // 如果这个函数属于 Oadin 类，需要是 async isOadinLatest() {}
    const platform = tools.getPlatform();
    let currentMainVersion = null;
    let currentSubVersion = null;
    let fullStdout = ''; // 用于存储完整的stdout，以便后续解析

    if (platform === 'win32') {
      try {
        const userDir = os.homedir();
        const oadinDir = path.join(userDir, 'Oadin');
        const oadinExecutable = path.join(oadinDir, 'oadin.exe');

        const originalPath = process.env.PATH;
        if (!process.env.PATH.includes(oadinDir)) {
          process.env.PATH = `${process.env.PATH}${path.delimiter}${oadinDir}`;
        }

        const { stdout } = await new Promise((resolve, reject) => {
          execFile(oadinExecutable, ['version'], { timeout: 5000 }, (error, stdout, stderr) => {
            process.env.PATH = originalPath; // 恢复 PATH

            if (error) {
              logAndConsole('error', `执行 'oadin version' 命令失败: ${error.message}, stderr: ${stderr.toString()}`);
              return reject(error);
            }
            resolve({ stdout: stdout.toString() });
          });
        });
        fullStdout = stdout.toString();

      } catch (err) {
        logAndConsole('error', `获取 Windows Oadin 版本失败: ${err.message}`);
        return false;
      }
    } else if (platform === 'darwin') {
      try {
        const oadinExecutable = tools.MAC_OADIN_PATH; // 确保 MAC_OADIN_PATH 是正确的
        const { stdout } = await new Promise((resolve, reject) => {
          execFile(oadinExecutable, ['version'], { timeout: 5000 }, (error, stdout, stderr) => {
            if (error) {
              logAndConsole('error', `执行 'oadin version' 命令失败: ${error.message}, stderr: ${stderr.toString()}`);
              return reject(error);
            }
            resolve({ stdout: stdout.toString() });
          });
        });
        fullStdout = stdout.toString();

      } catch (err) {
        logAndConsole('error', `获取 macOS Oadin 版本失败: ${err.message}`);
        return false;
      }
    } else {
      logAndConsole('warn', `不支持的平台，无法获取 Oadin 版本。`);
      return false;
    }

    // 解析主版本号
    const mainVersionMatch = fullStdout.match(/Oadin Version:\s*(v\d+\.\d+)/);
    if (mainVersionMatch && mainVersionMatch[1]) {
      currentMainVersion = mainVersionMatch[1];
    }

    // 解析子版本号
    const subVersionMatch = fullStdout.match(/Oadin SubVersion:\s*(\d+)/);
    if (subVersionMatch && subVersionMatch[1]) {
      currentSubVersion = subVersionMatch[1];
    }

    // 1. 如果没有子版本号，则认为不是最新
    if (!currentSubVersion) {
      logAndConsole('info', '当前 Oadin 版本只包含主版本号，不含子版本号，视为旧版本，需要更新。');
      return false;
    }

    // 2. 比较主版本号
    // 假设 `this.latestMainVersion` 和 `this.latestSubVersion` 可以从 Oadin 类中访问
    // 如果这个函数是独立的，你需要将它们作为参数传入或通过导入获取
    const latestMainVersion = this.latestMainVersion || "v0.2"; // 从 Oadin 类实例获取，或设置为常量
    const latestSubVersion = this.latestSubVersion || "20250701814"; // 从 Oadin 类实例获取，或设置为常量

    // 将版本字符串转换为可比较的数字（例如 v0.2 -> 0.2）
    const parseVersion = (versionStr) => parseFloat(versionStr.replace('v', ''));
    const currentMainNum = parseVersion(currentMainVersion);
    const latestMainNum = parseVersion(latestMainVersion);

    if (currentMainNum < latestMainNum) {
      logAndConsole('info', `Oadin 主版本不是最新 (当前: ${currentMainVersion}, 期望: ${latestMainVersion})，需要更新。`);
      return false;
    } else if (currentMainNum > latestMainNum) {
      logAndConsole('info', `Oadin 主版本 (当前: ${currentMainVersion}) 高于期望版本 (${latestMainVersion})。`);
      return true; // 主版本更高，通常认为是最新的
    }
    // 如果主版本相等，继续比较子版本

    // 3. 比较子版本号
    // 将子版本号转换为数字进行比较
    const currentSubNum = parseInt(currentSubVersion);
    const latestSubNum = parseInt(latestSubVersion);

    if (currentSubNum < latestSubNum) {
      logAndConsole('info', `Oadin 子版本不是最新 (当前: ${currentSubVersion}, 期望: ${latestSubVersion})，需要更新。`);
      return false;
    } else if (currentSubNum > latestSubNum) {
      logAndConsole('info', `Oadin 子版本 (当前: ${currentSubNum}) 高于期望版本 (${latestSubNum})。`);
      return true; // 子版本更高，通常认为是最新的
    }

    logAndConsole('info', `Oadin 主版本 ${currentMainVersion} 子版本 ${currentSubVersion}。`);
    // 如果主版本和子版本都匹配或更高，则认为是最新
    logAndConsole('info', '✅ Oadin 版本是最新。');
    return true;
  }

  getOadinInstallerPath() {
    const platform = tools.getPlatform();
    if (platform == "unsupported") { return null }
    else if (platform === 'win32') { return PLATFORM_CONFIG.win32.downloadUrl }
    else if (platform === 'darwin') { return PLATFORM_CONFIG.darwin.downloadUrl };
  }

  // 私有方法：仅下载
  async _downloadFile(url, dest, options, retries = 3) {
    logAndConsole('info', `准备下载文件: ${url} 到 ${dest}`);
    return await downloadFile(url, dest, options, retries);
  }

  // 运行安装包
  async _runOadinInstaller(installerPath) {
    const platform = tools.getPlatform();
    logAndConsole('info', `运行安装包: ${installerPath}，平台: ${platform}`);
    try {
      await runInstallerByPlatform(installerPath);
      logAndConsole('info', '安装包运行成功');
      return true;
    } catch (err) {
      logAndConsole('error', '安装包运行失败：' + err.message);
      return false;
    }
  }

  async downloadOadin(retries = 3) {
    try {
      const platform = tools.getPlatform();
      if (platform === 'unsupported' || !PLATFORM_CONFIG[platform]) {
        logAndConsole('error', '不支持的平台');
        return false;
      }
      const { downloadUrl, installerFileName, userAgent } = PLATFORM_CONFIG[platform];
      const userDir = os.homedir();
      const destDir = path.join(userDir, 'OadinInstaller');
      const dest = path.join(destDir, installerFileName);
      const options = {
        headers: {
          'User-Agent': userAgent,
        },
      };
      const downloadOk = await this._downloadFile(downloadUrl, dest, options, retries);
      if (downloadOk) {
        const installResult = await this._runOadinInstaller(dest);
        return installResult;
      } else {
        logAndConsole('error', '三次下载均失败，放弃安装。');
        return false;
      }
    } catch (err) {
      logAndConsole('error', '下载或安装 Oadin 失败: ' + err.message);
      return false;
    }
  }

  // 启动 Oadin 服务
  async startOadin() {
    const alreadyRunning = await this.isOadinAvailable(2, 1000);
    if (alreadyRunning) {
      logAndConsole('info', '[startOadin] Oadin 在运行中');
      return true;
    }
    return new Promise((resolve, reject) => {
      const platform = tools.getPlatform();
      const userDir = os.homedir();
      const oadinDir = path.join(userDir, 'Oadin');
      logAndConsole('info', `oadinDir: ${oadinDir}`);
      if (platform === "unsurported") return reject(new Error(`不支持的平台`));
      if (platform === 'win32') {
        if (!process.env.PATH.includes(oadinDir)) {
          process.env.PATH = `${process.env.PATH}${path.delimiter}${oadinDir}`;
          logAndConsole('info', '添加到临时环境变量');
        }
        const command = 'cmd.exe';
        const args = ['/c', 'start-oadin.bat'];
        logAndConsole('info', `正在运行命令: ${command} ${args.join(' ')}`);
        execFile(command, args, { windowsHide: true }, async (error, stdout, stderr) => {
          if (error) logAndConsole('error', 'oadin server start:error ' + error);
          if (stdout) logAndConsole('info', 'oadin server start:stdout: ' + stdout.toString());
          if (stderr) logAndConsole('error', 'oadin server start:stderr: ' + stderr.toString());
          const output = (stdout + stderr).toString().toLowerCase();
          if (error || output.includes('error')) {
            return resolve(false);
          }
          const available = await this.isOadinAvailable(5, 1500);
          return resolve(available);
        });
      } else if (platform === 'darwin') {
        try {
          if (!process.env.PATH.split(':').includes('/usr/local/bin')) {
            process.env.PATH = `/usr/local/bin:${process.env.PATH}`;
            logAndConsole('info', '已将 /usr/local/bin 添加到 PATH');
          }
          let child;
          let stderrContent = '';
          child = spawn('/usr/local/bin/oadin', ['server', 'start', '-d'], {
            stdio: ['ignore', 'pipe', 'pipe'],
            windowsHide: true,
          });
          child.stdout.on('data', (data) => {
            if (data.toString().includes('server start successfully')) {
              //TODO：获取退出状态码
              logAndConsole('info', 'Oadin 服务启动成功');
              resolve(true);
            }
            logAndConsole('info', `stdout: ${data}`);
          });
          child.stderr.on('data', (data) => {
            const errorMessage = data.toString().trim();
            stderrContent += errorMessage + '\n';
            logAndConsole('error', `stderr: ${errorMessage}`);
          });
          child.on('error', (err) => {
            logAndConsole('error', `❌ 启动失败: ${err.message}`);
            if (err.code === 'ENOENT') {
              logAndConsole('error', '未找到oadin可执行文件，请检查下载是否成功或环境变量未生效');
            }
            resolve(false);
          });
          child.on('close', (code) => {
            if (stderrContent.includes('Install model engine failed')){
              logAndConsole('error', '❌ 启动失败: 模型引擎安装失败。');
              resolve(false);
            } else if (code === 0) {
              logAndConsole('info', '进程退出，正在检查服务状态...');
            } else {
              logAndConsole('error', `❌ 启动失败，退出码: ${code}`);
              resolve(false);
            }
          });
          child.unref();
        } catch (error) {
          logAndConsole('error', '启动 Oadin 服务异常: ' + error.message);
          resolve(false);
        }
      }
    });
  }
  
  // 查看当前服务
  async getServices() {
    return this._requestWithSchema({
      method: 'get',
      url: '/service',
      schema: { response: schemas.getServicesSchema }
    });
  }

  // 创建新服务
  async installService(data) {
    return this._requestWithSchema({
      method: 'post',
      url: '/service',
      data,
      schema: { request: schemas.installServiceRequestSchema, response: schemas.ResponseSchema }
    });
  }

  // 更新服务
  async updateService(data) {
    return this._requestWithSchema({
      method: 'put',
      url: '/service',
      data,
      schema: { request: schemas.updateServiceRequestSchema, response: schemas.ResponseSchema }
    });
  }

  // 查看当前模型
  async getModels() {
    return this._requestWithSchema({
      method: 'get',
      url: '/model',
      schema: { response: schemas.getModelsSchema }
    });
  }

  // 安装新模型
  async installModel(data) {
    return this._requestWithSchema({
      method: 'post',
      url: '/model',
      data,
      schema: { request: schemas.installModelRequestSchema, response: schemas.ResponseSchema }
    });
  }

  async deleteModel(data) {
    return this._requestWithSchema({
      method: 'delete',
      url: '/model',
      data,
      schema: { request: schemas.deleteModelRequestSchema, response: schemas.ResponseSchema }
    });
  }

  async installModelStream(data) {
    const client = axios.create({
      baseURL: `http://localhost:16688/oadin/v0.2`,
      headers: {"Content-Type": "application/json" },
    })
    const config = { responseType: 'stream' };
    try {
      const res = await client.post('/model/stream', data, config);
      const eventEmitter = new EventEmitter();

      res.data.on('data', (chunk) => {
        try {
          // 解析流数据
          const rawData = chunk.toString().trim();
          const jsonString = rawData.startsWith('data:') ? rawData.slice(5) : rawData;
          const response = JSON.parse(jsonString);

          eventEmitter.emit('data', rawData);
          if (response.status === 'success') {
            eventEmitter.emit('end', rawData);
          }
          if (response.status === 'canceled') {
            eventEmitter.emit('canceled', rawData);
          }
          if (response.status === 'error') {
            eventEmitter.emit('end', rawData);
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

      return eventEmitter; // 返回 EventEmitter 实例
    } catch (error) {
      return {
        code: 400,
        msg: error.response?.data?.message || error.message || '请求失败',
        data: null,
      }
    }
  }


  async cancelInstallModel(data) {
    return this._requestWithSchema({
      method: 'post',
      url: '/model/stream/cancel',
      data,
      schema: { request: schemas.cancelModelStreamRequestSchema, response: schemas.ResponseSchema }
    });
  }

  // 查看服务提供商
  async getServiceProviders() {
    return this._requestWithSchema({
      method: 'get',
      url: '/service_provider',
      schema: { response: schemas.getServiceProvidersSchema }
    });
  }

  // 新增服务提供商
  async installServiceProvider(data) {
    return this._requestWithSchema({
      method: 'post',
      url: '/service_provider',
      data,
      schema: { request: schemas.installServiceProviderRequestSchema, response: schemas.ResponseSchema }
    });
  }

  // 更新服务提供商
  async updateServiceProvider(data) {
    return this._requestWithSchema({
      method: 'put',
      url: '/service_provider',
      data,
      schema: { request: schemas.updateServiceProviderRequestSchema, response: schemas.ResponseSchema }
    });
  }

  // 删除服务提供商
  async deleteServiceProvider(data) {
    return this._requestWithSchema({
      method: 'post',
      url: '/service_provider',
      data,
      schema: { request: schemas.deleteServiceProviderRequestSchema, response: schemas.ResponseSchema }
    });
  }

  // 导入配置文件
  async importConfig(filePath) {
    try {
      const data = await fsPromises.readFile(filePath, 'utf8');
      return this._requestWithSchema({
        method: 'post',
        url: '/service/import',
        data,
        schema: { response: schemas.ResponseSchema }
      });
    } catch (error) {
      return { code: 400, msg: error.message, data: null };
    }
  }

  // 导出配置文件
  async exportConfig(data = {}) {
    // 只做文件写入，http部分用统一schema校验
    const result = await this._requestWithSchema({
      method: 'post',
      url: '/service/export',
      data,
      schema: { request: schemas.exportRequestSchema, response: schemas.ResponseSchema }
    });
    if (result.code === 200) {
      try {
        const userDir = os.homedir();
        const destDir = path.join(userDir, 'Oadin');
        const dest = path.join(destDir, '.oadin');
        tools.ensureDirWritable(destDir);
        const fileContent = JSON.stringify(result.data, null, 2);
        fs.writeFileSync(dest, fileContent);
        console.log(`已将生成文件写入到 ${dest}`);
      } catch (error) {
        return { code: 400, msg: error.message, data: null };
      }
    }
    return result;
  }

  // 获取推荐模型列表
  async getModelsRecommended() {
    return this._requestWithSchema({
      method: 'get',
      url: '/model/recommend',
      schema: { response: schemas.recommendModelsResponse }
    });
  }

  // getModelsSupported
  async getModelsSupported(data) {
    return this._requestWithSchema({
      method: 'get',
      url: '/model/support',
      data,
      schema: { request: schemas.getModelsSupported, response: schemas.getSupportModelResponseSchema }
    });
  }

  // chat服务（支持流式和非流式）
  async chat(data) {
    const stream = data.stream;
    if (!stream) {
      // 非流式
      return this._requestWithSchema({ method: 'post', url: 'services/chat', data });
    }
    // 流式
    try {
      const client = axios.create({
        baseURL: `http://localhost:16688/oadin/v0.2`,
        headers: {"Content-Type": "application/json" },
      });
      const config = { responseType: 'stream' };
      const res = await client.post('services/chat', data, config);
      const eventEmitter = new EventEmitter();
      res.data.on('data', (chunk) => {
        try {
          let rawData = _.isString(chunk) ? _.trim(chunk) : _.trim(chunk.toString());
          let jsonString = _.startsWith(rawData, 'data:') ? rawData.slice(5) : rawData;
          jsonString = _.trim(jsonString);
          if (_.isEmpty(jsonString)) {
            throw new Error('收到空的流数据');
          }
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
        eventEmitter.emit('end'); // 触发结束事件
      });

      return eventEmitter;
    } catch (error) {
      return { code: 400, msg: error.response?.data?.message || error.message, data: null };
    }
  }

  // 生文服务（支持流式和非流式）
  async generate(data) {
    const stream = data.stream;
    if (!stream) {
      return this._requestWithSchema({ method: 'post', url: 'services/generate', data });
    }
    try {
      const client = axios.create({
        baseURL: `http://localhost:16688/oadin/v0.2`,
        headers: {"Content-Type": "application/json" },
      });
      const config = { responseType: 'stream' };
      const res = await client.post('services/generate', data, config);
      const eventEmitter = new EventEmitter();
      res.data.on('data', (chunk) => {
        try {
          let rawData = _.isString(chunk) ? _.trim(chunk) : _.trim(chunk.toString());
          let jsonString = _.startsWith(rawData, 'data:') ? rawData.slice(5) : rawData;
          jsonString = _.trim(jsonString);
          if (_.isEmpty(jsonString)) {
            throw new Error('收到空的流数据');
          }
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
        eventEmitter.emit('end'); // 触发结束事件
      });

      return eventEmitter;
    } catch (error) {
      return { code: 400, msg: error.response?.data?.message || error.message, data: null };
    }
  }
  
  // 生图服务
  async textToImage(data) {
    return this._requestWithSchema({
      method: 'post',
      url: '/services/text-to-image',
      data,
      schema: { request: schemas.textToImageRequest, response: schemas.textToImageResponse }
    });
  }

  // embed服务
  async embed(data) {
    return this._requestWithSchema({
      method: 'post',
      url: '/services/embed',
      data,
      schema: { request: schemas.embeddingRequest, response: schemas.embeddingResponse }
    });
  }

  // 用于一键安装 Oadin 和 导入配置
  // TODO：记录日志
  async OadinInit(path){
    const isOadinAvailable = await this.isOadinAvailable();
    if (isOadinAvailable) {
      logAndConsole('info','✅ Oadin 服务已启动，跳过安装。');
      return true;
    }
    
    const isOadinExisted = this.isOadinExisted();
    if (!isOadinExisted) {
      const downloadSuccess = await this.downloadOadin();
      if (!downloadSuccess) {
        logAndConsole('error','❌ 下载 Oadin 失败，请检查网络连接或手动下载。');
        return false;
      }
    } else {
      logAndConsole('info','✅ Oadin 已存在，跳过下载。');
    }

    const installSuccess = await this.startOadin();
    if (!installSuccess) {
      logAndConsole('error','❌ 启动 Oadin 服务失败，请检查配置或手动启动。');
      return false;
    } else {
      logAndConsole('info','✅ Oadin 服务已启动。');
    }

    const importSuccess = await this.importConfig(path);
    if (!importSuccess) {
      logAndConsole('error','❌ 导入配置文件失败。');
      return false;
    } else {
      logAndConsole('info','✅ 配置文件导入成功。');
    }
    return true;
  }
}

module.exports = Oadin;