import styles from './index.module.scss';
import { Collapse, CollapseProps, Input, List, Switch, Tooltip } from 'antd';
import { SearchIcon } from '@/components/icons';
import TagsRender from '@/components/tags-render';
import { useMcpTools } from '@/components/mcp-manage/mcp-detail/mcp-tools/useMcpTolls.ts';

const CollapseItemHeader = (props: { name: string; desc: string; tags: string[] }) => {
  const { name, desc, tags = [] } = props;
  return (
    <div className={styles.collapseItemHeader}>
      <div className={styles.collapseItemTitle}>
        <div className={styles.collapseItemName}>{name}</div>
        <TagsRender tags={tags} />
      </div>
      <div className={styles.collapseItemDesc}>{desc}</div>
    </div>
  );
};

// const

export default function McpTools() {
  const { handlePageChange, mcpTolls, pagination, changeTollStatus } = useMcpTools();

  return (
    <div className={styles.mcpTools}>
      <div className={styles.searchInput}>
        <Input
          suffix={<SearchIcon />}
          allowClear
          placeholder="请输入搜索工具名称"
          // value={modelSearchVal}
          // onChange={(e) => {}}
          style={{ width: 380 }}
        />
      </div>
      <List
        className={styles.tollsContent}
        itemLayout="horizontal"
        dataSource={mcpTolls}
        pagination={{
          onChange: handlePageChange,
          ...pagination,
          align: 'end',
          position: 'bottom',
        }}
        renderItem={(item) => (
          <List.Item>
            <Collapse
              style={{ flex: 1 }}
              items={[
                {
                  ...item,
                  label: (
                    <CollapseItemHeader
                      name={item.name}
                      desc={item.description}
                      tags={item.tags || []}
                    />
                  ),
                  extra: (
                    <Tooltip title={item.enabled ? '关闭后，大模型将不会调度已关闭函数' : ''}>
                      <Switch
                        checkedChildren="ON"
                        unCheckedChildren="OFF"
                        checked={item.enabled}
                        loading={item.changTollLoading}
                        onClick={(checked, e) => {
                          e.stopPropagation();
                          changeTollStatus(item, checked);
                        }}
                      />
                    </Tooltip>
                  ),
                  children: <div>345</div>,
                  collapsible: 'disabled',
                },
              ]}
            />
          </List.Item>
        )}
      />
    </div>
  );
}
