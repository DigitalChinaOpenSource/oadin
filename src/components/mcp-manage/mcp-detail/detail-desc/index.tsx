import Styles from './index.module.scss';
import TagsRender from '@/components/tags-render';
import { Tooltip } from 'antd';

export default function DetailDesc() {
  return (
    <div className={Styles.detailDescMain}>
      <div className={Styles.detailIcon}>
        <img
          src="http://120.232.136.73:31619/byzedev/model_avatar/qwen.png"
          alt="icon"
        />
      </div>
      <div className={Styles.detailContent}>
        <div className={Styles.detailTitle}>
          <div className={Styles.detailTitleName}>MCP 服务</div>
          <div className={Styles.tags}>
            <TagsRender
              tags={[
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
                '测试1',
                '测试2',
              ]}
            />
          </div>
        </div>

        <Tooltip title="测试">
          <div className={Styles.detailDesc}>
            这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内容这是内
          </div>
        </Tooltip>

        <div className={Styles.infoWrapper}>
          <div className={Styles.providerName}>深度求索</div>
          <div className={Styles.dot}>·</div>
          <div className={Styles.updateName}>2025-05-19 更新</div>
        </div>
      </div>
    </div>
  );
}
