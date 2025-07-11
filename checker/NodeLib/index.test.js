const Oadin = require('./index');

function createOadin() {
  return new Oadin('oadin/v0.2');
}

describe('Oadin 外部方法集成测试', () => {
  // ----- 下载启动 -----
  test('isOadinExisted 返回布尔值', () => {
    const oadin = createOadin();
    const result = oadin.isOadinExisted();
    expect(typeof result).toBe('boolean');
  });

  test('isOadinAvailable 能正常返回', async () => {
    const oadin = createOadin();
    const result = await oadin.isOadinAvailable(1, 1000);
    expect(typeof result).toBe('boolean');
  });

  test('downloadOadin 能正常执行', async () => {
    const oadin = createOadin();
    const result = await oadin.downloadOadin(1);
    expect(typeof result).toBe('boolean');
  });

  test('startOadin 能正常执行', async () => {
    const oadin = createOadin();
    const result = await oadin.startOadin();
    expect(typeof result).toBe('boolean');
  });

  // ----- GET 请求 -----
  test('getServices 能正常调用', async () => {
    const oadin = createOadin();
    const result = await oadin.getServices();
    expect(result).toHaveProperty('code');
  });

  test('getModels 能正常调用', async () => {
    const oadin = createOadin();
    const result = await oadin.getModels();
    expect(result).toHaveProperty('code');
  });

  test('getModelsRecommended 能正常调用', async () => {
    const oadin = createOadin();
    const result = await oadin.getModelsRecommended();
    expect(result).toHaveProperty('code');
  });

  test('getModelsSupported 能正常调用', async () => {
    const oadin = createOadin();
    const data = { service_source: 'local', flavor: 'ollama' };
    const result = await oadin.getModelsSupported(data);
    expect(result).toHaveProperty('code');
  });

  test('getServiceProviders 能正常调用', async () => {
    const oadin = createOadin();
    const result = await oadin.getServiceProviders();
    expect(result).toHaveProperty('code');
  });

  // ----- 同步请求 -----
  test('installService 能正常调用', async () => {
    const oadin = createOadin();
    // 增 service
    const data1 = {
      service_name: 'chat',
      service_source: 'local',
      hybrid_policy: 'default',
      flavor_name: 'ollama',
      provider_name: 'local_ollama_chat',
    };
    const result1 = await oadin.installService(data1);
    expect(result1).toHaveProperty('code');

    // 改 service
    const data2 = {
      service_name: 'chat',
      hybrid_policy: 'always_local',
    };
    const result2 = await oadin.updateService(data2);
    expect(result2).toHaveProperty('code');

    // 增 model
    const data3 = {
      model_name: 'qwen2:0.5b',
      service_name: 'chat',
      service_source: 'local',
      provider_name: 'local_ollama_chat',
    };
    const result3 = await oadin.installModel(data3);
    expect(result3).toHaveProperty('code');

    // 删 model
    const data4 = {
      model_name: 'qwen2:0.5b',
      service_name: 'chat',
      service_source: 'local',
      provider_name: 'local_ollama_chat',
    };
    const result4 = await oadin.deleteModel(data4);
    expect(result4).toHaveProperty('code');

    // 增 provider
    const data5 = {
      service_name: 'chat',
      service_source: 'remote',
      flavor_name: 'aliyun',
      provider_name: 'remote_aliyun_chat',
      auth_type: 'apikey',
      auth_key: '123456',
      models: ['deepseek-r1:7b'],
    };
    const result5 = await oadin.installServiceProvider(data5);
    expect(result5).toHaveProperty('code');

    // 改 provider
    const data6 = {
      service_name: 'chat',
      service_source: 'remote',
      flavor_name: 'baidu',
      provider_name: 'remote_baidu_chat',
    };
    const result6 = await oadin.updateServiceProvider(data6);
    expect(result6).toHaveProperty('code');

    const data7 = { provider_name: 'remote_baidu_chat' };
    const result7 = await oadin.deleteServiceProvider(data7);
    expect(result7).toHaveProperty('code');

  });

  test('importConfig 能正常调用', async () => {
    const oadin = createOadin();
    // 请确保本地有 .oadin 文件
    const result = await oadin.importConfig(require('os').homedir() + '/Oadin/.oadin');
    expect(result).toHaveProperty('code');
  });

  test('exportConfig 能正常调用', async () => {
    const oadin = createOadin();
    const result = await oadin.exportConfig({});
    expect(result).toHaveProperty('code');
  });

  test('chat 非流式调用', async () => {
    const oadin = createOadin();
    const data = {
      model: 'llama2',
      stream: false,
      messages: [
        { role: 'user', content: '你好' }
      ],
      temperature: 0.7,
      max_tokens: 100,
    };
    const result = await oadin.chat(data);
    expect(result).toHaveProperty('code');
  });

  test('generate 非流式调用', async () => {
    const oadin = createOadin();
    const data = {
      model: 'llama2',
      stream: false,
      prompt: '你好',
    };
    const result = await oadin.generate(data);
    expect(result).toHaveProperty('code');
  });

  test('textToImage 能正常调用', async () => {
    const oadin = createOadin();
    const data = {
      model: 'wanx2.1-t2i-turbo',
      prompt: '一间有着精致窗户的花店，漂亮的木质门，摆放着花朵',
    };
    const result = await oadin.textToImage(data);
    expect(result).toHaveProperty('code');
  });

  test('embed 能正常调用', async () => {
    const oadin = createOadin();
    const data = {
      model: 'llama2',
      input: '你好',
    };
    const result = await oadin.embed(data);
    expect(result).toHaveProperty('code');
  });

  test('OadinInit 能正常调用', async () => {
    const oadin = createOadin();
    // 请确保本地有 .oadin 文件
    const result = await oadin.OadinInit(require('os').homedir() + '/Oadin/.oadin');
    expect(typeof result).toBe('boolean');
  });
});