<script setup>
import { computed, inject, onMounted, reactive, ref } from "vue";
import Cookies from "js-cookie";
import { Post } from "@/js/fetch.js";
import { Notification } from "@/js/util.js";

const title = inject("title");

const keys = ref([]);
const loading = ref(false);
const creating = ref(false);
const selectedKey = ref(null);
const createdKey = ref("");
const showResult = ref(false);
const query = ref("");
const statusFilter = ref("all");
const chatExample = `curl -X POST http://localhost:2233/v1/chat/completions \\
  -H "Authorization: Bearer sk-替换为刚创建的Key" \\
  -H "Content-Type: application/json" \\
  -H "X-Provider: openai" \\
  -d '{
    "model": "gpt-5.5",
    "messages": [
      {"role": "user", "content": "写一句欢迎语"}
    ]
  }'`;

const form = reactive({
  name: "",
  description: "",
  role: "user",
  quota_day: 0,
  quota_month: 0,
  allowed_ips: "",
  expires_at: "",
});

const editForm = reactive({
  name: "",
  description: "",
  role: "user",
  quota_day: 0,
  quota_month: 0,
  allowed_ips: "",
  expires_at: "",
  enabled: true,
});

const stats = computed(() => {
  const total = keys.value.length;
  const active = keys.value.filter((key) => key.enabled).length;
  const expired = keys.value.filter((key) => isExpired(key.expires_at)).length;
  const limited = keys.value.filter((key) => key.quota_day > 0 || key.quota_month > 0).length;
  return { total, active, disabled: total - active, expired, limited };
});

const filteredKeys = computed(() => {
  const keyword = query.value.trim().toLowerCase();
  return keys.value.filter((key) => {
    const matchesKeyword = !keyword || [key.name, key.description, key.role]
      .some((value) => String(value || "").toLowerCase().includes(keyword));
    const matchesStatus =
      statusFilter.value === "all" ||
      (statusFilter.value === "enabled" && key.enabled) ||
      (statusFilter.value === "disabled" && !key.enabled) ||
      (statusFilter.value === "expired" && isExpired(key.expires_at));
    return matchesKeyword && matchesStatus;
  });
});

onMounted(() => {
  title.value = "API Key 管理";
  loadKeys();
});

async function loadKeys() {
  loading.value = true;
  try {
    const resp = await Post({
      Token: Cookies.get("token"),
      Send: { operation: "fetchAPIKeys" },
    });
    if (resp?.error) {
      Notification(resp.error);
      return;
    }
    keys.value = resp?.setting_body?.keys || resp?.keys || [];
    if (selectedKey.value) {
      selectedKey.value = keys.value.find((key) => key.id === selectedKey.value.id) || null;
      if (selectedKey.value) fillEditForm(selectedKey.value);
    }
  } catch (err) {
    Notification("获取 API Keys 失败: " + err.message);
  } finally {
    loading.value = false;
  }
}

async function submitCreate() {
  if (!form.name.trim()) {
    Notification("请输入 API Key 名称");
    return;
  }
  if (Number(form.quota_day) < 0 || Number(form.quota_month) < 0) {
    Notification("配额不能为负数");
    return;
  }

  creating.value = true;
  try {
    const resp = await Post({
      Token: Cookies.get("token"),
      Send: {
        operation: "createAPIKey",
        setting_body: {
          name: form.name.trim(),
          description: form.description.trim(),
          role: form.role,
          quota_day: Number(form.quota_day || 0),
          quota_month: Number(form.quota_month || 0),
          allowed_ips: parseList(form.allowed_ips),
          expires_at: toRFC3339(form.expires_at),
        },
      },
    });
    if (resp?.error) {
      Notification(resp.error);
      return;
    }
    createdKey.value = resp?.setting_body?.api_key || resp?.api_key || "";
    showResult.value = true;
    resetForm();
    await loadKeys();
  } catch (err) {
    Notification("创建失败: " + err.message);
  } finally {
    creating.value = false;
  }
}

