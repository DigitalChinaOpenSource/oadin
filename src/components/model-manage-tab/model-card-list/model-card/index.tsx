import { useRef, useEffect, useState } from 'react';
import styles from './index.module.scss';
import { Button, Tag } from 'antd';
import { useViewModel } from './view-model.ts';
import ModelDetailModal from '../model-detail-modal/index.tsx';
import modelLogo from '@/assets/modelLogo.png';

export interface IModelCardProps {
  // 是否用于详情展示
  isDetail?: boolean;
  // 本地还是云端
  isLocal?: boolean;
  content?: string;
  tags?: string[];
}

export default function ModelCard(props: IModelCardProps) {
  const { isDetail, isLocal } = props;
  const [isOverflow, setIsOverflow] = useState(false);
  const textRef = useRef<HTMLDivElement>(null);
  const vm = useViewModel(props);
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
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="#898ea3" viewBox="0 0 256 256">
                    <path d="M128,24h0A104,104,0,1,0,232,128,104.12,104.12,0,0,0,128,24Zm88,104a87.61,87.61,0,0,1-3.33,24H174.16a157.44,157.44,0,0,0,0-48h38.51A87.61,87.61,0,0,1,216,128ZM102,168H154a115.11,115.11,0,0,1-26,45A115.27,115.27,0,0,1,102,168Zm-3.9-16a140.84,140.84,0,0,1,0-48h59.88a140.84,140.84,0,0,1,0,48ZM40,128a87.61,87.61,0,0,1,3.33-24H81.84a157.44,157.44,0,0,0,0,48H43.33A87.61,87.61,0,0,1,40,128ZM154,88H102a115.11,115.11,0,0,1,26-45A115.27,115.27,0,0,1,154,88Zm52.33,0H170.71a135.28,135.28,0,0,0-22.3-45.6A88.29,88.29,0,0,1,206.37,88ZM107.59,42.4A135.28,135.28,0,0,0,85.29,88H49.63A88.29,88.29,0,0,1,107.59,42.4ZM49.63,168H85.29a135.28,135.28,0,0,0,22.3,45.6A88.29,88.29,0,0,1,49.63,168Zm98.78,45.6a135.28,135.28,0,0,0,22.3-45.6h35.66A88.29,88.29,0,0,1,148.41,213.6Z"></path>
                  </svg>
                  <div className={styles.localOrCloudText}>本地</div>
                </>
              ) : (
                <>
                  <svg width="16" height="16" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path
                      d="M9.75 6.375H2.25C2.05109 6.375 1.86032 6.45402 1.71967 6.59467C1.57902 6.73532 1.5 6.92609 1.5 7.125V9.375C1.5 9.57391 1.57902 9.76468 1.71967 9.90533C1.86032 10.046 2.05109 10.125 2.25 10.125H9.75C9.94891 10.125 10.1397 10.046 10.2803 9.90533C10.421 9.76468 10.5 9.57391 10.5 9.375V7.125C10.5 6.92609 10.421 6.73532 10.2803 6.59467C10.1397 6.45402 9.94891 6.375 9.75 6.375ZM9.75 9.375H2.25V7.125H9.75V9.375ZM9.75 1.875H2.25C2.05109 1.875 1.86032 1.95402 1.71967 2.09467C1.57902 2.23532 1.5 2.42609 1.5 2.625V4.875C1.5 5.07391 1.57902 5.26468 1.71967 5.40533C1.86032 5.54598 2.05109 5.625 2.25 5.625H9.75C9.94891 5.625 10.1397 5.54598 10.2803 5.40533C10.421 5.26468 10.5 5.07391 10.5 4.875V2.625C10.5 2.42609 10.421 2.23532 10.2803 2.09467C10.1397 1.95402 9.94891 1.875 9.75 1.875ZM9.75 4.875H2.25V2.625H9.75V4.875ZM9 3.75C9 3.86125 8.96701 3.97001 8.9052 4.06251C8.84339 4.15501 8.75554 4.22711 8.65276 4.26968C8.54998 4.31226 8.43688 4.3234 8.32776 4.30169C8.21865 4.27999 8.11842 4.22641 8.03975 4.14775C7.96109 4.06908 7.90751 3.96885 7.88581 3.85974C7.8641 3.75062 7.87524 3.63752 7.91782 3.53474C7.96039 3.43196 8.03249 3.34411 8.12499 3.2823C8.21749 3.22049 8.32625 3.1875 8.4375 3.1875C8.58668 3.1875 8.72976 3.24676 8.83525 3.35225C8.94074 3.45774 9 3.60082 9 3.75ZM9 8.25C9 8.36125 8.96701 8.47001 8.9052 8.56251C8.84339 8.65501 8.75554 8.72711 8.65276 8.76968C8.54998 8.81226 8.43688 8.8234 8.32776 8.80169C8.21865 8.77999 8.11842 8.72641 8.03975 8.64775C7.96109 8.56908 7.90751 8.46885 7.88581 8.35974C7.8641 8.25062 7.87524 8.13752 7.91782 8.03474C7.96039 7.93196 8.03249 7.84411 8.12499 7.7823C8.21749 7.72049 8.32625 7.6875 8.4375 7.6875C8.58668 7.6875 8.72976 7.74676 8.83525 7.85225C8.94074 7.95774 9 8.10082 9 8.25Z"
                      fill="#898ea3"
                    />
                  </svg>
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
                  {/* <div>
                          <span>已下载</span>
                          <DeleteOutlined />
                        </div> */}
                  {/* <Button type="text" style={{ color: '#344054', padding: 'unset' }}>
                        已下载
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="#344054" viewBox="0 0 256 256">
                          <path d="M216,48H176V40a24,24,0,0,0-24-24H104A24,24,0,0,0,80,40v8H40a8,8,0,0,0,0,16h8V208a16,16,0,0,0,16,16H192a16,16,0,0,0,16-16V64h8a8,8,0,0,0,0-16ZM96,40a8,8,0,0,1,8-8h48a8,8,0,0,1,8,8v8H96Zm96,168H64V64H192ZM112,104v64a8,8,0,0,1-16,0V104a8,8,0,0,1,16,0Zm48,0v64a8,8,0,0,1-16,0V104a8,8,0,0,1,16,0Z"></path>
                        </svg>
                      </Button>
                      <Button type="text" style={{ color: '#344054', padding: 'unset' }}>
                        下载中
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="#344054" viewBox="0 0 256 256">
                          <path d="M136,32V64a8,8,0,0,1-16,0V32a8,8,0,0,1,16,0Zm37.25,58.75a8,8,0,0,0,5.66-2.35l22.63-22.62a8,8,0,0,0-11.32-11.32L167.6,77.09a8,8,0,0,0,5.65,13.66ZM224,120H192a8,8,0,0,0,0,16h32a8,8,0,0,0,0-16Zm-45.09,47.6a8,8,0,0,0-11.31,11.31l22.62,22.63a8,8,0,0,0,11.32-11.32ZM128,184a8,8,0,0,0-8,8v32a8,8,0,0,0,16,0V192A8,8,0,0,0,128,184ZM77.09,167.6,54.46,190.22a8,8,0,0,0,11.32,11.32L88.4,178.91A8,8,0,0,0,77.09,167.6ZM72,128a8,8,0,0,0-8-8H32a8,8,0,0,0,0,16H64A8,8,0,0,0,72,128ZM65.78,54.46A8,8,0,0,0,54.46,65.78L77.09,88.4A8,8,0,0,0,88.4,77.09Z"></path>
                        </svg>
                      </Button> */}
                  {/* 下载失败也回到下载 */}
                  <Button type="primary" size="small">
                    下载
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="#fff" viewBox="0 0 256 256">
                      <path d="M224,144v64a8,8,0,0,1-8,8H40a8,8,0,0,1-8-8V144a8,8,0,0,1,16,0v56H208V144a8,8,0,0,1,16,0Zm-101.66,5.66a8,8,0,0,0,11.32,0l40-40a8,8,0,0,0-11.32-11.32L136,124.69V32a8,8,0,0,0-16,0v92.69L93.66,98.34a8,8,0,0,0-11.32,11.32Z"></path>
                    </svg>
                  </Button>
                </>
              ) : (
                <>{/* 云端下载相关内容 */}</>
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
            <Button type="link" size="small" className={styles.moreButton} onClick={vm.handleDetailVisible}>
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
      {vm.isDetailVisible && <ModelDetailModal onDetailClose={vm.onDetailClose} />}
    </>
  );
}
