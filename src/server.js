import express from 'express';
import diskusage from 'diskusage';
import cors from 'cors';
import path from 'path';
import { fileURLToPath } from 'url';
import fs from 'fs';
import dotenv from 'dotenv';

dotenv.config();

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const app = express();
app.use(cors());
const PORT = process.env.PORT || 33333;

app.use(express.static(path.join(__dirname, '../dist')));

// 处理所有非API请求，返回index.html
app.get('*', (req, res) => {
  if (!req.path.startsWith('/api')) {
    res.sendFile(path.join(__dirname, '../dist/index.html'));
  }
});

// 允许跨域请求
app.use((req, res, next) => {
  res.header('Access-Control-Allow-Origin', '*');
  res.header('Access-Control-Allow-Headers', 'Origin, X-Requested-With, Content-Type, Accept');
  next();
});

// 获取指定路径的磁盘信息
app.get('/api/disk-info/path', async (req, res) => {
  try {
    const targetPath = req.query.path || __dirname;

    // 检查路径是否存在
    if (!fs.existsSync(targetPath)) {
      return res.status(404).json({ error: '指定的路径不存在' });
    }

    // 获取磁盘使用情况
    const usage = await diskusage.check(targetPath);

    // 计算总大小、已用空间和可用空间
    const total = usage.total;
    const free = usage.free;
    const used = total - free;

    // 计算使用率百分比
    const capacityPercent = ((used / total) * 100).toFixed(2) + '%';

    // 获取文件系统信息（在 macOS 上可能需要额外处理）
    let filesystem = '';
    let mounted = '';

    try {
      // 尝试获取挂载点信息（这部分在不同操作系统上可能需要调整）
      const mountInfo = getMountInfo(targetPath);
      filesystem = mountInfo.filesystem;
      mounted = mountInfo.mounted;
    } catch (err) {
      console.error('获取挂载点信息失败:', err);
      // 使用默认值
      filesystem = 'unknown';
      mounted = path.parse(targetPath).root;
    }

    res.json({
      filesystem: filesystem,
      mounted: mounted,
      size: total.toString(),
      used: used.toString(),
      available: free.toString(),
      capacity: capacityPercent,
      path: targetPath,
    });
  } catch (error) {
    console.error('获取磁盘信息失败:', error);
    res.status(500).json({ error: error.message });
  }
});

// 辅助函数：获取挂载点信息（支持 Windows 和 macOS）
function getMountInfo(targetPath) {
  const { execSync } = require('child_process');
  const os = require('os');

  // 检测操作系统类型
  const platform = os.platform();
  try {
    if (platform === 'win32') {
      // 获取盘符（如 C:）
      const driveLetter = path.parse(targetPath).root;
      // 使用 wmic 命令获取文件系统信息
      const output = execSync(`wmic logicaldisk where "DeviceID='${driveLetter.replace('\\', '')}'" get FileSystem, DeviceID, VolumeName /format:csv`, { encoding: 'utf8' });
      const lines = output.trim().split('\n');
      if (lines.length >= 2) {
        const parts = lines[1].split(',');
        if (parts.length >= 3) {
          return {
            filesystem: parts[2] || 'NTFS', // 文件系统类型，通常是 NTFS
            mounted: parts[1] || driveLetter, // 挂载点（盘符）
          };
        }
      }
      // 如果无法获取详细信息，返回默认值
      return {
        filesystem: 'NTFS',
        mounted: driveLetter,
      };
    } else {
      // macOS/Linux 系统
      // 使用 df 命令获取文件系统信息
      const output = execSync(`df -T "${targetPath}"`, { encoding: 'utf8' });

      const lines = output.trim().split('\n');
      if (lines.length >= 2) {
        const info = lines[1].split(/\s+/);

        // df 命令输出格式可能因系统而异，需要适当调整索引
        return {
          filesystem: info[0] || 'local',
          mounted: info[info.length - 1] || '/',
        };
      }
    }
  } catch (error) {
    console.error('获取挂载点信息失败:', error);
  }

  // 默认返回值
  return {
    filesystem: 'unknown',
    mounted: platform === 'win32' ? path.parse(targetPath).root : '/',
  };
}

app.listen(PORT, () => {
  console.log(`服务器运行在 http://localhost:${PORT}`);
});
