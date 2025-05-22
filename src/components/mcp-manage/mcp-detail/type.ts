export type McpDetailType = {
  id: string | number;
  status:number;
  name: {
    src: string;
    zh: string;
  };
  serverName?:string;
  supplier: string;
  hosted:boolean;
  category?: string;
  tags: string[];
  abstract: {
    src: string;
    zh: string;
  };
  envSchema:Record<string, Record<string, any>>;
  logo: string;
  serverConfig:Record<string, any>[];
  summary:{
    src: string;
    zh: string;
  };
  tools?:Record<string, any>[];
  popularity?: number;
  createdAt?: number;
  updatedAt?: number;
}