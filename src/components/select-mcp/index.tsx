import { Button } from 'antd';
import styles from './index.module.scss';
import { McpSelectBtn } from '@/components/icons';

enum selectMcpType {
  'default' = 'default',
  'selected' = 'selected',
  'disabled' = 'disabled',
}

export const SelectMcp = () => {
  const selectType: selectMcpType = selectMcpType['default'];
  const selectTypeClass = styles[`select_mcp_${selectMcpType[selectType]}`];
  return (
    <div className={`${styles.select_mcp} ${selectTypeClass}`}>
      <Button
        icon={
          <McpSelectBtn
            width={20}
            height={20}
            color="#27272A"
          />
        }
      >
        已选<span className={styles.select_mcp_count_warp}>0</span>
      </Button>
    </div>
  );
};
