import { IModelSelectCardItem } from '@/pages/AppManagement/AppConfig/types.ts';
import { IModelDataItem } from '@/types/model.ts';
import { Tag } from '@/pages/AppManagement/AppConfig/TagFilter/TagFilter.tsx';
import { defaultTag } from '@/pages/AppManagement/AppConfig/index.tsx';
import { message } from 'antd';

export const transformedAll2Card = (allList: IModelDataItem[]): IModelSelectCardItem[] => {
  return allList.map((item) => {
    return {
      id: item.id,
      name: item.name,
      avatar: item.avatar || 'https://byzer.ai/byzer-logo.png',
      class: item.class || [],
    };
  });
};

export const transformedAll2Ids = (allList: IModelDataItem[]): string[] => {
  return allList.map((item) => {
    return item.id;
  });
};
export const transformedIds2Card = (allList: IModelDataItem[], ids: string[]): IModelSelectCardItem[] => {
  return allList
    .filter((item) => {
      return ids.includes(item.id);
    })
    .map((item) => {
      return {
        id: item.id,
        name: item.name,
        avatar: item.avatar || 'https://byzer.ai/byzer-logo.png',
        class: item.class || [],
      };
    });
};
export const transformedCard2Ids = (cardList: IModelSelectCardItem[]): string[] => {
  return cardList.map((item) => {
    return item.id;
  });
};

export const transformedCard2Tags = (cardList: IModelSelectCardItem[]): Tag[] => {
  // 用于统计每个标签出现的次数
  const tagCount: Record<string, number> = {};

  // 遍历所有卡片，统计每个标签出现的次数
  cardList.forEach((card) => {
    if (card.class && Array.isArray(card.class)) {
      card.class.forEach((tag) => {
        if (tag) {
          tagCount[tag] = (tagCount[tag] || 0) + 1;
        }
      });
    }
  });

  // 将统计结果转换为 Tag 数组
  const tags: Tag[] = Object.keys(tagCount).map((label) => ({
    label,
    count: tagCount[label],
  }));

  // 按照出现次数降序排序，使常见标签排在前面
  const sortedTags = tags.sort((a, b) => (b.count || 0) - (a.count || 0));

  // 添加"全部"标签，计数为所有卡片的数量
  const allTag: Tag = {
    label: defaultTag,
    count: cardList.length,
  };

  // 将"全部"标签添加到列表最前面
  return sortedTags.length > 0 ? [allTag, ...sortedTags] : [];
};

/**
 * 检查MCP列表长度是否超过限制
 * @param mcpListLength MCP列表长度
 * @returns 如果长度小于等于4则返回true，否则返回false
 */
export const checkMcpLength = (mcpListLength: number): boolean => {
  if (mcpListLength > 4) {
    message.warning('为保障服务稳定运行与优质体验，建议您选择的MCP工具不要超过5个。');
    return false;
  }
  return true;
};
