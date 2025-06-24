import React from 'react';
import { Typography } from 'antd';
import styles from './TagFilter.module.scss';

export interface Tag {
  label: string;
  count?: number;
}

interface ModelTagFilterProps {
  tags: Tag[];
  selectedTag: string;
  onTagSelect: (tagLabel: string) => void;
}

const TagFilter: React.FC<ModelTagFilterProps> = ({ tags, selectedTag, onTagSelect }) => {
  // 处理标签点击
  const handleTagClick = (tagLabel: string) => {
    onTagSelect(tagLabel);
  };

  return (
    <div className={styles.container}>
      <div className={styles.filterRow}>
        <div className={styles.filterLabel}>{/* 可以添加标签筛选的图标和标题 */}</div>
      </div>
      <div className={styles.tagsRow}>
        {tags.map((tag, index) => (
          <div
            key={index}
            className={`${styles.tag} ${selectedTag === tag.label ? styles.tagSelected : ''}`}
            onClick={() => handleTagClick(tag.label)}
          >
            <Typography.Text>
              {tag.label} {tag.count !== undefined && `${tag.count}`}
            </Typography.Text>
          </div>
        ))}
      </div>
    </div>
  );
};

export default TagFilter;
