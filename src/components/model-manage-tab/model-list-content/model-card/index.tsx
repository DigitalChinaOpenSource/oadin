import { useRef, useEffect, useState } from 'react';
import styles from './index.module.scss';
import { Button, Tag } from 'antd';
import { IModelAuthType, ModelDataItem } from '../../types';
import { LoadingIcon, DownloadIcon, LocalIcon, CloudIcon, DeleteIcon } from '../../../icons';

export interface IModelCardProps {
  // 是否用于详情展示
  isDetail?: boolean;
  // 模型数据
  modelData?: ModelDataItem;
  onDetailModalVisible?: (visible: boolean) => void;
  onModelAuthVisible?: (visible: boolean, type: IModelAuthType) => void;
  deleteConfirm?: (modelData: ModelDataItem) => void;
  downloadConfirm?: (modelData: ModelDataItem) => void;
}

export default function ModelCard(props: IModelCardProps) {
  const { isDetail, onDetailModalVisible, onModelAuthVisible, deleteConfirm, downloadConfirm, modelData } = props;

  const [isOverflow, setIsOverflow] = useState(false);
  const textRef = useRef<HTMLDivElement>(null);

  const tags = ['深度思考', '文本模型', '语言理解', '逻辑思维', '情感分析', '信息检索'];

  useEffect(() => {
    const checkOverflow = () => {
      const element = textRef.current;
      if (element) {
        // 检查内容是否超过两行
        const lineHeight = parseInt(window.getComputedStyle(element).lineHeight);
        const maxHeight = lineHeight * 2;
        setIsOverflow(element.scrollHeight > maxHeight);
      }
    };
    checkOverflow();
    window.addEventListener('resize', checkOverflow);
    return () => window.removeEventListener('resize', checkOverflow);
  }, [modelData]);

  return (
    <>
      <div className={`${isDetail ? '' : styles.modelCardStyle} ${styles.modelCard}`}>
        <div className={styles.cardHeader}>
          <div className={styles.cardTitle}>
            <img
              src={modelData?.avatar}
              width={24}
            />
            <div className={styles.title}>{modelData?.name}</div>

            {/* 本地还是云端 */}
            <div className={styles.localOrCloud}>
              {modelData?.source === 'local' ? (
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

            {/* 推荐使用 */}
            {modelData?.is_recommended && <div className={styles.recommend}>🔥 推荐使用</div>}

            {/* 已授权 */}
            {/* <div className={styles.authorized}>已授权</div> */}
          </div>

          {!isDetail && (
            <div className={styles.cardDownload}>
              {modelData?.source === 'local' ? (
                <>
                  {/* 本地下载相关内容 */}
                  {modelData?.can_select ? (
                    <Button
                      type="text"
                      className={styles.hasLoaded}
                      onClick={() => deleteConfirm?.(modelData!)}
                    >
                      已下载
                      <DeleteIcon />
                    </Button>
                  ) : (
                    // 下载失败也回到下载
                    <Button
                      type="primary"
                      size="small"
                      onClick={() => downloadConfirm?.(modelData!)}
                    >
                      下载
                      <DownloadIcon />
                    </Button>
                  )}
                  {/* <Button type="text" style={{ color: '#344054', padding: 'unset' }} >
                    下载中
                    <LoadingIcon />
                  </Button> */}
                </>
              ) : (
                <>
                  {/* 云端下载相关内容 */}
                  <Button
                    type="primary"
                    size="small"
                    onClick={() => onModelAuthVisible?.(true, 'config')}
                  >
                    配置授权
                  </Button>
                  {/* <Button className={styles.updateSetting} variant="filled" size="small" onClick={() => onModelAuthorizeVisible?.('update')}>
                    更新配置
                  </Button> */}
                </>
              )}
            </div>
          )}
        </div>

        <div className={styles.contentWrapper}>
          <div
            ref={textRef}
            className={`${isDetail ? '' : styles.textContent}`}
          >
            {modelData?.desc}
          </div>
          {isOverflow && !isDetail && (
            <Button
              type="link"
              size="small"
              className={styles.moreButton}
              onClick={() => onDetailModalVisible?.(true)}
            >
              更多
            </Button>
          )}
        </div>

        <div className={styles.infoWrapper}>
          <div className={styles.providerName}>深度求索</div>
          <div className={styles.contextLength}>
            <span className={styles.splitLine}>｜</span>
            <span>上下文长度：</span>
            <span>测试暂无</span>
          </div>
          <div className={styles.modelSize}>
            <span className={styles.splitLine}>｜</span>
            <span>大小：</span>
            <span>{modelData?.size}</span>
          </div>
          <div className={styles.beUsed}>
            <span className={styles.splitLine2}>｜</span>
            <span>使用中：</span>
            <span>测试2个</span>
          </div>
        </div>

        <div className={styles.tagWrapper}>
          {tags.map((tag, index) => (
            <Tag key={index}>{tag}</Tag>
          ))}
        </div>
      </div>
    </>
  );
}
