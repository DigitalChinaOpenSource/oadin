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

import { Flex, Modal, Select } from 'antd';
import type { LabeledValue } from 'antd/es/select';
import { useState, useMemo } from 'react';
import type { Service } from '@/hooks/useServices';
import CheckIcon from '@/components/icons/Check.svg?react';
import { upFirstLetter } from '@/utils';
type InstallServiceProps = {
  isModalOpen: boolean;
  setIsModalOpen: (isOpen: boolean) => void;
  install: (service: Service) => Promise<void>;
  services: Service[];
};
const InstallService: React.FC<InstallServiceProps> = (props: InstallServiceProps) => {
  const [selectedService, setSelectedService] = useState<LabeledValue>();
  const [status, setStatus] = useState<'error' | 'warning'>();

  const options = useMemo(() => {
    return props.services.length
      ? props.services.map((item) => ({
        value: item.service_name,
        label: upFirstLetter(item.service_name),
      }))
      : [
        {
          value: 'none',
          label: 'There are no selectable services for now.',
          disabled: true,
        },
      ];
  }, [props.services]);
  const handleCancel = () => {
    props.setIsModalOpen(false);
    setSelectedService(undefined);
    setStatus(undefined);
  };
  const handleOk = () => {
    if (selectedService) {
      const service = props.services?.find((service) => service.service_name === selectedService.value);
      if (service) {
        props.install(service);
        props.setIsModalOpen(false);
        setSelectedService(undefined);
      } else {
        setStatus('error');
      }
    } else {
      setStatus('error');
    }
  };
  const handleChange = (value: LabeledValue) => {
    setSelectedService(value);
    setStatus(undefined);
  };
  return (
    <Modal
      width={480}
      centered
      title="Install Services"
      open={props.isModalOpen}
      onOk={handleOk}
      onCancel={handleCancel}
      cancelButtonProps={{ style: { width: '50%', height: '56px' }, size: 'large' }}
      okButtonProps={{ style: { width: '50%', height: '56px' }, size: 'large' }}
      okText="Install"
      cancelText="Cancel"
      footer={(_, { OkBtn, CancelBtn }) => (
        <Flex>
          <CancelBtn />
          <OkBtn />
        </Flex>
      )}
    >
      <Select
        status={status}
        labelInValue
        optionRender={(option) => (
          <Flex
            gap={8}
            align="center"
          >
            {option.value === selectedService?.value && <CheckIcon />}
            {option.label}
          </Flex>
        )}
        style={{ width: '100%' }}
        size="large"
        value={selectedService}
        placeholder="AOG Services"
        onChange={handleChange}
        options={options}
      />
    </Modal>
  );
};
export default InstallService;
