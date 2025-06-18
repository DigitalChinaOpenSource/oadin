import styles from './index.module.scss';
import ModelSearch from './model-search';
import ModallistContent from './model-list-content';
import { useViewModel, IUseSearchViewModelReturn } from './view-model';
import { IUseViewModel } from '@/components/model-manage-tab/model-list-content/view-model.ts';

export interface ModelManageProps {
  isDialog?: boolean;
  vmContent: IUseViewModel;
  vmSearch: IUseSearchViewModelReturn;
}
export default function ModelManageTab(props: ModelManageProps) {
  const { vmContent, isDialog, vmSearch } = props;
  return (
    <div className={styles.modelManageTab}>
      <ModelSearch
        total={vmContent.pagination.total ?? 0}
        modelSearchVal={vmSearch.modelSearchVal}
        modelSourceVal={vmSearch.modelSourceVal}
        onModelSearch={vmSearch.onModelSearch}
        onModelSourceChange={vmSearch.onModelSourceChange}
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
