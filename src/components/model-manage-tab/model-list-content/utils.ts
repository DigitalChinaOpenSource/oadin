import { RECOMMEND_MODEL, PIORITY_MODEL } from '@/constants';
import { ModelDataItem } from '@/types';
// 处理问学模型列表的数据
export const dealSmartVisionModels = (data: ModelDataItem[]) => {
  function removeDuplicates(arr: any) {
    return [...new Set(arr)];
  }

  const recommendedSeq = JSON.parse(JSON.stringify(RECOMMEND_MODEL));
  const pioritySeq = JSON.parse(JSON.stringify(PIORITY_MODEL));
  const seq = recommendedSeq.concat(pioritySeq);

  const map = data.reduce((acc: any, model) => {
    acc[model.name] = model;
    return acc;
  }, {});

  // 先用id进行排序，然后去重 返回排序的数据 并增加 isRecommended 字段
  const wholeSeq = seq.concat(data.map((item) => item.name));
  const dedup = removeDuplicates(wholeSeq);

  // 过滤掉 null 的数据
  return dedup
    .map((modelname: any) => {
      const model = map[modelname];
      if (model) {
        const { introduce, tags, ...rest } = model;
        const result = {
          ...rest,
          is_recommended: recommendedSeq.includes(modelname),
        };
        return result;
      }
      return null;
    })
    .filter(Boolean);
};
