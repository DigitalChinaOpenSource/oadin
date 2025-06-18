import TagsRender from '@/components/tags-render';
import { McpDetailType } from '@/components/mcp-manage/mcp-detail/type.ts';
import { GlobeIcon, HardDrivesIcon } from '@phosphor-icons/react';
import styles from './index.module.scss';
import defaultPng from '@/assets/favicon.png';

export default function DrawerDetailTitle(props: { mcpDetail: McpDetailType }) {
  const { logo, tags, name, hosted } = props.mcpDetail;

  return (
    <div className={styles.detailDescMain}>
      <div className={styles.detailIcon}>
        <img
          src={logo || defaultPng}
          alt=""
        />
      </div>
      <div className={styles.detailTitleContent}>
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
        </div>

        <div className={styles.tags}>
          <TagsRender tags={tags || []} />
        </div>
      </div>
    </div>
  );
}
