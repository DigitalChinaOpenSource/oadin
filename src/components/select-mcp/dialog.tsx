import React, { useState } from 'react';
import { List, Checkbox, Button, Input } from 'antd';
import { CheckboxChangeEvent } from 'antd/es/checkbox';
import styles from './index.module.scss';
import { SearchIcon } from '@/components/icons';
import { useViewModel } from '@/components/mcp-manage/mcp-square-tab/view-model.ts';
import { IMcpListItem } from '@/components/mcp-manage/mcp-square-tab/types.ts';
import TagsRender from '@/components/tags-render';

export const SelectMcpDialog: React.FC = () => {
  const vm = useViewModel();
  console.info(vm, 'vmvm');
  // 示例数据
  // const [dataSource, setDataSource] = useState<ListItem[]>([
  //   { id: 1, name: '项目 1', description: '这是第一个项目描述' },
  //   { id: 2, name: '项目 2', description: '这是第二个项目描述' },
  //   { id: 3, name: '项目 3', description: '这是第三个项目描述' },
  //   { id: 4, name: '项目 4', description: '这是第四个项目描述' },
  //   { id: 5, name: '项目 5', description: '这是第五个项目描述' },
  // ]);
  const dataSource: IMcpListItem[] = vm?.mcpListData.map((item) => {
    return {
      ...item,
      id: Number(item.id),
    };
  });

  // 存储选中项的ID
  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  // 控制是否只显示已选中的项
  const [showOnlySelected, setShowOnlySelected] = useState<boolean>(false);
  // 搜索关键字
  const [searchValue, setSearchValue] = useState<string>('');

  // 处理单个项目的选择
  const handleItemSelect = (id: number, checked: boolean) => {
    if (checked) {
      setSelectedIds([...selectedIds, id]);
    } else {
      setSelectedIds(selectedIds.filter((itemId) => itemId !== id));
    }
  };

  // 处理筛选复选框
  const handleFilterChange = (e: CheckboxChangeEvent) => {
    setShowOnlySelected(e.target.checked);
  };

  // 处理搜索
  const handleSearch = (value: string) => {
    setSearchValue(value);
  };

  // 根据筛选状态和搜索值过滤数据
  const filteredData = dataSource.filter((item) => {
    const matchesSearch = item.name.zh.includes(searchValue);
    return showOnlySelected ? matchesSearch && selectedIds.includes(item.id as number) : matchesSearch;
  });

  return (
    <div className={styles.dialog_mcp}>
      <div className={styles.dialog_mcp_title}>
        <div>选择 MCP 工具</div>
        <Button type="link">添加更多MCP工具</Button>
      </div>
      <div className={styles.dialog_mcp_search}>
        <div className={styles.dialog_mcp_search_input}>
          <Input
            style={{ width: '260px' }}
            allowClear
            placeholder="搜索MCP"
            suffix={
              <div
                className={styles.searchIcon}
                onClick={() => handleSearch(searchValue)}
              >
                <SearchIcon />
              </div>
            }
            onChange={(e) => handleSearch(e.target.value)}
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
        style={{ height: '440px', overflowY: 'scroll' }}
        dataSource={filteredData}
        renderItem={(item) => (
          <List.Item
            actions={[
              <Button
                size="small"
                onClick={() => {}}
              >
                查看详情
              </Button>,
            ]}
          >
            <div>
              <Checkbox
                checked={selectedIds.includes(Number(item.id))}
                onChange={(e) => handleItemSelect(item.id as number, e.target.checked)}
                style={{ marginRight: 16 }}
              />
              <div>
                <div>{item.name.zh}</div>
                <TagsRender tags={item.tags} />
              </div>
            </div>
          </List.Item>
        )}
      />
    </div>
  );
};
