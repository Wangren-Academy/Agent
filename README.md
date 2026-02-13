# AgentForge - 智能体协作编排与可视化平台

## 将AI的“黑箱思考”转化为白箱流程图

[项目概述](#项目概述) • [创新点](#创新点) • [快速开始](#快速开始) • [技术栈](#技术栈) • [贡献指南](#贡献指南)

---

## 项目概述

**AgentForge** 是一个**无内置AI**的智能体协作编排平台，专注于**连接外部AI服务**（OpenAI、Anthropic、本地模型等），通过可视化界面让开发者编排多个智能体的分工与依赖关系。系统负责**调度、通信、状态持久化**，并提供**业界领先的协作过程可视化沙盘**，使复杂的多智能体协作变得可观察、可调试、可重放。

## 创新点

### 智能体协作沙盘

1. **实时协作时序图** - 每个智能体的思考过程、工具调用、返回结果以动态时序图实时渲染
2. **调用链追踪与上下文快照** - 点击任意步骤查看完整Prompt、Messages、Token消耗、延迟等
3. **"What-if"沙盘重放** - 修改任意中间步骤输出，自动重新执行后续依赖节点，对比路径差异

## 快速开始

### 前置要求

- Docker & Docker Compose
- Go 1.22+ (开发时)
- Node.js 18+ (开发时)

### 一键启动

```bash
# 克隆仓库
git clone https://github.com/Wangren-Academy/Agent.git
cd Agent

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件，填入你的 OPENAI_API_KEY

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f
```

访问地址：

- 🌐 前端界面: [http://localhost:3000](http://localhost:3000)
- 🔌 后端 API: [http://localhost:8080](http://localhost:8080)
- 📊 API 文档: [http://localhost:8080/swagger](http://localhost:8080/swagger)

### 本地开发

```bash
# 后端
cd backend
go mod download
go run cmd/api/main.go

# 前端
cd frontend
npm install
npm run dev
```

## 技术栈

| 层次   | 技术                                                    |
| ------ | ------------------------------------------------------- |
| 后端   | Go 1.22+, Gin, gorilla/websocket, pgx, pgvector         |
| 前端   | Next.js 14+, TypeScript, React Flow, Zustand, shadcn/ui |
| 数据库 | PostgreSQL 15+, pgvector                                |
| 部署   | Docker, Docker Compose                                  |

## 项目结构

```text
agent-forge/
├── frontend/           # Next.js 前端应用
├── backend/            # Go 后端服务
├── docs/               # 详细文档
├── docker-compose.yml  # Docker 编排配置
└── README.md
```

## 贡献指南

我们欢迎所有形式的贡献！请查看 [CONTRIBUTING.md](./CONTRIBUTING.md) 了解详情。

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 许可证

本项目采用 Apache 2.0 许可证 - 详见 [LICENSE](./LICENSE) 文件。

---

Made with ❤️ by [Wangren Academy](https://github.com/Wangren-Academy)
