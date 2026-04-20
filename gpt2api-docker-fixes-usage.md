# gpt2api Docker 部署补丁使用说明

## 补丁内容

这个补丁文件是：

```text
gpt2api-docker-fixes.patch
```

包含以下修复：

- `deploy/Dockerfile`：改为 Docker 多阶段构建，不再要求提前准备 `deploy/bin/gpt2api`、`deploy/bin/goose`、`web/dist`。
- `internal/account/refresher.go`：RT 刷 AT 改用 `https://auth0.openai.com/oauth/token`。
- `internal/account/importer_tokens.go`：RT 导入换 AT 同样改用 `https://auth0.openai.com/oauth/token`。
- 账号默认 `client_id`：改为 `pdlLIX2Y72MIl2rhLhTE9VV9bN905kBh`，避免 RT 刷新时报 `RT 已失效(401)`。
- `internal/gateway/images_proxy.go`：图片代理增加短重试，减少刚生成成功但预览空白的问题。
- `web/src/views/admin/Accounts.vue`：新建/编辑账号弹窗新增“抓包 JSON”粘贴解析，可自动填入 AT、RT、邮箱、`client_id` 等字段。

## 新服务器使用方式

先克隆原项目：

```bash
git clone https://github.com/432539/gpt2api.git
cd gpt2api
```

把 `gpt2api-docker-fixes.patch` 复制到这个目录，然后执行：

```bash
git apply gpt2api-docker-fixes.patch
```

如果想先检查补丁能否应用：

```bash
git apply --check gpt2api-docker-fixes.patch
```

## Docker 启动

进入部署目录：

```bash
cd deploy
cp .env.example .env
```

编辑 `.env`，至少修改：

```env
JWT_SECRET=至少32位随机字符串
CRYPTO_AES_KEY=64位hex字符串
MYSQL_ROOT_PASSWORD=强密码
MYSQL_PASSWORD=强密码
HTTP_PORT=18080
```

启动：

```bash
docker compose up -d --build
```

查看状态：

```bash
docker compose ps
docker compose logs -f server
```

访问：

```text
http://服务器IP:18080/
```

## 注意事项

- 不要把旧抓包文件里的 RT 反复导入；RT 刷新后可能轮换，旧 RT 会失效。
- 如果已有数据库里账号的 `client_id` 还是旧值，需要在后台编辑账号改成 `pdlLIX2Y72MIl2rhLhTE9VV9bN905kBh`，或重新用这个 `client_id` 导入最新 RT。
- 新建账号时可以直接点“新建账号”，把抓包 JSON 粘贴到“抓包 JSON”，再点“解析并填入”。
- 如果服务重启，旧的 `/p/img/...sig=...` 图片链接会失效，需要重新从任务接口或页面刷新获取新链接。
- 如果服务器使用代理，后台代理地址不要写 `127.0.0.1`，除非代理也运行在同一个容器内；通常应填写宿主机或代理服务器的局域网 IP。
