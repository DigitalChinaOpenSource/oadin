import ChatContainer from '@/components/chat-container';
import ModelChecking from '@/components/model-checking';
import styles from './index.module.scss';
import useSelectedModelStore from '@/store/useSelectedModel';
export default function ModelChat() {
  const { selectedModel } = useSelectedModelStore();

  return <div className={styles.modelChat}>{selectedModel?.id ? <ChatContainer /> : <ModelChecking />}</div>;
}
