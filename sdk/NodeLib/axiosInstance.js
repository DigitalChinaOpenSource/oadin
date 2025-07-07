//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

const axios = require('axios');
const { AOG_VERSION } = require('./constants.js');
const { error } = require('console');
const Ajv = require('ajv');
const addFormats = require('ajv-formats');

const instance = axios.create({
  baseURL: `http://localhost:16688/${AOG_VERSION}`,
  headers: { "Content-Type": "application/json" },
});

function createAxiosInstance(version) {
  return axios.create({
    baseURL: `http://localhost:16688/${version || AOG_VERSION}`,
    headers: { "Content-Type": "application/json" },
  });
}

instance.interceptors.response.use(
  config => {
    // 处理响应数据
    return config.data;
  },
  error => Promise.reject(error)
);

instance.interceptors.request.use(
  config => {
    // 在发送请求之前做些什么
    return config;
  },
  error => {
    // 处理请求错误
    return Promise.reject(error);
  }
);

const get = (url, params, config) => instance.get(url, { ...config, params });
const post = (url, data, config) => instance.post(url, data, config);
const put = (url, data, config) => instance.put(url, data, config);
const del = (url, params, config) => instance.delete(url, { ...config, params });

const ajv = new Ajv();
addFormats(ajv);

/**
 * 通用请求方法，支持请求和响应schema校验及统一返回格式
 * @param {Object} param0
 * @param {'get'|'post'|'put'|'delete'} param0.method
 * @param {string} param0.url
 * @param {any} param0.data
 * @param {object} [param0.schema] - { request: 请求schema, response: 响应schema }
 * @returns {Promise<{code:number,msg:string,data:any}>}
 */
async function requestWithSchema({ method, url, data, schema }) {
  // 1. 请求参数校验（如果有）
  if (schema && schema.request) {
    const validateReq = ajv.compile(schema.request);
    if (!validateReq(data)) {
      return { code: 400, msg: `Request schema validation failed: ${JSON.stringify(validateReq.errors)}`, data: null };
    }
  }
  try {
    let res;
    if (method === 'get') {
      res = await instance.get(url, { params: data });
    } else if (method === 'post') {
      res = await instance.post(url, data);
    } else if (method === 'put') {
      res = await instance.put(url, data);
    } else if (method === 'delete') {
      res = await instance.delete(url, { params: data });
    } else {
      throw new Error('不支持的请求方法');
    }
    // 2. 响应schema校验（如果有）
    if (schema && schema.response) {
      const validateRes = ajv.compile(schema.response);
      if (!validateRes(res)) {
        throw new Error(`Response schema validation failed: ${JSON.stringify(validateRes.errors)}`);
      }
    }
    return { code: 200, msg: res.message || null, data: res.data };
  } catch (error) {
    return { code: 400, msg: error.response?.data?.message || error.message, data: null };
  }
}

module.exports = {
  get,
  post,
  put,
  del,
  request: instance.request.bind(instance),
  instance,
  requestWithSchema,
  createAxiosInstance
};