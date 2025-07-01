import { Input } from 'antd';
import { MagnifyingGlassIcon } from '@phosphor-icons/react';
import React, { Dispatch, FC, SetStateAction } from 'react';
import styles from './ModelSelectModal.module.scss';
import { searchFunc } from '@/pages/AppManagement/AppConfig/types.ts';

interface IModelSearchProps {
  setSearchText: Dispatch<SetStateAction<string>>;
  searchText?: string;
  onSearch: searchFunc;
}

export const ModelSearch: FC<IModelSearchProps> = ({ setSearchText, searchText, onSearch }) => {
  const handleSearch = async () => {
    await onSearch({
      keyword: searchText,
    });
  };

  return (
    <div className={styles.searchInput}>
      <Input
        allowClear
        placeholder="请输入 MCP 服务名称"
        suffix={
          <div
            className={styles.searchIcon}
            onClick={async () => {
              await handleSearch();
            }}
          >
            <MagnifyingGlassIcon
              width={16}
              height={16}
              fill="#808899"
            />
          </div>
        }
        value={searchText}
        onChange={(e) => setSearchText(e.target.value.trim())}
        onClear={async () => {
          setSearchText('');
          await onSearch({
            keyword: '',
          });
        }}
        onPressEnter={async () => {
          await handleSearch();
        }}
        style={{ width: 380 }}
      />
    </div>
  );
};
