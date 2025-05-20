import { useNavigate } from 'react-router-dom';

const MyMcp = () => {
  const navigate = useNavigate();
  const handleClick = () => {
    navigate(`/mcp-service-detail?serviceId=3&mcpFrom=myMcp`);
  };
  return <div onClick={handleClick}>{`我的mcp${Math.random()}`}</div>;
};

export default MyMcp;
