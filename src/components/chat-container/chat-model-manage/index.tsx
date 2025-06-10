import { useState } from 'react';
import sample from '@/components/icons/sample.svg';
import thinkSvg from '@/components/icons/think.svg';
import exchangeSvg from '@/components/icons/exchange.svg';
import { Tooltip } from 'antd';
import { ChooseModelDialog } from '@/components/choose-model-dialog';
import styles from './index.module.scss';

interface IChatModelManageProps {
  currModelData?: any;
  onChooseModel?: (modelData: any) => void;
}

export default function ChatModelManage(props: IChatModelManageProps) {
  const modelTags = ['支持 MCP', '深度思考'];
  const [open, setOpen] = useState<boolean>(false);

  return (
    <>
      <div className={styles.chatModelManage}>
        <div className={styles.left}>
          <div className={styles.modelIcon}>
            <img
              src={sample}
              alt="模型图标"
            />
          </div>
          <div className={styles.modelName}>
            <div className={styles.name}>模型名称</div>
            {modelTags.map((tag, index) => (
              <span
                key={index}
                className={styles.modelTag}
              >
                {tag}
              </span>
            ))}
          </div>
        </div>

        <div className={styles.right}>
          {/* TODO */}
          <div className={styles.think}>
            <img
              src={thinkSvg}
              alt="思考图标"
            />
            <span>深度思考</span>
          </div>
          <div className={styles.fill}></div>
          <Tooltip title="切换模型后，将开启新会话">
            <div
              className={styles.changeModel}
              onClick={() => setOpen(true)}
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