async function saveSelectedKey() {
  if (!selectedKey.value) return;
  if (!editForm.name.trim()) {
    Notification("名称不能为空");
    return;
  }
  const data = {
    name: editForm.name.trim(),
    description: editForm.description.trim(),
    role: editForm.role,
    quota_day: Number(editForm.quota_day || 0),
    quota_month: Number(editForm.quota_month || 0),
    allowed_ips: JSON.stringify(parseList(editForm.allowed_ips)),
    expires_at: toRFC3339(editForm.expires_at),
    enabled: editForm.enabled,
  };
  if (!data.expires_at) delete data.expires_at;

  try {
    const resp = await Post({
      Token: Cookies.get("token"),
      Send: {
        operation: "updateAPIKey",
        setting_body: {
          api_key_id: selectedKey.value.id,
          api_key_data: data,
        },
      },
    });
    if (resp?.error) {
      Notification(resp.error);
      return;
    }
    Notification("已保存");
    await loadKeys();
  } catch (err) {
    Notification("保存失败: " + err.message);
  }
}

async function toggleKey(key) {
  await updateKey(key.id, { enabled: !key.enabled }, key.enabled ? "已禁用" : "已启用");
}

async function deleteKey(key) {
  if (!confirm(`确定删除 "${key.name || "未命名"}" 吗？此操作不可撤销。`)) return;
  try {
    const resp = await Post({
      Token: Cookies.get("token"),
      Send: {
        operation: "deleteAPIKey",
        setting_body: { api_key_id: key.id },
      },
    });
    if (resp?.error) {
      Notification(resp.error);
      return;
    }
    Notification("已删除");
    if (selectedKey.value?.id === key.id) selectedKey.value = null;
    await loadKeys();
  } catch (err) {
    Notification("删除失败: " + err.message);
  }
}

async function updateKey(id, data, message) {
  try {
    const resp = await Post({
      Token: Cookies.get("token"),
      Send: {
        operation: "updateAPIKey",
        setting_body: {
          api_key_id: id,
          api_key_data: data,
        },
      },
    });
    if (resp?.error) {
      Notification(resp.error);
      return;
    }
    Notification(message);
    await loadKeys();
  } catch (err) {
    Notification("操作失败: " + err.message);
  }
}

function selectKey(key) {
  selectedKey.value = key;
  fillEditForm(key);
}

function fillEditForm(key) {
  Object.assign(editForm, {
    name: key.name || "",
    description: key.description || "",
    role: key.role || "user",
    quota_day: key.quota_day || 0,
    quota_month: key.quota_month || 0,
    allowed_ips: formatIPList(key.allowed_ips),
    expires_at: toLocalInput(key.expires_at),
    enabled: key.enabled !== false,
  });
}

function resetForm() {
  Object.assign(form, {
    name: "",
    description: "",
    role: "user",
    quota_day: 0,
    quota_month: 0,
    allowed_ips: "",
    expires_at: "",
  });
}

