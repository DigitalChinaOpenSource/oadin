import { useState } from 'react';
import { httpRequest } from '@/utils/httpRequest';
import { useRequest } from 'ahooks';
import thinkSvg from '@/components/icons/think.svg';
import nothinkSvg from '@/components/icons/no-think.svg';
import TagsRender from '@/components/tags-render';
import useSelectedModelStore from '@/store/useSelectedModel';
import { getSessionIdFromUrl } from '@/utils/sessionParamUtils';
import styles from './index.module.scss';

export default function ChatModelManage() {
  const { selectedModel } = useSelectedModelStore();
  const currentSessionId = getSessionIdFromUrl();

  const [isThinking, setIsThinking] = useState<boolean>(true);

  const { run: fetchThinkingEnable } = useRequest(
    async (params: { sessionId: string; enabled: boolean }) => {
      const data = await httpRequest.post('/playground/session/thinking', params);
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data: any) => {
        setIsThinking(data.thinking_active);
      },
      onError: (error: any) => {
        console.log('思考状态更新失败:', error);
      },
    },
  );

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
                  if (!selectedModel.think_switch) return;
                  fetchThinkingEnable({
                    sessionId: currentSessionId,
                    enabled: !isThinking,
                  });
                }}
              >
                <img
                  src={isThinking ? thinkSvg : nothinkSvg}
                  alt="思考图标"
                />
                <span>深度思考</span>
              </div>
            </>
          )}
        </div>
      </div>
    </>
  );
}
