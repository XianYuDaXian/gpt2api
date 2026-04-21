<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import * as meApi from '@/api/me'
import { useSiteStore } from '@/stores/site'
import { formatCredit, formatDateTime, formatErrorCode } from '@/utils/format'
import { resolvePublicUrl } from '@/utils/url'

const siteStore = useSiteStore()
const apiBaseURL = computed(() => siteStore.apiBaseURL())

const loading = ref(false)
const cleanupLoading = ref(false)
const items = ref<meApi.ImageTask[]>([])
const total = ref(0)
const filter = reactive({
  limit: 20,
  offset: 0,
})

const page = computed<number>({
  get: () => Math.floor(filter.offset / filter.limit) + 1,
  set: (v) => {
    filter.offset = (v - 1) * filter.limit
    load()
  },
})

const statusMap: Record<string, { tag: 'success' | 'danger' | 'warning' | 'info'; label: string }> = {
  queued: { tag: 'info', label: '排队中' },
  dispatched: { tag: 'warning', label: '已调度' },
  running: { tag: 'warning', label: '生成中' },
  success: { tag: 'success', label: '成功' },
  failed: { tag: 'danger', label: '失败' },
}

function statusTag(s: string) {
  return statusMap[s]?.tag || 'info'
}

function statusLabel(s: string) {
  return statusMap[s]?.label || s || '-'
}

function imageURLs(task?: meApi.ImageTask | null) {
  return (task?.image_urls || []).map((u) => resolvePublicUrl(u, apiBaseURL.value))
}

function taskErrorText(task?: meApi.ImageTask | null) {
  if (!task) return ''
  return task.error_message || task.error || ''
}

async function load() {
  loading.value = true
  try {
    const d = await meApi.listMyImageTasks({
      limit: filter.limit,
      offset: filter.offset,
    })
    items.value = d.items || []
    total.value = d.total || 0
  } finally {
    loading.value = false
  }
}

function refresh() {
  filter.offset = 0
  load()
}

async function cleanupFailed() {
  cleanupLoading.value = true
  try {
    const res = await meApi.cleanupFailedImageTasks()
    ElMessage.success(`已清理 ${res.deleted || 0} 条失败记录`)
    refresh()
  } finally {
    cleanupLoading.value = false
  }
}

const detailVisible = ref(false)
const detailLoading = ref(false)
const current = ref<meApi.ImageTask | null>(null)

async function openDetail(task: meApi.ImageTask) {
  detailVisible.value = true
  detailLoading.value = true
  try {
    current.value = await meApi.getMyImageTask(task.task_id)
  } finally {
    detailLoading.value = false
  }
}

function copyText(text?: string) {
  const raw = text || ''
  if (!raw) return
  navigator.clipboard.writeText(raw)
    .then(() => ElMessage.success('已复制'))
    .catch(() => ElMessage.warning('复制失败'))
}

function openImage(url: string) {
  window.open(resolvePublicUrl(url, apiBaseURL.value), '_blank', 'noopener')
}

onMounted(load)
</script>

