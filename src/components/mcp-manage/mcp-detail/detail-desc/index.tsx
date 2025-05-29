import styles from './index.module.scss';
import TagsRender from '@/components/tags-render';
import { Tooltip } from 'antd';
import { McpDetailType } from '@/components/mcp-manage/mcp-detail/type.ts';
import { CloudIcon, LocalIcon } from '@/components/icons';
import dayjs from 'dayjs';
import { forwardRef, useEffect, useImperativeHandle, useRef, useState } from 'react';
import defaultPng from '@/assets/favicon.png';

export default forwardRef(function DetailDesc(props: { mcpDetail: McpDetailType }, ref) {
  let { logo, tags, name, abstract, updatedAt, supplier, hosted } = props.mcpDetail;
  const formatUnixTime = (unixTime: number) => {
    if (!unixTime) return unixTime;
    const date = dayjs.unix(unixTime);
    return date.format('YYYY-MM-DD');
  };
  const contentRef = useRef<HTMLDivElement>(null);
  const [showTooltip, setShowTooltip] = useState(false);

  useEffect(() => {
    if (contentRef.current) {
      const isOverflowing = contentRef.current.scrollHeight > contentRef.current.offsetHeight;
      console.log('isOverflowing', isOverflowing);
      setShowTooltip(isOverflowing);
    }
  }, [abstract?.zh]);

  useImperativeHandle(ref, () => showTooltip);

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
                <LocalIcon />
                <div className={styles.localOrCloudText}>本地</div>
              </>
            ) : (
              <>
                <CloudIcon />
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

        <Tooltip title={showTooltip && abstract?.zh}>
          <div
            ref={contentRef}
            className={styles.detailDesc}
          >
            {abstract?.zh || '暂无描述'}
          </div>
        </Tooltip>

        <div className={styles.infoWrapper}>
          <div className={styles.providerName}>{supplier}</div>
          <div className={styles.dot}>{updatedAt && '·'}</div>
          <div className={styles.updateName}>{updatedAt && formatUnixTime(updatedAt as number) + '更新'} </div>
        </div>
      </div>
    </div>
  );
});
