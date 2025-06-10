import styles from './index.module.scss';
import ModelSearch from './model-search';
import ModallistContent from './model-list-content';
import { useViewModel } from './view-model';
import { useViewModel as useViewModelContent } from '@/components/model-manage-tab/model-list-content/view-model.ts';

export interface ModelManageProps {
  isMine?: boolean;
  isDialog?: boolean;
}
export default function ModelManageTab(props: ModelManageProps) {
  const vm = useViewModel();
  const { isMine, isDialog } = props;
  // 获取接口数据内容
  const vmContent = useViewModelContent({
    onModelSearch: vm.onModelSearch,
    modelSearchVal: vm.modelSearchVal,
    modelSourceVal: vm.modelSourceVal,
    mine: isMine,
  });
  return (
    <div className={styles.modelManageTab}>
      <ModelSearch
        total={vmContent.pagination.total ?? 0}
        modelSearchVal={vm.modelSearchVal}
        modelSourceVal={vm.modelSourceVal}
        onModelSearch={vm.onModelSearch}
        onModelSourceChange={vm.onModelSourceChange}
      />
      {isDialog ? (
        <div className={styles.chooseModelList}>
          <ModallistContent
            vm={vmContent}
            isSelectable={isDialog}
          />
        </div>
      ) : (
        <ModallistContent
          vm={vmContent}
          isSelectable={isDialog}
        />
      )}
    </div>
  );
}
