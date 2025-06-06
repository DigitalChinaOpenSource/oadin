import { Button, Upload, Tooltip } from 'antd';
import type { UploadProps } from 'antd';
import uploadSvg from '@/components/icons/upload.svg';
export default function UploadTool() {
  return (
    <Upload
      beforeUpload={(file, fileList) => {
        console.log('beforeUpload', file, fileList);
        return false;
      }}
      showUploadList={true}
    >
      <Tooltip title="文件格式支持 txt、HTML、Markdown、PDF、DOC、DOCX、XLS、XLSX、MP4、MP3、AVI、WMV。单个文件限制 50MB">
        <Button icon={<img src={uploadSvg} />} />
      </Tooltip>
    </Upload>
  );
}
