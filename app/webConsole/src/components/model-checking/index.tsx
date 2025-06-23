import { Button, message, Skeleton, Tooltip } from 'antd';
import styles from './index.module.scss';
import { ModelCheckingNodata } from './model-checking-nodata';
import moreModel from '@/components/icons/moreModel.png';
import { ChooseModelDialog } from '@/components/choose-model-dialog';
import { useEffect, useMemo, useState } from 'react';
import { useViewModel, IMyModelListViewModel } from './view-model.ts';
import { ModelCheckingHasdata } from '@/components/model-checking/model-checking-data';
import useSelectedModelStore, { selectedModelType } from '@/store/useSelectedModel.ts';
import { getMessageByModel } from '@/i18n';
import useChatStore from '@/components/chat-container/store/useChatStore';

export default function ModelChecking() {
  const { setIsSelectedModel, setSelectedModel } = useSelectedModelStore();
  const { currentSessionId } = useChatStore();

  const [selectedStateModel, setSelecteStatedModel] = useState<selectedModelType>(null);
  const vm: IMyModelListViewModel = useViewModel();
  const { modelSupportLoading, fetchModelSupport, modelListData } = vm;
  const [open, setOpen] = useState<boolean>(false);
  /// 初始化选择模型
  const onOk = () => {
    if (selectedStateModel && Object.keys(selectedStateModel).length > 0) {
      setIsSelectedModel(true);
      setSelectedModel(selectedStateModel);

      const tempParams = { service_name: selectedStateModel.service_name } as any;
      if (selectedStateModel.source === 'local') {
        tempParams.local_provider = selectedStateModel.service_provider_name;
      } else if (selectedStateModel.source === 'remote') {
        tempParams.remote_provider = selectedStateModel.service_provider_name;
      }
      vm.fetchChooseModelNotify(tempParams);
    } else {
      message.warning(
        getMessageByModel('noSelectModel', {
          msg: '请先选择模型，再体验。',
        }),
      );
    }
  };
  const isHasMedel = useMemo(() => {
    return modelListData.length > 0;
  }, [modelListData]);
  useEffect(() => {
    fetchModelSupport({
      service_source: 'local',
    });
  }, []);
  return (
    <Skeleton loading={modelSupportLoading}>
      <div>
        <div className={styles.warp}>
          <div className={styles.title}>
            <h2>
              <span>请选择模型</span> <span className={styles.title_weight}>立即体验</span>
            </h2>
            <p>体验不同模型和MCP工具的能力组合</p>
          </div>
          {isHasMedel ? (
            <ModelCheckingHasdata
              vm={vm}
              selectedStateModel={selectedStateModel}
              setSelecteStatedModel={setSelecteStatedModel}
            />
          ) : (
            <ModelCheckingNodata />
          )}
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
    </Skeleton>
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
