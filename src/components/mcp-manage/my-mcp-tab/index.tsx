import { useViewModel } from './view-model';
import McpList from '../mcp-list';
import { renderSearch, renderTitle, renderWarp } from '@/components/mcp-manage/mcp-common';
export interface MyMcpTabProps {
  isDialog?: boolean; // 是否是对话框模式
}
export default function MyMcpTab(props: MyMcpTabProps) {
  const vm = useViewModel();
  const { isDialog } = props;

  return renderWarp({
    isDialog,
    childrenNode: (
      <McpList
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
