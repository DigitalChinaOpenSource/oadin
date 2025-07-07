import mitt from 'mitt';

/** 词嵌入模型下载完成后 */
const embedDownloadEventBus = mitt();

export default embedDownloadEventBus;
