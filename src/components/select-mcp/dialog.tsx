import React, { useEffect, useState } from 'react';
import { List, Checkbox, Button, Input } from 'antd';
import { CheckboxChangeEvent } from 'antd/es/checkbox';
import styles from './index.module.scss';
import { MagnifyingGlassIcon } from '@phosphor-icons/react';
import { useViewModel } from '@/components/mcp-manage/mcp-square-tab/view-model.ts';
import { IMcpListItem } from '@/components/mcp-manage/mcp-square-tab/types.ts';
import TagsRender from '@/components/tags-render';
import defaultLogo from '@/assets/favicon.png';
import useSelectMcpStore from '@/store/useSelectMcpStore.ts';
import { ChooseMcpDialog } from '@/components/choose-mcp-dialog';
import { DetailDrawer } from '@/components/detail_drawer';

interface ISelectMcpDialogProps {
  setSelectMcpPopOpen: (bool: boolean) => void;
}

export const SelectMcpDialog = (props: ISelectMcpDialogProps) => {
  const { handlePageChange, mcpListData, pagination, mcpListLoading, onMcpInputSearch } = useViewModel();
  const { setSelectMcpPopOpen } = props;
  const [allList, setAllList] = useState<IMcpListItem[]>([]);
  const [filteredData, setFilteredData] = useState<IMcpListItem[]>([]);
  const { setSelectMcpList, selectMcpList, drawerOpenId, setDrawerOpenId } = useSelectMcpStore();
  // 打开选择更多MCP工具弹窗
  const [open, setOpen] = useState<boolean>(false);
  // 控制是否只显示已选中的项
  const [showOnlySelected, setShowOnlySelected] = useState<boolean>(false);
  // 搜索关键字
  const [searchValue, setSearchValue] = useState<string>('');

  // 处理单个项目的选择
  const handleItemSelect = (item: IMcpListItem, checked: boolean) => {
    if (checked) {
      setSelectMcpList([...selectMcpList, item]);
    } else {
      setSelectMcpList(selectMcpList.filter((mcpItem) => mcpItem?.id !== item?.id));
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
  };
  /// 通过mcp的分页数据的变化来对数据进行组装
  useEffect(() => {
    const _allList = allList.concat(mcpListData);
    setAllList(_allList);

    setFilteredData(_allList);
  }, [mcpListData]);

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
    pagination?.total > 12 && !mcpListLoading && !showOnlySelected ? (
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
            onChange={(e) => setSearchValue(e.target.value)}
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
        style={{ height: '440px', overflowY: 'scroll', overflowX: 'hidden' }}
        dataSource={filteredData}
        renderItem={(item) => (
          <List.Item
            actions={[
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
              <Checkbox
                checked={selectMcpList
                  .map((item) => {
                    return item?.id;
                  })
                  .includes(item.id)}
                onChange={(e) => handleItemSelect(item, e.target.checked)}
                style={{ marginRight: 16 }}
              />
              <div className={styles.select_mcp_title_logo}>
                <img
                  src={item?.logo?.trim() ? item.logo : defaultLogo}
                  alt=""
                />
              </div>
              <div>
                <div>{item.name.zh}</div>
                <div style={{ width: '260px' }}>
                  <TagsRender tags={item.tags} />
                </div>
              </div>
            </div>
          </List.Item>
        )}
      />

      <ChooseMcpDialog
        open={open}
        onCancelProps={() => {
          setOpen(false);
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
