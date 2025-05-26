import styles from './index.module.scss';
import TagsRender from '@/components/tags-render';
import { Tooltip } from 'antd';
import { McpDetailType } from '@/components/mcp-manage/mcp-detail/type.ts';
import { CloudIcon, LocalIcon } from '@/components/icons';

export default function DetailDesc(props: { mcpDetail: McpDetailType }) {
  const { logo, tags, name, abstract, updatedAt, supplier, hosted } = props.mcpDetail;
  return (
    <div className={styles.detailDescMain}>
      <div className={styles.detailIcon}>
        <img
          src={logo}
          alt="logo"
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
            <TagsRender tags={tags} />
          </div>
        </div>

        <Tooltip title="测试">
          <div className={styles.detailDesc}>{abstract.zh}</div>
        </Tooltip>

        <div className={styles.infoWrapper}>
          <div className={styles.providerName}>{supplier}</div>
          <div className={styles.dot}>·</div>
          <div className={styles.updateName}>{updatedAt} 更新</div>
        </div>
      </div>
    </div>
  );
}
