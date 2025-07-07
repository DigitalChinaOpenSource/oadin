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

import { Breadcrumb } from 'antd';
import { useMatches, NavLink } from 'react-router-dom';
import type { BreadcrumbItemType } from 'antd/es/breadcrumb/Breadcrumb';
import React from 'react';

function useBreadcrumbItems() {
  const matches = useMatches();
  return matches
    .filter((route) => route.handle && (route.handle as { breadcrumb?: unknown }).breadcrumb)
    .map((route) => {
      const handle = route.handle as { breadcrumb?: string | ((args: { params: Record<string, string>; route: typeof route }) => string) };
      let label = handle.breadcrumb;
      if (typeof label === 'function') {
        label = label({ params: (route as { params?: Record<string, string> }).params ?? {}, route });
      }
      const pathname = 'pathname' in route ? (route as { pathname?: string }).pathname : '';
      const id = 'id' in route ? (route as { id?: string }).id : '';
      return {
        title: label,
        key: pathname || id || '',
        path: pathname || '',
      };
    });
}

const CustomBreadcrumb: React.FC = () => {
  const breadcrumbItems = useBreadcrumbItems();
  const itemRender = (route: Partial<BreadcrumbItemType>, params: Record<string, string>, routes: Partial<BreadcrumbItemType>[], paths: string[]): React.ReactNode => {
    if (routes.indexOf(route) === routes.length - 1) {
      return <span>{route.title}</span>;
    }
    return (
      <NavLink
        style={{ color: 'var(--color-primary)', fontWeight: '400' }}
        to={route.path as string}
      >
        {route.title}
      </NavLink>
    );
  };
  return (
    <Breadcrumb
      items={breadcrumbItems}
      itemRender={itemRender}
    />
  );
};

export default CustomBreadcrumb;
