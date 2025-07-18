import styles from './index.module.scss';
import ModelSearch from './model-search';
import ModallistContent from './model-list-content';
import { IUseSearchViewModelReturn } from './view-model';
import { IUseViewModel } from '@/components/model-manage-tab/model-list-content/view-model.ts';
import { ISelectedDialogProps } from '@/components/choose-model-dialog';

export interface ModelManageProps extends ISelectedDialogProps {
  vmContent: IUseViewModel;
  vmSearch: IUseSearchViewModelReturn;
  currentTab?: string;
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
            {...props}
            vmContent={vmContent}
            isSelectable={isDialog}
          />
        </div>
      ) : (
        <ModallistContent
          {...props}
          vmContent={vmContent}
          isSelectable={isDialog}
        />
      )}
    </div>
  );
}
