import React from 'react';
import styles from './index.module.scss'
import { Input } from 'antd';

const { Search } = Input;

interface IModelTitleSearchProps {
  onModelTitleSearch: (val: string) => void;
  modelNums: number;
  modelSearchVal: string;
}

export default function ModelTitleSearch(props: IModelTitleSearchProps) {
  const { onModelTitleSearch, modelNums, modelSearchVal } = props;
  const onSearch = (val: string) => {
    onModelTitleSearch(val);
  };
  return (
    <div className={styles.modelTitleSearch}>
      <div className={styles.modelTitleSearch_left}>
        <span className={styles.title}>模型管理</span>
        <span className={styles.nums}>{modelNums || 99}</span>
      </div>

      <div className={styles.modelTitleSearch_right}>
        <div className={styles.search}>
          <Search placeholder='请输入模型名称' onSearch={onSearch} style={{ width: 380 }} />
        </div>
      </div>
    </div>
  );
};