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
import { Input } from 'antd';
import type { InputProps } from 'antd/lib/input';
import type { PasswordProps } from 'antd/lib/input/Password';
import styles from './index.module.scss';

export interface FloatInputProps extends InputProps, PasswordProps {
  placeholder: string;
}

const FloatInput: React.FC<FloatInputProps> = ({ value = '', onChange, placeholder, type = 'text', style, className, disabled, ...rest }) => {
  const [focused, setFocused] = useState(false);

  const InputComponent = type === 'password' ? Input.Password : Input;

  return (
    <div
      className={styles.floatInputWrap + (className ? ' ' + className : '')}
      style={{ ...style, borderColor: focused ? 'var(--color-primary)' : '#808080', borderWidth: focused ? '2px' : '1px', borderStyle: 'solid' }}
    >
      <InputComponent
        value={value}
        onChange={onChange}
        onFocus={() => setFocused(true)}
        onBlur={() => setFocused(false)}
        type={type}
        disabled={disabled}
        {...rest}
      />
      <label className={styles.floatingLabel + (focused || value ? ' ' + styles.floating : '')}>{placeholder}</label>
    </div>
  );
};

export default FloatInput;
