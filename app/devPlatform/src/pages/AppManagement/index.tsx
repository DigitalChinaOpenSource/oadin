import React from 'react';
import { Card, Typography, Row, Col, Statistic, Space, Button, Divider, List } from 'antd';
import { AppstoreAddOutlined, CloudOutlined, RocketOutlined, SettingOutlined } from '@ant-design/icons';

const { Title, Paragraph } = Typography;

const AppManagement: React.FC = () => {
  const appList = [
    { id: 1, name: '应用 1', description: '这是一个示例应用', status: '运行中' },
    { id: 2, name: '应用 2', description: '这是一个示例应用', status: '停止' },
    { id: 3, name: '应用 3', description: '这是一个示例应用', status: '部署中' },
    { id: 4, name: '应用 4', description: '这是一个示例应用', status: '运行中' },
  ];

  return (
    <div style={{ width: '100%' }}>
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <Card>
            <Title level={2}>应用管理</Title>
            <Paragraph>欢迎使用应用管理平台，您可以在这里管理所有的应用。</Paragraph>
          </Card>
        </Col>
      </Row>

      <Row
        gutter={[16, 16]}
        style={{ marginTop: 16 }}
      >
        <Col
          xs={24}
          sm={12}
          md={6}
        >
          <Card>
            <Statistic
              title="总应用数"
              value={8}
              prefix={<AppstoreAddOutlined />}
            />
          </Card>
        </Col>
        <Col
          xs={24}
          sm={12}
          md={6}
        >
          <Card>
            <Statistic
              title="运行中"
              value={4}
              prefix={<CloudOutlined />}
            />
          </Card>
        </Col>
        <Col
          xs={24}
          sm={12}
          md={6}
        >
          <Card>
            <Statistic
              title="部署中"
              value={1}
              prefix={<RocketOutlined />}
            />
          </Card>
        </Col>
        <Col
          xs={24}
          sm={12}
          md={6}
        >
          <Card>
            <Statistic
              title="已停止"
              value={3}
              prefix={<SettingOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Card style={{ marginTop: 16 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <Title level={4}>应用列表</Title>
          <Button type="primary">创建应用</Button>
        </div>
        <Divider />
        <List
          dataSource={appList.concat(appList)} // 使用 concat 以确保有数据
          renderItem={(item) => (
            <List.Item
              key={item.id}
              actions={[
                <Button
                  key="view"
                  type="link"
                >
                  查看
                </Button>,
                <Button
                  key="edit"
                  type="link"
                >
                  编辑
                </Button>,
                <Button
                  key="delete"
                  type="link"
                  danger
                >
                  删除
                </Button>,
              ]}
            >
              <List.Item.Meta
                title={item.name}
                description={item.description}
              />
              <div>状态: {item.status}</div>
            </List.Item>
          )}
        />
      </Card>
    </div>
  );
};

export default AppManagement;
