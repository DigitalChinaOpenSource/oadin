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
const OADIN_VERSION = 'oadin/v0.4';
const WIN_OADIN_PATH = 'OADIN';
const WIN_OADIN_EXE = 'oadin.exe';
const MAC_OADIN_PATH = '/usr/local/bin';
const MAC_OADIN_EXE = 'oadin';
//TODO: 把下载域名拆开
const WIN_INSTALLER_URL = 'https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/oadin/windows/oadin-installer-latest.exe';
const MAC_INSTALLER_URL = 'https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/oadin/windows/oadin-installer-latest.pkg';
const WIN_INSTALLER_NAME = 'oadin-installer-latest.exe';
const MAC_INSTALLER_NAME = 'oadin-installer-latest.pkg';
const OADIN_INSTALLER_DIR = 'OADINInstaller';
const OADIN_CONFIG_FILE = '.oadin';
const OADIN_HEALTH = "http://localhost:16688/health";
const OADIN_ENGINE_PATH = "http://localhost:16688/engine/health";

const PLATFORM_CONFIG = {
  win32: {
    downloadUrl: 'https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/oadin/windows/oadin-installer-latest.exe',
    installerFileName: 'oadin-installer-latest.exe',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)'
  },
  darwin: {
    downloadUrl: 'https://smartvision-aipc-open.oss-cn-hangzhou.aliyuncs.com/oadin/windows/oadin-installer-latest.pkg',
    installerFileName: 'oadin-installer-latest.pkg',
    userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)'
  }
};

module.exports = {
  OADIN_VERSION,
  WIN_OADIN_PATH,
  WIN_OADIN_EXE,
  MAC_OADIN_PATH,
  MAC_OADIN_EXE,
  WIN_INSTALLER_URL,
  MAC_INSTALLER_URL,
  WIN_INSTALLER_NAME,
  MAC_INSTALLER_NAME,
  OADIN_INSTALLER_DIR,
  OADIN_CONFIG_FILE,
  OADIN_HEALTH,
  OADIN_ENGINE_PATH,
  PLATFORM_CONFIG,
};
