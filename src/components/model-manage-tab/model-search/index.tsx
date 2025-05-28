import { useState } from 'react';
import { Input, Select } from 'antd';
import { SearchIcon } from '@/components/icons';
import { IModelSourceType } from '@/types';
import styles from './index.module.scss';

export interface IModelTitleSearchProps {
  modelSearchVal: string;
  modelSourceVal: string;
  onModelSearch: (val: string) => void;
  onModelSourceChange: (val: IModelSourceType) => void;
}

export default function ModelSearch(props: IModelTitleSearchProps) {
  const { onModelSearch, modelSearchVal, modelSourceVal, onModelSourceChange } = props;
  const [searchVal, setSearchVal] = useState<string>(modelSearchVal);

  const options = [
    { value: 'local', label: '本地模型' },
    { value: 'remote', label: '云端模型' },
  ];
  return (
    <div className={styles.modelSearch}>
      <div className={styles.searchInput}>
        <Input
          allowClear
          placeholder="请输入模型名称"
          suffix={
            <div
              className={styles.searchIcon}
              onClick={() => onModelSearch(searchVal)}
            >
              <SearchIcon />
            </div>
          }
          value={searchVal}
          onChange={(e) => setSearchVal(e.target.value.trim())}
          onPressEnter={() => onModelSearch(searchVal)}
          onClear={() => onModelSearch('')}
          style={{ width: 380 }}
        />
      </div>
      <div className={styles.modelSelect}>
        <Select
          style={{ width: 160 }}
          value={modelSourceVal as IModelSourceType}
          options={options}
          onChange={onModelSourceChange}
        />
      </div>
    </div>
  );
}
