import { Button, message, Popover } from 'antd';
import styles from './index.module.scss';
import { McpSelectBtn } from '@/components/icons';
import { SelectMcpDialog } from './dialog.tsx';
import useSelectMcpStore from '@/store/useSelectMcpStore.ts';
import { useState } from 'react';

enum selectMcpType {
  'default' = 'default',
  'selected' = 'selected',
  'disabled' = 'disabled',
}

export const SelectMcp = () => {
  const { selectMcpList } = useSelectMcpStore();
  const [selectMcpPopOpen, setSelectMcpPopOpen] = useState<boolean>(false);
  // 根据selectMcpList长度确定类型
  const selectType = selectMcpList.length > 0 ? selectMcpType.selected : selectMcpType.default;
  const selectTypeClass = styles[`select_mcp_${selectMcpType[selectType]}`];
  return (
    <div className={`${styles.select_mcp} ${selectTypeClass}`}>
      <Popover
        open={selectMcpPopOpen}
        arrow={false}
        content={<SelectMcpDialog setSelectMcpPopOpen={setSelectMcpPopOpen} />}
        trigger="click"
        onOpenChange={(open) => {
          setSelectMcpPopOpen(open);
          if (!open) {
            message.info('调用启动MCP服务');
          }
        }}
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
