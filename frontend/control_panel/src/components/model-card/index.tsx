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

import React from 'react';
import styles from './index.module.scss';
import { NavLink, useLocation } from 'react-router-dom';
import { Checkbox } from 'antd';
import type { CheckboxProps } from 'antd';
import { formatContentLength } from '@/utils/index';
export interface ModelItem {
  id: string;
  service_name: string;
  api_flavor: string;
  flavor: string;
  method: string;
  desc: string;
  url: string;
  auth_type: string;
  auth_apply_url: string;
  auth_fields: string[];
  name: string;
  service_provider_name: string;
  size: string;
  is_recommended: boolean;
  status: string;
  avatar: string;
  can_select: boolean;
  class: string[];
  ollama_id: string;
  params_size: number;
  input_length: number;
  output_length: number;
  source: string;
  smartvision_provider: string;
  smartvision_model_key: string;
  is_default: string;
}
interface ModelCardProps {
  model: ModelItem;
  showSet?: boolean;
  setDefault?: () => void;
  loading?: boolean;
}

const ModelCard: React.FC<ModelCardProps> = ({ model, showSet, setDefault, loading }) => {
  const { avatar, name, can_select, desc, is_default, flavor, size, source, input_length, output_length } = model;
  const location = useLocation();
  const onChange: CheckboxProps['onChange'] = (e) => {
    setDefault?.();
  };
  const status = can_select ? (source === 'local' ? 'Downloaded' : 'Authorized') : source === 'local' ? 'Undownloaded' : 'Unauthorized ';
  return (
    <div className={styles.card}>
      <div className={styles.header}>
        <div className={styles.headerLeft}>
          <div className={styles.avatarWrap}>
            <img
              className={styles.avatar}
              src={avatar}
              alt={name}
            />
          </div>
          {showSet ? (
            <span className={styles.name}> {name}</span>
          ) : (
            <NavLink
              to={`${location.pathname}/${encodeURIComponent(name)}`}
              className={`${styles.name} ${styles.link}`}
            >
              {name}
            </NavLink>
          )}

          {is_default === 'true' && <span className="defaultBtn">Default</span>}
        </div>
        {showSet && can_select && (
          <div className={styles.headerRight}>
            <Checkbox
              defaultChecked={is_default === 'true'}
              onChange={onChange}
            >
              Set as default model
            </Checkbox>
          </div>
        )}
      </div>
      <div className={styles.infoRow}>
        <span className={styles.vendor}>{flavor}</span>
        <span className={styles.divider} />
        <span className={styles.context}>{`Context length: ${input_length + output_length === 0 ? 'None' : formatContentLength(input_length + output_length)}`}</span>
        <span className={styles.divider} />
        <span className={styles.status}>{status}</span>
      </div>
      <div className={styles.desc}>{desc}</div>
    </div>
  );
};

export default ModelCard;
