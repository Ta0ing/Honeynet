import { useEffect, useRef, useState } from 'react';
import axios from 'axios';
import {
  Alert, Button, Card, Form, Grid, Input, Message, Popconfirm, Skeleton, Space, Tag,
  Typography, Upload, type UploadProps,
} from '@arco-design/web-react';
import {
  IconDelete, IconEmail, IconLink, IconPhone, IconRefresh, IconSave, IconUpload,
} from '@arco-design/web-react/icon';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, errorMessage } from '../api';
import {
  DEFAULT_BRANDING, normalizeBranding, platformOEMQueryKey, publicExternalURL,
  type PlatformBranding,
} from '../branding';
import BrandLogo from '../components/BrandLogo';
import PageHeader from '../components/PageHeader';

type LogoKind = 'system' | 'company';
type UploadItem = NonNullable<UploadProps['fileList']>[number];
type RequestOptions = Parameters<NonNullable<UploadProps['customRequest']>>[0];
type TextSettings = Pick<PlatformBranding,
  'system_name' | 'system_version' | 'copyright' | 'customer_service_phone' |
  'customer_service_email' | 'official_website_url' | 'product_documentation_url'>;

const settingsQueryKey = ['platform-settings'] as const;
const textSettings = (value: PlatformBranding): TextSettings => ({
  system_name: value.system_name,
  system_version: value.system_version,
  copyright: value.copyright,
  customer_service_phone: value.customer_service_phone,
  customer_service_email: value.customer_service_email,
  official_website_url: value.official_website_url,
  product_documentation_url: value.product_documentation_url,
});

const httpURLRule = {
  validator: (value: string | undefined, callback: (error?: string) => void) => {
    if (!value) return callback();
    try {
      const parsed = new URL(value);
      if (parsed.protocol === 'http:' || parsed.protocol === 'https:') return callback();
    } catch { /* handled below */ }
    callback('请输入完整的 HTTP 或 HTTPS 地址');
  },
};

function VersionText({ version }: { version: string }) {
  if (!version) return <>版本未设置</>;
  return <>{/^v/i.test(version) ? version : `v${version}`}</>;
}

