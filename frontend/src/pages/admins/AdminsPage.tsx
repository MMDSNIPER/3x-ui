import { useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Card,
  Checkbox,
  Col,
  ConfigProvider,
  Form,
  Input,
  Layout,
  Modal,
  Row,
  Space,
  Spin,
  Table,
  Tag,
  message,
} from 'antd';
import { DeleteOutlined, EditOutlined, PlusOutlined, TeamOutlined } from '@ant-design/icons';

import { useTheme } from '@/hooks/useTheme';
import { useMediaQuery } from '@/hooks/useMediaQuery';
import { HttpUtil } from '@/utils';
import AppSidebar from '@/components/AppSidebar';
import '@/styles/page-cards.css';

const basePath = window.X_UI_BASE_PATH || '';
const requestUri = window.location.pathname;

interface SubAdmin {
  id: number;
  username: string;
  role: string;
  allowedInbounds: number[];
}

interface InboundOption {
  id: number;
  remark: string;
  port: number;
  protocol: string;
}

interface AdminFormValues {
  username: string;
  password: string;
  allowedInbounds: number[];
}

const EMPTY_FORM: AdminFormValues = {
  username: '',
  password: '',
  allowedInbounds: [],
};

export default function AdminsPage() {
  const { t } = useTranslation();
  const { isDark, isUltra, antdThemeConfig } = useTheme();
  const { isMobile } = useMediaQuery();
  const [modal, modalContextHolder] = Modal.useModal();
  const [messageApi, messageContextHolder] = message.useMessage();

  const [admins, setAdmins] = useState<SubAdmin[]>([]);
  const [inbounds, setInbounds] = useState<InboundOption[]>([]);
  const [fetched, setFetched] = useState(false);

  const [formOpen, setFormOpen] = useState(false);
  const [formMode, setFormMode] = useState<'add' | 'edit'>('add');
  const [editingId, setEditingId] = useState<number | null>(null);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState<AdminFormValues>(EMPTY_FORM);

  const fetchData = useCallback(async () => {
    const [adminsRes, inboundsRes] = await Promise.all([
        HttpUtil.get('/panel/api/admins', undefined, { silent: true }),
        HttpUtil.get('/panel/api/inbounds/list', undefined, { silent: true }),
    ]);
    if (adminsRes?.success) {
      setAdmins(
        (adminsRes.obj as SubAdmin[]).map((a) => ({
          ...a,
          allowedInbounds: Array.isArray(a.allowedInbounds) ? a.allowedInbounds : [],
        })),
      );
    }
    if (inboundsRes?.success) {
      setInbounds(
        (inboundsRes.obj as InboundOption[]).map((ib) => ({
          id: ib.id,
          remark: ib.remark,
          port: ib.port,
          protocol: ib.protocol,
        })),
      );
    }
    setFetched(true);
  }, []);

  useEffect(() => { fetchData(); }, [fetchData]);

  const openAdd = useCallback(() => {
    setFormMode('add');
    setEditingId(null);
    setForm(EMPTY_FORM);
    setFormOpen(true);
  }, []);

  const openEdit = useCallback((admin: SubAdmin) => {
    setFormMode('edit');
    setEditingId(admin.id);
    setForm({ username: admin.username, password: '', allowedInbounds: admin.allowedInbounds });
    setFormOpen(true);
  }, []);

  const confirmDelete = useCallback((admin: SubAdmin) => {
    modal.confirm({
      title: t('pages.admins.deleteConfirmTitle', { username: admin.username }),
      content: t('pages.admins.deleteConfirmContent'),
      okText: t('delete'),
      okType: 'danger',
      cancelText: t('cancel'),
      onOk: async () => {
        const res = await HttpUtil.post(`/panel/api/admins/delete/${admin.id}`);
        if (res?.success) {
          messageApi.success(t('pages.admins.toasts.deleted'));
          fetchData();
        }
      },
    });
  }, [modal, t, messageApi, fetchData]);

  const onSave = useCallback(async () => {
    if (!form.username.trim()) {
      messageApi.error(t('pages.admins.toasts.usernameRequired'));
      return;
    }
    if (formMode === 'add' && !form.password.trim()) {
      messageApi.error(t('pages.admins.toasts.passwordRequired'));
      return;
    }
    setSaving(true);
    const payload = {
      username: form.username.trim(),
      password: form.password,
      allowedInbounds: form.allowedInbounds,
    };
    const url = formMode === 'add'
      ? '/panel/api/admins/create'
      : `/panel/api/admins/update/${editingId}`;
    const res = await HttpUtil.post(url, payload);
    setSaving(false);
    if (res?.success) {
      messageApi.success(
        formMode === 'add' ? t('pages.admins.toasts.created') : t('pages.admins.toasts.updated'),
      );
      setFormOpen(false);
      fetchData();
    }
  }, [form, formMode, editingId, messageApi, t, fetchData]);

  const inboundLabel = useCallback((ib: InboundOption) =>
    `${ib.remark || ib.protocol} :${ib.port}`,
  []);

  const columns = useMemo(() => [
    {
      title: t('pages.admins.username'),
      dataIndex: 'username',
      key: 'username',
      render: (val: string) => <strong>{val}</strong>,
    },
    {
      title: t('pages.admins.access'),
      key: 'access',
      render: (_: unknown, record: SubAdmin) => {
        if (record.allowedInbounds.length === 0) {
          return <Tag color="green">{t('pages.admins.fullAccess')}</Tag>;
        }
        return (
          <Space wrap size={4}>
            {record.allowedInbounds.map((id) => {
              const ib = inbounds.find((i) => i.id === id);
              return (
                <Tag key={id} color="blue">
                  {ib ? inboundLabel(ib) : `#${id}`}
                </Tag>
              );
            })}
          </Space>
        );
      },
    },
    {
      title: t('action'),
      key: 'action',
      width: 100,
      render: (_: unknown, record: SubAdmin) => (
        <Space>
          <Button
            size="small"
            icon={<EditOutlined />}
            onClick={() => openEdit(record)}
          />
          <Button
            size="small"
            danger
            icon={<DeleteOutlined />}
            onClick={() => confirmDelete(record)}
          />
        </Space>
      ),
    },
  ], [t, inbounds, inboundLabel, openEdit, confirmDelete]);

  const pageClass = useMemo(() => {
    const parts = ['admins-page'];
    if (isDark) parts.push('is-dark');
    if (isUltra) parts.push('is-ultra');
    return parts.join(' ');
  }, [isDark, isUltra]);

  return (
    <ConfigProvider theme={antdThemeConfig}>
      {messageContextHolder}
      {modalContextHolder}
      <Layout className={pageClass}>
        <AppSidebar basePath={basePath} requestUri={requestUri} />

        <Layout className="content-shell">
          <Layout.Content id="content-layout" className="content-area">
            <Spin spinning={!fetched} delay={200} size="large">
              {!fetched ? (
                <div className="loading-spacer" />
              ) : (
                <Row gutter={[isMobile ? 8 : 16, isMobile ? 8 : 12]}>
                  <Col span={24}>
                    <Card
                      hoverable
                      title={
                        <Space>
                          <TeamOutlined />
                          {t('pages.admins.title')}
                        </Space>
                      }
                      extra={
                        <Button type="primary" icon={<PlusOutlined />} onClick={openAdd}>
                          {!isMobile && t('pages.admins.addAdmin')}
                        </Button>
                      }
                    >
                      <Table
                        dataSource={admins}
                        columns={columns}
                        rowKey="id"
                        pagination={false}
                        size="small"
                      />
                    </Card>
                  </Col>
                </Row>
              )}
            </Spin>
          </Layout.Content>
        </Layout>
      </Layout>

      <Modal
        open={formOpen}
        title={formMode === 'add' ? t('pages.admins.addAdmin') : t('pages.admins.editAdmin')}
        okText={t('save')}
        cancelText={t('cancel')}
        confirmLoading={saving}
        onOk={onSave}
        onCancel={() => setFormOpen(false)}
        destroyOnHidden
      >
        <Form layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item label={t('pages.admins.username')} required>
            <Input
              value={form.username}
              onChange={(e) => setForm((f) => ({ ...f, username: e.target.value }))}
              autoComplete="off"
            />
          </Form.Item>

          <Form.Item
            label={t('pages.admins.password')}
            extra={formMode === 'edit' ? t('pages.admins.passwordHint') : undefined}
            required={formMode === 'add'}
          >
            <Input.Password
              value={form.password}
              onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
              autoComplete="new-password"
            />
          </Form.Item>

          <Form.Item
            label={t('pages.admins.allowedInbounds')}
            extra={t('pages.admins.allowedInboundsHint')}
          >
            <Checkbox.Group
              value={form.allowedInbounds}
              onChange={(vals) => setForm((f) => ({ ...f, allowedInbounds: vals as number[] }))}
              style={{ display: 'flex', flexDirection: 'column', gap: 6 }}
            >
              {inbounds.map((ib) => (
                <Checkbox key={ib.id} value={ib.id}>
                  {inboundLabel(ib)}
                </Checkbox>
              ))}
            </Checkbox.Group>
          </Form.Item>
        </Form>
      </Modal>
    </ConfigProvider>
  );
}