import { Modal } from 'antd';
interface ICopyrightNoticeProps {
  open: boolean;
  onClose: () => void;
}

export default function CopyrightNotice({ open, onClose }: ICopyrightNoticeProps) {
  return (
    <Modal
      centered={true}
      closable={true}
      title={'版权声明'}
      onCancel={onClose}
      open={open}
      onOk={onClose}
    >
      这里是版权声明内容
    </Modal>
  );
}
