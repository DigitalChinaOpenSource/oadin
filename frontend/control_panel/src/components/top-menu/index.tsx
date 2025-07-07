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

// 顶部菜单组件
import React, { useEffect, useMemo } from 'react';
import { NavLink, useLocation } from 'react-router-dom';
import { Menu } from 'antd';

const items = [
  {
    key: 'dashboard',
    label: <NavLink to="/dashboard">Dashboard</NavLink>,
  },
  {
    key: 'about-aog',
    label: <NavLink to="/about-aog/choose-service">About AOG</NavLink>,
  },
];

const TopMenu: React.FC = () => {
  const location = useLocation();
  const selectedKey = useMemo(() => {
    // 计算逻辑
    return location.pathname.split('/').filter(Boolean)[0];
  }, [location.pathname]);
  return (
    <Menu
      mode="horizontal"
      selectedKeys={selectedKey ? [selectedKey] : []}
      items={items}
      style={{ borderBottom: 0 }}
    />
  );
};

export default TopMenu;
