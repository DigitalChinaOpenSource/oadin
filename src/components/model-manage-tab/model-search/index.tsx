import styles from './index.module.scss';
import { Input, Select } from 'antd';
import { SearchIcon } from '../../icons';
import { IModelSourceType } from '../types';

interface IModelTitleSearchProps {
  onModelSearch: (val: string) => void;
  modelSearchVal: string;
  modelSourceVal: string;
  onModelSourceChange: (val: IModelSourceType) => void;
}

export default function ModelSearch(props: IModelTitleSearchProps) {
  const { onModelSearch, modelSearchVal, modelSourceVal, onModelSourceChange } = props;
  const options = [
    { value: 'local', label: '本地模型' },
    { value: 'remote', label: '云端模型' },
    { value: 'all', label: '全部模型' },
  ];
  return (
    <div className={styles.modelSearch}>
      <div className={styles.searchInput}>
        <Input
          prefix={<SearchIcon />}
          allowClear
          placeholder="请输入模型名称"
          value={modelSearchVal}
          onChange={(e) => onModelSearch(e.target.value.trim())}
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
