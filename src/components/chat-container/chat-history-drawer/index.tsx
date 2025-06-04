import { Drawer } from 'antd';

export default function ChatHistoryDrawer() {
  return (
    <Drawer
      title="历史对话"
      placement="right"
      closable={true}
      onClose={() => {}}
      open={true}
    >
      <div>
        <h2>历史对话</h2>
        <p>这将显示聊天历史记录。</p>
      </div>
    </Drawer>
  );
}
