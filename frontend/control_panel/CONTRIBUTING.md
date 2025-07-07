## 开发规范

### 代码风格

- 本项目使用 ESLint 和 Prettier 进行代码格式化和检查
- 提交代码前请确保通过 lint 检查： yarn lint

### 提交规范

- 请遵守 Git 提交规范
- 请遵守 Git 分支管理规范
- 请遵守 Git 提交信息规范

- `<type>` 是提交的类型，包括 `feat`（新增功能）、`fix`（修复问题）、`docs`（文档更新）等
- `<subject>` 是提交的简要描述
- 示例：

```
feat: add new feature
fix: fix bug in login page
docs: update README.md
```

### 分支管理

- 分支命名应遵循以下格式：

- `<type>` 是分支的类型，包括 `feat`（新增功能）、`fix`（修复问题）、`docs`（文档更新）等
- `<subject>` 是分支的简要描述
- 示例：

```
feature/add-new-feature
fix/fix-bug-in-login-page
docs/update-readme
```

### 其他

工程内如需使用 icon，均从 https://phosphoricons.com 拷贝至本地使用，参考 src/components/icons，使用前请先查看是否有同名 icon，如有则使用已有 icon

## 项目结构

vanta/
├── src/ # 源代码目录
│ ├── assets/ # 静态资源
│ ├── components/ # 组件
│ ├── constants/ # 常量定义
│ ├── hooks/ # 自定义 React Hooks
│ ├── pages/ # 页面组件
│ ├── routes/ # 路由配置
│ ├── store/ # 状态管理
│ ├── styles/ # 全局样式
│ ├── types/ # TypeScript 类型定义
│ ├── utils/ # 工具函数
│ ├── App.tsx # 应用入口组件
│ └── main.tsx # 应用入口文件
├── public/ # 公共资源
└── ...配置文件

```

```
