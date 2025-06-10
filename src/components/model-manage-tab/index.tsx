import styles from './index.module.scss';
import ModelSearch from './model-search';
import ModallistContent from './model-list-content';
import { useViewModel } from './view-model';

export interface ModelManageProps {
  isMine?: boolean;
  isDialog?: boolean;
}
export default function ModelManageTab(props: ModelManageProps) {
  const vm = useViewModel();
  const { isMine, isDialog } = props;

  return (
    <div className={styles.modelManageTab}>
      <ModelSearch
        modelSearchVal={vm.modelSearchVal}
        modelSourceVal={vm.modelSourceVal}
        onModelSearch={vm.onModelSearch}
        onModelSourceChange={vm.onModelSourceChange}
      />
      {isDialog ? (
        <div className={styles.chooseModelList}>
          <ModallistContent
            mine={isMine}
            isSelectable={isDialog}
          />
        </div>
      ) : (
        <ModallistContent
          mine={isMine}
          isSelectable={isDialog}
        />
      )}
    </div>
  );
}
