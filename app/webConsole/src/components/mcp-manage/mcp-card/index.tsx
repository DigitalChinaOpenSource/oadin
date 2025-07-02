import React, { Dispatch, SetStateAction, useEffect, useState } from 'react';
import styles from './index.module.scss';
import { Button, Checkbox, Popover } from 'antd';
import { IMcpListItem } from '../mcp-square-tab/types';
import { DotsThreeCircleIcon, GlobeIcon, HardDrivesIcon } from '@phosphor-icons/react';
import TagsRender from '@/components/tags-render';
import dayjs from 'dayjs';
import defaultPng from '@/assets/favicon.png';
import { ExclamationCircleOutlined, PlusCircleOutlined } from '@ant-design/icons';
import McpAuthModal from '@/components/mcp-manage/mcp-detail/mcp-auth-modal';
import { useMcpDetail } from '@/components/mcp-manage/mcp-detail/useMcpDetail.ts';
import { McpDetailType } from '@/components/mcp-manage/mcp-detail/type.ts';
import EllipsisTooltip from '@vanta/ellipsis-tooltip';
import useMcpDownloadStore from '@/store/useMcpDownloadStore.ts';
import { checkMcpLength } from '@/components/select-mcp/lib/useSelectMcpHelper';

export interface IMcpCardProps {
  // 模型数据
  mcpData: IMcpListItem;
  handelMcpCardClick: (mcpId: string | number) => void;
  isSelectable?: boolean;
  selectTemporaryMcpItems?: IMcpListItem[];
  setSelectTemporaryMcpItems?: Dispatch<SetStateAction<IMcpListItem[]>>;
  handleMcpListToPage?: (page?: number) => void;
  isMyMcp?: boolean; // 是否是我的MCP
}

