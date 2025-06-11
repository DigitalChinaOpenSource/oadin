import styles from './index.module.scss';
import ModelSearch from './model-search';
import ModallistContent from './model-list-content';
import { useViewModel } from './view-model';
import { IUseViewModel } from '@/components/model-manage-tab/model-list-content/view-model.ts';

export interface ModelManageProps {
  isDialog?: boolean;
  vmContent: IUseViewModel;
}
export default function ModelManageTab(props: ModelManageProps) {
  const vm = useViewModel();
  const { vmContent, isDialog } = props;
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
