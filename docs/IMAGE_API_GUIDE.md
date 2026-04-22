# GPT2API 生图接入指南

## 接口

```http
POST /v1/images/generations
```

## 鉴权

```http
Authorization: Bearer YOUR_API_KEY
Content-Type: application/json
```

## 请求参数

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `model` | string | 是 | 固定使用 `gpt-image-2` |
| `prompt` | string | 是 | 生图提示词，比例、构图、画幅直接写进提示词 |
| `n` | number | 否 | 生成张数，建议 `1` |
| `thinking` | boolean | 否 | 是否使用官网图片思考模式；不传或 `false` 表示普通生成 |
| `thinking_effort` | string | 否 | 思考强度，仅 `thinking=true` 时生效；`standard`=标准，`extended`=进阶 |
| `reference_images` | string[] | 否 | 图生图参考图，支持图片 URL、data URL 或 base64 |

不要传 `size`、`width`、`height`、`aspect_ratio` 等字段。比例和画幅统一写入 `prompt`，例如 `16:9 横版电影海报`、`9:16 竖版手机壁纸`、`1:1 方图构图`。

`thinking` 和 `thinking_effort` 都是可选字段。只传 `thinking=true` 且不传 `thinking_effort` 时，默认使用 `standard`。只传 `thinking_effort` 但不传 `thinking=true` 时，不启用思考模式。

## 文生图示例

```bash
curl http://你的服务器:端口/v1/images/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-image-2",
    "prompt": "生成一张 16:9 横版电影海报，未来城市雨夜，霓虹灯，赛博朋克风格，电影级光影，高细节",
    "n": 1
  }'
```

## 思考模式示例

标准思考：

```bash
curl http://你的服务器:端口/v1/images/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-image-2",
    "prompt": "生成一张小米和华为旗舰手机联动海报，21:9 横版，未来科技感，电影海报构图",
    "n": 1,
    "thinking": true
  }'
```

进阶思考：

```json
{
  "model": "gpt-image-2",
  "prompt": "生成一张小米和华为旗舰手机联动海报，21:9 横版，未来科技感，电影海报构图",
  "n": 1,
  "thinking": true,
  "thinking_effort": "extended"
}
```

## 图生图示例

使用图片 URL：

```bash
curl http://你的服务器:端口/v1/images/generations \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-image-2",
    "prompt": "参考输入图片，生成一张 16:9 横版电影海报，保留主体姿态，改成未来科幻城市背景，电影级光影",
    "n": 1,
    "reference_images": [
      "https://example.com/input.png"
    ]
  }'
```

使用 data URL：

```json
{
  "model": "gpt-image-2",
  "prompt": "参考输入图片，生成一张 9:16 竖版手机壁纸，保留主体，改成赛博朋克风格",
  "n": 1,
  "reference_images": [
    "data:image/png;base64,iVBORw0KGgoAAAANSUhEUg..."
  ]
}
```

## 返回格式

```json
{
  "created": 1770000000,
  "task_id": "img_xxxxxxxxxxxxxxxxxxxxxxxx",
  "data": [
    {
      "url": "http://你的服务器:端口/p/img/img_xxx/0?expires=...",
      "file_id": "file-xxx"
    }
  ]
}
```

客户端只需要读取并下载：

```text
data[0].url
```

## 失败处理

| code | 处理方式 |
| --- | --- |
| `no_available_account` | 稍后重试 |
| `rate_limited` | 延迟重试 |
| `content_policy` | 修改提示词 |
| `upstream_error` | 可重试一次 |
| `poll_timeout` | 可重试一次 |

建议客户端超时设置为 `5-8 分钟`。
