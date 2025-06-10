import { Input } from 'antd';
import { MagnifyingGlassIcon } from '@phosphor-icons/react';
import styles from './index.module.scss';
import { IUseViewModelReturn } from '@/components/mcp-manage/my-mcp-tab/view-model.ts';

interface IRenderTitleProps {
  count: number;
  title?: string;
}

interface IRenderSerchProps {
  vm: IUseViewModelReturn;
}
interface IRenderWarpProps {
  isDialog?: boolean; // 是否是对话框模式
  filterNode?: React.ReactNode;
  titleNode?: React.ReactNode;
  childrenNode?: React.ReactNode;
}

export const renderTitle = (props: IRenderTitleProps) => {
  const { count, title } = props;
  return (
    <div className={styles.mcpTitleText}>
      <span>{title}</span>
      <span className={styles.mcpCount}>共 {count} 个</span>
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
  const { filterNode, titleNode, childrenNode } = props;
  return (
    <div className={styles.mcp_common}>
      {/* 列表区域 */}
      <div className={styles.mcp_common_content}>
        <div className={styles.mcpTitle}>{titleNode}</div>
        {childrenNode}
      </div>
      {/* 过滤器 */}
      {filterNode}
    </div>
  );
};
