import { useState } from "react";

interface IUserInfo {
  id: string;
  username: string;
  // 账号类型，个人账号或企业账号
  accountType: 'person' | 'enterprise';
  
  email: string;
  phoneNumber: string;
  wechatInfo:Record<string, any>;
  wechatBind: boolean; // 是否绑定微信
  isHaveAuth: boolean;
  authInfo: Record<string, any>;
}

export const useUserCenterView = () => {

    const [userInfo, setUserInfo] = useState<IUserInfo | null>(null);

    // 获取用户信息
    const getUserInfo = async () => {
        await new Promise((resolve) => {setTimeout(resolve, 2000)});
        return {
            id: 'user-123',
            username: 'testuser',
            email: '',
            phoneNumber: '1234567890',
            accountType: 'person', // 个人账号
            wechatBind:false,
            wechatInfo: {
                openId: 'wx-123456',
                unionId: 'wx-union-123456',
                nickname: 'Test WeChat User',
                avatarUrl: 'https://example.com/avatar.jpg',
            },
            isHaveAuth: false,
            authInfo: {}
        } 
    }

    // 修改用户名
    const changeUsername = async (newUsername: string,phoneNumber:string,email?:string) => {
        // 模拟 API 调用
        await new Promise((resolve) => setTimeout(resolve, 1000));
        
        // 更新用户信息
        setUserInfo((prev) => ({
            ...prev,
            username: newUsername
        } as IUserInfo));
    }

    return {
        userInfo,
        getUserInfo,
        changeUsername
    }
}