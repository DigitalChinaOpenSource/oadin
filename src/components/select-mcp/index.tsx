import { Button, Popover } from 'antd';
import styles from './index.module.scss';
import { McpSelectBtn } from '@/components/icons';
import { SelectMcpDialog } from './dialog.tsx';
import useSelectMcpStore from '@/store/useSelectMcpStore.ts';

enum selectMcpType {
  'default' = 'default',
  'selected' = 'selected',
  'disabled' = 'disabled',
}

export const SelectMcp = () => {
  const { selectMcpList } = useSelectMcpStore();
  // 根据selectMcpList长度确定类型
  const selectType = selectMcpList.length > 0 ? selectMcpType.selected : selectMcpType.default;
  const selectTypeClass = styles[`select_mcp_${selectMcpType[selectType]}`];
  return (
    <div className={`${styles.select_mcp} ${selectTypeClass}`}>
      <Popover
        arrow={false}
        content={<SelectMcpDialog />}
        trigger="click"
      >
        <Button
          icon={
            <McpSelectBtn
              width={20}
              height={20}
            />
          }
        >
          已选<span className={styles.select_mcp_count_warp}>{selectMcpList.length || 0}</span>
        </Button>
      </Popover>
    </div>
  );
};
