import React, { useState } from 'react';

interface LocalIconProps {
  width?: number;
  height?: number;
  fill?: string;
  hoverFill?: string; // 新增 hoverFill 属性
}

const LocalIcon: React.FC<LocalIconProps> = ({
  width = 24,
  height = 24,
  fill = '#4cbd6e',
  hoverFill = '#000', // 默认 hover 颜色
}) => {
  const [currentFill, setCurrentFill] = useState(fill);

  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width={width}
      height={height}
      fill={currentFill}
      viewBox="0 0 256 256"
      // onMouseEnter={() => setCurrentFill(hoverFill)} // 鼠标移入时设置 hoverFill
      // onMouseLeave={() => setCurrentFill(fill)} // 鼠标移出时恢复默认 fill
    >
      <path d="M173.66,98.34a8,8,0,0,1,0,11.32l-56,56a8,8,0,0,1-11.32,0l-24-24a8,8,0,0,1,11.32-11.32L112,148.69l50.34-50.35A8,8,0,0,1,173.66,98.34ZM224,48V208a16,16,0,0,1-16,16H48a16,16,0,0,1-16-16V48A16,16,0,0,1,48,32H208A16,16,0,0,1,224,48ZM208,208V48H48V208H208Z"></path>
    </svg>
  );
};

export default LocalIcon;
