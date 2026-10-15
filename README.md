# 个人简历动态更新系统

一个面向个人的简历管理工具：上传/预览简历、维护 JD、根据 JD 动态优化简历、简历评价、面试备战建议，以及多格式导出。

## 功能

1. **简历管理**：上传并解析 `docx / md / pdf / txt`，结构化预览与在线编辑。
2. **JD 管理**：维护职位描述（职责、要求、加分项）。
3. **简历优化**：选中 JD 后动态调整简历，展示命中关键词与调整说明，可对比差异。
4. **简历评价**：针对不同 JD 打分并给出改进意见。
5. **面试备战**：输入 JD 后生成面试问题、复习知识点与项目准备建议。
6. **导出**：支持 `docx / md / pdf / txt`。
7. **系统设置**：可配置大模型（任意 OpenAI 兼容接口），未配置时自动使用内置规则引擎。

## 技术栈

- 后端：Go 1.24（`net/http` + `pgx`）
- 数据库：PostgreSQL
- 前端：React 18 + Vite + TypeScript + Tailwind CSS

## 目录结构

```
resume/
├── backend/          # Go 后端（REST API）
│   ├── main.go       # 入口
│   ├── db.go         # 数据库连接与建表
│   ├── handlers*.go  # HTTP 处理
│   ├── ai.go         # 规则引擎（优化/评价/备战）
│   ├── llm.go        # 大模型调用（OpenAI 兼容）
│   ├── parse.go      # 文件解析（docx/pdf/md/txt）
│   └── export.go     # 导出（docx/pdf/md/txt）
└── frontend/         # React 前端
```

## 运行

### 1. 准备 PostgreSQL

```bash
createdb resume   # 或自行创建数据库
```

默认连接串：`postgres://gaolei@localhost:5432/resume?sslmode=disable`
可通过环境变量 `DATABASE_URL` 覆盖。服务启动时自动建表。

### 2. 启动后端

```bash
cd backend
go run .
# 监听 http://localhost:4000
```

> 若依赖下载缓慢，可设置 `GOPROXY=https://goproxy.cn,direct`。

### 3. 启动前端

```bash
cd frontend
npm install
npm run dev
# 监听 http://localhost:5173，/api 代理到 4000
```

打开 http://localhost:5173 即可使用。

## 大模型配置

进入「系统设置」页，填入：

- 接口地址（OpenAI 兼容），如 `https://api.deepseek.com`
- 模型名称，如 `deepseek-chat`
- API Key

保存后即可启用大模型。未启用或调用失败时，自动回退到内置规则引擎。

## PDF 导出字体

PDF 导出需要单个 `.ttf` 中文字体。默认按以下顺序自动检测：

- `PDF_FONT_PATH` 环境变量
- macOS：`/System/Library/Fonts/Supplemental/Arial Unicode.ttf`
- Linux：常见 wqy / Noto 字体路径

也可在「系统设置」页手动指定字体路径。

