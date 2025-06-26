import React, { Dispatch, FC, SetStateAction } from 'react';
import { List } from 'antd';
import { ICardDeatilItem, searchFunc } from '@/pages/AppManagement/AppConfig/types.ts';
import styles from './ModelSelectModal.module.scss';
import ModelDeatilCard from '@/pages/AppManagement/AppConfig/ModelDeatilCard';

interface IModelContentProps {
  onSearch: searchFunc;
  filterSearchList: ICardDeatilItem[];
  showOnlySelecte?: boolean;
  selectedModelIds?: string[];
  setSelectedModelIds?: Dispatch<SetStateAction<string[]>>;
  setDrawerOpenId?: Dispatch<SetStateAction<string>>;
}

export const ModelContent: FC<IModelContentProps> = ({ setDrawerOpenId, selectedModelIds, setSelectedModelIds, showOnlySelecte, onSearch, filterSearchList }) => {
  const handleSearch = async () => {
    await onSearch({});
  };
  console.info(filterSearchList, '当前渲染的数据列表');

  return (
    <div className={styles.mcpCardList}>
      <List
        rowKey="id"
        grid={{ gutter: 16, column: 2, xs: 1, sm: 1, md: 2, lg: 2, xl: 2, xxl: 2 }}
        dataSource={filterSearchList}
        pagination={
          showOnlySelecte
            ? false // 筛选状态下强制关闭分页
            : // : pagination.total >= pagination.pageSize
              //   ? { className: styles.mcpListPagination, align: 'end', ...pagination, pageSizeOptions: [12, 24, 48, 96], showSizeChanger: true, onChange: onPageChange }
              false
        }
        renderItem={(item) => (
          <List.Item key={item.id}>
            <ModelDeatilCard
              setSelectedModelIds={setSelectedModelIds}
              setDrawerOpenId={setDrawerOpenId}
              selectedModelIds={selectedModelIds}
              deatilData={item}
            />
          </List.Item>
        )}
      />
    </div>
  );
};
