import styles from './index.module.scss';
import ModelSearch from './model-search';
import ModallistContent from './model-list-content';
import { useViewModel } from './view-model';

export default function ModelManageTab() {
  const vm = useViewModel();

  return (
    <div className={styles.modelManageTab}>
      <ModelSearch
        modelSearchVal={vm.modelSearchVal}
        modelSourceVal={vm.modelSourceVal}
        onModelSearch={vm.onModelSearch}
        onModelSourceChange={vm.onModelSourceChange}
      />

      <ModallistContent
        modelSearchVal={vm.modelSearchVal}
        modelSourceVal={vm.modelSourceVal}
      />
    </div>
  );
}
