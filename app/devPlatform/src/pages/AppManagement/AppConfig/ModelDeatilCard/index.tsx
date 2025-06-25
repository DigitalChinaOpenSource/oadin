import React, { Dispatch, SetStateAction } from 'react';
import styles from './index.module.scss';
import { Button, Checkbox } from 'antd';
import { GlobeIcon, HardDrivesIcon } from '@phosphor-icons/react';
import TagsRender from '@/components/tags-render';
import dayjs from 'dayjs';
import defaultPng from '@/assets/favicon.png';
import { ExclamationCircleOutlined } from '@ant-design/icons';
import { checkMcpLength } from '@/pages/AppManagement/AppConfig/uitls.ts';
import { IModelDataItem } from '@/types/model.ts';
import EllipsisTooltip from '@/pages/AppManagement/AppConfig/ModelDeatilCard/ellipsisTooltip.tsx';

export interface IMcpCardProps {
  // 模型数据
  deatilData: IModelDataItem;
  handelMcpCardClick: (mcpId: string | number) => void;
  selectedModelIds?: string[];
  setSelectedModelIds?: Dispatch<SetStateAction<string[]>>;
}

const formatUnixTime = (unixTime: number) => {
  const date = dayjs.unix(unixTime);
  return date.format('YYYY-MM-DD');
};
export default function McpCard(props: IMcpCardProps) {
  const { deatilData, handelMcpCardClick, setSelectedModelIds, selectedModelIds = [] } = props;
  // 处理单个项目的选择
  const handleItemSelect = (item: IModelDataItem, checked: boolean) => {
    if (checked) {
      if (checkMcpLength(selectedModelIds.length)) {
        setSelectedModelIds?.([...(selectedModelIds ?? []), item.id]);
      }
    } else {
      setSelectedModelIds?.(selectedModelIds.filter((id) => id !== item?.id));
    }
  };

  return (
    <div className={styles.mcpCard}>
      <div className={styles.cardHeader}>
        <div className={styles.cardTitle}>
          {/* logo */}
          <div className={styles.avatar}>
            <img
              alt=""
              src={deatilData?.avatar || defaultPng}
              width={24}
            />
          </div>
          {/* 名称 */}
          <EllipsisTooltip
            className={styles.title}
            title={deatilData?.name}
          >
            {deatilData?.name}
          </EllipsisTooltip>
          {/* 本地还是云端 */}
          <div className={styles.localOrCloud}>
            {deatilData?.source === 'local' ? (
              <>
                <HardDrivesIcon
                  width={16}
                  height={16}
                  fill="#898ea3"
                />
                <div className={styles.localOrCloudText}>本地</div>
              </>
            ) : (
              <>
                <GlobeIcon
                  width={16}
                  height={16}
                  fill="#898ea3"
                />
                <div className={styles.localOrCloudText}>云端</div>
              </>
            )}
          </div>
        </div>

        <Checkbox
          checked={selectedModelIds.includes(deatilData.id)}
          onChange={(e) => handleItemSelect(deatilData, e.target.checked)}
        />
      </div>
      <div style={{ height: '24px' }}>
        <TagsRender tags={deatilData?.class || []} />
      </div>
      {/* 修改：使用 EllipsisTooltip 组件 */}
      <EllipsisTooltip
        title={deatilData?.desc}
        className={styles.contentWrapper}
        maxWidth={400}
      >
        {deatilData?.desc}
      </EllipsisTooltip>
      <div className={styles.infoWrapper}>
        <div className={styles.providerName}>{deatilData?.flavor}</div>
        <div className={styles.dot}>·</div>
        <div className={styles.updateName}>{deatilData?.update_time && formatUnixTime(deatilData?.update_time) + '更新'}</div>
      </div>
      <div className={styles.cardOperate}>
        <Button
          type={'text'}
          icon={<ExclamationCircleOutlined />}
          onClick={() => handelMcpCardClick(deatilData?.id)}
        >
          查看详情
        </Button>
      </div>
    </div>
  );
}
