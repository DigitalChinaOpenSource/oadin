import styles from './index.module.scss';
import { Collapse, CollapseProps, Input, List, Switch, Tooltip } from 'antd';
import { SearchIcon } from '@/components/icons';
import TagsRender from '@/components/tags-render';

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
  const mcpTollData = [
    { id: '1', key: '1', label: '测试1', description: '测试内容1', tags: ['测试1', '测算出'], enabled: true },
    { id: '2', key: '2', label: '测试2', description: '测试内容2', tags: [], enabled: true },
    { id: '3', key: '3', label: '测试3', description: '测试内容3', tags: [], enabled: true },
    { id: '3', key: '3', label: '测试3', description: '测试内容3', tags: [], enabled: true },
    { id: '3', key: '3', label: '测试3', description: '测试内容3', tags: [], enabled: true },
    { id: '3', key: '3', label: '测试3', description: '测试内容3', tags: [], enabled: true },
    { id: '3', key: '3', label: '测试3', description: '测试内容3', tags: [], enabled: true },
    { id: '3', key: '3', label: '测试3', description: '测试内容3', tags: [], enabled: true },
  ];

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
        dataSource={mcpTollData}
        // pagination={{
        //   onChange: (page) => {
        //     console.log(page);
        //   },
        //   pageSize: 10,
        //   align: 'end',
        //   position: 'top',
        // }}
        renderItem={(item) => (
          <List.Item>
            <Collapse
              style={{ flex: 1 }}
              items={[
                {
                  ...item,
                  label: (
                    <CollapseItemHeader
                      name={item.label}
                      desc={item.description}
                      tags={item.tags || []}
                    />
                  ),
                  extra: (
                    <Tooltip title={true ? '关闭后，大模型将不会调度已关闭函数' : '未启用'}>
                      <Switch
                        checkedChildren="ON"
                        unCheckedChildren="OFF"
                        defaultChecked={item.enabled}
                        onClick={(checked, e) => {
                          e.stopPropagation();
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
