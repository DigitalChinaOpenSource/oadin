import { useSearchParams } from 'react-router-dom';
import { httpRequest } from '@/utils/httpRequest.ts';
import { useRequest } from 'ahooks';
import { useEffect, useState } from 'react';
import { CardItemType } from '@/components/mcp-manage/mcp-detail/recommended-client/type.ts';
import { Modal, notification } from 'antd';

export function useRecommendedClient() {
  // const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const serviceId = searchParams.get('serviceId');
  const [clients, setClients] = useState<Record<string, any>[]>([]);

  // 获取推荐客户端
  const { loading: clientLoading, run: getClients } = useRequest(
    async () => {
      // await new Promise((resolve) => setTimeout(resolve, 2000));
      const data = await httpRequest.get(`/mcp/${serviceId}/clients`);
      if (!data) throw new Error('获取推荐客户端失败');
      return data;
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('获取推荐客户端===>', data);
        setClients(data);
      },
      onError: (error) => {
        console.error('获取推荐客户端失败:', error);
      },
    },
  );

  // 检查是否调起第三方应用
  const checkProtocol = (protocol: string | number) => {
    return new Promise((resolve) => {
      const timeout = 2000; // 2秒超时检测
      let timer: any = null;

      // 监听窗口是否失去焦点（如果跳转成功，页面会失去焦点）
      const onBlur = () => {
        clearTimeout(timer);
        window.removeEventListener('blur', onBlur);
        resolve(true); // 协议可用
      };

      window.addEventListener('blur', onBlur);

      // 尝试打开协议
      window.location.href = `${protocol}://`; // 或 window.open()

      // 如果 2 秒后未跳转，则认为协议未注册
      timer = setTimeout(() => {
        window.removeEventListener('blur', onBlur);
        resolve(false); // 协议不可用
      }, timeout);
    });
  };

  // 客户端点击
  const handleClick = (protocol: string | number, data: CardItemType) => {
    if (!protocol) return;
    checkProtocol(protocol).then((isAvailable) => {
      if (isAvailable) {
        window.open(`${protocol}://`);
      } else {
        if (data?.linkCommand) {
          Modal.confirm({
            title: '未检测到应用，即将前往官网，是否确认？',
            okText: '确认',
            centered: true,
            okButtonProps: {
              style: { backgroundColor: '#5429ff' },
            },
            onOk() {
              window.open(data?.linkCommand, '_blank');
            },
            onCancel() {
              console.log('Cancel');
            },
          });
        }
      }
    });
  };

  useEffect(() => {
    getClients();
  }, [serviceId]);

  return { clientLoading, clients, handleClick };
}
