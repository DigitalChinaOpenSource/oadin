export interface IDownParseData {
  progress: number;
  status: string;
  completedsize: number;
  totalsize: number;
  message?: string;
  error?: string;
}

export interface IProgressData {
  status: string;
  digest?: string;
  completed?: number;
  total?: number;
  message?: string;
}

export interface IDownloadCallbacks {
  onmessage?: (data: IDownParseData | any) => void;
  onerror?: (error: Error) => void;
  onopen?: () => void;
  onclose?: () => void;
}
