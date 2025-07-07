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
import { Flex, Button, notification, Form, Skeleton, message } from 'antd';
import ModelCard from '@/components/model-card';
import { useEffect } from 'react';
import FloatInput from '@/components/float-input';
import { httpRequest } from '@/utils/httpRequest.ts';
import { useRequest } from 'ahooks';
import { useParams } from 'react-router-dom';

const RemoteModelDetail = () => {
  const params = useParams();
  const [messageApi, messageContextHolder] = message.useMessage();
  const [api, contextHolder] = notification.useNotification();
  const [form] = Form.useForm();
  const service_name = params.type as string;
  const { loading, data, run } = useRequest(
    () =>
      httpRequest.get('/control_panel/modellist', {
        service_name,
        service_source: 'remote',
        page_size: 1,
        page: 1,
        search_name: params.name,
      }),
    {
      manual: true,
      onSuccess: (res) => {
        if (res.data[0].can_select) {
          getKey({
            model_name: res.data[0].name,
            provider_name: res.data[0].service_provider_name,
          });
        }
      },
    },
  );
  const { run: setDefaultModel, loading: setting } = useRequest((data) => httpRequest.post('/control_panel/set_default', data), {
    manual: true,
    onSuccess: (res) => {
      run();
    },
  });
  const { run: getKey } = useRequest((data) => httpRequest.post('/control_panel/modelkey', data), {
    manual: true,
    onSuccess: (res) => {
      const model_key = JSON.parse(res.model_key);
      form.setFieldsValue(model_key);
    },
  });
  const { loading: saving, run: save } = useRequest(
    (auth_key: { url?: string; auth_key?: string }) =>
      httpRequest.put('/service_provider', {
        service_name,
        service_source: 'remote',
        api_flavor: model?.flavor,
        provider_name: model?.service_provider_name,
        auth_type: model?.auth_type,
        models: [model?.name],
        auth_key: JSON.stringify(auth_key),
      }),
    {
      manual: true,
      onSuccess: () => {
        messageApi.success('Save successfully');
        run();
      },
      onError: (error) => {
        run();
        // messageApi.error('Authorization verification failed,please fill in authorization information again');
        form.resetFields();
      },
    },
  );
  const submit = () => {
    const values = form.getFieldsValue();
    for (const key in values) {
      if (Object.prototype.hasOwnProperty.call(values, key)) {
        const element = values[key];
        if (!element) {
          messageApi.error('Please fill in authorization information');
          return;
        }
      }
    }

    save(values);
  };
  const model = data?.data[0];
  useEffect(() => {
    run();
  }, [run]);
  return (
    // <Spin spinning={loading}>
    <Flex
      vertical
      gap={24}
      align="start"
    >
      {contextHolder}
      {messageContextHolder}
      <CustomBreadcrumb />
      <Skeleton
        loading={loading || setting}
        active
        avatar
      >
        {model && (
          <ModelCard
            model={model}
            showSet
            setDefault={() => {
              setDefaultModel({
                model_name: model.name,
                service_name: model.service_name,
                service_source: model.source,
                provider_name: model.service_provider_name,
              });
            }}
          />
        )}
      </Skeleton>

      {model?.auth_fields.length > 0 && model?.auth_fields[0] && (
        <>
          <div>Authorization</div>
          <Form
            layout={'inline'}
            form={form}
            size="large"
          >
            {model?.auth_fields.includes('url') && (
              <Form.Item
                label=""
                name="url"
              >
                <FloatInput
                  style={{ width: 300 }}
                  placeholder="API-Host"
                />
              </Form.Item>
            )}
            {model?.auth_fields.includes('api_key') && (
              <Form.Item
                label=""
                name="api_key"
              >
                <FloatInput
                  // visibilityToggle={false}
                  style={{ width: 300 }}
                  placeholder="API-Key"
                  type="password"
                  autoComplete="off"
                />
              </Form.Item>
            )}
          </Form>
          <Button
            size="large"
            type="primary"
            loading={saving}
            onClick={submit}
          >
            Save
          </Button>
        </>
      )}
    </Flex>
    // </Spin>
  );
};
export default RemoteModelDetail;
