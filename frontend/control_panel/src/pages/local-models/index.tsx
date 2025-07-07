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

import CustomBreadcrumb from '@/components/custom-breadcrumb';
import { Flex, Input, Col, Pagination, Row, Spin } from 'antd';
import SearchIcon from '@/components/icons/search.svg?react';
import ModelCard from '@/components/model-card';
import { Outlet, useLocation, useParams } from 'react-router-dom';
import { httpRequest } from '@/utils/httpRequest.ts';
import { useRequest } from 'ahooks';
import { useEffect, useState } from 'react';
import type { ModelItem } from '@/components/model-card';

export default function LocalModels() {
  const location = useLocation();
  const params = useParams();
  const showOutlet = !!params.name;
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState('');
  const PAGE_SIZE = 9;

  // 计算 service_name
  const service_name = params.type;

  const { loading, data, run } = useRequest(
    (pageNum: number, searchVal: string) =>
      httpRequest.get('/control_panel/modellist', {
        service_name,
        service_source: 'local',
        page_size: PAGE_SIZE,
        page: pageNum,
        search_name: searchVal.trim() || undefined,
      }),
    {
      manual: true,
    },
  );

  useEffect(() => {
    if (!showOutlet) {
      // setPage(1);
      run(page, search);
    }
  }, [showOutlet]);

  if (showOutlet) {
    return <Outlet />;
  }

  // 处理分页
  const handlePageChange = (pageNum: number) => {
    setPage(pageNum);
    run(pageNum, search);
  };

  // 处理搜索
  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setSearch(val);
    setPage(1);
    run(1, val);
  };

  // 渲染模型卡片

  const models: ModelItem[] = data?.data || [];
  const total = data?.total || 0;

  return (
    <Flex
      vertical
      gap={24}
    >
      <Flex
        justify="space-between"
        align="center"
      >
        <CustomBreadcrumb />
        <Input
          size="small"
          style={{ width: '450px' }}
          placeholder="Search for local models"
          prefix={<SearchIcon />}
          value={search}
          onChange={handleSearch}
          allowClear
        />
      </Flex>
      <Spin spinning={loading}>
        <Row gutter={[20, 20]}>
          {models.length === 0 && !loading ? (
            <Col
              span={24}
              style={{ textAlign: 'center', color: '#aaa', padding: '40px 0' }}
            >
              No Data
            </Col>
          ) : null}
          {models.map((item) => (
            <Col
              span={8}
              key={item.id || item.name}
            >
              <ModelCard
                model={item}
                showSet={false}
              />
            </Col>
          ))}
        </Row>
      </Spin>
      <Pagination
        hideOnSinglePage
        current={page}
        pageSize={PAGE_SIZE}
        total={total}
        onChange={handlePageChange}
        align="end"
      />
    </Flex>
  );
}
