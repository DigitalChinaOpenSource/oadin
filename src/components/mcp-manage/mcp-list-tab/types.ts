export interface IMcpListRequestParams {
  id?: number[];
  /** 搜索关键词，模糊匹配 title 和 abstract */
  keyword?: string;
  /** 类别筛选 */
  category?: string[];
  /** 标签筛选，取交集 */
  tags?: string[];
  /** 部署类型："local" / "hosted" */
  deployment?: 'local' | 'hosted';
  /** 页码，默认 1 */
  page?: number;
  /**  每页条数，默认 10 */
  size?: number;
}

export interface IMcpListItem {
  id: number;
  name: {
    src: string;
    zh: string;
  };
  abstract: {
    src: string;
    zh: string;
  };
  supplier: string;
  logo: string;
  popularity: number;
  tags: string[];
  hosted: boolean;
  updatedAt: number;
}

export interface IMcpListData {
  data: {
    total: number;
    list: IMcpListItem[];
  };
}

export type cardType = {
  title: string;
  content: string;
  icon?: string;
  tags: string[];
  serviceId: string;
};