function parseList(value) {
  return String(value || "")
    .split(/[\n,]/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function formatIPList(value) {
  if (!value) return "";
  if (Array.isArray(value)) return value.join(", ");
  try {
    const parsed = JSON.parse(value);
    return Array.isArray(parsed) ? parsed.join(", ") : String(value);
  } catch {
    return String(value);
  }
}

function toRFC3339(value) {
  if (!value) return "";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toISOString();
}

function toLocalInput(value) {
  if (!value || isZeroDate(value)) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  const offset = date.getTimezoneOffset() * 60000;
  return new Date(date.getTime() - offset).toISOString().slice(0, 16);
}

function isZeroDate(value) {
  return !value || String(value).startsWith("0001-01-01");
}

function isExpired(value) {
  if (isZeroDate(value)) return false;
  const date = new Date(value);
  return !Number.isNaN(date.getTime()) && date.getTime() < Date.now();
}

function formatDate(value) {
  if (isZeroDate(value)) return "永不过期";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString("zh-CN");
}

function shortDate(value) {
  if (isZeroDate(value)) return "永不过期";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleDateString("zh-CN");
}

function usagePercent(usage, quota) {
  if (!quota || quota <= 0) return 0;
  return Math.min(100, Math.round((usage / quota) * 100));
}

function copyKey() {
  navigator.clipboard.writeText(createdKey.value).then(() => Notification("已复制到剪贴板"));
}

function copyExample() {
  navigator.clipboard.writeText(chatExample).then(() => Notification("示例已复制到剪贴板"));
}
</script>

<template>
  <mdui-layout-main class="apikey-page">
    <div class="page-content">
      <header class="page-header">
        <div>
          <h1>API Key 管理</h1>
          <p>创建、限制和停用访问密钥。配额为 0 时表示不限制，IP 留空表示允许所有来源。</p>
        </div>
        <mdui-button class="inline-button" variant="filled" icon="refresh" @click="loadKeys">刷新</mdui-button>
      </header>

      <section class="metric-grid">
        <div class="metric-card">
          <span>总数</span>
          <strong>{{ stats.total }}</strong>
        </div>
        <div class="metric-card">
          <span>启用</span>
          <strong>{{ stats.active }}</strong>
        </div>
        <div class="metric-card">
          <span>停用</span>
          <strong>{{ stats.disabled }}</strong>
        </div>
        <div class="metric-card">
          <span>有配额</span>
          <strong>{{ stats.limited }}</strong>
        </div>
      </section>

      <section class="usage-panel">
        <div class="usage-copy">
          <div class="section-title">
            <mdui-icon name="terminal"></mdui-icon>
            <h2>使用说明</h2>
          </div>
          <p>创建后的 Key 用于访问 OpenAI 兼容接口，推荐放在 Authorization: Bearer sk-... Header 中。请求体里的 model 优先生效，提供方默认模型只在未传 model 时兜底。</p>
          <div class="usage-grid">
            <div>
              <span>对话接口</span>
              <strong>POST /v1/chat/completions</strong>
            </div>
            <div>
              <span>模型列表</span>
              <strong>GET /v1/models?provider=openai</strong>
            </div>
            <div>
              <span>指定提供方</span>
              <strong>X-Provider: openai</strong>
            </div>
          </div>
        </div>
        <div class="example-box">
          <div class="example-head">
            <span>请求示例</span>
            <mdui-button class="inline-button" variant="text" icon="content_copy" @click="copyExample">复制</mdui-button>
          </div>
          <pre><code>{{ chatExample }}</code></pre>
        </div>
      </section>

      <div class="workspace">
        <section class="list-panel">
          <div class="panel-toolbar">
            <mdui-text-field class="search-field" variant="outlined" label="搜索 Key" :value="query"
              @input="query = $event.target.value"></mdui-text-field>
            <mdui-segmented-button-group :value="statusFilter" @change="statusFilter = $event.target.value">
              <mdui-segmented-button value="all">全部</mdui-segmented-button>
              <mdui-segmented-button value="enabled">启用</mdui-segmented-button>
              <mdui-segmented-button value="disabled">停用</mdui-segmented-button>
              <mdui-segmented-button value="expired">过期</mdui-segmented-button>
            </mdui-segmented-button-group>
          </div>

        <div v-if="loading" class="empty-state">
          <mdui-circular-progress></mdui-circular-progress>
          <span>加载中</span>
        </div>
        <div v-else-if="filteredKeys.length === 0" class="empty-state">
          <mdui-icon name="key_off"></mdui-icon>
          <span>没有匹配的 API Key</span>
        </div>

        <div v-else class="key-list">
          <button v-for="key in filteredKeys" :key="key.id" class="key-row"
            :class="{ active: selectedKey?.id === key.id, disabled: !key.enabled }" @click="selectKey(key)">
            <div class="key-main">
              <span class="key-title">{{ key.name || "未命名" }}</span>
              <span class="key-subtitle">{{ key.description || "无描述" }}</span>
            </div>
            <div class="key-meta">
              <span class="badge" :class="key.role === 'admin' ? 'admin' : 'user'">{{ key.role === "admin" ? "Admin" : "User" }}</span>
              <span class="expire" :class="{ danger: isExpired(key.expires_at) }">{{ shortDate(key.expires_at) }}</span>
            </div>
            <div class="usage-line">
              <span>日 {{ key.usage_day }}/{{ key.quota_day > 0 ? key.quota_day : "不限" }}</span>
              <div><i :style="{ width: usagePercent(key.usage_day, key.quota_day) + '%' }"></i></div>
            </div>
          </button>
        </div>
      </section>

      <aside class="side-panel">
        <section class="create-panel">
          <div class="section-title">
            <mdui-icon name="add_circle"></mdui-icon>
            <h2>新建 Key</h2>
          </div>
          <mdui-text-field class="field" variant="outlined" label="名称" :value="form.name"
            @input="form.name = $event.target.value"></mdui-text-field>
          <mdui-text-field class="field" variant="outlined" label="描述" :value="form.description"
            @input="form.description = $event.target.value"></mdui-text-field>
          <div class="two-col">
            <mdui-select class="field" variant="outlined" label="角色" :value="form.role"
              @change="form.role = $event.target.value">
              <mdui-menu-item value="user">普通用户</mdui-menu-item>
              <mdui-menu-item value="admin">管理员</mdui-menu-item>
            </mdui-select>
            <mdui-text-field class="field" variant="outlined" type="datetime-local" label="过期时间"
              :value="form.expires_at" @input="form.expires_at = $event.target.value"></mdui-text-field>
          </div>
          <div class="two-col">
            <mdui-text-field class="field" variant="outlined" type="number" label="日配额"
              :value="form.quota_day" @input="form.quota_day = Number($event.target.value || 0)"></mdui-text-field>
            <mdui-text-field class="field" variant="outlined" type="number" label="月配额"
              :value="form.quota_month" @input="form.quota_month = Number($event.target.value || 0)"></mdui-text-field>
          </div>
          <mdui-text-field class="field" variant="outlined" label="允许 IP，逗号或换行分隔"
            :value="form.allowed_ips" @input="form.allowed_ips = $event.target.value"></mdui-text-field>
          <mdui-button class="full-button" variant="filled" icon="key" :loading="creating" @click="submitCreate">创建 API Key</mdui-button>
        </section>

        <section class="detail-panel" v-if="selectedKey">
          <div class="section-title">
            <mdui-icon name="manage_accounts"></mdui-icon>
            <h2>Key 设置</h2>
          </div>
          <div class="detail-status">
            <span :class="{ off: !editForm.enabled }">{{ editForm.enabled ? "已启用" : "已停用" }}</span>
            <mdui-switch :checked="editForm.enabled" @change="editForm.enabled = $event.target.checked"></mdui-switch>
          </div>
          <mdui-text-field class="field" variant="outlined" label="名称" :value="editForm.name"
            @input="editForm.name = $event.target.value"></mdui-text-field>
          <mdui-text-field class="field" variant="outlined" label="描述" :value="editForm.description"
            @input="editForm.description = $event.target.value"></mdui-text-field>
          <div class="two-col">
            <mdui-select class="field" variant="outlined" label="角色" :value="editForm.role"
              @change="editForm.role = $event.target.value">
              <mdui-menu-item value="user">普通用户</mdui-menu-item>
              <mdui-menu-item value="admin">管理员</mdui-menu-item>
            </mdui-select>
            <mdui-text-field class="field" variant="outlined" type="datetime-local" label="过期时间"
              :value="editForm.expires_at" @input="editForm.expires_at = $event.target.value"></mdui-text-field>
          </div>
          <div class="two-col">
            <mdui-text-field class="field" variant="outlined" type="number" label="日配额"
              :value="editForm.quota_day" @input="editForm.quota_day = Number($event.target.value || 0)"></mdui-text-field>
            <mdui-text-field class="field" variant="outlined" type="number" label="月配额"
              :value="editForm.quota_month" @input="editForm.quota_month = Number($event.target.value || 0)"></mdui-text-field>
          </div>
          <mdui-text-field class="field" variant="outlined" label="允许 IP" :value="editForm.allowed_ips"
            @input="editForm.allowed_ips = $event.target.value"></mdui-text-field>
          <div class="info-grid">
            <span>创建</span><strong>{{ formatDate(selectedKey.created_at) }}</strong>
            <span>最后使用</span><strong>{{ formatDate(selectedKey.last_used_at) }}</strong>
            <span>月用量</span><strong>{{ selectedKey.usage_month }}/{{ selectedKey.quota_month > 0 ? selectedKey.quota_month : "不限" }}</strong>
          </div>
          <div class="action-row">
            <mdui-button class="inline-button" variant="outlined" @click="toggleKey(selectedKey)">
              {{ selectedKey.enabled ? "停用" : "启用" }}
            </mdui-button>
            <mdui-button class="inline-button danger-button" variant="outlined" @click="deleteKey(selectedKey)">删除</mdui-button>
            <mdui-button class="inline-button" variant="filled" icon="save" @click="saveSelectedKey">保存</mdui-button>
          </div>
        </section>
        </aside>
      </div>
    </div>

    <mdui-dialog :open="showResult" headline="API Key 已创建" description="密钥只显示一次，请立即复制。"
      @close="showResult = false">
      <div class="created-key">
        <code>{{ createdKey }}</code>
      </div>
      <mdui-button class="dialog-action-button" slot="action" variant="text" @click="showResult = false">关闭</mdui-button>
      <mdui-button class="dialog-action-button" slot="action" variant="filled" icon="content_copy" @click="copyKey">复制</mdui-button>
    </mdui-dialog>
  </mdui-layout-main>
</template>

<style scoped>
.apikey-page {
  background: #f6f8fb;
  box-sizing: border-box;
  display: block;
  min-height: 100%;
  width: 100%;
}

.page-content {
  box-sizing: border-box;
  padding: 0 1.5rem 1.5rem;
}

.page-header,
.workspace,
.metric-grid,
.usage-panel {
  margin-left: auto;
  margin-right: auto;
  max-width: 1280px;
}

.page-header {
  align-items: flex-end;
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;
}

h1,
h2,
p {
  margin: 0;
}

h1 {
  color: #202124;
  font-size: 1.9rem;
  font-weight: 760;
}

.page-header p {
  color: #5f6368;
  line-height: 1.55;
  margin-top: 0.35rem;
}

.inline-button {
  width: auto;
}

.metric-grid {
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  margin-bottom: 1rem;
}

.metric-card {
  background: #fff;
  border: 1px solid #dde3ea;
  border-radius: 8px;
  padding: 0.9rem 1rem;
}

.metric-card span {
  color: #5f6368;
  display: block;
  font-size: 0.82rem;
}

.metric-card strong {
  color: #202124;
  display: block;
  font-size: 1.6rem;
  margin-top: 0.25rem;
}

.usage-panel {
  background: #fff;
  border: 1px solid #dde3ea;
  border-radius: 8px;
  box-sizing: border-box;
  display: grid;
  gap: 1rem;
  grid-template-columns: minmax(0, 0.85fr) minmax(20rem, 1.15fr);
  margin-bottom: 1rem;
  padding: 1rem;
}

.usage-copy p {
  color: #5f6368;
  font-size: 0.9rem;
  line-height: 1.55;
  margin-bottom: 0.9rem;
}

.usage-grid {
  display: grid;
  gap: 0.65rem;
}

.usage-grid div {
  background: #f8fafc;
  border-radius: 8px;
  padding: 0.7rem 0.75rem;
}

.usage-grid span {
  color: #6b7280;
  display: block;
  font-size: 0.78rem;
  margin-bottom: 0.2rem;
}

.usage-grid strong {
  color: #202124;
  display: block;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.84rem;
  font-weight: 650;
  overflow-wrap: anywhere;
}

.example-box {
  background: #111827;
  border-radius: 8px;
  color: #e5e7eb;
  min-width: 0;
  overflow: hidden;
}

.example-head {
  align-items: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.12);
  display: flex;
  justify-content: space-between;
  padding: 0.55rem 0.75rem;
}

.example-head span {
  font-size: 0.84rem;
  font-weight: 700;
}

.example-box pre {
  margin: 0;
  overflow: auto;
  padding: 0.85rem;
}

.example-box code {
  color: #e5e7eb;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
  font-size: 0.82rem;
  line-height: 1.5;
  white-space: pre;
}

.workspace {
  align-items: start;
  display: grid;
  gap: 1rem;
  grid-template-columns: minmax(0, 1fr) 27rem;
}

.list-panel,
.create-panel,
.detail-panel {
  background: #fff;
  border: 1px solid #dde3ea;
  border-radius: 8px;
  box-sizing: border-box;
  padding: 1rem;
}

.panel-toolbar {
  align-items: center;
  display: grid;
  gap: 0.75rem;
  grid-template-columns: minmax(14rem, 1fr) auto;
  margin-bottom: 0.75rem;
}

.search-field,
.field {
  box-sizing: border-box;
  margin: 0;
  width: 100%;
}

.empty-state {
  align-items: center;
  color: #6b7280;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  min-height: 18rem;
  justify-content: center;
}

.empty-state mdui-icon {
  color: #9aa0a6;
  font-size: 3rem;
}

.key-list {
  display: grid;
  gap: 0.55rem;
}

.key-row {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  cursor: pointer;
  display: grid;
  font: inherit;
  gap: 0.65rem;
  grid-template-columns: minmax(0, 1fr) auto;
  padding: 0.85rem;
  text-align: left;
}

.key-row:hover,
.key-row.active {
  background: #f8fbff;
  border-color: #9bbcff;
}

.key-row.disabled {
  opacity: 0.62;
}

.key-main {
  min-width: 0;
}

.key-title {
  color: #202124;
  display: block;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.key-subtitle {
  color: #6b7280;
  display: block;
  font-size: 0.84rem;
  margin-top: 0.2rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.key-meta {
  align-items: flex-end;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.badge {
  border-radius: 999px;
  font-size: 0.74rem;
  font-weight: 700;
  padding: 0.22rem 0.55rem;
}

.badge.admin {
  background: #e8f0fe;
  color: #174ea6;
}

.badge.user {
  background: #edf2f7;
  color: #334155;
}

.expire {
  color: #6b7280;
  font-size: 0.78rem;
}

.expire.danger {
  color: #b42318;
}

.usage-line {
  align-items: center;
  color: #6b7280;
  display: grid;
  font-size: 0.78rem;
  gap: 0.6rem;
  grid-column: 1 / -1;
  grid-template-columns: 8rem 1fr;
}

.usage-line div {
  background: #edf2f7;
  border-radius: 999px;
  height: 0.42rem;
  overflow: hidden;
}

.usage-line i {
  background: #1a73e8;
  display: block;
  height: 100%;
}

.side-panel {
  display: grid;
  gap: 1rem;
}

.section-title {
  align-items: center;
  display: flex;
  gap: 0.55rem;
  margin-bottom: 0.9rem;
}

.section-title mdui-icon {
  color: #1a73e8;
}

.section-title h2 {
  color: #202124;
  font-size: 1.1rem;
  font-weight: 720;
}

.two-col {
  display: grid;
  gap: 0.65rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin-bottom: 0.65rem;
}

.create-panel > .field,
.detail-panel > .field {
  margin-bottom: 0.65rem;
}

.full-button {
  margin-top: 0.15rem;
  width: 100%;
}

.detail-status {
  align-items: center;
  background: #f8fafc;
  border-radius: 8px;
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.75rem;
  padding: 0.65rem 0.75rem;
}

.detail-status span {
  color: #137333;
  font-weight: 700;
}

.detail-status span.off {
  color: #b42318;
}

.info-grid {
  background: #f8fafc;
  border-radius: 8px;
  display: grid;
  gap: 0.45rem 0.75rem;
  grid-template-columns: auto minmax(0, 1fr);
  margin-top: 0.75rem;
  padding: 0.75rem;
}

.info-grid span {
  color: #6b7280;
}

.info-grid strong {
  color: #202124;
  font-weight: 650;
  min-width: 0;
  overflow-wrap: anywhere;
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  justify-content: flex-end;
  margin-top: 0.85rem;
}

.danger-button {
  color: #b42318;
}

.created-key {
  background: #f8fafc;
  border: 1px solid #dde3ea;
  border-radius: 8px;
  margin-top: 0.75rem;
  overflow-wrap: anywhere;
  padding: 1rem;
}

.created-key code {
  color: #202124;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
}

.dialog-action-button {
  width: auto;
}

@media (max-width: 1024px) {
  .workspace {
    grid-template-columns: 1fr;
  }

  .usage-panel {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .apikey-page {
    padding: 0;
  }

  .page-content {
    padding: 0 0.75rem 1rem;
  }

  .page-header,
  .panel-toolbar {
    align-items: stretch;
    grid-template-columns: 1fr;
    flex-direction: column;
  }

  .metric-grid,
  .two-col {
    grid-template-columns: 1fr;
  }

  .key-row {
    grid-template-columns: 1fr;
  }

  .key-meta {
    align-items: flex-start;
    flex-direction: row;
  }
}
</style>
