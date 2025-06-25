import React, { useEffect, useRef, useState } from 'react';
import { Avatar, Button, Input } from 'antd';
import { PencilSimpleLineIcon, UserIcon, CheckIcon, CheckCircleIcon, WarningCircleIcon, BuildingOfficeIcon } from '@phosphor-icons/react';
import styles from './index.module.scss';
import UserNameModal from './userNameModal';
import { IAccountInfo } from '@/pages/UserCenter/types';
import DefaultUserIcon from '@/assets/userIcon.svg';
import CompanyIcon from '@/assets/companyIcon.svg';

/**
 * 用户中心账号信息组件
 */
const AccountInfo = ({ accountInfo }: { accountInfo: IAccountInfo }) => {
  const { userType, userName, companyName, email, phoneNumber, avatarUrl, wechatInfo, wechatBind } = accountInfo;
  const [userEditShow, setUserEditShow] = useState<boolean>(false);

  const isPerson = userType === 'person';

  const renderBindStatus = (isBound: boolean) => {
    return isBound ? (
      <>
        <CheckCircleIcon
          size={14}
          fill="#00CC39"
          weight={'fill'}
        />
        <span className={styles.bindingText}>已绑定</span>
      </>
    ) : (
      <>
        <WarningCircleIcon
          size={14}
          fill="#ff9300"
          weight={'fill'}
        />
        <span className={styles.noBindingText}>未绑定</span>
      </>
    );
  };

  const renderAccountTypeIcon = () => {
    return isPerson ? (
      <>
        <UserIcon
          size={16}
          weight={'fill'}
          fill={'#4F4DFF'}
        />
        <span className={styles.infoText}>个人账号</span>
      </>
    ) : (
      <>
        <BuildingOfficeIcon
          size={16}
          weight={'fill'}
          fill={'#4F4DFF'}
        />
        <span className={styles.infoText}>企业账号</span>
      </>
    );
  };

  return (
    <div className={styles.accountInfoContainer}>
      {/* 用户头像和名称部分 */}
      <div className={styles.headerSection}>
        <div className={styles.userInfoBlock}>
          <img
            src={isPerson ? avatarUrl || DefaultUserIcon : CompanyIcon}
            alt=""
          />

          {/* 用户名和编辑按钮 */}
          <div className={styles.userInfo}>
            <div className={styles.userNameBlock}>
              <span className={styles.userName}>{isPerson ? userName : companyName}</span>

              <Button
                type="text"
                className={styles.editButton}
                icon={<PencilSimpleLineIcon size={16} />}
                onClick={() => setUserEditShow(true)}
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
          <div className={styles.typeWithIcon}>{renderAccountTypeIcon()}</div>
        </div>
      </div>

      {/* 分割线 */}
      <div className={styles.divider}></div>

      {/* 手机号码 */}
      <div className={styles.infoRow}>
        <span className={styles.infoLabel}>手机号码</span>
        <div className={styles.infoContent}>
          <span className={styles.infoText}>{phoneNumber}</span>
          <div className={styles.bindingStatus}>
            {phoneNumber ? (
              <>
                <CheckCircleIcon
                  size={14}
                  fill="#00CC39"
                  weight={'fill'}
                />
                <span className={styles.bindingText}>已绑定</span>
              </>
            ) : (
              <>
                <WarningCircleIcon
                  size={14}
                  fill="#ff9300"
                  weight={'fill'}
                />
                <span className={styles.noBindingText}>未绑定</span>
              </>
            )}
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
            {wechatBind && <span className={styles.infoText}>{wechatInfo?.userName}</span>}
            <div className={styles.bindingStatus}>{renderBindStatus(wechatBind as boolean)}</div>
          </div>
        </div>

        <Button className={styles.changeButton}>{wechatBind ? '更换' : '绑定'}</Button>
      </div>
      <UserNameModal
        visible={userEditShow}
        title={isPerson ? '用户名修改' : '企业名称修改'}
        userType="person"
        onCancel={() => setUserEditShow(false)}
        onConfirm={() => setUserEditShow(false)}
      />
    </div>
  );
};

export default AccountInfo;
