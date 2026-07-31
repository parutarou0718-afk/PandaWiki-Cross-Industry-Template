# OpenAI-compatible chat API

## Self-hosted API Token and user access

The self-hosted control console supports the existing knowledge-base API Token
screen without a commercial license. Open a knowledge base, go to its settings,
then use **API Token** to create, list, update, or revoke tokens. These actions
remain restricted to a signed-in member with `full_control` for that knowledge
base; API Tokens themselves can never administer users or other knowledge bases.

The system administration screen can also create normal `user` accounts. When
creating a normal user, assign that user to a knowledge base and choose
`full_control`, `doc_manage`, or `data_operate`. Account creation remains an
administrator-only action, and server-side knowledge-base and node permissions
continue to be enforced.

Upgrading this capability needs no database migration. Rebuild and recreate the
`api` service after pulling the new commit; do not remove PostgreSQL, NATS,
MinIO, or vector volumes.

管理员在目标知识库中启用“问答机器人 API”并生成 API Token 后，标准 OpenAI-compatible 客户端可直接调用；普通单知识库 Token 不需要 `X-KB-ID`：

```bash
curl http://<server-ip>:2444/share/v1/chat/completions \
  -H "Authorization: Bearer <API_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"model":"knowledge-base","messages":[{"role":"user","content":"总结知识库资料。"}],"stream":false}'
```

若同一个 Token 被明确授权给多个知识库，才需要增加 `X-KB-ID` 来选择其中一个已授权知识库。请勿将 Token 写入源代码、截图或日志。完整用法与错误说明见 [`docs/openai-compatible-chat-api.md`](../docs/openai-compatible-chat-api.md)。

# Ubuntu 全新部署指南

此目录部署的是本仓库的自托管无限版：旧的开源版、专业版、商业版、企业版功能门槛与数量配额均已取消；用户、知识库、用户组和节点权限仍然有效。

## 1. Ubuntu 前置条件

安装 Docker Engine、Docker Compose v2、Git、Node.js 22 与 pnpm 10。确认 Docker 可运行：

```bash
docker compose version
node --version
pnpm --version
```

## 2. 获取源码并配置秘密

```bash
git clone <你的仓库地址> panda-wiki
cd panda-wiki
git checkout feature/self-hosted-unlimited
cp deploy/.env.example deploy/.env
chmod 600 deploy/.env
```

编辑 `deploy/.env`，将每一个 `CHANGE_ME` 替换为独立的高强度随机值。尤其必须设置 `ADMIN_PASSWORD`；它会创建或修复正常的 `admin` 管理员账号。该文件不可提交或发送给他人。

如果主机已使用 `169.254.15.0/24`，修改 `SUBNET_PREFIX` 为未占用的私有网段前三段，例如 `172.30.15`。

## 3. 构建并启动

```bash
chmod +x deploy/scripts/*.sh
./deploy/scripts/build-images.sh
./deploy/scripts/up.sh
./deploy/scripts/healthcheck.sh
```

首次构建会下载 Go、Node 与容器依赖，时间较长。首次 API 启动会运行项目内置数据库迁移；不要在迁移期间关闭容器。

管理端地址：

```text
http://<服务器IP>:2444
```

使用账号 `admin` 和 `.env` 中的 `ADMIN_PASSWORD` 登录。随后应立即在管理端创建普通用户，并按知识库、用户组和节点权限分配访问范围。

## 4. 日常命令

```bash
cd deploy
docker compose ps
docker compose logs -f api
docker compose logs -f consumer
docker compose down
docker compose up -d
```

## 5. 升级本分支

```bash
cd /path/to/panda-wiki
git pull --ff-only
./deploy/scripts/build-images.sh
cd deploy
docker compose up -d --remove-orphans
./scripts/healthcheck.sh
```

升级前备份 `deploy/data/postgres`、`deploy/data/minio`、`deploy/data/qdrant` 与 `deploy/.env`。不要把这些文件放进 Git。

## 6. 完全重置（会永久删除全部数据）

仅确认不再需要任何知识库、文件、用户和向量数据时执行：

```bash
cd deploy
docker compose down
sudo rm -rf data
```

下次执行 `./scripts/up.sh` 会创建全新的空白系统。

## 7. 故障排查

若健康检查失败，先运行：

```bash
cd deploy
docker compose ps
docker compose logs --tail=200 api
docker compose logs --tail=200 raglite
docker compose logs --tail=200 postgres
```

不要在日志、截图或工单中粘贴 `.env` 内容、JWT、模型 API Key 或管理员密码。
