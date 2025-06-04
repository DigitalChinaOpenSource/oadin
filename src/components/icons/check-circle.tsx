import React, { useState } from 'react';

interface CheckCircleIconProps {
  width?: number;
  height?: number;
  fill?: string;
  hoverFill?: string;
}

const CheckCircleIcon: React.FC<CheckCircleIconProps> = ({ width = 24, height = 24, fill = '#4f4dff' }) => {
  const [currentFill, setCurrentFill] = useState(fill);

  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width={width}
      height={height}
      fill={currentFill}
      viewBox="0 0 256 256"
    >
      <path d="M128,24A104,104,0,1,0,232,128,104.11,104.11,0,0,0,128,24Zm45.66,85.66-56,56a8,8,0,0,1-11.32,0l-24-24a8,8,0,0,1,11.32-11.32L112,148.69l50.34-50.35a8,8,0,0,1,11.32,11.32Z"></path>
    </svg>
  );
};

export default CheckCircleIcon;
