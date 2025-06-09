import { Drawer } from 'antd';

export interface IChatHistoryDrawerProps {
  onHistoryDrawerClose?: () => void;
}
export default function ChatHistoryDrawer({ onHistoryDrawerClose }: IChatHistoryDrawerProps) {
  return (
    <Drawer
      title="历史对话"
      placement="right"
      closable={true}
      maskClosable={true}
      onClose={onHistoryDrawerClose}
      open={true}
    >
      <div>
        <h2>历史对话</h2>
        <p>这将显示聊天历史记录。</p>
      </div>
    </Drawer>
  );
}
