import { RECOMMEND_MODEL, PIORITY_MODEL } from '@/constants';
import { IModelDataItem } from '@/types';
// 处理问学模型列表的数据
export const dealSmartVisionModels = (data: IModelDataItem[]) => {
  function removeDuplicates(arr: string[]) {
    return [...new Set(arr)];
  }

  const recommendedSeq = JSON.parse(JSON.stringify(RECOMMEND_MODEL));
  const pioritySeq = JSON.parse(JSON.stringify(PIORITY_MODEL));
  const seq = recommendedSeq.concat(pioritySeq);

  console.log('dealSmartVisionModels', data);
  const map = data.reduce((acc: { [key: string]: IModelDataItem }, model) => {
    acc[model.name] = model;
    return acc;
  }, {});

  // 先用id进行排序，然后去重 返回排序的数据 并增加 isRecommended 字段
  const wholeSeq = seq.concat(data.map((item) => item.name));
  const dedup = removeDuplicates(wholeSeq);

  // 过滤掉 null 的数据
  return dedup
    .map((modelname: string) => {
      const model = map[modelname];
      if (model) {
        const result = {
          ...model,
          is_recommended: recommendedSeq.includes(modelname),
        };
        return result;
      }
      return null;
    })
    .filter(Boolean);
};
