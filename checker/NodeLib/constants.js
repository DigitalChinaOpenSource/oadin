// 常量定义
const VERSION = 'v0.2.1';
const OADIN_VERSION = 'oadin/v0.2';
const WIN_OADIN_PATH = 'Oadin';
const WIN_OADIN_EXE = 'oadin.exe';
const MAC_OADIN_PATH = '/usr/local/bin/oadin';
const MAC_OADIN_EXE = 'oadin';
//TODO: 把下载域名拆开
// TEST
const WIN_INSTALLER_URL = 'http://10.3.70.145:32018/repository/raw-hosted/intel-ai-pc/oadin/releases/win/oadin-installer-latest.exe';
const MAC_INSTALLER_URL = 'http://10.3.70.145:32018/repository/raw-hosted/intel-ai-pc/oadin/releases/mac/oadin-installer-latest.pkg'
// const WIN_INSTALLER_URL = 'https://oss-aipc.dcclouds.com/oadin/releases/windows/oadin-installer-latest.exe';
// const MAC_INSTALLER_URL = 'https://oss-aipc.dcclouds.com/oadin/releases/macos/oadin-installer-latest.pkg';
const WIN_INSTALLER_NAME = 'oadin-installer-latest.exe';
const MAC_INSTALLER_NAME = 'oadin-installer-latest.pkg';
const OADIN_INSTALLER_DIR = 'OadinInstaller';
const OADIN_CONFIG_FILE = '.oadin';
const OADIN_HEALTH = "http://localhost:16688/health";
const OADIN_ENGINE_PATH = "http://localhost:16688/engine/health";

const PLATFORM_CONFIG = {
  win32: {
    downloadUrl: WIN_INSTALLER_URL,
    installerFileName: 'oadin-installer-latest.exe',
    userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)'
  },
  darwin: {
    downloadUrl: MAC_INSTALLER_URL,
    installerFileName: 'oadin-installer-latest.pkg',
    userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)'
  }
};

module.exports = {
  VERSION,
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
