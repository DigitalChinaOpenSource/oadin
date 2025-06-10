import styles from '../index.module.scss';
import ModelSearch from '../model-search';
import ModallistContent from '../model-list-content';
import { useViewModel } from '../view-model.ts';

export default function MyModelManageTab() {
  const vm = useViewModel();

  return (
    <div className={styles.modelManageTab}>
      <ModelSearch
        modelSearchVal={vm.modelSearchVal}
        modelSourceVal={vm.modelSourceVal}
        onModelSearch={vm.onModelSearch}
        onModelSourceChange={vm.onModelSourceChange}
      />

      <ModallistContent mine={true} />
    </div>
  );
}