export default function McpCard(props: IMcpCardProps) {
  const { mcpData, handelMcpCardClick, isSelectable, setSelectTemporaryMcpItems, selectTemporaryMcpItems = [], isMyMcp = false } = props;

  const { fetchMcpDetail, mcpDetail, handleAddMcp, handleCancelMcp, cancelMcpLoading, downMcpLoading, authMcpLoading, handleAuthMcp, showMcpModal, setShowMcpModal } = useMcpDetail(mcpData.id);

  // 获取全局的下载中和失败的mcp
  const { mcpDownloadList } = useMcpDownloadStore();

  const formatUnixTime = (unixTime: number) => {
    const date = dayjs.unix(unixTime);
    return date.format('YYYY-MM-DD');
  };
  // 处理单个项目的选择
  const handleItemSelect = (item: IMcpListItem, checked: boolean) => {
    if (checked) {
      if (checkMcpLength(selectTemporaryMcpItems.length)) {
        setSelectTemporaryMcpItems?.([...(selectTemporaryMcpItems ?? []), item]);
      }
    } else {
      setSelectTemporaryMcpItems?.(selectTemporaryMcpItems.filter((mcpItem) => mcpItem?.id !== item?.id));
    }
  };

  const [showOperate, setShowOperate] = useState(false);

  // 根据全局store的MCP设置状态
  const [showLoading, setShowLoading] = useState(false);

  const clickBefore = async () => {
    if (!mcpDetail) {
      return fetchMcpDetail();
    } else {
      return mcpDetail;
    }
  };
  useEffect(() => {
    const currentItem = mcpDownloadList.find((item) => item.mcpDetail?.id === mcpData.id);
    if (!currentItem) return setShowLoading(false);
    // 如果找到当前mcp的下载状态，则设置对应的loading状态
    if (currentItem.downStatus === 'downloading') {
      setShowLoading(true);
    } else {
      setShowLoading(false);
      if (currentItem.downStatus === 'success') {
        fetchMcpDetail();
      }
    }
  }, [mcpDownloadList]);

  // status为0表示未添加，1表示已添加---isAdd为true表示未添加，false表示已添加
  const isAdd = mcpDetail ? mcpDetail?.status === 0 : mcpData?.status === 0;

  return (
    <div
      className={styles.mcpCard}
      onClick={(e) => {
        // 添加检查，如果已经有Popover或Modal操作正在进行，则不触发卡片点击
        if (showOperate || showMcpModal) {
          return;
        }
        // 如果是弹窗
        if (isSelectable) {
          // 如果是已添加的MCP
          if (!isAdd) {
            if (selectTemporaryMcpItems.find((item) => item.id === mcpData.id)) {
              handleItemSelect(mcpData, false);
            } else {
              handleItemSelect(mcpData, true);
            }
          }
        }
      }}
    >
      <div className={styles.cardHeader}>
        <div className={styles.cardTitle}>
          {/* logo */}
          <div className={styles.avatar}>
            <img
              alt=""
              src={mcpData?.logo || defaultPng}
              width={24}
            />
          </div>
          {/* 名称 */}
          <EllipsisTooltip
            className={styles.title}
            title={mcpData?.name?.zh}
          >
            {mcpData?.name?.zh}
          </EllipsisTooltip>
          {/* 本地还是云端 */}
          <div className={styles.localOrCloud}>
            {mcpData?.hosted ? (
              <>
                <HardDrivesIcon
                  width={16}
                  height={16}
                  fill="#898ea3"
                />
                <div className={styles.localOrCloudText}>本地</div>
              </>
            ) : (
              <>
                <GlobeIcon
                  width={16}
                  height={16}
                  fill="#898ea3"
                />
                <div className={styles.localOrCloudText}>云端</div>
              </>
            )}
          </div>
        </div>

        {isSelectable ? (
          <div>
            <Checkbox
              checked={selectTemporaryMcpItems
                .map((item) => {
                  return item?.id;
                })
                .includes(mcpData.id)}
              disabled={isAdd}
              onChange={(e) => handleItemSelect(mcpData, e.target.checked)}
            />
          </div>
        ) : null}
      </div>
      <div style={{ height: '24px' }}>
        <TagsRender tags={mcpData?.tags || []} />
      </div>

      {/* 修改：使用 EllipsisTooltip 组件 */}
      <EllipsisTooltip
        title={mcpData?.abstract?.zh}
        className={styles.contentWrapper}
        maxWidth={400}
      >
        {mcpData?.abstract?.zh}
      </EllipsisTooltip>

      <div className={styles.infoWrapper}>
        <div className={styles.providerName}>{mcpData?.supplier}</div>
        <div className={styles.dot}>·</div>
        <div className={styles.updateName}>{mcpData?.updatedAt && formatUnixTime(mcpData?.updatedAt) + '更新'}</div>
        {/*{!!mcpData.status && <div className={styles.mcpStatus}>已添加</div>}*/}
      </div>
      <div
        className={styles.cardOperate}
        onClick={(e) => {
          // 防止卡片点击事件触发
          e.stopPropagation();
        }}
      >
        <Button
          type={'text'}
          icon={<ExclamationCircleOutlined />}
          onClick={() => {
            handelMcpCardClick(mcpData?.id);
          }}
        >
          查看详情
        </Button>
        <div className={styles.splitLine}></div>
        {isAdd && (
          <Button
            type={'link'}
            icon={<PlusCircleOutlined />}
            loading={showLoading || authMcpLoading || downMcpLoading || cancelMcpLoading}
            onClick={async (e) => {
              e.stopPropagation();
              e.preventDefault();
              const data = await clickBefore();
              if (data) await handleAddMcp(data);
            }}
          >
            立即添加
          </Button>
        )}
        {(mcpDetail ? mcpDetail?.status === 1 : mcpData?.status === 1) && (
          <Popover
            trigger="click"
            arrow={false}
            open={showOperate}
            onOpenChange={(visible) => {
              setShowOperate(visible);
              // 给事件处理一个延时，避免冒泡到卡片
              // if (!visible) {
              //   setTimeout(() => {}, 100);
              // }
            }}
            styles={{ body: { padding: 0 } }}
            content={
              <div
                className={styles.moreOperate}
                onClick={(e) => {
                  // 阻止所有点击事件冒泡到卡片
                  e.stopPropagation();
                }}
              >
                <div
                  className={styles.operateBtn}
                  onClick={async () => {
                    setShowOperate(false);
                    await handleAddMcp(mcpDetail as McpDetailType);
                  }}
                >
                  更新
                </div>
                <div
                  className={styles.operateBtn}
                  onClick={() => {
                    setShowOperate(false);
                    if (isMyMcp) {
                      handleCancelMcp(isMyMcp, props.handleMcpListToPage);
                    } else {
                      handleCancelMcp();
                    }
                  }}
                >
                  取消添加
                </div>
              </div>
            }
          >
            <Button
              type={'text'}
              icon={<DotsThreeCircleIcon />}
              onClick={async () => {
                setShowOperate(!showOperate);
                await clickBefore();
              }}
              disabled={mcpData?.envRequired === 0}
              loading={showLoading || authMcpLoading || downMcpLoading || cancelMcpLoading}
            >
              更多操作
            </Button>
          </Popover>
        )}
      </div>
      {mcpDetail && (
        <McpAuthModal
          mcpDetail={mcpDetail}
          handleAuthMcp={handleAuthMcp}
          setShowMcpModal={setShowMcpModal}
          showMcpModal={showMcpModal}
          operateType={mcpDetail?.status === 0 ? 'add' : 'edit'}
        />
      )}
    </div>
  );
}
