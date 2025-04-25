'use client';

import React from 'react';
import { Radio } from 'antd';
import styles from './index.module.scss';

interface IModelSourceProps {
  onModelSourceChange: (val: string) => void;
  modelSourceVal: string;
}

export default function ModelSource(props: IModelSourceProps) {
	const { onModelSourceChange, modelSourceVal } = props;
  return (
    <div className={styles.modelSource}>
      <div className={styles.modelSourceTitle}>模型来源</div>
      <Radio.Group value={modelSourceVal || 'local'} onChange={(e) => onModelSourceChange(e.target.value)}>
        <Radio value="local">查看本地模型</Radio>
        <Radio value="remote">查看云端模型</Radio>
      </Radio.Group>
    </div>
  );
};