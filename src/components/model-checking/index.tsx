import { Button, message, Tooltip } from 'antd';
import styles from './index.module.scss';
import { ModelCheckingNodata } from './model-checking-nodata';
import moreModel from '@/components/icons/moreModel.png';
import { ChooseModelDialog } from '@/components/choose-model-dialog';
import { useState } from 'react';
import { IModelListContent, IUseViewModel, useViewModel } from '@/components/model-manage-tab/model-list-content/view-model.ts';
import { ModelCheckingHasdata } from '@/components/model-checking/model-checking-data';
import useSelectedModelStore from '@/store/useSelectedModel.ts';

export default function ModelChecking() {
  const vmProps: IModelListContent = {
    modelSearchVal: '',
    modelSourceVal: 'local',
    onModelSearch: () => {},
    mine: true,
  };
  const { selectedModel, setIsSelectedModel } = useSelectedModelStore();
  const vm: IUseViewModel = useViewModel(vmProps);
  const isHasMedel = vm.pagination.total > 0;
  const [open, setOpen] = useState<boolean>(false);
  const onOk = () => {
    if (selectedModel && Object.keys(selectedModel).length > 0) {
      setIsSelectedModel(true);
    } else {
      message.warning('请先选择一个模型');
    }
  };
  return (
    <div>
      <div className={styles.warp}>
        <div className={styles.title}>
          <h2>
            <span>请选择模型</span> <span className={styles.title_weight}>立即体验</span>
          </h2>
          <p>体验不同模型和MCP工具的能力组合</p>
        </div>
        {isHasMedel ? <ModelCheckingHasdata vm={vm} /> : <ModelCheckingNodata />}
        <div
          className={styles.more_model_warp}
          onClick={() => {
            setOpen(true);
          }}
        >
          <img
            src={moreModel}
            alt=""
          />
          更多模型
        </div>
        <div className={styles.button}>
          {isHasMedel ? (
            renderButton({
              onOk,
            })
          ) : (
            <Tooltip
              title="下载模型后即可开始体验"
              color="#fff"
            >
              {renderButton({
                onOk,
              })}
            </Tooltip>
          )}
        </div>
      </div>
      <ChooseModelDialog
        open={open}
        onCancel={() => setOpen(false)}
      />
    </div>
  );
}
const renderButton = ({ onOk }: { onOk: () => void }) => {
  return (
    <Button
      type="primary"
      onClick={() => {
        if (onOk) {
          onOk();
        }
      }}
    >
      立即体验
    </Button>
  );
};
