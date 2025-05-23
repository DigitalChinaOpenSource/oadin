import { Checkbox } from 'antd';
import { useEffect, useState } from 'react';

export default function McpTagsCheck() {
  const testData = [
    {
      category: '图像处理',
      tags: ['增强', '压缩', '分割'],
    },
    {
      category: '文本分析',
      tags: ['情感分析', '关键词提取'],
    },
  ];

  const initData = testData.reduce((acc: Record<string, any>, item) => {
    acc[item.category] = [];
    return acc;
  }, {});

  const [checkedValues, setCheckedValues] = useState(initData);
  console.log('checkedValues', checkedValues);

  const handleChange = (category: string, list: any) => {
    setCheckedValues({
      ...checkedValues,
      [category]: list,
    });
  };

  return (
    <>
      {testData.map((item: Record<string, any>) => {
        return (
          <div
            key={item.category}
            style={{ marginBottom: '16px' }}
          >
            <h4>{item.category}</h4> {/* 添加标题 */}
            <Checkbox.Group
              options={item.tags}
              style={{ display: 'flex', flexDirection: 'column', gap: '8px', marginTop: '8px' }}
              value={checkedValues[item.category]}
              onChange={(list) => handleChange(item.category, list)} // 修复拼写错误
            />
          </div>
        );
      })}
    </>
  );
}
