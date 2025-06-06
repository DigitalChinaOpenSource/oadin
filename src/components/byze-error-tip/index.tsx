import { Alert, Button } from 'antd';
import useByzeServerCheckStore from '@/store/useByzeServerCheckStore.ts';

export default function ByzeErrorTip() {
  // 白泽服务状态
  const { checkByzeStatus, fetchByzeServerStatus } = useByzeServerCheckStore();

  return (
    <>
      {!checkByzeStatus ? (
        <Alert
          type="error"
          showIcon
          banner={true}
          closable={true}
          message={
            <div>
              服务状态异常，点击
              <Button
                type={'link'}
                onClick={fetchByzeServerStatus}
              >
                重启服务
              </Button>{' '}
            </div>
          }
        />
      ) : null}
    </>
  );
}
