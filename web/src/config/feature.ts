// 兼容旧导入路径。
//
// 真实开关已迁移到后端系统设置 `site.enable_chat_model`,前端请从 `useSiteStore()`
// 读取。这里保留一个静态回退值,避免尚未迁移的旧页面直接报错。
export const ENABLE_CHAT_MODEL = false
