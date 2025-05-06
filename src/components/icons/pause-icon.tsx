import React from 'react';

interface PauseIconProps {
  width?: number;
  height?: number;
  fill?: string;
}

const PauseIcon: React.FC<PauseIconProps> = ({ width = 16, height = 16, fill = '#9daabb' }) => {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" width={width} height={height} fill={fill} viewBox="0 0 256 256">
      <path d="M200,32H160a16,16,0,0,0-16,16V208a16,16,0,0,0,16,16h40a16,16,0,0,0,16-16V48A16,16,0,0,0,200,32Zm0,176H160V48h40ZM96,32H56A16,16,0,0,0,40,48V208a16,16,0,0,0,16,16H96a16,16,0,0,0,16-16V48A16,16,0,0,0,96,32Zm0,176H56V48H96Z"></path>
    </svg>
  );
};

export default PauseIcon;
