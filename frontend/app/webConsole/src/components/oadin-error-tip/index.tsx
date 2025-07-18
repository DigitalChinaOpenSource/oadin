import { Alert, Button } from 'antd';
import useOadinServerCheckStore from '@/store/useOadinServerCheckStore';
import { useEffect, useRef } from 'react';

export default function OadinErrorTip() {
  // 奥丁服务状态
  const { checkOadinStatus } = useOadinServerCheckStore();

  // 使用 ref 记录上一次的状态
  const prevStatusRef = useRef(checkOadinStatus);

  useEffect(() => {
    // 如果之前状态为 false，现在变为 true，则刷新页面
    if (!prevStatusRef.current && checkOadinStatus) {
      window.location.reload();
    }

    // 更新上一次状态的引用
    prevStatusRef.current = checkOadinStatus;
  }, [checkOadinStatus]);

  return (
    <>
      {!checkOadinStatus ? (
        <Alert
          type="error"
          showIcon
          banner={true}
          closable={true}
          message={
            <div>
              {/*奥丁服务状态异常，点击*/}
              奥丁服务状态异常
              {/*<Button*/}
              {/*  type={'link'}*/}
              {/*  onClick={fetchOadinServerStatus}*/}
              {/*>*/}
              {/*  重启服务*/}
              {/*</Button>{' '}*/}
            </div>
          }
        />
      ) : null}
    </>
  );
}
