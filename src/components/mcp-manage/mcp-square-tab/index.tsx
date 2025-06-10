import { Tooltip } from 'antd';
import { useViewModel } from './view-model';
import expandSvg from '@/components/icons/expand.svg';
import McpAdvanceFilter from '../mcp-advance-filter';
import McpList from '../mcp-list';
import commonStyles from '@/components/mcp-manage/mcp-common/index.module.scss';
import { MyMcpTabProps } from '@/components/mcp-manage/my-mcp-tab';
import { renderSearch, renderTitle, renderWarp } from '@/components/mcp-manage/mcp-common';

export default function McpSquareTab(props: MyMcpTabProps) {
  const vm = useViewModel();
  const { isDialog } = props;
  return renderWarp({
    isDialog,
    titleNode: (
      <>
        {isDialog
          ? null
          : renderTitle({
              title: 'MCP广场',
              count: vm.pagination.total ?? 0,
            })}
        {renderSearch({
          vm,
        })}
        {vm.collapsed && (
          <Tooltip title="展开筛选">
            <div
              className={commonStyles.expandIcon}
              onClick={() => vm.setCollapsed(false)}
            >
              <img
                src={expandSvg}
                alt="折叠筛选面板"
              />
            </div>
          </Tooltip>
        )}
      </>
    ),
    childrenNode: (
      <McpList
        mcpListData={vm.mcpListData}
        onPageChange={vm.handlePageChange}
        handelMcpCardClick={vm.handelMcpCardClick}
        pagination={vm.pagination}
        pageLoading={vm.mcpListLoading}
      />
    ),
    filterNode: (
      <McpAdvanceFilter
        collapsed={vm.collapsed}
        setCollapsed={vm.setCollapsed}
        handleClearTags={vm.handleClearTags}
        tagsData={vm.tagsData}
        checkedValues={vm.checkedValues}
        handleTagsChange={vm.handleTagsChange}
      />
    ),
  });
}
