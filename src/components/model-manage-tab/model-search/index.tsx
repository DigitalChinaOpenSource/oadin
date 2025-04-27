import styles from './index.module.scss'
import { Input, Select } from 'antd';

interface IModelTitleSearchProps {
  onModelSearch: (val: string) => void;
  modelSearchVal: string;
  modelSourceVal: string;
  onModelSourceChange: (val: string) => void;
}

export default function ModelSearch(props: IModelTitleSearchProps) {
  const { onModelSearch, modelSearchVal, modelSourceVal, onModelSourceChange } = props;
  const options = [
    { value: 'local', label: '本地模型' },
    { value: 'remote', label: '云端模型' },
    { value: 'all', label: '全部模型' }
  ];
  return (
    <div className={styles.modelSearch}>
      <div className={styles.searchInput}>
        <Input
          prefix={<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="#808899" viewBox="0 0 256 256"><path d="M229.66,218.34l-50.07-50.06a88.11,88.11,0,1,0-11.31,11.31l50.06,50.07a8,8,0,0,0,11.32-11.32ZM40,112a72,72,0,1,1,72,72A72.08,72.08,0,0,1,40,112Z"></path></svg>}
          placeholder='请输入模型名称'
          value={modelSearchVal}
          onChange={(e) => onModelSearch(e.target.value)}
          style={{ width: 380 }}
        />
      </div>
      <div className={styles.modelSelect}>
        <Select
          style={{ width: 160 }}
          value={modelSourceVal}
          options={options}
          onChange={onModelSourceChange}
        />
      </div>
      
    </div>
  );
};