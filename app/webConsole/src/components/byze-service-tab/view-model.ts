import { useState, useEffect } from 'react';
import { healthRequest } from '@/utils/httpRequest';
import { message } from 'antd';
import { useRequest } from 'ahooks';
import useByzeServerCheckStore from '@/store/useByzeServerCheckStore';
export function useViewModel() {
  const { checkByzeStatus, checkByzeServerLoading, fetchByzeServerStatus } = useByzeServerCheckStore();

  const handleRefresh = () => {
    fetchByzeServerStatus();
  };
  return { handleRefresh, checkByzeServerLoading, checkByzeStatus };
}
