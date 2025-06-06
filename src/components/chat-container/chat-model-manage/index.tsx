import React from 'react';
import sample from '@/components/icons/sample.svg';
import thinkSvg from '@/components/icons/think.svg';
import exchangeSvg from '@/components/icons/exchange.svg';
import { Tooltip } from 'antd';
import styles from './index.module.scss';

interface IChatModelManageProps {
  currModelData?: any;
  onChangeModel?: (modelData: any) => void;
}

export default function ChatModelManage(props: IChatModelManageProps) {
  const modelTags = ['支持 MCP', '深度思考'];
  return (
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
        <div className={styles.think}>
          <img
            src={thinkSvg}
            alt="思考图标"
          />
          <span>深度思考</span>
        </div>
        <div className={styles.fill}></div>
        {/* TODO */}
        <Tooltip title="切换模型后，将开启新会话">
          <div
            className={styles.changeModel}
            onClick={() => console.log('切换模型')}
          >
            <img
              src={exchangeSvg}
              alt="切换模型图标"
            />
          </div>
        </Tooltip>
      </div>
    </div>
  );
}
