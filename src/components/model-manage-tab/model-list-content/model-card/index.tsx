import { useRef, useEffect, useState } from 'react';
import styles from './index.module.scss';
import { Button, Tag } from 'antd';
import modelLogo from '@/assets/modelLogo.png';
import { IModelAuthType } from '../../types';
import {LoadingIcon, DownloadIcon, LocalIcon, CloudIcon, DeleteIcon} from '../../../icons';

export interface IModelCardProps {
  // 是否用于详情展示
  isDetail?: boolean;
  // 本地还是云端
  isLocal?: boolean;
  content?: string;
  tags?: string[];
  onDetailModalVisible?: (visible: boolean) => void;
  onModelAuthVisible?: (visible: boolean, type: IModelAuthType) => void;
  deleteConfirm?: (modelData: any) => void;
  downloadConfirm?: (modelData: any) => void;
}

export default function ModelCard(props: IModelCardProps) {
  const { isDetail, isLocal = true, onDetailModalVisible, onModelAuthVisible, deleteConfirm, downloadConfirm } = props;
  const [isOverflow, setIsOverflow] = useState(false);
  const textRef = useRef<HTMLDivElement>(null);
  const content =
    '这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容';
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
  }, []);

  return (
    <>
      <div className={`${isDetail ? '' : styles.modelCardStyle} ${styles.modelCard}`}>
        {/* title */}
        <div className={styles.cardHeader}>
          <div className={styles.cardTitle}>
            <img src={modelLogo} width={24} />
            <div className={styles.title}>模型名称待替换</div>

            {/* 本地还是云端 */}
            <div className={styles.localOrCloud}>
              {isLocal ? (
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
            <div className={styles.recommend}>🔥 推荐使用</div>
            {/* 已授权 */}
            {/* <div className={styles.authorized}>已授权</div> */}
          </div>

          {!isDetail && (
            <div className={styles.cardDownload}>
              {isLocal ? (
                <>
                  {/* 本地下载相关内容 */}
                  <Button type="text" className={styles.hasLoaded} onClick={() => deleteConfirm?.({})}>
                    已下载
                    <DeleteIcon />
                  </Button>
                  {/* <Button type="text" style={{ color: '#344054', padding: 'unset' }} >
                    下载中
                    <LoadingIcon />
                  </Button> */}
                  {/* 下载失败也回到下载 */}
                  {/* <Button type="primary" size="small" onClick={() => downloadConfirm?.({})}>
                    下载
                    <DownloadIcon />
                  </Button> */}
                </>
              ) : (
                <>
                  {/* 云端下载相关内容 */}
                  <Button type="primary" size="small" onClick={() => onModelAuthVisible?.(true, 'config')}>
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

        {/* content */}
        <div className={styles.contentWrapper}>
          <div ref={textRef} className={`${isDetail ? '' : styles.textContent}`}>
            {content}
          </div>
          {isOverflow && !isDetail && (
            <Button type="link" size="small" className={styles.moreButton} onClick={() => onDetailModalVisible?.(true)}>
              更多
            </Button>
          )}
        </div>

        <div className={styles.infoWrapper}>
          <div className={styles.providerName}>深度求索</div>
          <div className={styles.contextLength}>
            <span className={styles.splitLine}>｜</span>
            <span>上下文长度：</span>
            <span>4096k</span>
          </div>
          <div className={styles.modelSize}>
            <span className={styles.splitLine}>｜</span>
            <span>大小：</span>
            <span>3.8GB</span>
          </div>
          <div className={styles.beUsed}>
            <span className={styles.splitLine2}>｜</span>
            <span>使用中：</span>
            <span>2个应用</span>
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
