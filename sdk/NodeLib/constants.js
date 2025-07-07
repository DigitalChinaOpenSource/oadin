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

// 常量定义
const AOG_VERSION = 'aog/v0.4';
const WIN_AOG_PATH = 'AOG';
const WIN_AOG_EXE = 'aog.exe';
const MAC_AOG_PATH = '/usr/local/bin';
const MAC_AOG_EXE = 'aog';
//TODO: 把下载域名拆开
const WIN_INSTALLER_URL = 'https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/aog/windows/aog-installer-latest.exe';
const MAC_INSTALLER_URL = 'https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/aog/windows/aog-installer-latest.pkg';
const WIN_INSTALLER_NAME = 'aog-installer-latest.exe';
const MAC_INSTALLER_NAME = 'aog-installer-latest.pkg';
const AOG_INSTALLER_DIR = 'AOGInstaller';
const AOG_CONFIG_FILE = '.aog';
const AOG_HEALTH = "http://localhost:16688/health";
const AOG_ENGINE_PATH = "http://localhost:16688/engine/health";

const PLATFORM_CONFIG = {
  win32: {
    downloadUrl: 'https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/aog/windows/aog-installer-latest.exe',
    installerFileName: 'aog-installer-latest.exe',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)'
  },
  darwin: {
    downloadUrl: 'https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/aog/windows/aog-installer-latest.pkg',
    installerFileName: 'aog-installer-latest.pkg',
    userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)'
  }
};

module.exports = {
  AOG_VERSION,
  WIN_AOG_PATH,
  WIN_AOG_EXE,
  MAC_AOG_PATH,
  MAC_AOG_EXE,
  WIN_INSTALLER_URL,
  MAC_INSTALLER_URL,
  WIN_INSTALLER_NAME,
  MAC_INSTALLER_NAME,
  AOG_INSTALLER_DIR,
  AOG_CONFIG_FILE,
  AOG_HEALTH,
  AOG_ENGINE_PATH,
  PLATFORM_CONFIG,
};
