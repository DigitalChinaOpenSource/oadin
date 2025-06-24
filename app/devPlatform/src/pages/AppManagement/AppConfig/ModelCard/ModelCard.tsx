import React from 'react';
import { IModelSelectCardItem } from '@/pages/AppManagement/AppConfig/types.ts';
import TagsRender from '@/components/tags-render';
import { Card } from 'antd';
import styles from './ModelCard.module.scss';
import { XIcon } from '@phosphor-icons/react';

// 选择模型之后的卡片样式
export interface ModelCardProps {
  model: IModelSelectCardItem;
  onRemove?: (model: IModelSelectCardItem) => void;
}

const ModelCard: React.FC<ModelCardProps> = ({ model, onRemove }) => {
  const _onRemove = () => {
    onRemove && onRemove(model);
  };
  return (
    <Card className={styles.modelCard}>
      <div className={styles.headerContainer}>
        <div className={styles.modelInfo}>
          <div className={styles.modelInfoImg}>
            <img
              src={model.avatar}
              alt=""
            />
          </div>
          <div className={styles.modelName}>{model.name}</div>
        </div>
        {onRemove && (
          <div
            onClick={_onRemove}
            className={styles.removeWarp}
          >
            <XIcon
              width={16}
              height={16}
            />
          </div>
        )}
      </div>
      <TagsRender tags={model.class} />
    </Card>
  );
};

export default ModelCard;
