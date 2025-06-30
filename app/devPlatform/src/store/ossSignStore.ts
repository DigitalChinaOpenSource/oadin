//  前端不用获取阿里直传签名了，由后台上传

// import { create } from 'zustand';
// import { persist } from 'zustand/middleware';
// import { httpRequest } from '@/utils/httpRequest.ts';
// import { getUserToken } from '@/utils/getUserToken.ts';
//
// export interface IOssSignProps {
//   accessKeyId: string;
//   accessKeySecret: string;
//   securityToken: string; // stsToken
//   bucket: string;
//   region: string;
//   dir: string;
//   expire: number; //过期时间
//   expiration: string; //过期时间
// }
//
// interface IOssSignStore {
//   ossSign: IOssSignProps | null;
//   setOssSign: (sign: IOssSignProps) => void;
//   getOssSign: () => Promise<IOssSignProps | null>;
// }
//
// export const useOssSignStore = create<IOssSignStore>()(
//   persist(
//     (set) => ({
//       ossSign: null,
//       setOssSign: (sign: IOssSignProps) => set({ ossSign: sign }),
//       getOssSign: () => getSign(), // 获取签名的方法
//     }),
//     {
//       name: 'oss-sign-storage', // 持久化存储的键名
//       partialize: (state) => ({ ossSign: state.ossSign }), // 只持久化ossSign字段
//     },
//   ),
// );
//
// export const getSign = async (): Promise<IOssSignProps | null> => {
//   const currentStore = useOssSignStore.getState();
//   const currentOssSign = currentStore.ossSign;
//   // 检查是否有缓存的签名且未过期
//   if (currentOssSign) {
//     const nowTimestamp = Math.floor(Date.now() / 1000);
//     if (currentOssSign.expire > nowTimestamp) {
//       // 签名未过期，直接返回缓存的签名
//       return currentOssSign;
//     }
//   }
//
//   // 签名不存在或已过期，重新获取
//   try {
//     const res = await httpRequest.get('/enterprise/oss-sts', null, { headers: { Authorization: getUserToken() } });
//     if (res.data) {
//       currentStore.setOssSign(res.data);
//       return res.data;
//     } else {
//       console.error('获取OSS签名失败:', res.message);
//       return null;
//     }
//   } catch (error) {
//     console.error('获取OSS签名出错:', error);
//     return null;
//   }
// };
