import { Button, Tooltip, List, message, Radio } from 'antd';
import styles from './index.module.scss';
import GeneralCard from '@/components/model-manage-tab/model-list-content/general-card';
import ModelPathModal from '../modelpath-modal';
import ModelAuthorizeModal from '../model-authorize-modal';
import ModelDetailModal from '../model-detail-modal';
import { useViewModel } from './view-model';
import { SettingIcon, FailedIcon } from '../../icons';
import realLoadingSvg from '@/components/icons/real-loading.svg';
import noDataSvg from '@/components/icons/no-data.svg';
import { IModelSourceType } from '@/types';
import useModelDownloadStore from '@/store/useModelDownloadStore';
import useModelPathChangeStore from '@/store/useModelPathChangeStore';
import { ModelList } from '@/components/model-manage-tab/model-list-content/ModelList.tsx';
import { useViewModel as useViewTabModel } from '@/components/model-manage-tab/view-model.ts';
import { useEffect } from 'react';
export interface IModelListContent {
  modelSearchVal: string;
  modelSourceVal: IModelSourceType;
  onModelSearch: (val: string) => void;
  // onModelSourceChange: (val: IModelSourceType) => void;
}

export default function ModelListContent(props: IModelListContent) {
  const { onModelSearch, modelSearchVal, modelSourceVal, onModelSourceChange } = useViewTabModel();
  const vm = useViewModel({
    onModelSearch,
    modelSearchVal,
    modelSourceVal,
  });
  console.info(vm, 'content的值');
  return (
    <>
      {vm.modelSupportLoading ? (
        <div className={styles.loading}>
          <img
            src={realLoadingSvg}
            alt="loading"
          />
        </div>
      ) : (
        <div className={styles.modelListContent}>
          <div className={styles.contentContainer}>
            <ModelList
              {...props}
              grid={{ gutter: 16, column: 3, xs: 1, sm: 1, md: 2, lg: 2, xl: 3, xxl: 3 }}
            />
          </div>
          {/* 模型路径弹窗 */}
          {vm.modalPathVisible && (
            <ModelPathModal
              modalPath={vm.modelPath}
              onModelPathVisible={vm.onModelPathVisible}
              onModalPathChangeSuccess={vm.onModalPathChangeSuccess}
            />
          )}
          {/* 配置授权弹窗 */}
          {vm.modelAuthVisible && (
            <ModelAuthorizeModal
              modelDataItem={vm.selectModelData}
              modelAuthType={vm.modelAuthType}
              onModelAuthVisible={vm.onModelAuthVisible}
              onModelAuthSuccess={vm.onModelAuthSuccess}
            />
          )}
          {/* 模型详情弹窗 */}
          {vm.isDetailVisible && (
            <ModelDetailModal
              modelSourceVal={vm.modelSourceVal}
              onDetailModalVisible={vm.onDetailModalVisible}
              selectModelData={vm.selectModelData}
            />
          )}
        </div>
      )}
    </>
  );
}
