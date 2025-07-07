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

import { useEffect, useState } from 'react';
import { Outlet } from 'react-router-dom';
import styles from './index.module.scss';
import useServerCheckStore from '@/store/useServerCheckStore';
import { Layout, Tooltip } from 'antd';
import TopHeader from '@/components/main-layout/top-header';
import { CaretDoubleLeftIcon, CaretDoubleRightIcon } from '@phosphor-icons/react';
import TopMenu from '../top-menu';

export default function MainLayout() {
  const { Header, Content, Sider } = Layout;
  const [collapsed, setCollapsed] = useState(false);

  // 获取和更新服务的健康状态
  const { fetchServerStatus } = useServerCheckStore();
  useEffect(() => {
    const interval = setInterval(() => {
      fetchServerStatus();
    }, 30000);

    // 清除定时器
    return () => clearInterval(interval);
  }, [fetchServerStatus]);

  const handleCollapse = (collapsed: boolean) => {
    setCollapsed(collapsed);
  };

  const TriggerIcon = () => {
    return (
      <div className={styles.triggerContent}>
        <div className={styles.triggerIconContent}>
          <Tooltip
            title={collapsed ? '展开' : '收起'}
            placement={'top'}
          >
            {collapsed ? (
              <CaretDoubleRightIcon
                width={20}
                height={20}
                fill="#71717D"
              />
            ) : (
              <CaretDoubleLeftIcon
                width={20}
                height={20}
                fill="#71717D"
              />
            )}
          </Tooltip>
        </div>
      </div>
    );
    // return collapsed ? <span className={styles.triggerIcon}>▶</span> : <span className={styles.triggerIcon}>◀</span>;
  };

  return (
    <Layout className={styles.mainLayout}>
      <Header className={styles.header}>
        <div className={styles.logo}>LOGO</div>
        <div className={styles.topMenu}>
          <TopMenu />
        </div>

        {/* <TopHeader /> */}
      </Header>

      <Content className={styles.content}>
        <Outlet />
      </Content>
    </Layout>
  );
}
