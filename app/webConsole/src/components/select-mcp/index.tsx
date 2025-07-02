import { Button, Popover, Tooltip } from 'antd';
import styles from './index.module.scss';
import { McpSelectBtn } from '@/components/icons';
import { SelectMcpDialog } from './dialog.tsx';
import useSelectMcpStore from '@/store/useSelectMcpStore.ts';
import useSelectedModelStore from '@/store/useSelectedModel.ts';
import { useEffect, useState } from 'react';
import { useViewModel as useMyMcpViewModel } from '@/components/mcp-manage/my-mcp-tab/view-model.ts';

enum selectMcpType {
  'default' = 'default',
  'selected' = 'selected',
  'disabled' = 'disabled',
}

export const SelectMcp = () => {
  const { selectMcpList } = useSelectMcpStore();
  const { selectedModel } = useSelectedModelStore();
  const [selectMcpPopOpen, setSelectMcpPopOpen] = useState<boolean>(false);
  const myMcpViewModel = useMyMcpViewModel();

  const { handleMcpListToPage } = myMcpViewModel;
  // 根据selectMcpList长度确定类型
  const selectType = selectMcpList.length > 0 ? selectMcpType.selected : selectMcpType.default;
  const selectTypeClass = styles[`select_mcp_${selectMcpType[selectType]}`];
  useEffect(() => {
    handleMcpListToPage();
  }, [selectMcpPopOpen]);
  return (
    <div className={`${styles.select_mcp} ${selectTypeClass}`}>
      <Popover
        placement="topLeft"
        open={selectMcpPopOpen}
        arrow={false}
        content={
          <SelectMcpDialog
            myMcpViewModel={myMcpViewModel}
            setSelectMcpPopOpen={setSelectMcpPopOpen}
          />
        }
        trigger="click"
        onOpenChange={(open) => {
          setSelectMcpPopOpen(open);
          if (!open) {
            // message.info('调用启动MCP服务');
          }
        }}
      >
        <Tooltip
          title={!selectedModel?.tools && '所选模型不支持该服务'}
          placement="top"
        >
          <Button
            disabled={!selectedModel?.tools}
            icon={
              <McpSelectBtn
                width={20}
                height={20}
              />
            }
          >
            已选<span className={styles.select_mcp_count_warp}>{selectMcpList.length || 0}</span>
          </Button>
        </Tooltip>
      </Popover>
    </div>
  );
};
