import { useState } from 'react';
export function useViewModel() {
  const [mcpListData, setMcpListData] = useState([]);

  return {
    mcpListData,
  };
}
