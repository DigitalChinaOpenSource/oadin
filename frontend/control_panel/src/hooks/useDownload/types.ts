//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

export interface IDownParseData {
  progress: number;
  status: string;
  completedsize: number;
  totalsize: number;
  message?: string;
  error?: string;
  data?: string;
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
