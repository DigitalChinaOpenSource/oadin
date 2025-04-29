import React, { useCallback } from 'react';
import { IModelCardProps } from './index';

export function useViewModel(props: IModelCardProps) {
  const [isDetailVisible, setIsDetailVisible] = React.useState(false);

  const handleDetailVisible = useCallback(() => {
    setIsDetailVisible(true);
  }, []);

  const onDetailClose = useCallback(() => {
    setIsDetailVisible(false);
  }, []);

  return {
    isDetailVisible,
    handleDetailVisible,
    onDetailClose,
  };
}