<template>
  <div class="page-container image-history">
    <div class="card-block hero">
      <div>
        <h2 class="page-title hero-title">图片记录</h2>
        <div class="page-sub">
          在线体验与 OpenAI 兼容生图接口都会写入这里。列表只展示当前账号自己的记录,详情页可查看 prompt、任务状态、上游会话与图片文件信息。
        </div>
      </div>
      <div class="hero-actions">
        <el-button type="danger" plain :loading="cleanupLoading" @click="cleanupFailed">
          清理失败记录
        </el-button>
        <el-button type="primary" :loading="loading" @click="refresh">
          <el-icon><Refresh /></el-icon> 刷新
        </el-button>
      </div>
    </div>

    <div class="card-block">
      <el-table
        v-loading="loading"
        :data="items"
        row-key="task_id"
        class="image-table"
        empty-text="暂无图片生成记录"
      >
        <el-table-column label="图片" width="170">
          <template #default="{ row }">
            <div v-if="imageURLs(row).length" class="thumbs">
              <el-image
                v-for="(url, idx) in imageURLs(row).slice(0, 3)"
                :key="url"
                :src="url"
                :preview-src-list="imageURLs(row)"
                :initial-index="idx"
                fit="cover"
                preview-teleported
                class="thumb"
              />
              <span v-if="imageURLs(row).length > 3" class="more">+{{ imageURLs(row).length - 3 }}</span>
            </div>
            <span v-else class="muted">未产出</span>
          </template>
        </el-table-column>
        <el-table-column label="任务" min-width="260">
          <template #default="{ row }">
            <div class="task-main">
              <span class="task-id">{{ row.task_id }}</span>
              <el-tooltip
                v-if="row.status === 'failed' && taskErrorText(row)"
                :content="taskErrorText(row)"
                placement="top"
              >
                <el-tag :type="statusTag(row.status)" size="small" effect="plain">
                  {{ statusLabel(row.status) }}
                </el-tag>
              </el-tooltip>
              <el-tag v-else :type="statusTag(row.status)" size="small" effect="plain">
                {{ statusLabel(row.status) }}
              </el-tag>
            </div>
            <div class="prompt" :title="row.prompt">{{ row.prompt }}</div>
            <div v-if="row.status === 'failed' && taskErrorText(row)" class="task-error">
              {{ taskErrorText(row) }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="规格" width="130">
          <template #default="{ row }">
            <div>{{ row.size || '-' }}</div>
            <div class="muted">{{ row.n || 0 }} 张</div>
          </template>
        </el-table-column>
        <el-table-column label="扣费" width="110">
          <template #default="{ row }">{{ formatCredit(row.credit_cost) }}</template>
        </el-table-column>
        <el-table-column label="创建时间" width="180">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pager">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="filter.limit"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          background
          @size-change="refresh"
        />
      </div>
    </div>

    <el-drawer v-model="detailVisible" title="图片任务详情" size="720px" class="detail-drawer">
      <div v-loading="detailLoading">
        <el-empty v-if="!current && !detailLoading" description="未找到记录" />
        <template v-else-if="current">
          <div class="detail-head">
            <div>
              <div class="detail-title">{{ current.task_id }}</div>
              <div class="muted">{{ formatDateTime(current.created_at) }}</div>
            </div>
            <el-tag :type="statusTag(current.status)" effect="plain">
              {{ statusLabel(current.status) }}
            </el-tag>
          </div>

          <div v-if="imageURLs(current).length" class="detail-images">
            <el-image
              v-for="(url, idx) in imageURLs(current)"
              :key="url"
              :src="url"
              :preview-src-list="imageURLs(current)"
              :initial-index="idx"
              fit="cover"
              preview-teleported
              class="detail-img"
            />
          </div>

          <el-descriptions :column="2" border class="detail-desc">
            <el-descriptions-item label="模型 ID">{{ current.model_id || '-' }}</el-descriptions-item>
            <el-descriptions-item label="上游账号 ID">{{ current.account_id || '-' }}</el-descriptions-item>
            <el-descriptions-item label="画面比例">{{ current.size || '-' }}</el-descriptions-item>
            <el-descriptions-item label="张数">{{ current.n || 0 }}</el-descriptions-item>
            <el-descriptions-item label="扣费">{{ formatCredit(current.credit_cost) }}</el-descriptions-item>
            <el-descriptions-item label="完成时间">{{ formatDateTime(current.finished_at) }}</el-descriptions-item>
            <el-descriptions-item label="上游会话" :span="2">
              <span class="mono">{{ current.conversation_id || '-' }}</span>
              <el-button
                v-if="current.conversation_id"
                type="primary"
                link
                @click="copyText(current.conversation_id)"
              >
                复制
              </el-button>
            </el-descriptions-item>
            <el-descriptions-item v-if="current.error" label="错误" :span="2">
              <el-tag type="danger" effect="plain">{{ formatErrorCode(current.error) }}</el-tag>
              <span class="error-raw">{{ current.error }}</span>
            </el-descriptions-item>
            <el-descriptions-item v-if="current.error_message" label="上游返回" :span="2">
              <div class="upstream-text">{{ current.error_message }}</div>
              <el-button
                size="small"
                class="copy-upstream"
                @click="copyText(current.error_message)"
              >
                复制文本
              </el-button>
            </el-descriptions-item>
          </el-descriptions>

          <div class="section">
            <div class="section-title">Prompt</div>
            <div class="prompt-box">{{ current.prompt || '-' }}</div>
            <el-button size="small" class="copy-prompt" @click="copyText(current.prompt)">复制 Prompt</el-button>
          </div>

          <div class="section">
            <div class="section-title">图片文件</div>
            <el-table :data="current.file_ids || []" size="small" border empty-text="暂无文件 ID">
              <el-table-column label="#" width="60">
                <template #default="{ $index }">{{ $index + 1 }}</template>
              </el-table-column>
              <el-table-column label="File ID">
                <template #default="{ row }">
                  <span class="mono">{{ row }}</span>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="160">
                <template #default="{ $index }">
                  <el-button
                    v-if="imageURLs(current)[$index]"
                    type="primary"
                    link
                    @click="openImage(imageURLs(current)[$index])"
                  >
                    打开图片
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </template>
      </div>
    </el-drawer>
  </div>
</template>

<style scoped lang="scss">
.hero {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
}
.hero-title { margin: 0 0 6px; }
.page-sub {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  line-height: 1.7;
}
.hero-actions {
  display: flex;
  gap: 10px;
  flex-shrink: 0;
}
.image-table {
  width: 100%;
}
.thumbs {
  display: flex;
  align-items: center;
  gap: 8px;
}
.thumb {
  width: 46px;
  height: 46px;
  border-radius: 10px;
  overflow: hidden;
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-fill-color-light);
}
.more {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.task-main {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}
.task-id,
.mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
}
.task-id {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.prompt {
  max-width: 560px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.task-error {
  max-width: 720px;
  margin-top: 4px;
  color: var(--el-color-danger);
  font-size: 12px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.muted {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.pager {
  display: flex;
  justify-content: flex-end;
  padding-top: 16px;
}
.detail-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}
.detail-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--el-text-color-primary);
}
.detail-images {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 12px;
  margin-bottom: 18px;
}
.detail-img {
  width: 100%;
  aspect-ratio: 1 / 1;
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid var(--el-border-color-lighter);
  background: var(--el-fill-color-light);
}
.detail-desc {
  margin-bottom: 18px;
}
.error-raw {
  margin-left: 8px;
  color: var(--el-text-color-secondary);
}
.upstream-text {
  white-space: pre-wrap;
  line-height: 1.7;
  color: var(--el-text-color-primary);
}
.copy-upstream {
  margin-top: 8px;
}
.section + .section {
  margin-top: 18px;
}
.section-title {
  font-weight: 700;
  margin-bottom: 8px;
}
.prompt-box {
  white-space: pre-wrap;
  line-height: 1.7;
  padding: 12px;
  border-radius: 10px;
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color-lighter);
}
.copy-prompt {
  margin-top: 8px;
}

@media (max-width: 720px) {
  .hero {
    align-items: flex-start;
    flex-direction: column;
  }
  .hero-actions {
    width: 100%;
  }
  .hero-actions .el-button {
    width: 100%;
  }
}
</style>
