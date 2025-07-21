import ChatContainer from '@/components/chat-container';
import ModelChecking from '@/components/model-checking';
import styles from './index.module.scss';
import useSelectedModelStore from '@/store/useSelectedModel';
export default function ModelChat() {
  const { selectedModel } = useSelectedModelStore();
  console.log('当前选中的模型', selectedModel);
  return <div className={styles.modelChat}>{selectedModel?.name ? <ChatContainer /> : <ModelChecking />}</div>;
}
