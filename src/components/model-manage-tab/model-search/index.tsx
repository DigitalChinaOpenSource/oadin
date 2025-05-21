import styles from './index.module.scss';
import { Input, Select } from 'antd';
import { SearchIcon } from '@/components/icons';
import { IModelSourceType } from '@/types';

export interface IModelTitleSearchProps {
  modelSearchVal: string;
  modelSourceVal: string;
  onModelSearch: (val: string) => void;
  onModelSourceChange: (val: IModelSourceType) => void;
}

const { Search } = Input;

export default function ModelSearch(props: IModelTitleSearchProps) {
  const { onModelSearch, modelSearchVal, modelSourceVal, onModelSourceChange } = props;
  const options = [
    { value: 'local', label: '本地模型' },
    { value: 'remote', label: '云端模型' },
  ];
  return (
    <div className={styles.modelSearch}>
      <div className={styles.searchInput}>
        <Search
          allowClear
          placeholder="请输入模型名称"
          value={modelSearchVal}
          onSearch={(value) => onModelSearch(value.trim())}
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
