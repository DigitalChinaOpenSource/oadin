import styles from './index.module.scss';
import TagsRender from '@/components/tags-render';
import { Tooltip } from 'antd';

export default function DetailDesc() {
  return (
    <div className={styles.detailDescMain}>
      <div className={styles.detailIcon}>
        <img
          src="http://120.232.136.73:31619/byzedev/model_avatar/qwen.png"
          alt="icon"
        />
      </div>
      <div className={styles.detailContent}>
        <div className={styles.detailTitle}>
          <div className={styles.detailTitleName}>MCP 服务</div>
          <div className={styles.tags}>
            <TagsRender
              tags={[
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
              ]}
            />
          </div>
        </div>

        <Tooltip title="测试">
          <div className={styles.detailDesc}>
            这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内
          </div>
        </Tooltip>

        <div className={styles.infoWrapper}>
          <div className={styles.providerName}>深度求索</div>
          <div className={styles.dot}>·</div>
          <div className={styles.updateName}>2025-05-19 更新</div>
        </div>
      </div>
    </div>
  );
}
