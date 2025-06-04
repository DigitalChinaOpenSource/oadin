import React, { useState } from 'react';

interface LocalIconProps {
  width?: number;
  height?: number;
  fill?: string;
  hoverFill?: string; // 新增 hoverFill 属性
}

const ArrowLineRight: React.FC<LocalIconProps> = ({
  width = 16,
  height = 16,
  fill = '#71717D',
  hoverFill = 'blue', // 默认 hover 颜色
}) => {
  const [currentFill, setCurrentFill] = useState(fill);

  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width={width}
      height={height}
      fill={currentFill}
      viewBox="0 0 256 256"
      onMouseEnter={() => setCurrentFill(hoverFill)} // 鼠标移入时设置 hoverFill
      onMouseLeave={() => setCurrentFill(fill)} // 鼠标移出时恢复默认 fill
    >
      <path d="M189.66,122.34a8,8,0,0,1,0,11.32l-72,72a8,8,0,0,1-11.32-11.32L164.69,136H32a8,8,0,0,1,0-16H164.69L106.34,61.66a8,8,0,0,1,11.32-11.32ZM216,32a8,8,0,0,0-8,8V216a8,8,0,0,0,16,0V40A8,8,0,0,0,216,32Z"></path>
    </svg>
  );
};

export default ArrowLineRight;
