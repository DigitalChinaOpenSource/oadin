//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

import React, { useState } from 'react';

interface CopySuccessProps {
  width?: number;
  height?: number;
  fill?: string;
  hoverFill?: string; // 新增 hoverFill 属性
}

const CopySuccess: React.FC<CopySuccessProps> = ({
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
    >
      <path d="M173.66,98.34a8,8,0,0,1,0,11.32l-56,56a8,8,0,0,1-11.32,0l-24-24a8,8,0,0,1,11.32-11.32L112,148.69l50.34-50.35A8,8,0,0,1,173.66,98.34ZM224,48V208a16,16,0,0,1-16,16H48a16,16,0,0,1-16-16V48A16,16,0,0,1,48,32H208A16,16,0,0,1,224,48ZM208,208V48H48V208H208Z"></path>
    </svg>
  );
};

export default CopySuccess;
