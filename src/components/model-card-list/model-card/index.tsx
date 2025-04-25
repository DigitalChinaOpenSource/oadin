import React, { useRef, useEffect, useState } from 'react';
import styles from './index.module.scss';
import { OpenAIOutlined, DeleteOutlined } from '@ant-design/icons';
import { Card, Button, Tag } from 'antd';
import { useViewModel } from './view-model';
import ModelDetailModal from '../model-detail-modal.tsx';
export interface IModelCardProps {
  // 是否用于详情展示
  isDetail: boolean;
  content?: string;
  tags?: string[];
}

export default function ModelCard(props: IModelCardProps) {
  const { isDetail } = props;
  const [isOverflow, setIsOverflow] = useState(false);
  const textRef = useRef<HTMLDivElement>(null);
  const vm = useViewModel(props)
  const content = "这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容";
  const tags = [
    '深度思考',
    '文本模型',
    '语言理解',
    '逻辑思维',
    '情感分析',
    '信息检索',
  ]
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
      <Card className={styles.modelCard}>
        {/* title */}
        <div className={styles.cardHeader}>
          <div className={styles.cardTitle}>
            <OpenAIOutlined />
            {/* <img src={icon} width={24} /> */}
            <span>模型名称待替换</span>
          </div>
          {
            !isDetail && (
              <div className={styles.cardDownload}>
                {/* 这里有多种状态，包括 下载、下载中，下载失败、已下载 */}
                <div>
                  <span>已下载</span>
                  <DeleteOutlined />
                </div>
              </div>
            )
          }
        </div>

        {/* content */}
        <div className={`${isDetail ? '' : styles.contentWrapper}`}>
          <div 
            ref={textRef} 
            className={`${isDetail ? '' : styles.textContent}`}
          >
            {content}
          </div>
          {isOverflow && !isDetail && (
            <Button 
              type="link" 
              size="small" 
              className={styles.moreButton}
              onClick={vm.handleDetailVisible}
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
            <span>4096k</span>
          </div>
          <div className={styles.beUsed}>
            <span className={styles.splitLine}>｜</span>
            <span>使用中：</span>
            <span>2个应用</span>
          </div>
        </div>

        <div className={styles.tagWrapper}>
          {tags.map((tag, index) => (
            <Tag key={index}>{tag}</Tag>
          ))}
        </div>
      </Card>
      {
        vm.isDetailVisible && (
          <ModelDetailModal onDetailClose={vm.onDetailClose}/>
        )
      }
    </>
  );
};