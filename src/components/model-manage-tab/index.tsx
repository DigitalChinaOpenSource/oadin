'use client';

/** 模型管理 */
import React from 'react';
import styles from './index.module.scss';
import ModelSource from './model-source';
import ModelPath from './model-path';
import ModelTitleSearch from './model-title-search';
import ModelCardList from '../model-card-list';
import { useViewModel } from "./view-model";

export default function ModelManageTab() {
  const vm = useViewModel();

  return (
    <div className={styles.modelManageTab}>
      <ModelSource modelSourceVal={vm.modelSourceVal} onModelSourceChange={vm.onModelSourceChange} />
      <ModelPath modelPathVal={vm.modelPathVal} onModelPathChange={vm.onModelPathChange} />
      <ModelTitleSearch modelNums={vm.modelNums} modelSearchVal={vm.modelSearchVal} onModelTitleSearch={vm.onModelTitleSearch} />
      <ModelCardList />
    </div>
  );
};