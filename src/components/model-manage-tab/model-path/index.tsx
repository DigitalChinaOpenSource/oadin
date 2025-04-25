import React from 'react';
import { SettingOutlined } from '@ant-design/icons';
import { Input, Button } from 'antd';
import styles from './index.module.scss';

interface IModelPathProps {
  onModelPathChange: (val: string) => void;
  modelPathVal: string;
}

export default function ModelPath(props: IModelPathProps) {
	const {
		onModelPathChange,
		modelPathVal,
	} = props;

	return (	
		<div className={styles.modelPath}>
			<div className={styles.modelPathTitle}>模型存储路径</div>
			<Input
				className={styles.modelPathInput}
				size='small'
				value={modelPathVal}
				onChange={(e) => onModelPathChange(e.target.value)}
				addonAfter={<Button block type='text' icon={<SettingOutlined />} />}
			/>
		</div>
		
	)
}