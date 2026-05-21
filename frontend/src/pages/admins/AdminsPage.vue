<script setup>
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  SafetyOutlined,
} from '@ant-design/icons-vue';
import { message, Modal } from 'ant-design-vue';
import AppSidebar from '@/components/AppSidebar.vue';
import { HttpUtil } from '@/utils';
import { antdThemeConfig, theme } from '@/composables/useTheme.js';

const { t } = useI18n();

const basePath = window.X_UI_BASE_PATH ?? '/';
const requestUri = window.location.pathname;

const admins = ref([]);
const inboundOptions = ref([]);
const loading = ref(false);
const modalOpen = ref(false);
const saving = ref(false);
const isEdit = ref(false);
const editingId = ref(null);

const form = ref({
  username: '',
  password: '',
  confirmPassword: '',
  allowedInbounds: [],
});

const columns = [
  { title: '#', dataIndex: 'id', width: 60 },
  { title: t('username'), dataIndex: 'username' },
  { title: t('pages.admins.allowedInbounds'), key: 'allowedInbounds' },
  { title: t('pages.admins.actions'), key: 'actions', width: 120, align: 'center' },
];

function parseAllowedIds(admin) {
  if (!admin.allowedInbounds) return [];
  if (Array.isArray(admin.allowedInbounds)) return admin.allowedInbounds;
  try { return JSON.parse(admin.allowedInbounds); } catch { return []; }
}

function inboundLabel(id) {
  const opt = inboundOptions.value.find((o) => o.id === id);
  return opt ? `${opt.remark || opt.protocol}:${opt.port}` : `#${id}`;
}

async function fetchAdmins() {
  loading.value = true;
  try {
    const msg = await HttpUtil.get(basePath + 'panel/api/admins');
    if (msg?.success) admins.value = msg.obj ?? [];
  } finally {
    loading.value = false;
  }
}

async function fetchInbounds() {
    const msg = await HttpUtil.get(basePath + 'panel/api/inbounds/options');
    if (msg?.success) inboundOptions.value = msg.obj ?? [];
}

function openCreate() {
  isEdit.value = false;
  editingId.value = null;
  form.value = { username: '', password: '', confirmPassword: '', allowedInbounds: [] };
  modalOpen.value = true;
}

function openEdit(admin) {
  isEdit.value = true;
  editingId.value = admin.id;
  form.value = {
    username: admin.username,
    password: '',
    confirmPassword: '',
    allowedInbounds: parseAllowedIds(admin),
  };
  modalOpen.value = true;
}

async function save() {
  if (!form.value.username) {
    message.error(t('pages.login.toasts.emptyUsername'));
    return;
  }
  if (!isEdit.value && !form.value.password) {
    message.error(t('pages.login.toasts.emptyPassword'));
    return;
  }
  if (form.value.password && form.value.password !== form.value.confirmPassword) {
    message.error(t('pages.admins.passwordMismatch'));
    return;
  }
  saving.value = true;
  try {
    const payload = {
      username: form.value.username,
      password: form.value.password,
      allowedInbounds: form.value.allowedInbounds,
    };
    const msg = isEdit.value
      ? await HttpUtil.post(basePath + 'panel/api/admins', payload)
      : await HttpUtil.put(basePath + 'panel/api/admins/' + editingId.value, payload);
    if (msg?.success) {
      modalOpen.value = false;
      await fetchAdmins();
    }
  } finally {
    saving.value = false;
  }
}

function confirmDelete(admin) {
  Modal.confirm({
    title: t('pages.admins.deleteTitle'),
    content: t('pages.admins.deleteContent', { username: admin.username }),
    okType: 'danger',
    onOk: () => doDelete(admin.id),
  });
}

async function doDelete(id) {
    const msg = await HttpUtil.delete(basePath + 'panel/api/admins/' + id);
    if (msg?.success) await fetchAdmins();
}

onMounted(async () => {
  await Promise.all([fetchAdmins(), fetchInbounds()]);
});
</script>

<template>
  <a-config-provider :theme="antdThemeConfig">
    <a-layout :class="{ 'is-dark': theme.isDark }" style="min-height: 100vh">
      <AppSidebar :base-path="basePath" :request-uri="requestUri" />
      <a-layout-content style="padding: 24px">

        <div style="display:flex; align-items:center; justify-content:space-between; margin-bottom:16px">
          <h2 style="margin:0; display:flex; align-items:center; gap:8px">
            <SafetyOutlined />
            {{ t('menu.admins') }}
          </h2>
          <a-button type="primary" @click="openCreate">
            <template #icon><PlusOutlined /></template>
            {{ t('pages.admins.addAdmin') }}
          </a-button>
        </div>

        <a-table
          :columns="columns"
          :data-source="admins"
          :loading="loading"
          row-key="id"
          :pagination="false"
          bordered
        >
          <template #bodyCell="{ column, record }">
            <template v-if="column.key === 'allowedInbounds'">
              <a-tag v-if="!parseAllowedIds(record).length" color="warning">
                {{ t('pages.admins.noInbounds') }}
              </a-tag>
              <template v-else>
                <a-tag
                  v-for="id in parseAllowedIds(record)"
                  :key="id"
                  color="blue"
                  style="margin-bottom:2px"
                >
                  {{ inboundLabel(id) }}
                </a-tag>
              </template>
            </template>

            <template v-else-if="column.key === 'actions'">
              <a-space>
                <a-tooltip :title="t('edit')">
                  <a-button size="small" @click="openEdit(record)">
                    <template #icon><EditOutlined /></template>
                  </a-button>
                </a-tooltip>
                <a-tooltip :title="t('delete')">
                  <a-button size="small" danger @click="confirmDelete(record)">
                    <template #icon><DeleteOutlined /></template>
                  </a-button>
                </a-tooltip>
              </a-space>
            </template>
          </template>
        </a-table>

      </a-layout-content>
    </a-layout>

    <!-- Create / Edit modal -->
    <a-modal
      v-model:open="modalOpen"
      :title="isEdit ? t('pages.admins.editAdmin') : t('pages.admins.addAdmin')"
      :ok-text="t('save')"
      :confirm-loading="saving"
      :destroy-on-close="true"
      @ok="save"
    >
      <a-form layout="vertical" style="margin-top:8px">
        <a-form-item :label="t('username')">
          <a-input v-model:value="form.username" :placeholder="t('username')" />
        </a-form-item>

        <a-form-item :label="t('password')">
          <a-input-password
            v-model:value="form.password"
            :placeholder="isEdit ? t('pages.admins.leaveBlankToKeep') : t('password')"
          />
        </a-form-item>

        <a-form-item :label="t('pages.admins.confirmPassword')">
          <a-input-password
            v-model:value="form.confirmPassword"
            :placeholder="t('pages.admins.confirmPassword')"
          />
        </a-form-item>

        <a-form-item :label="t('pages.admins.allowedInbounds')">
          <a-select
            v-model:value="form.allowedInbounds"
            mode="multiple"
            :placeholder="t('pages.admins.selectInbounds')"
            :options="inboundOptions.map((o) => ({
              value: o.id,
              label: `${o.remark || o.protocol}:${o.port}`,
            }))"
            style="width:100%"
          />
          <div style="margin-top:4px; opacity:0.65; font-size:12px">
            {{ t('pages.admins.inboundsHint') }}
          </div>
        </a-form-item>
      </a-form>
    </a-modal>
  </a-config-provider>
</template>