/** 智能模型切换 */
import React from 'react';
import { Card, Switch } from 'antd';
import styles from './index.module.scss';

export default function ModelChangeTab() {
  return (
    <div className={styles.modelChangeTab}>
      <Card className={styles.card}>
        <div className={styles.title}>智能模型切换</div>
        <div className={styles.content}>
          <div>
            为了确保您在使用应用时获得最佳体验,系统会根据您设备的资源情况(如内存、CPU、GPU等)自动选择使用本地模型或云端模型。
          </div>
          <div>
            1.
            当您的本地资源紧张(比如内存、磁盘空间或GPU显存占用过高时),系统会自动将任务转移到云端模型上,确保您仍然能够流畅使通。
          </div>
          <div>
            2.
            当本地资源恢复到正常水平,系统会自动切换回本地模型,让您的操作更加高效、稳定。
          </div>
        </div>
        <div className={styles.modelSwitch}>
          <Switch
            checkedChildren="开启"
            unCheckedChildren="关闭"
            defaultChecked
          />
        </div>
      </Card>
    </div>
  );
}
