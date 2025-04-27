import { useState, useEffect } from 'react';
import { Card, Spin, Progress, Input, Button, message } from 'antd';
import axios from 'axios';

interface DiskInfoType {
  filesystem: string;
  mounted: string;
  size: string;
  used: string;
  available: string;
  capacity: string;
  path: string;
}

const DiskInfo = () => {
  const [loading, setLoading] = useState(false);
  const [diskInfo, setDiskInfo] = useState<DiskInfoType | null>(null);
  const [path, setPath] = useState('/');
  const [error, setError] = useState('');

  const fetchDiskInfo = async (targetPath: string) => {
    setLoading(true);
    setError('');
    try {
      console.log('请求路径:', targetPath);
      const response = await axios.get(`/api/disk-info/path?path=${encodeURIComponent(targetPath)}`);
      console.log('服务器响应:', response.data);
      
      if (response.data && response.data.error) {
        setError(response.data.error);
        message.error(`获取磁盘信息失败: ${response.data.error}`);
        setDiskInfo(null);
      } else {
        setDiskInfo(response.data);
      }
    } catch (err: any) {
      console.error('获取磁盘信息失败:', err);
      setError(`获取磁盘信息失败: ${err.message || '请确保路径正确且服务器正在运行'}`);
      message.error(`获取磁盘信息失败: ${err.message || '请确保路径正确且服务器正在运行'}`);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchDiskInfo(path);
  }, []);

  const formatBytes = (bytes: string) => {
    const size = parseInt(bytes, 10);
    if (size === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(size) / Math.log(k));
    return parseFloat((size / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const handleSubmit = () => {
    fetchDiskInfo(path);
  };

  // 处理回车键提交
  const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      handleSubmit();
    }
  };

  return (
    <Card title="磁盘空间信息" style={{ width: '100%', maxWidth: 600, margin: '20px auto' }}>
      <div style={{ marginBottom: 16 }}>
        <Input 
          placeholder="输入路径，例如: /Users/lc" 
          value={path} 
          onChange={(e) => setPath(e.target.value)}
          onKeyPress={handleKeyPress}
          style={{ width: 'calc(100% - 88px)' }}
        />
        <Button type="primary" onClick={handleSubmit} style={{ marginLeft: 8 }}>
          查询
        </Button>
      </div>
      
      {error && <div style={{ color: 'red', marginBottom: 16 }}>{error}</div>}
      
      {loading ? (
        <div style={{ textAlign: 'center', padding: '20px 0' }}>
          <Spin />
        </div>
      ) : diskInfo ? (
        <div>
          <p><strong>文件系统:</strong> {diskInfo.filesystem}</p>
          <p><strong>挂载点:</strong> {diskInfo.mounted}</p>
          <p><strong>路径:</strong> {diskInfo.path}</p>
          <p><strong>总空间:</strong> {formatBytes(diskInfo.size)}</p>
          <p><strong>已用空间:</strong> {formatBytes(diskInfo.used)}</p>
          <p><strong>可用空间:</strong> {formatBytes(diskInfo.available)}</p>
          <p><strong>使用率:</strong></p>
          <Progress 
            percent={parseFloat(String(diskInfo?.capacity || '0').replace('%', ''))} 
            status={parseFloat(String(diskInfo?.capacity || '0').replace('%', '')) > 90 ? 'exception' : 'normal'}
          />
        </div>
      ) : null}
    </Card>
  );
};

export default DiskInfo;