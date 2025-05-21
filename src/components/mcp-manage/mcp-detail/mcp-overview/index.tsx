import styles from './index.module.scss';
import ReactMarkdown from 'react-markdown';
import testData from '../mcp_schema.json'

export default function McpOverview() {

  return <div className={styles.mcpOverview}>
     <ReactMarkdown >{testData.summary.zh}</ReactMarkdown>
  </div>;
}
