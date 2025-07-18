import { useNavigate, useSearchParams } from 'react-router-dom';
import { useCallback, useEffect, useState } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';
import { McpDetailType } from '@/components/mcp-manage/mcp-detail/type.ts';
import testDta from './mcp_schema.json';
import { message, Modal } from 'antd';
import useMcpDownloadStore from '@/store/useMcpDownloadStore.ts';
import { IMcpListItem } from '@/components/mcp-manage/mcp-square-tab/types.ts';
import useSelectMcpStore from '@/store/useSelectMcpStore.ts';
import { useSelectRemoteHelper } from '@/components/select-mcp/lib/useSelectMcpHelper.ts';

export const useMcpDetail = (id?: string | number, setMcpListData?: (value: ((prevState: IMcpListItem[]) => IMcpListItem[]) | IMcpListItem[]) => void) => {
  const navigate = useNavigate();
  const serviceId = id;
  const { addMcpDownloadItemInit, addMcpDownloadItem } = useMcpDownloadStore();
  const { selectMcpList, setSelectMcpList } = useSelectMcpStore();
  const { stopMcps } = useSelectRemoteHelper();
  const [mcpDetail, setMcpDetail] = useState<McpDetailType | null>(null);
  const [showMcpModal, setShowMcpModal] = useState(false);

  // 获取 mcp 详情
  const { loading: mcpDetailLoading, runAsync: fetchMcpDetail } = useRequest(
    async () => {
      const data = await httpRequest.get<McpDetailType>(`/mcp/${serviceId}`);
      return data || testDta;
    },
    {
      manual: true,
      onSuccess: (data) => {
        setMcpDetail(data);
      },
      onError: (error) => {
        console.error('获取mcp详情失败:', error);
      },
    },
  );

  // 下载mcp
  const { loading: downMcpLoading, run: downMcp } = useRequest(
    async (curMcpDetail?: McpDetailType) => {
      addMcpDownloadItemInit({
        mcpDetail: curMcpDetail || mcpDetail,
        downStatus: 'downloading',
      });
      return await httpRequest.put<McpDetailType>(`/mcp/${curMcpDetail ? curMcpDetail.id : serviceId}/download`, null, {
        timeout: 2 * 60 * 1000,
        needMcpStore: true,
        mcpId: curMcpDetail ? curMcpDetail.id : serviceId,
        addMcpDownloadItem,
      });
    },
    {
      manual: true,
      onSuccess: (data) => {
        console.log('下载mcp===>', data);
        // 下载成功修改 下载状态值为1   1为已授权（若需授权）已下载 --- 一切就绪
        setMcpDetail({
          ...(mcpDetail as McpDetailType),
          status: 1,
        });
        // 如果是列表页操作，且mcp添加成功，则更新mcp列表的对应状态
        setMcpListData?.((preList: IMcpListItem[]) =>
          preList.map((item: IMcpListItem) => {
            return item.id === serviceId ? { ...item, status: 1 } : item;
          }),
        );
        message.success('mcp添加成功');
        // 添加成功后，判断是否是对话中已选的mcp，如果是则清除该已选mcp
        if (selectMcpList.find((item) => item.id === serviceId)) {
          // 如果是已选的mcp，则先停止mcp
          stopMcps({ ids: [serviceId as string] }); // 停止远端mcp
          setSelectMcpList((preList) => preList.filter((item) => item.id !== serviceId));
        }
      },
      onError: (error) => {
        console.error('下载本地mcp失败:', error);
      },
    },
  );

  // 远端mcp授权
  const { loading: authMcpLoading, runAsync: authMcp } = useRequest(
    async (authParams: any, curMcpDetail?: McpDetailType) => {
      return await httpRequest.put<McpDetailType>(`/mcp/${curMcpDetail ? curMcpDetail.id : serviceId}/auth`, authParams);
    },
    {
      manual: true,
      onSuccess: (data, params) => {
        console.log('授权mcp===>', data);
        setMcpDetail({
          ...(mcpDetail as McpDetailType),
          authorized: 1, // 1为已授权
        });
        // 授权成功 开始下载
        if (params[1]) {
          downMcp(params[1]);
        } else {
          downMcp();
        }
      },
      onError: (error) => {
        console.error('授权mcp失败:', error);
      },
    },
  );

  //添加mcp点击
  const handleAddMcp = async (mcpDetail: McpDetailType) => {
    try {
      const { hosted, status, envRequired } = mcpDetail || {};
      if (envRequired === 0) {
        // 不需要授权
        Modal.confirm({
          title: '确认添加吗？',
          okText: '确认',
          centered: true,
          okButtonProps: {
            style: { backgroundColor: '#4f4dff' },
          },
          onOk() {
            downMcp(mcpDetail); // 直接下载mcp
          },
          onCancel() {
            console.log('Cancel');
          },
        });
      } else {
        // 显示授权弹窗
        setShowMcpModal(true);
      }
    } catch (error) {
      console.error('添加mcp失败:', error);
    }
  };

  // 取消mcp
  const { loading: cancelMcpLoading, run: handleCancelMcp } = useRequest(
    async (isMyMcp?: boolean, onPageChange?: (page: number) => void) => {
      return await httpRequest.put(`/mcp/${serviceId}/reverse`);
    },
    {
      manual: true,
      onSuccess: (data, params) => {
        console.log('授权mcp===>', data);
        setMcpDetail({
          ...(mcpDetail as McpDetailType),
          status: 0,
        });
        // 取消添加mcp成功后，更新mcp列表， 将对应item的状态改为0
        setMcpListData?.((preList: IMcpListItem[]) =>
          preList.map((item: IMcpListItem) => {
            return item.id === serviceId ? { ...item, status: 0 } : item;
          }),
        );
        // 取消添加成功后，判断是否是对话中已选的mcp，如果是则清除该已选mcp
        if (selectMcpList.find((item) => item.id === serviceId)) {
          // 如果是已选的mcp，则先停止mcp
          stopMcps({ ids: [serviceId as string] }); // 停止远端mcp
          setSelectMcpList((preList) => preList.filter((item) => item.id !== serviceId));
        }

        message.success('取消添加mcp成功');
        // 如果是在我的mcp列表页，删除mcp时刷新页面到第一页
        if (params[0] && params[1]) {
          params[1]?.(1); // 如果是我的mcp 则回到第一页
        }
      },
      onError: (error) => {},
    },
  );

  // 授权mcp 确认
  const handleAuthMcp = async (authParams: any, curMcpDetail?: McpDetailType) => {
    setShowMcpModal(false);
    try {
      if (curMcpDetail) {
        await authMcp(authParams, curMcpDetail);
      } else {
        await authMcp(authParams);
      }
    } catch (error) {
      console.error('授权mcp失败:', error);
    }
  };

  // 页面返回
  const handleGoBack = (): void => {
    navigate(-1);
  };

  return {
    mcpDetailLoading,
    downMcpLoading,
    handleGoBack,
    mcpDetail,
    setMcpDetail,
    handleAddMcp,
    handleCancelMcp,
    cancelMcpLoading,
    showMcpModal,
    setShowMcpModal,
    authMcpLoading,
    authMcp,
    handleAuthMcp,
    fetchMcpDetail,
    serviceId,
  };
};
