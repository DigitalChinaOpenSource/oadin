import { useViewModel } from './view-model';
import McpList from '../mcp-list';
import { genGrid, IMcpCommonProps, renderSearch, renderTitle, renderWarp } from '@/components/mcp-manage/mcp-common';

export interface IMyMcpTabProps extends IMcpCommonProps {}

export default function MyMcpTab(props: IMyMcpTabProps) {
  const vm = useViewModel();
  const { isDialog, showOnlySelectedMcpList, showOnlySelectedMyMcp, activeKey } = props;

  return renderWarp({
    isDialog,
    childrenNode: (
      <McpList
        showOnlySelectedMcpList={showOnlySelectedMcpList}
        showOnlySelectedMyMcp={showOnlySelectedMyMcp}
        activeKey={activeKey}
        isSelectable={isDialog}
        isDialog={isDialog}
        grid={genGrid({ isDialog })}
        mcpListData={vm.mcpListData}
        pagination={vm.pagination}
        onPageChange={vm.handlePageChange}
        handelMcpCardClick={vm.handelMcpCardClick}
        pageLoading={vm.mcpListLoading}
      />
    ),
    titleNode: (
      <>
        {isDialog
          ? null
          : renderTitle({
              title: '我的MCP',
              count: vm.pagination.total ?? 0,
            })}
        {renderSearch({ vm })}
      </>
    ),
  });
}
