import { useState, useEffect } from 'react';
import axios from 'axios';
import { message } from 'antd';

// 定义磁盘信息类型
export interface DiskInfoType {
  filesystem: string;
  mounted: string;
  size: string;
  used: string;
  available: string;
  capacity: string;
  path: string;
}

// 定义返回类型
interface UseDiskInfoReturn {
  diskInfo: DiskInfoType | null;
  loading: boolean;
  error: string;
  fetchDiskInfo: (targetPath: string) => Promise<void>;
  formatBytes: (bytes: string) => string;
}

/**
 * 获取当前操作系统的默认根路径
 * @returns 默认根路径
 */
const getDefaultRootPath = (): string => {
  // 检测是否在浏览器环境中
  if (typeof window !== 'undefined' && window.navigator) {
    const userAgent = window.navigator.userAgent.toLowerCase();
    // 检测是否为 Windows 系统
    if (userAgent.indexOf('win') !== -1) {
      return 'C:\\'; // Windows 默认盘符
    }
  }
  // 默认返回 Unix/Linux/macOS 的根路径
  return '/';
};

/**
 * 用于获取指定路径的磁盘信息
 * @param initialPath 初始路径，默认为系统根目录
 * @returns 包含磁盘信息、加载状态、错误信息和获取方法的对象
 */
export default function useDiskInfo(initialPath?: string): UseDiskInfoReturn {
  // 如果没有提供初始路径，则使用系统默认根路径
  const defaultPath = initialPath || getDefaultRootPath();
  const [loading, setLoading] = useState<boolean>(false);
  const [diskInfo, setDiskInfo] = useState<DiskInfoType | null>(null);
  const [error, setError] = useState<string>('');

  /**
   * 获取指定路径的磁盘信息
   * @param targetPath 目标路径
   */
  const fetchDiskInfo = async (targetPath: string): Promise<void> => {
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
      const errorMessage = err.message || '请确保路径正确且服务器正在运行';
      setError(`获取磁盘信息失败: ${errorMessage}`);
      message.error(`获取磁盘信息失败: ${errorMessage}`);
    } finally {
      setLoading(false);
    }
  };

  /**
   * 格式化字节数为可读格式
   * @param bytes 字节数字符串
   * @returns 格式化后的字符串，如 "1.5 GB"
   */
  const formatBytes = (bytes: string): string => {
    const size = parseInt(bytes, 10);
    if (size === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(size) / Math.log(k));
    return parseFloat((size / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  // 组件挂载时获取初始路径的磁盘信息
  useEffect(() => {
    fetchDiskInfo(defaultPath);
  }, [defaultPath]);

  // 返回所有需要的状态和方法
  return {
    diskInfo,
    loading,
    error,
    fetchDiskInfo,
    formatBytes,
  };
}
