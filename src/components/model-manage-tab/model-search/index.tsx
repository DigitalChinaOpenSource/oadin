import { useEffect, useState } from 'react';
import { Input, Select } from 'antd';
import { MagnifyingGlassIcon } from '@phosphor-icons/react';
import { IModelSourceType } from '@/types';
import styles from './index.module.scss';
import { useViewModel } from '@/components/model-manage-tab/model-list-content/view-model.ts';
import { IModelListContent } from '@/components/model-manage-tab/model-list-content';

export interface IModelTitleSearchProps extends IModelListContent {
  onModelSourceChange: (val: IModelSourceType) => void;
}

export default function ModelSearch(props: IModelTitleSearchProps) {
  const { onModelSearch, modelSearchVal, modelSourceVal, onModelSourceChange } = props;

  const vmContent = useViewModel({
    onModelSearch,
    modelSearchVal,
    modelSourceVal,
  });
  const { pagination } = vmContent;
  const [searchVal, setSearchVal] = useState<string>('');

  const options = [
    { value: 'local', label: '本地模型' },
    { value: 'remote', label: '云端模型' },
  ];
  useEffect(() => {
    setSearchVal('');
  }, [modelSourceVal]);

  useEffect(() => {
    setSearchVal(modelSearchVal || '');
  }, [modelSearchVal]);

  return (
    <div className={styles.modelSearch}>
      <div className={styles.modelSelect}>
        <Select
          style={{ width: 160 }}
          value={modelSourceVal as IModelSourceType}
          options={options}
          onChange={onModelSourceChange}
        />
        <div className={styles.modelTotalWarp}>共{pagination?.total}条</div>
      </div>
      <div className={styles.searchInput}>
        <Input
          allowClear
          placeholder="搜索模型"
          suffix={
            <div
              className={styles.searchIcon}
              onClick={() => onModelSearch(searchVal)}
            >
              <MagnifyingGlassIcon
                width={16}
                height={16}
                fill="#808899"
              />
            </div>
          }
          value={searchVal}
          onChange={(e) => {
            setSearchVal(e.target.value.trim());
          }}
          onPressEnter={() => {
            onModelSearch(searchVal.replace(/'/g, ''));
          }}
          onClear={() => onModelSearch('')}
          style={{ width: 380 }}
        />
      </div>
    </div>
  );
}
