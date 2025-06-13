import { Button, Form, Input, Modal } from 'antd';
import { McpDetailType } from '@/components/mcp-manage/mcp-detail/type.ts';
type McpAddModalProps = {
  showMcpModal: boolean;
  setShowMcpModal: (isShow: boolean) => void;
  handleAuthMcp: (authParams: any, curMcpDetail?: McpDetailType) => void;
  mcpDetail: McpDetailType;
  operateType?: 'add' | 'edit';
};

export default function McpAuthModal(props: McpAddModalProps) {
  const { showMcpModal, setShowMcpModal, handleAuthMcp, mcpDetail, operateType } = props;
  const { envSchema } = mcpDetail;
  const { properties } = envSchema;
  const [form] = Form.useForm();

  const onFinish = (formValues: Record<string, any>) => {
    handleAuthMcp(formValues, mcpDetail);
  };
  return (
    <Modal
      title={operateType === 'edit' ? 'MCP 服务更新' : 'MCP 服务添加'}
      open={showMcpModal}
      onCancel={() => setShowMcpModal(false)}
      footer={null}
      centered
      destroyOnHidden={true}
      // width={600}
      styles={{
        header: {
          marginBottom: 16,
        },
      }}
    >
      {/*<div style={{ margin: '16px 0' }}>这里是描述</div>*/}
      <Form
        form={form}
        onFinish={onFinish}
        layout="vertical"
        labelAlign="right" // 设置 label 右对齐
        clearOnDestroy={true}
      >
        {Object.entries(properties || {}).map(([key, config]) => (
          <Form.Item
            key={key}
            name={key}
            label={key}
            rules={[{ required: config.required, message: `${key}是必填项` }]}
          >
            <Input placeholder={`请输入${config.description}`} />
          </Form.Item>
        ))}
        <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
          <Button
            htmlType="button"
            onClick={() => setShowMcpModal(false)}
            style={{ marginRight: 8 }}
          >
            取消
          </Button>
          <Button
            type="primary"
            htmlType="submit"
          >
            确认授权
          </Button>
        </Form.Item>
      </Form>
    </Modal>
  );
}
