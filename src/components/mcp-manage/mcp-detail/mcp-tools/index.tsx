import styles from './index.module.scss';
import { Collapse, CollapseProps, List } from 'antd';
// type CollapseItemType = {;

export default function McpTools() {
  const mcpTollData = [
    { id: '1', key: '1', label: '测试1', description: '测试内容1' },
    { id: '2', key: '2', label: '测试2', description: '测试内容2' },
    { id: '3', key: '3', label: '测试3', description: '测试内容3' },
  ];
  return (
    <div className={styles.mcpTools}>
      <List
        // className={}
        itemLayout="horizontal"
        // grid={{ gutter: 16, column: 3, xs: 1, sm: 1, md: 2, lg: 2, xl: 2, xxl: 3 }}
        dataSource={mcpTollData}
        pagination={{
          onChange: (page) => {
            console.log(page);
          },
          pageSize: 3,
          align: 'end',
        }}
        renderItem={(item) => (
          <List.Item>
            <Collapse
              style={{ width: '100%' }}
              items={[{ ...item, extra: <div>123</div>, children: <div>345</div> }]}
            />
          </List.Item>
        )}
      />
    </div>
  );
}
