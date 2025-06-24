import React, { useEffect, useRef, useState } from 'react';
import { Avatar, Button, Input } from 'antd';
import { PencilSimpleLineIcon, UserIcon, CheckIcon, CheckCircleIcon, WarningCircleIcon } from '@phosphor-icons/react';
import styles from './index.module.scss';

/**
 * 用户中心账号信息组件
 */
const PersonInfo: React.FC = () => {
  const [userName, setUserName] = useState<string>('用户名');
  const [userEdit, setUserEdit] = useState<boolean>(false);
  const userRef: any = useRef();

  useEffect(() => {
    if (userEdit) {
      setTimeout(() => {
        userRef.current.focus();
      }, 10);
    }
  }, [userEdit]);

  return (
    <div className={styles.accountInfoContainer}>
      {/* 用户头像和名称部分 */}
      <div className={styles.headerSection}>
        <div className={styles.userInfoBlock}>
          {/* 头像 */}
          <Avatar
            size={48}
            icon={<UserIcon />}
          />

          {/* 用户名和编辑按钮 */}
          <div className={styles.userInfo}>
            <div className={styles.userNameBlock}>
              {userEdit ? (
                <Input
                  ref={userRef}
                  value={userName}
                  onChange={(e) => setUserName(e.target.value)}
                  onBlur={() => setUserEdit(false)}
                  onPressEnter={() => setUserEdit(false)}
                />
              ) : (
                <span className={styles.userName}>{userName}</span>
              )}

              <Button
                type="text"
                className={styles.editButton}
                icon={<PencilSimpleLineIcon size={16} />}
                onClick={() => {
                  console.log('userRef.current', userRef.current);

                  userRef.current?.focus();
                  setUserEdit(!userEdit);
                }}
              />
            </div>
          </div>
        </div>
      </div>

      {/* 分割线 */}
      <div className={styles.divider}></div>

      {/* 账号属性 */}
      <div className={styles.infoRow}>
        <div className={styles.accountTypeBlock}>
          <span className={styles.infoLabel}>账号属性</span>
          <div className={styles.typeWithIcon}>
            <UserIcon
              size={16}
              fill="#4F4DFF"
            />
            <span className={styles.infoText}>个人账号</span>
          </div>
        </div>
      </div>

      {/* 分割线 */}
      <div className={styles.divider}></div>

      {/* 手机号码 */}
      <div className={styles.infoRow}>
        <span className={styles.infoLabel}>手机号码</span>
        <div className={styles.infoContent}>
          <span className={styles.infoText}>123****1234</span>
          <div className={styles.bindingStatus}>
            <CheckCircleIcon
              size={14}
              fill="#00CC39"
            />
            <span className={styles.bindingText}>已绑定</span>
            <WarningCircleIcon
              size={14}
              fill="#ff9300"
            />
            <span className={styles.noBindingText}>未绑定</span>
          </div>
        </div>
      </div>

      {/* 分割线 */}
      <div className={styles.divider}></div>

      {/* 微信号 */}
      <div className={styles.wechatRow}>
        <div className={styles.infoRow}>
          <span className={styles.infoLabel}>微信号</span>
          <div className={styles.infoContent}>
            <span className={styles.infoText}>xxxxxxxxx</span>
            <div className={styles.bindingStatus}>
              <CheckIcon
                size={14}
                color="#00CC39"
              />
              <span className={styles.bindingText}>已绑定</span>
            </div>
          </div>
        </div>

        <Button className={styles.changeButton}>更换</Button>
      </div>
    </div>
  );
};

export default PersonInfo;
