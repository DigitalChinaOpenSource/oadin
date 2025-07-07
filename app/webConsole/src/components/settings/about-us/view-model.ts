import { useState } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest.ts';
import favicon from '@/assets/favicon.png';

interface IAboutUsView {
  name: string;
  enName?: string;
  version: string;
  description: string;
  logo: string;
  officialWebsite: string;
  copyright: string;
}

const testData = {
  name: 'OADIN',
  enName: 'Open AIPC Development INfrastructure',
  version: '',
  description:
    '奥丁是基于AOG （AIPC OPEN GATEWAY）框架设计的创新式端侧能力平台，它通过智能化管理本地模型及云端模型、提供丰富的MCP工具生态和强大的服务监控功能，帮助开发者高效构建和优化Al应用，赋予开发者智慧的工具和无限的可能性',
  logo: favicon,
  officialWebsite: '',
  copyright:
    '© 2025 奥丁团队 版权所有。保留所有权利。  \n' +
    '\n' +
    '本产品由 **XXX科技有限公司** 开发与维护。使用本产品即表示您同意我们的 [《用户协议》](https://www.baize.ai/user-agreement)和 [《隐私政策》](https://www.baize.ai/privacy-policy)。  \n' +
    '\n' +
    '部分功能可能使用第三方开源组件，其版权归原作者所有，详见 [《开源组件声明》](https://www.baize.ai/open-source-notice)。  \n' +
    '\n' +
    '联系邮箱：support@baize.ai ｜ 官网：[https://www.baize.ai](https://www.baize.ai)',
};

export function useAboutUsView() {
  const [aboutDetails, setAboutDetails] = useState<IAboutUsView>();

  // 获取关于我们的详细信息
  const { loading: aboutUsLoading, run: fetchAboutUsDetail } = useRequest(
    async () => {
      return await httpRequest.get<IAboutUsView>('/system/about');
    },
    {
      manual: true,
      onSuccess: (data) => {
        setAboutDetails({ ...data, copyright: testData.copyright });
      },
      onError: (error) => {
        console.error('获取模型列表失败:', error);
        setAboutDetails(testData);
      },
    },
  );
  return { aboutUsLoading, aboutDetails, fetchAboutUsDetail };
}
