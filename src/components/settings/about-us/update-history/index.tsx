import { Drawer } from 'antd';
import { useUpdateHistory } from '@/components/settings/about-us/update-history/view.module.ts';
interface IUpdateHistoryProps {
  open: boolean;
  onClose: () => void;
}

export default function UpdateHistory({ open, onClose }: IUpdateHistoryProps) {
  const { historyLoading } = useUpdateHistory(open); // Replace with actual loading state logic
  return (
    <Drawer
      closable={true}
      width={'50%'}
      destroyOnHidden={true}
      title={'更新日志'}
      placement={'right'}
      loading={historyLoading}
      open={open}
      onClose={onClose}
    >
      <div>1</div>
      <div>12</div>
      <div>123</div>
    </Drawer>
  );
}
