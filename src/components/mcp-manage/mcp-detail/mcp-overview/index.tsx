import styles from './index.module.scss';
import ReactMarkdown from 'react-markdown';
type markDownDataType = {
  src?: string;
    zh: string;
    en?: string;
} | undefined

export default function McpOverview({ markDownData }  : { markDownData:markDownDataType } ) {

  return <div className={styles.mcpOverview}>
     <ReactMarkdown >{markDownData?.zh}</ReactMarkdown>
  </div>;
}
