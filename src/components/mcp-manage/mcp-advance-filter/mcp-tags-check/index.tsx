import { Checkbox } from 'antd';

export default function McpTagsCheck(props: any) {
  const { vm } = props;
  const { tagsData, checkedValues, handleTagsChange } = vm;
  return (
    <>
      {tagsData.map((item: Record<string, any>) => {
        return (
          <div
            key={item.category}
            style={{ marginBottom: '16px' }}
          >
            <h4>{item.category}</h4> {/* 添加标题 */}
            <Checkbox.Group
              options={item.tags.map((tag: Record<string, any>) => ({ ...tag, label: tag?.name, value: tag?.name }))}
              style={{ display: 'flex', flexDirection: 'column', gap: '8px', marginTop: '8px' }}
              value={checkedValues[item.category]}
              onChange={(list) => handleTagsChange(item.category, list)} // 修复拼写错误
            />
          </div>
        );
      })}
    </>
  );
}
