import React, { useEffect, useRef, useState } from 'react';
import { Button, Checkbox, Input, List, message, Spin } from 'antd';
import { CheckboxChangeEvent } from 'antd/es/checkbox';
import styles from './index.module.scss';
import { MagnifyingGlassIcon } from '@phosphor-icons/react';
import { IUseViewModelReturn } from '@/components/mcp-manage/my-mcp-tab/view-model.ts';
import { IMcpListItem } from '@/components/mcp-manage/mcp-square-tab/types.ts';
import TagsRender from '@/components/tags-render';
import defaultLogo from '@/assets/favicon.png';
import useSelectMcpStore from '@/store/useSelectMcpStore.ts';
import { ChooseMcpDialog } from '@/components/choose-mcp-dialog';
import { DetailDrawer } from '@/components/detail-drawer';
import { checkMcpLength, useSelectRemoteHelper } from '@/components/select-mcp/lib/useSelectMcpHelper';
import realLoadingSvg from '@/components/icons/real-loading.svg';
import { LoadingOutlined } from '@ant-design/icons';
import { getMessageByMcp } from '@/i18n';

interface ISelectMcpDialogProps {
  setSelectMcpPopOpen: (bool: boolean) => void;
  myMcpViewModel: IUseViewModelReturn;
}

export const SelectMcpDialog = (props: ISelectMcpDialogProps) => {
  const { mcpListData, handlePageChange, pagination, mcpListLoading, onMcpInputSearch } = props.myMcpViewModel;
  const { startMcps, stopMcps } = useSelectRemoteHelper();
  const { setSelectMcpPopOpen } = props;
  const [allList, setAllList] = useState<IMcpListItem[]>([]);
  const [filteredData, setFilteredData] = useState<IMcpListItem[]>([]);

  // 组件挂载时重置列表数据
  useEffect(() => {
    setAllList(mcpListData);
    setFilteredData(mcpListData);
    return () => {
      // 组件卸载时清空数据
      setAllList([]);
      setFilteredData([]);

      // 取消所有未完成的请求
      abortControllersRef.current.forEach((controller) => {
        controller.abort();
      });
      abortControllersRef.current.clear();
    };
  }, []);
  const { setSelectMcpList, selectMcpList, drawerOpenId, setDrawerOpenId } = useSelectMcpStore();
  // 打开选择更多MCP工具弹窗
  const [open, setOpen] = useState<boolean>(false);
  // 控制是否只显示已选中的项
  const [showOnlySelected, setShowOnlySelected] = useState<boolean>(false);
  // 搜索关键字
  const [searchValue, setSearchValue] = useState<string>('');
  const [startMcpNow, setStartMcpNow] = useState<Record<string, number>>({});
  // 存储每个 MCP 的 AbortController，用于取消请求
  const abortControllersRef = useRef<Map<string, AbortController>>(new Map());

  // 处理单个项目的选择
  const handleItemSelect = (item: IMcpListItem, checked: boolean) => {
    console.log(Object.values(startMcpNow).filter((item: number) => item === 2).length, '当前正在启动的数量');
    // 当前正在启动的数量
    const isStarting = Object.values(startMcpNow).filter((item: number) => item === 2).length;
    if (checked) {
      console.info('选择了MCP:', item);
      // 如果没有超过最大数量限制并且正在启动的mcp和已启动成功的mcp数量不超过限制则能够继续添加
      if (checkMcpLength(selectMcpList.length) && checkMcpLength(isStarting + selectMcpList.length)) {
        /// 1 失败， 0成功， 2进行中
        setStartMcpNow((prevState) => ({
          ...prevState,
          [item?.id as string]: 2,
        }));

        // 创建 AbortController 用于取消请求
        const controller = new AbortController();
        // 存储 AbortController 以便后续取消
        abortControllersRef.current.set(item.id as string, controller);

        // 启动MCP
        const startPromise = startMcps({
          ids: [item?.id as string],
        });

        // 设置当请求被取消时的处理
        const cancelHandler = () => {
          console.info('MCP启动请求已取消:', item.id);
          setStartMcpNow((prevState) => ({
            ...prevState,
            [item?.id as string]: 0, // 重置状态
          }));
        };

        // 将请求和取消处理关联起来
        controller.signal.addEventListener('abort', cancelHandler);

        startPromise
          .then((res) => {
            if (controller.signal.aborted) {
              // 如果请求已被取消，不做任何处理
              return;
            }

            const { successIds = [], errorIds = [] } = res;
            /// 当前这条启用成功
            if (successIds && successIds.indexOf(item.id) > -1) {
              setStartMcpNow((prevState) => ({
                ...prevState,
                [item?.id as string]: 0,
              }));

              setSelectMcpList((prevList) => [...prevList, item]);
            }
            if (errorIds && errorIds.indexOf(item.id) > -1) {
              setStartMcpNow((prevState) => ({
                ...prevState,
                [item?.id as string]: 1,
              }));
            }
          })
          .catch((error) => {
            if (controller.signal.aborted) {
              // 如果请求已被取消，不处理错误
              return;
            }

            console.error('启动MCP失败:', error);
            setStartMcpNow((prevState) => ({
              ...prevState,
              [item?.id as string]: 1,
            }));
          })
          .finally(() => {
            if (!controller.signal.aborted) {
              console.info('加载完毕');
              // 请求完成后移除 AbortController
              abortControllersRef.current.delete(item.id as string);
            }
          });
      } else {
        message.warning(
          getMessageByMcp('maxSelectMcp', {
            msg: '为保障服务稳定运行与优质体验，建议您选择的MCP工具不要超过5个。',
          }),
        );
      }
    } else {
      setSelectMcpList((prevList) => prevList.filter((mcpItem) => mcpItem?.id !== item?.id));
      // 停止MCP
      stopMcps({
        ids: [item?.id as string],
      });
    }
  };

  // 处理筛选复选框
  const handleFilterChange = (e: CheckboxChangeEvent) => {
    setShowOnlySelected(e.target.checked);
  };

  // 处理搜索
  const handleSearch = (value: string) => {
    setAllList([]);
    onMcpInputSearch(value);
  }; /// 通过mcp的分页数据的变化来对数据进行组装
  useEffect(() => {
    // 仅在分页时累加数据，避免重复添加
    if (pagination.current > 1) {
      const _allList = allList.concat(mcpListData);
      setFilteredData(_allList);
    } else {
      // 首次加载或弹窗打开时，直接使用新数据
      setAllList(mcpListData);
      setFilteredData(mcpListData);
    }
  }, [mcpListData, pagination.current]);

  useEffect(() => {
    // 根据筛选状态和搜索值过滤数据
    const _filteredData = allList.filter((item) => {
      return showOnlySelected
        ? selectMcpList
            .map((mcpItem) => {
              return mcpItem?.id;
            })
            .includes(item.id)
        : true;
    });
    setFilteredData(_filteredData);
  }, [showOnlySelected]);

  const onLoadMore = () => {
    handlePageChange(pagination.current + 1, 12);
  };

  const loadMore =
    pagination?.total > pagination.current * 12 && !mcpListLoading && !showOnlySelected ? (
      <div
        style={{
          textAlign: 'center',
          marginTop: 12,
          height: 32,
          lineHeight: '32px',
        }}
      >
        <Button
          type="link"
          loading={mcpListLoading}
          onClick={onLoadMore}
        >
          显示更多
        </Button>
      </div>
    ) : null;
  return (
    <div className={styles.dialog_mcp}>
      <div className={styles.dialog_mcp_title}>
        <div>选择 MCP 工具</div>
        <Button
          type="link"
          onClick={() => {
            setOpen(true);
          }}
        >
          添加更多MCP工具
        </Button>
      </div>
      <div className={styles.dialog_mcp_search}>
        <div className={styles.dialog_mcp_search_input}>
          <Input
            style={{ width: '260px' }}
            allowClear
            placeholder="搜索MCP"
            onClear={() => {
              setSearchValue('');
              handleSearch('');
            }}
            suffix={
              <div
                className={styles.searchIcon}
                onClick={() => handleSearch(searchValue)}
              >
                <MagnifyingGlassIcon
                  width={16}
                  height={16}
                  fill="#808899"
                />
              </div>
            }
            value={searchValue}
            onChange={(e) => setSearchValue(e.target.value.trim())}
            onPressEnter={() => handleSearch(searchValue)}
          />
        </div>
        <div>
          <Checkbox
            checked={showOnlySelected}
            onChange={handleFilterChange}
          >
            仅显示已选
          </Checkbox>
        </div>
      </div>

      <List
        loadMore={loadMore}
        style={{ height: '440px', overflowY: 'auto', overflowX: 'hidden' }}
        dataSource={filteredData}
        loading={{
          spinning: mcpListLoading,
          indicator: (
            <img
              src={realLoadingSvg}
              alt="loading"
            />
          ),
        }}
        renderItem={(item) => (
          <List.Item
            actions={[
              ...(startMcpNow[item?.id as string] === 2
                ? [
                    <Button
                      type="link"
                      size="small"
                      onClick={() => {
                        /// 取消当前这个id对应的接口
                        abortControllersRef.current.get(item.id as string)?.abort();
                        abortControllersRef.current.delete(item.id as string); // 移除已取消的请求
                        setStartMcpNow((prevState) => ({
                          ...prevState,
                          [item?.id as string]: 0,
                        }));
                      }}
                    >
                      取消
                    </Button>,
                  ]
                : []),
              <Button
                type="link"
                size="small"
                onClick={() => {
                  setDrawerOpenId(item?.id as string);
                }}
              >
                查看详情
              </Button>,
            ]}
          >
            <div className={styles.select_mcp_title_warp}>
              {startMcpNow[item?.id as string] === 2 ? (
                <Spin
                  style={{ marginRight: 16 }}
                  indicator={<LoadingOutlined spin />}
                  size="small"
                ></Spin>
              ) : (
                <Checkbox
                  checked={selectMcpList
                    .map((item) => {
                      return item?.id;
                    })
                    .includes(item.id)}
                  onChange={(e) => {
                    e.stopPropagation();
                    handleItemSelect(item, e.target.checked);
                  }}
                  style={{ marginRight: 16 }}
                />
              )}
              <div className={styles.select_mcp_title_logo}>
                <img
                  src={item?.logo?.trim() ? item.logo : defaultLogo}
                  alt=""
                />
              </div>
              <div>
                <div>{item.name.zh}</div>
                <div style={{ width: '220px' }}>
                  <TagsRender tags={item.tags} />
                </div>
              </div>
            </div>
          </List.Item>
        )}
      />

      <ChooseMcpDialog
        open={open}
        onSelectMcpOkProps={() => {
          setOpen(false);
          setSelectMcpPopOpen(false);
        }}
        onCancelProps={() => {
          setOpen(false);
          // 直接关闭最底层的选择mcp弹窗（mcp添加有延时导致不能立即更新列表，和产品确认可以有延时）
          setSelectMcpPopOpen(false);
        }}
      />
      <DetailDrawer
        id={drawerOpenId}
        open={!!(drawerOpenId && drawerOpenId !== '')}
        onClose={() => setDrawerOpenId('')}
      />
    </div>
  );
};
