import { useNavigate } from 'react-router-dom';

const MyMcpTab = () => {
  const navigate = useNavigate();
  const handleClick = () => {
    navigate(`/mcp-detail?serviceId=3&mcpFrom=myMcp`);
  };
  return <div onClick={handleClick}>{`我的mcp${Math.random()}`}</div>;
};

export default MyMcpTab;
