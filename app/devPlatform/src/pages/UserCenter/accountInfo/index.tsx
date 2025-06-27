import React, { useEffect, useRef, useState } from 'react';
import { Avatar, Button, Input } from 'antd';
import { PencilSimpleLineIcon, UserIcon, CheckIcon, CheckCircleIcon, WarningCircleIcon, BuildingOfficeIcon } from '@phosphor-icons/react';
import styles from './index.module.scss';
import UserNameModal from './userNameModal';
import { IAccountInfo, IAccountInfoProps } from '@/pages/UserCenter/types';
import DefaultUserIcon from '@/assets/userIcon.svg';
import CompanyIcon from '@/assets/companyIcon.svg';
import ChangePasswordModal from '@/pages/UserCenter/accountInfo/changePasswordModal';
import WechatBind from '@/pages/UserCenter/accountInfo/wechatBind';
import { useUserCenterView } from '@/pages/UserCenter/useUserCenterView.ts';

/**
 * 用户中心账号信息组件
 */
const AccountInfo = ({ accountInfo, setUserInfo }: IAccountInfoProps) => {
  const { type: userType, username, nickname, wechatName, enterpriseName, email, phone, avatar, wechatInfo, wechatBind } = accountInfo;

  const { changeUsername, userInfo } = useUserCenterView();
  // 修改用户名的弹窗
  const [userEditShow, setUserEditShow] = useState<boolean>(false);
  // 修改密码的弹框
  const [changePasswordShow, setChangePasswordShow] = useState<boolean>(false);
  // 绑定微信的弹窗
  const [wechatBindShow, setWechatBindShow] = useState<boolean>(false);
  // 是否是个人账户
  const isPerson = userType === 'person';

  // 渲染绑定状态
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

  // 渲染账号类型图标和文本
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

  // 公共渲染卡片
  const renderCommonRow = ({ label, content, action }: { label: string; content?: React.ReactNode; action?: React.ReactNode }) => {
    return (
      <div className={styles.commonRow}>
        <div className={styles.infoRow}>
          <span className={styles.infoLabel}>{label}</span>
          <div className={styles.infoContent}>{content}</div>
        </div>
        {action}
      </div>
    );
  };

  // 修改用户名
  const handleChangeUserName = async (newUserName: string) => {
    console.log('handleChangeUserName:', newUserName);
    const changeRes: any = await changeUsername(newUserName, accountInfo.phone);
    // TODO 需要根据实际接口修改
    if (changeRes) {
      setUserInfo({
        ...accountInfo,
        [isPerson ? 'userName' : 'enterpriseName']: newUserName,
      });
      setUserEditShow(false);
    }
  };

  return (
    <div className={styles.accountInfoContainer}>
      {/* 用户头像和名称部分 */}
      <div className={styles.headerSection}>
        <div className={styles.userInfoBlock}>
          <img
            src={isPerson ? avatar || DefaultUserIcon : CompanyIcon}
            alt=""
          />

          {/* 用户名和编辑按钮 */}
          <div className={styles.userInfo}>
            <div className={styles.userNameBlock}>
              <span className={styles.userName}>{isPerson ? nickname : enterpriseName}</span>

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
      {renderCommonRow({
        label: '手机号码',
        content: (
          <>
            <span className={styles.infoText}>{phone}</span>
            <div className={styles.bindingStatus}>{renderBindStatus(Boolean(phone))}</div>
          </>
        ),
        // action: <Button className={styles.changeButton}>{phoneNumber ? '更换' : '绑定'}</Button>,
      })}

      {/* 分割线 */}
      <div className={styles.divider}></div>

      {/* 个人- 微信号 */}
      {isPerson &&
        renderCommonRow({
          label: '微信号',
          content: (
            <>
              {wechatBind && <span className={styles.infoText}>{wechatName}</span>}
              <div className={styles.bindingStatus}>{renderBindStatus(wechatBind as boolean)}</div>
            </>
          ),
          action: !wechatBind && (
            <Button
              className={styles.changeButton}
              onClick={() => setWechatBindShow(true)}
            >
              {wechatBind ? '更换' : '绑定'}
            </Button>
          ),
        })}
      {/* 企业- 企业邮箱*/}
      {!isPerson &&
        // 企业邮箱
        renderCommonRow({
          label: '企业邮箱',
          content: (
            <>
              <span className={styles.infoText}>{email}</span>
              <div className={styles.bindingStatus}>{renderBindStatus(Boolean(email))}</div>
            </>
          ),
          // action: <Button className={styles.changeButton}>{email ? '更换' : '绑定'}</Button>,
        })}
      {/*企业 - 企业密码设置*/}
      {!isPerson && (
        <>
          {/*分割线*/}
          <div className={styles.divider}></div>
          {/*密码设置*/}
          {renderCommonRow({
            label: '密码设置',
            action: (
              <Button
                className={styles.changeButton}
                onClick={() => setChangePasswordShow(true)}
              >
                修改
              </Button>
            ),
          })}
        </>
      )}

      {/*修改用户名的弹框*/}
      <UserNameModal
        visible={userEditShow}
        title={isPerson ? '用户名修改' : '企业名称修改'}
        userType="person"
        onCancel={() => setUserEditShow(false)}
        onConfirm={(newUserName) => handleChangeUserName(newUserName)}
      />
      {/*修改密码的弹框*/}
      <ChangePasswordModal
        visible={changePasswordShow}
        onCancel={() => setChangePasswordShow(false)}
        onConfirm={() => setChangePasswordShow(false)}
      />
      {/*绑定微信的弹窗*/}
      <WechatBind
        visible={wechatBindShow}
        onCancel={() => setWechatBindShow(false)}
      />
    </div>
  );
};

export default AccountInfo;
