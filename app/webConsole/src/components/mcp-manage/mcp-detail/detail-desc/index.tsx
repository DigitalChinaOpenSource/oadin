import styles from './index.module.scss';
import TagsRender from '@/components/tags-render';
import { Tooltip } from 'antd';
import { McpDetailType } from '@/components/mcp-manage/mcp-detail/type.ts';
import { GlobeIcon, HardDrivesIcon } from '@phosphor-icons/react';
import dayjs from 'dayjs';
import { forwardRef, useEffect, useImperativeHandle, useRef, useState } from 'react';
import defaultPng from '@/assets/favicon.png';
import EllipsisTooltip from '@/components/ellipsis-tooltip';

export default function DetailDesc(props: { mcpDetail: McpDetailType }) {
  const { logo, tags, name, abstract, updatedAt, supplier, hosted } = props.mcpDetail;
  const formatUnixTime = (unixTime: number) => {
    if (!unixTime) return unixTime;
    const date = dayjs.unix(unixTime);
    return date.format('YYYY-MM-DD');
  };

  return (
    <div className={styles.detailDescMain}>
      <div className={styles.detailIcon}>
        <img
          src={logo || defaultPng}
          alt=""
        />
      </div>
      <div className={styles.detailContent}>
        <div className={styles.detailTitle}>
          <div className={styles.detailTitleName}>{name.zh}</div>
          <div className={styles.localOrCloud}>
            {hosted ? (
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
          <div className={styles.tags}>
            <TagsRender
              tags={tags || []}
              // className={'descTag'}
            />
          </div>
        </div>

        <EllipsisTooltip
          title={abstract?.zh}
          className={styles.detailDesc}
          maxWidth={'60%'}
        >
          {abstract?.zh || '暂无描述'}
        </EllipsisTooltip>

        <div className={styles.infoWrapper}>
          <div className={styles.providerName}>{supplier}</div>
          <div className={styles.dot}>{updatedAt && '·'}</div>
          <div className={styles.updateName}>{updatedAt && formatUnixTime(updatedAt as number) + '更新'} </div>
        </div>
      </div>
    </div>
  );
}
