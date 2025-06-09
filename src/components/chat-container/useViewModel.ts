import { useState, useEffect } from 'react';
export default function useViewModel() {
  const [historyVisible, setHistoryVisible] = useState(false);

  return {
    historyVisible,
    setHistoryVisible,
  };
}
