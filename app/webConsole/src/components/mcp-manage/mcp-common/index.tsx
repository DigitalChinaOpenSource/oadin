import { Input } from 'antd';
import { MagnifyingGlassIcon } from '@phosphor-icons/react';
import styles from './index.module.scss';
import { IUseViewModelReturn } from '@/components/mcp-manage/my-mcp-tab/view-model.ts';
import type { ListGridType } from 'antd/es/list';
import { IMcpListItem } from '@/components/mcp-manage/mcp-square-tab/types.ts';
import { Dispatch, SetStateAction } from 'react';
export interface IMcpCommonProps {
  isDialog?: boolean; // 是否是对话框模式
  showOnlySelectedMyMcp?: boolean;
  showOnlySelectedMcpList?: boolean;
  activeKey?: string;
  selectTemporaryMcpItems?: IMcpListItem[];
  setSelectTemporaryMcpItems?: Dispatch<SetStateAction<IMcpListItem[]>>;
}
interface IRenderTitleProps extends IMcpCommonProps {
  count: number;
  title?: string;
}

interface IRenderSerchProps {
  vm: IUseViewModelReturn;
}
interface IRenderWarpProps extends IMcpCommonProps {
  filterNode?: React.ReactNode;
  titleNode?: React.ReactNode;
  childrenNode?: React.ReactNode;
}

export const renderTitle = (props: IRenderTitleProps) => {
  const { count, title, isDialog } = props;
  return isDialog ? null : (
    <div className={styles.mcpTitleText}>
      <span>{title}</span>
      {!!count && <span className={styles.mcpCount}>共 {count} 个</span>}
    </div>
  );
};

export const renderSearch = (props: IRenderSerchProps) => {
  const { vm } = props;
  return (
    <div className={styles.searchInput}>
      <Input
        allowClear
        placeholder="请输入 MCP 服务名称"
        suffix={
          <div
            className={styles.searchIcon}
            onClick={() => vm.onMcpInputSearch(vm.searchVal)}
          >
            <MagnifyingGlassIcon
              width={16}
              height={16}
              fill="#808899"
            />
          </div>
        }
        value={vm.searchVal}
        onChange={(e) => vm.setSearchVal(e.target.value.trim())}
        // onSearch={(value) => vm.onMcpInputSearch(value)}
        onClear={() => {
          vm.setSearchVal('');
          vm.onMcpInputSearch('');
        }}
        onPressEnter={() => vm.onMcpInputSearch(vm.searchVal)}
        style={{ width: 380 }}
      />
    </div>
  );
};

export const renderWarp = (props: IRenderWarpProps) => {
  const { filterNode, titleNode, childrenNode, isDialog } = props;
  return (
    <div className={styles.mcp_common}>
      {/* 列表区域 */}
      <div className={isDialog ? styles.mcp_common_content_dialog : styles.mcp_common_content}>
        <div className={styles.mcpTitle}>{titleNode}</div>
        {childrenNode}
      </div>
      {/* 过滤器 */}
      {filterNode}
    </div>
  );
};

export const genGrid = (props: IMcpCommonProps): ListGridType => {
  return props.isDialog ? { gutter: 16, column: 2, xs: 1, sm: 1, md: 2, lg: 2, xl: 2, xxl: 2 } : { gutter: 16, column: 3, xs: 1, sm: 1, md: 2, lg: 2, xl: 2, xxl: 3 };
};