export default function PlatformConfig() {
  const [form] = Form.useForm<TextSettings>();
  const queryClient = useQueryClient();
  const revisionRef = useRef(0);
  const [revision, setRevision] = useState(1);
  const [draft, setDraft] = useState<PlatformBranding>(normalizeBranding(DEFAULT_BRANDING));
  const [dirty, setDirty] = useState(false);
  const [logoBusy, setLogoBusy] = useState<Record<LogoKind, boolean>>({ system: false, company: false });
  const [fileLists, setFileLists] = useState<Record<LogoKind, UploadItem[]>>({ system: [], company: [] });

  const settings = useQuery<PlatformBranding>({
    queryKey: settingsQueryKey,
    queryFn: async () => normalizeBranding(await api.get('/platform/settings')),
    retry: false,
  });

  const replaceServerState = (value: unknown, preserveForm = false) => {
    const next = normalizeBranding(value);
    revisionRef.current = next.revision;
    setRevision(next.revision);
    queryClient.setQueryData(settingsQueryKey, next);
    queryClient.setQueryData(platformOEMQueryKey, next);
    void queryClient.invalidateQueries({ queryKey: platformOEMQueryKey });
    if (preserveForm) {
      setDraft({ ...next, ...form.getFieldsValue() });
    } else {
      form.setFieldsValue(textSettings(next));
      setDraft(next);
      setDirty(false);
    }
  };

  useEffect(() => {
    if (settings.data && settings.data.revision !== revisionRef.current) replaceServerState(settings.data);
    // replaceServerState is deliberately tied to the server revision only.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [settings.data?.revision]);

  const refreshAfterConflict = async (error: unknown) => {
    if (axios.isAxiosError(error) && error.response?.status === 409) {
      Message.warning('配置已被其他管理员更新，已重新加载最新版本');
      const latest = await settings.refetch();
      if (latest.data) replaceServerState(latest.data);
      return;
    }
    Message.error(errorMessage(error));
  };

  const save = useMutation({
    mutationFn: async () => {
      const values = await form.validate();
      return api.put('/platform/settings', { ...values, revision: revisionRef.current });
    },
    onSuccess: (value) => { replaceServerState(value); Message.success('平台配置已保存并全局生效'); },
    onError: refreshAfterConflict,
  });

  const uploadLogo = (kind: LogoKind, options: RequestOptions) => {
    const controller = new AbortController();
    const body = new FormData();
    body.append('file', options.file);
    body.append('revision', String(revisionRef.current));
    setLogoBusy((current) => ({ ...current, [kind]: true }));
    void api.post(`/platform/assets/${kind}`, body, {
      signal: controller.signal,
      onUploadProgress: (event) => options.onProgress(event.total ? Math.round(event.loaded / event.total * 100) : 0),
    }).then((value) => {
      replaceServerState(value, true);
      options.onSuccess(value as object);
      Message.success(kind === 'system' ? '系统 Logo 已更新' : '公司 Logo 已更新');
    }).catch((error) => {
      options.onError(error as object);
      if (!axios.isCancel(error)) void refreshAfterConflict(error);
    }).finally(() => setLogoBusy((current) => ({ ...current, [kind]: false })));
    return { abort: () => controller.abort() };
  };

  const validateLogo = (file: File) => {
    const supported = ['image/png', 'image/jpeg', 'image/webp'].includes(file.type.toLowerCase());
    if (!supported) { Message.error('仅支持 PNG、JPEG 或 WebP 图片'); return false; }
    if (file.size > 2 * 1024 * 1024) { Message.error('图片不能超过 2 MiB'); return false; }
    return true;
  };

  const clearLogo = async (kind: LogoKind) => {
    setLogoBusy((current) => ({ ...current, [kind]: true }));
    try {
      const result = await api.delete(`/platform/assets/${kind}`, { params: { revision: revisionRef.current } });
      replaceServerState(result, true);
      setFileLists((current) => ({ ...current, [kind]: [] }));
      Message.success(kind === 'system' ? '已恢复内置系统 Logo' : '已恢复内置公司 Logo');
    } catch (error) {
      await refreshAfterConflict(error);
    } finally {
      setLogoBusy((current) => ({ ...current, [kind]: false }));
    }
  };

  const restoreDefaults = useMutation({
    mutationFn: async () => {
      let version = settings.data?.system_version || draft.system_version;
      try {
        const summary: any = await api.get('/dashboard/summary');
        version = summary?.version || version;
      } catch { /* retaining the known version is safe */ }
      let current: any;
      try {
        current = await api.put('/platform/settings', {
          ...textSettings({ ...normalizeBranding(DEFAULT_BRANDING), system_version: version }),
          revision: revisionRef.current,
        });
        current = await api.delete('/platform/assets/system', { params: { revision: current.revision } });
        current = await api.delete('/platform/assets/company', { params: { revision: current.revision } });
        return current;
      } catch (error) {
        throw Object.assign(error instanceof Error ? error : new Error('恢复默认失败'), { restorePartiallyApplied: !!current });
      }
    },
    onSuccess: (value) => {
      replaceServerState(value);
      setFileLists({ system: [], company: [] });
      Message.success('已恢复默认品牌配置');
    },
    onError: async (error: any) => {
      if (error?.restorePartiallyApplied) {
        Message.warning('部分默认配置已生效，已重新加载服务端最终状态；请确认后再次恢复');
        const latest = await settings.refetch();
        if (latest.data) replaceServerState(latest.data);
        return;
      }
      await refreshAfterConflict(error);
    },
  });

  const logoCard = (kind: LogoKind, title: string, description: string) => {
    const logoURL = kind === 'system' ? draft.system_logo_url : draft.company_logo_url;
    return <Card size="small" title={title} className="platform-logo-card">
      <div className="platform-logo-row">
        <BrandLogo src={logoURL} alt={title} kind={kind} className="platform-logo-image" />
        <div className="platform-logo-actions">
          <Typography.Text type="secondary">{description}</Typography.Text>
          <Space wrap>
            <Upload
              accept="image/png,image/jpeg,image/webp"
              fileList={fileLists[kind]}
              showUploadList={false}
              beforeUpload={validateLogo}
              customRequest={(options) => uploadLogo(kind, options)}
              onChange={(list) => setFileLists((current) => ({ ...current, [kind]: list.slice(-1) }))}
              disabled={logoBusy[kind]}
            >
              <Button icon={<IconUpload />} loading={logoBusy[kind]}>上传或替换</Button>
            </Upload>
            <Popconfirm title={`确定恢复内置${title}吗？`} onOk={() => clearLogo(kind)}>
              <Button status="warning" icon={<IconDelete />} disabled={logoBusy[kind]}>恢复内置</Button>
            </Popconfirm>
          </Space>
          <Typography.Text type="secondary" className="platform-upload-hint">PNG、JPEG 或 WebP，最大 2 MiB；服务端会安全解码并统一存储。</Typography.Text>
        </div>
      </div>
    </Card>;
  };

  if (settings.isLoading) return <Card><Skeleton text={{ rows: 8 }} animation /></Card>;
  if (settings.isError && !settings.data) return <Alert type="error" title="无法读取平台配置" content="请确认当前账号具有管理员权限后重试。" action={<Button icon={<IconRefresh />} onClick={() => settings.refetch()}>重试</Button>} />;

  const website = publicExternalURL(draft.official_website_url);
  const documentation = publicExternalURL(draft.product_documentation_url);
  return <>
    <PageHeader
      title="配置管理"
      description="统一维护系统品牌、联系方式与对外入口，保存后登录页、控制台和大屏同步生效"
      extra={<Space wrap>
        <Tag color="arcoblue">配置版本 r{revision}</Tag>
        {dirty && <Tag color="orange">有未保存更改</Tag>}
        <Button icon={<IconRefresh />} loading={settings.isFetching} onClick={() => void settings.refetch().then((result) => { if (result.data) replaceServerState(result.data); })}>刷新</Button>
        <Popconfirm title="恢复默认配置并清除自定义 Logo？" content="该操作会立即全局生效。" onOk={() => restoreDefaults.mutate()}>
          <Button status="warning" loading={restoreDefaults.isPending}>恢复默认</Button>
        </Popconfirm>
        <Button type="primary" icon={<IconSave />} loading={save.isPending} onClick={() => save.mutate()}>保存配置</Button>
      </Space>}
    />
    <Alert className="platform-config-alert" type="info" content="Logo 上传与文本保存均使用配置版本校验；多人同时修改时会拒绝覆盖并加载最新版。" />
    <Grid.Row gutter={16} align="stretch">
      <Grid.Col xs={24} xl={15}>
        <Card title="基础品牌信息" className="panel-card platform-config-card">
          <Form<TextSettings>
            form={form}
            layout="vertical"
            onValuesChange={(_, values) => {
              setDirty(true);
              setDraft((current) => ({ ...current, ...values }));
            }}
          >
            <Grid.Row gutter={16}>
              <Grid.Col xs={24} md={12}><Form.Item label="系统名称" field="system_name" rules={[{ required: true, message: '请输入系统名称' }, { maxLength: 128 }]}><Input maxLength={128} showWordLimit placeholder="例如：企业蜜网平台" /></Form.Item></Grid.Col>
              <Grid.Col xs={24} md={12}><Form.Item label="系统版本" field="system_version" rules={[{ required: true, message: '请输入系统版本' }, { maxLength: 64 }]}><Input maxLength={64} showWordLimit placeholder="例如：1.0.0" /></Form.Item></Grid.Col>
            </Grid.Row>
            <Form.Item label="版权信息" field="copyright" rules={[{ required: true, message: '请输入版权信息' }, { maxLength: 512 }]}><Input.TextArea maxLength={512} showWordLimit autoSize={{ minRows: 2, maxRows: 4 }} /></Form.Item>
            <Grid.Row gutter={16}>
              <Grid.Col xs={24} md={12}><Form.Item label="客服热线" field="customer_service_phone" rules={[{ maxLength: 64 }]}><Input prefix={<IconPhone />} maxLength={64} placeholder="例如：400-000-0000" /></Form.Item></Grid.Col>
              <Grid.Col xs={24} md={12}><Form.Item label="客服邮箱" field="customer_service_email" rules={[{ type: 'email', message: '请输入有效邮箱地址' }, { maxLength: 254 }]}><Input prefix={<IconEmail />} maxLength={254} placeholder="support@example.com" /></Form.Item></Grid.Col>
            </Grid.Row>
            <Form.Item label="官方网站" field="official_website_url" rules={[httpURLRule, { maxLength: 1024 }]}><Input prefix={<IconLink />} maxLength={1024} placeholder="https://www.example.com" /></Form.Item>
            <Form.Item label="产品文档" field="product_documentation_url" rules={[httpURLRule, { maxLength: 1024 }]}><Input prefix={<IconLink />} maxLength={1024} placeholder="https://docs.example.com" /></Form.Item>
          </Form>
        </Card>
        <Grid.Row gutter={16}>
          <Grid.Col xs={24} md={12}>{logoCard('system', '系统 Logo', '用于浏览器图标、登录页和侧边栏。')}</Grid.Col>
          <Grid.Col xs={24} md={12}>{logoCard('company', '公司 Logo', '用于登录页、品牌预览和控制台页脚。')}</Grid.Col>
        </Grid.Row>
      </Grid.Col>
      <Grid.Col xs={24} xl={9}>
        <Card title="全局实时预览" className="panel-card platform-preview-card">
          <div className="platform-sidebar-preview">
            <BrandLogo src={draft.system_logo_url} alt={draft.system_name} className="preview-system-logo" />
            <div><strong>{draft.system_name || DEFAULT_BRANDING.system_name}</strong><span><VersionText version={draft.system_version} /></span></div>
          </div>
          <div className="platform-login-preview">
            <BrandLogo src={draft.system_logo_url} alt={draft.system_name} className="preview-login-logo" />
            <Typography.Title heading={5}>{draft.system_name || DEFAULT_BRANDING.system_name}</Typography.Title>
            <Typography.Text type="secondary">登录管理中心</Typography.Text>
          </div>
          <div className="platform-contact-preview">
            <Typography.Text bold>联系与支持</Typography.Text>
            {draft.customer_service_phone ? <a href={`tel:${draft.customer_service_phone}`}><IconPhone />{draft.customer_service_phone}</a> : <Typography.Text type="secondary">客服热线未配置</Typography.Text>}
            {draft.customer_service_email ? <a href={`mailto:${draft.customer_service_email}`}><IconEmail />{draft.customer_service_email}</a> : <Typography.Text type="secondary">客服邮箱未配置</Typography.Text>}
            {website && <a href={website} target="_blank" rel="noreferrer"><IconLink />官方网站</a>}
            {documentation && <a href={documentation} target="_blank" rel="noreferrer"><IconLink />产品文档</a>}
          </div>
          <div className="platform-footer-preview">
            <BrandLogo src={draft.company_logo_url} alt="公司 Logo" kind="company" className="preview-company-logo" />
            <span>{draft.copyright || DEFAULT_BRANDING.copyright}</span>
          </div>
        </Card>
      </Grid.Col>
    </Grid.Row>
  </>;
}
