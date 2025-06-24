import { useState, useEffect } from 'react';
import { Tooltip } from 'antd';
import thinkSvg from '@/components/icons/think.svg';
import nothinkSvg from '@/components/icons/no-think.svg';
import exchangeSvg from '@/components/icons/exchange.svg';
import TagsRender from '@/components/tags-render';
import { ChooseModelDialog } from '@/components/choose-model-dialog';
import useSelectedModelStore from '@/store/useSelectedModel';
import useChatStore from '../store/useChatStore';

import styles from './index.module.scss';

interface IChatModelManageProps {
  currModelData?: any;
  onChooseModel?: (modelData: any) => void;
}

export default function ChatModelManage(props: IChatModelManageProps) {
  const { selectedModel } = useSelectedModelStore();
  const { isLoading } = useChatStore();
  const [open, setOpen] = useState<boolean>(false);
  const [isThinking, setIsThinking] = useState<boolean>(true);
  return (
    <>
      <div className={styles.chatModelManage}>
        <div className={styles.left}>
          <div className={styles.modelIcon}>
            <img
              src={selectedModel?.avatar}
              alt="模型图标"
            />
          </div>
          <div className={styles.modelName}>
            <div className={styles.name}>{selectedModel?.name}</div>
            <TagsRender
              tags={selectedModel?.class || []}
              highlightNums={selectedModel?.class?.length || 0}
            />
          </div>
        </div>

        <div className={styles.right}>
          {selectedModel?.think && (
            <>
              <div
                className={isThinking ? styles.think : styles.noThink}
                onClick={() => {
                  setIsThinking(!isThinking);
                  // TODO 调接口去关闭
                }}
              >
                <img
                  src={isThinking ? thinkSvg : nothinkSvg}
                  alt="思考图标"
                />
                <span>深度思考</span>
              </div>
              <div className={styles.fill}></div>
            </>
          )}

          <Tooltip title={isLoading ? '对话中不可修改模型' : '切换模型后，将开启新会话'}>
            <div
              className={styles.changeModel}
              onClick={() => {
                if (isLoading) return;
                setOpen(true);
              }}
            >
              <img
                src={exchangeSvg}
                alt="切换模型图标"
              />
            </div>
          </Tooltip>
        </div>
      </div>
      {open && (
        <ChooseModelDialog
          open={true}
          onCancel={() => setOpen(false)}
        />
      )}
    </>
  );
}
