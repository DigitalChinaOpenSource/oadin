import { memo, useCallback } from 'react';
import { Button } from 'antd';
import { StopIcon } from '@phosphor-icons/react';
import sendSvg from '@/components/icons/send.svg';
import { useChatStream } from '../useChatStream';
export const InputButton = memo(
  (props: { isLoading: boolean; onClick: () => void }) => {
    const { isLoading, onClick } = props;
    const { cancelRequest } = useChatStream();
    console.log('InputButton==>', props);

    const handleCancelRequest = useCallback(() => {
      console.log('stop');
      cancelRequest();
    }, [cancelRequest]);

    return (
      <>
        {isLoading ? (
          <Button
            icon={
              <StopIcon
                width={24}
                weight="fill"
                fill="#4f4dff"
              />
            }
            onClick={() => {
              console.log('stop');
              handleCancelRequest();
            }}
          />
        ) : (
          <Button
            type="primary"
            style={{ borderRadius: 8, cursor: 'pointer' }}
            icon={
              <img
                src={sendSvg}
                alt="发送"
              />
            }
            onClick={onClick}
          />
        )}
      </>
    );
  },
  (prevProps: any, nextProps: any) => {
    console.log('====>', prevProps.isLoading === nextProps.isLoading);
    return nextProps.isLoading;
  },
);
