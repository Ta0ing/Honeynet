import { useState } from 'react';
import { Alert, Button, Card, Descriptions, Form, Grid, Input, InputNumber, Message, Modal, Popconfirm, Select, Space, Table, Tag, Typography } from '@arco-design/web-react';
import { IconDelete, IconEdit, IconPlus, IconRefresh, IconStorage } from '@arco-design/web-react/icon';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import axios from 'axios';
import { api, errorMessage, formatTime } from '../api';
import PageHeader from '../components/PageHeader';
import RiskTag from '../components/RiskTag';
import { eventTypeLabel } from '../presentation';
import { useAuth } from '../store';
import { usePlatformBranding } from '../branding';

const channelTypes = [
  { label: '通用 Webhook', value: 'webhook' },
  { label: '邮件 SMTP', value: 'email' },
  { label: 'Syslog', value: 'syslog' },
  { label: '企业微信机器人', value: 'wecom' },
  { label: '钉钉机器人', value: 'dingtalk' },
  { label: '飞书机器人', value: 'feishu' },
];
const typeLabels = Object.fromEntries(channelTypes.map((item) => [item.value, item.label]));
const deliveryColors: Record<string, string> = { sent: 'green', pending: 'blue', sending: 'arcoblue', retrying: 'orange', failed: 'red', skipped: 'gray' };

type EngineState = 'healthy' | 'degraded' | 'disabled' | 'unknown';
type EngineView = {
  name: string;
  role: string;
  state: EngineState;
  version?: string;
  schema?: string;
  namespace?: string;
  lastWrite?: string;
  backlog?: number;
  detail?: string;
};

const stateMeta: Record<EngineState, { label: string; color: string }> = {
  healthy: { label: '运行正常', color: 'green' },
  degraded: { label: '降级运行', color: 'orange' },
  disabled: { label: '未启用', color: 'gray' },
  unknown: { label: '状态待上报', color: 'arcoblue' },
};

function objectValue(value: unknown): Record<string, any> {
  return value && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, any> : {};
}

function firstDefined(...values: any[]) {
  return values.find((value) => value !== undefined && value !== null && value !== '');
}

function sanitizeStatusDetail(value: unknown) {
  return String(value || '')
    .replace(/:\/\/[^@\s/]+@/g, '://***@')
    .replace(/(password|token|secret|key)=([^&\s;]+)/gi, '$1=***')
    .slice(0, 220);
}

function engineState(raw: Record<string, any>, available: boolean, enabledByDefault?: boolean): EngineState {
  const enabled = firstDefined(raw.enabled, raw.configured, enabledByDefault);
  if (enabled === false || raw.status === 'disabled') return 'disabled';
  if (raw.degraded === true || raw.healthy === false || ['degraded', 'error', 'failed', 'unhealthy'].includes(String(raw.status || '').toLowerCase())) return 'degraded';
  if (raw.healthy === true || ['ok', 'healthy', 'online', 'ready', 'connected'].includes(String(raw.status || '').toLowerCase())) return 'healthy';
  return available ? 'healthy' : 'unknown';
}

function formatBacklog(value?: number) {
  return Number.isFinite(value) ? Number(value).toLocaleString('zh-CN') : '未上报';
}

function optionalNumber(value: unknown) {
  if (value === undefined || value === null || value === '') return undefined;
  const numeric = Number(value);
  return Number.isFinite(numeric) ? numeric : undefined;
}

function EngineValue({ value, copyable = false, lines = 1 }: { value?: string; copyable?: boolean; lines?: number }) {
  const text = String(value || '').trim();
  if (!text) return <Typography.Text type="secondary">未上报</Typography.Text>;
  return <Typography.Paragraph
    className={`engine-description-value${copyable ? ' engine-namespace' : ''}`}
    ellipsis={{ rows: lines, showTooltip: true }}
    copyable={copyable ? { text } : false}
  >{text}</Typography.Paragraph>;
}

function EngineCard({ engine }: { engine: EngineView }) {
  const meta = stateMeta[engine.state];
  return <Card className={`panel-card engine-card engine-${engine.state}`}>
    <div className="engine-card-head">
      <div className="engine-identity"><span className="engine-icon"><IconStorage /></span><div><Typography.Title heading={6}>{engine.name}</Typography.Title><Typography.Text type="secondary">{engine.role}</Typography.Text></div></div>
      <Tag color={meta.color}>{meta.label}</Tag>
    </div>
    <Descriptions
      border
      size="small"
      layout="vertical"
      column={{ xs: 1, sm: 2 }}
      tableLayout="fixed"
      data={[
        { label: '引擎版本', value: <EngineValue value={engine.version} lines={2} /> },
        { label: 'Schema 版本', value: <EngineValue value={engine.schema} /> },
        { label: '数据空间', value: engine.namespace ? <EngineValue value={engine.namespace} copyable /> : <Typography.Text type="secondary">使用默认配置</Typography.Text> },
        { label: '最后写入', value: engine.lastWrite ? formatTime(engine.lastWrite) : '暂无写入记录' },
        { label: '待处理积压', value: formatBacklog(engine.backlog) },
        { label: '运行模式', value: meta.label },
        { label: '数据职责', value: engine.role },
      ]}
    />
    {engine.state === 'degraded' && <Alert type="warning" showIcon content={engine.detail || '分析引擎当前不可用，请检查连接；业务管理数据不受影响。'} />}
  </Card>;
}

function channelFormConfig(type: string, values: any) {
  if (type === 'webhook') return { url: values.url, secret: values.secret };
  if (['wecom', 'dingtalk', 'feishu'].includes(type)) return { webhook_url: values.webhook_url, secret: values.secret };
  if (type === 'email') return {
    host: values.host,
    port: values.port,
    username: values.username,
    password: values.password,
    from: values.from,
    to: String(values.to || '').split(',').map((item) => item.trim()).filter(Boolean),
    tls_mode: values.tls_mode,
  };
  return { address: values.address, network: values.network, facility: values.facility };
}

export default function System() {
  const authUser = useAuth((state) => state.user);
  const branding = usePlatformBranding().data!;
  const role = authUser?.role;
  const queryClient = useQueryClient();
  const [channelModal, setChannelModal] = useState(false);
  const [editingChannel, setEditingChannel] = useState<any>();
  const [channelType, setChannelType] = useState('webhook');
  const [userModal, setUserModal] = useState(false);
  const [editingUser, setEditingUser] = useState<any>();
  const [form] = Form.useForm();
  const [userForm] = Form.useForm();
  const users = useQuery<any[]>({ queryKey: ['users'], queryFn: () => api.get('/users'), enabled: role === 'admin' });
  const auditLogs = useQuery<any>({ queryKey: ['audit-logs'], queryFn: () => api.get('/audit-logs?page_size=50'), enabled: role === 'admin' });
  const rules = useQuery<any[]>({ queryKey: ['rules'], queryFn: () => api.get('/alert-rules') });
  const channels = useQuery<any[]>({ queryKey: ['alert-channels'], queryFn: () => api.get('/alert-channels') });
  const deliveries = useQuery<any>({ queryKey: ['alert-deliveries'], queryFn: () => api.get('/alert-deliveries?page_size=30'), refetchInterval: 5000 });
  const system = useQuery<any>({ queryKey: ['dashboard', 'summary'], queryFn: () => api.get('/dashboard/summary') });
  const analytics = useQuery<any>({
    queryKey: ['analytics', 'status'],
    queryFn: async () => {
      try { return await api.get('/analytics/status'); }
      catch (error) {
        if (axios.isAxiosError(error) && [404, 405].includes(error.response?.status || 0)) return null;
        throw error;
      }
    },
    retry: false,
    refetchInterval: 15_000,
  });

  const summary = objectValue(system.data);
  const statusPayload = objectValue(analytics.data || summary.analytics || summary.storage);
  const mysqlRaw = objectValue(statusPayload.mysql || summary.mysql);
  const reportedClickhouse = objectValue(statusPayload.engine || summary.analytics || statusPayload.clickhouse || summary.clickhouse || (statusPayload.role === 'security-analytics' ? statusPayload : null));
  const clickhouseRaw = analytics.isError ? { ...reportedClickhouse, healthy: false, error: errorMessage(analytics.error) } : reportedClickhouse;
  const mysqlEngine: EngineView = {
    name: 'MySQL',
    role: '业务数据库',
    state: engineState(mysqlRaw, Boolean(system.data), true),
    version: firstDefined(mysqlRaw.version, mysqlRaw.server_version, summary.mysql_version),
    schema: firstDefined(mysqlRaw.schema_version, mysqlRaw.migration_version, summary.mysql_schema_version),
    lastWrite: firstDefined(mysqlRaw.last_write_at, mysqlRaw.last_write, summary.mysql_last_write_at),
    backlog: optionalNumber(firstDefined(mysqlRaw.backlog, mysqlRaw.pending, summary.mysql_backlog)),
    detail: sanitizeStatusDetail(firstDefined(mysqlRaw.error, mysqlRaw.message)),
  };
  const clickhouseAvailable = Boolean(analytics.data || summary.analytics || summary.clickhouse);
  const clickhouseEngine: EngineView = {
    name: '安全分析引擎',
    role: '蜜罐安全数据分析引擎',
    state: engineState(clickhouseRaw, clickhouseAvailable, typeof clickhouseRaw.enabled === 'boolean' ? clickhouseRaw.enabled : undefined),
    version: firstDefined(clickhouseRaw.version, clickhouseRaw.server_version, statusPayload.version, summary.clickhouse_version),
    schema: firstDefined(clickhouseRaw.schema_version, clickhouseRaw.migration_version, statusPayload.schema_version, summary.clickhouse_schema_version),
    namespace: [clickhouseRaw.database, clickhouseRaw.table].filter(Boolean).join('.') || undefined,
    lastWrite: firstDefined(clickhouseRaw.last_write_at, clickhouseRaw.last_write, statusPayload.last_write_at, summary.analytics_last_write_at),
    backlog: optionalNumber(firstDefined(clickhouseRaw.backlog, clickhouseRaw.queue_depth, clickhouseRaw.pending_events, statusPayload.backlog, summary.analytics_backlog)),
    detail: sanitizeStatusDetail(firstDefined(clickhouseRaw.error, statusPayload.error, clickhouseRaw.message)),
  };

  const saveChannel = useMutation({
    mutationFn: async () => {
      const values = await form.validate();
      const payload = { name: values.name, type: values.type, enabled: values.enabled === 'true', config: channelFormConfig(values.type, values) };
      return editingChannel ? api.put(`/alert-channels/${editingChannel.id}`, payload) : api.post('/alert-channels', payload);
    },
    onSuccess: () => { Message.success(editingChannel ? '告警通道已更新' : '告警通道已创建'); setChannelModal(false); queryClient.invalidateQueries({ queryKey: ['alert-channels'] }); },
    onError: (error) => Message.error(errorMessage(error)),
  });
  const deleteChannel = useMutation({ mutationFn: (id: string) => api.delete(`/alert-channels/${id}`), onSuccess: () => { Message.success('通道已删除'); queryClient.invalidateQueries({ queryKey: ['alert-channels'] }); }, onError: (error) => Message.error(errorMessage(error)) });
  const retryDelivery = useMutation({ mutationFn: (id: string) => api.post(`/alert-deliveries/${id}/retry`), onSuccess: () => { Message.success('已重新加入投递队列'); queryClient.invalidateQueries({ queryKey: ['alert-deliveries'] }); }, onError: (error) => Message.error(errorMessage(error)) });
  const saveUser = useMutation({
    mutationFn: async () => {
      const values = await userForm.validate();
      const payload = { username: values.username, display_name: values.display_name, role: values.role, enabled: values.enabled === 'true', password: values.password || '' };
      return editingUser ? api.put(`/users/${editingUser.id}`, payload) : api.post('/users', payload);
    },
    onSuccess: () => { Message.success(editingUser ? '账号已更新，权限变更会立即撤销旧会话' : '账号已创建'); setUserModal(false); queryClient.invalidateQueries({ queryKey: ['users'] }); queryClient.invalidateQueries({ queryKey: ['audit-logs'] }); },
    onError: (error) => Message.error(errorMessage(error)),
  });
  const deleteUser = useMutation({ mutationFn: (id: string) => api.delete(`/users/${id}`), onSuccess: () => { Message.success('账号已删除并撤销全部会话'); queryClient.invalidateQueries({ queryKey: ['users'] }); queryClient.invalidateQueries({ queryKey: ['audit-logs'] }); }, onError: (error) => Message.error(errorMessage(error)) });

  const openChannel = (channel?: any) => {
    const config = channel?.config || {};
    const type = channel?.type || 'webhook';
    setEditingChannel(channel);
    setChannelType(type);
    setChannelModal(true);
    form.resetFields();
    setTimeout(() => form.setFieldsValue({
      name: channel?.name || '', type, enabled: (channel?.enabled ?? true) ? 'true' : 'false',
      ...config, to: Array.isArray(config.to) ? config.to.join(', ') : config.to,
      port: config.port || 587, tls_mode: config.tls_mode || 'starttls',
      network: config.network || 'udp', facility: config.facility ?? 16,
    }), 0);
  };
  const openUser = (user?: any) => {
    setEditingUser(user);
    setUserModal(true);
    userForm.resetFields();
    setTimeout(() => userForm.setFieldsValue({ username: user?.username || '', display_name: user?.display_name || '', role: user?.role || 'viewer', enabled: (user?.enabled ?? true) ? 'true' : 'false', password: '' }), 0);
  };
  const testChannel = async (id: string) => {
    try { const result: any = await api.post(`/alert-channels/${id}/test`); Message.success(result.message); }
    catch (error) { Message.error(errorMessage(error)); }
  };
  const updateRuleChannels = async (rule: any, channelIDs: string[]) => {
    try { await api.put(`/alert-rules/${rule.id}`, { channel_ids: channelIDs }); Message.success('规则投递通道已更新'); queryClient.invalidateQueries({ queryKey: ['rules'] }); }
    catch (error) { Message.error(errorMessage(error)); }
  };

  const channelColumns: any[] = [
    { title: '通道名称', dataIndex: 'name', render: (value: string) => <Typography.Text bold>{value}</Typography.Text> },
    { title: '类型', dataIndex: 'type', render: (value: string) => <Tag color="arcoblue">{typeLabels[value] || value}</Tag> },
    { title: '状态', dataIndex: 'enabled', render: (value: boolean) => <Tag color={value ? 'green' : 'gray'}>{value ? '启用' : '停用'}</Tag> },
    { title: '更新时间', dataIndex: 'updated_at', render: formatTime },
    { title: '操作', width: 230, render: (_: any, row: any) => <Space>
      <Button type="text" icon={<IconRefresh />} onClick={() => testChannel(row.id)}>测试</Button>
      {role === 'admin' && <Button type="text" icon={<IconEdit />} onClick={() => openChannel(row)}>编辑</Button>}
      {role === 'admin' && <Popconfirm title="确认删除该告警通道？" onOk={() => deleteChannel.mutate(row.id)}><Button type="text" status="danger" icon={<IconDelete />} /></Popconfirm>}
    </Space> },
  ];
  const deliveryColumns: any[] = [
    { title: '时间', dataIndex: 'created_at', render: formatTime, width: 180 },
    { title: '通道', render: (_: any, row: any) => <div>{row.channel_name}<br /><Typography.Text type="secondary">{typeLabels[row.channel_type] || row.channel_type}</Typography.Text></div> },
    { title: '状态', dataIndex: 'status', render: (value: string) => <Tag color={deliveryColors[value] || 'gray'}>{value}</Tag> },
    { title: '尝试', dataIndex: 'attempt', width: 80 },
    { title: '结果', dataIndex: 'last_error', render: (value: string, row: any) => value ? <Typography.Text type="secondary" ellipsis={{ showTooltip: true }}>{value}</Typography.Text> : formatTime(row.delivered_at) },
    { title: '操作', width: 90, render: (_: any, row: any) => ['failed', 'skipped'].includes(row.status) && role !== 'viewer' ? <Button type="text" onClick={() => retryDelivery.mutate(row.id)}>重试</Button> : null },
  ];
  const userColumns: any[] = [
    { title: '用户名', dataIndex: 'username', render: (value: string, row: any) => <Space><Typography.Text bold>{value}</Typography.Text>{row.id === authUser?.id && <Tag color="blue">当前账号</Tag>}</Space> },
    { title: '显示名称', dataIndex: 'display_name', render: (value: string) => value || '—' },
    { title: '角色', dataIndex: 'role', render: (value: string) => <Tag color={value === 'admin' ? 'red' : value === 'operator' ? 'arcoblue' : 'gray'}>{value === 'admin' ? '管理员' : value === 'operator' ? '运营人员' : '只读用户'}</Tag> },
    { title: '状态', render: (_: any, row: any) => row.locked_until && new Date(row.locked_until) > new Date() ? <Tag color="orange">登录锁定</Tag> : <Tag color={row.enabled ? 'green' : 'red'}>{row.enabled ? '启用' : '禁用'}</Tag> },
    { title: '最后登录', dataIndex: 'last_login_at', render: (value: string) => value ? formatTime(value) : '从未登录' },
    { title: '操作', width: 130, render: (_: any, row: any) => <Space><Button type="text" icon={<IconEdit />} onClick={() => openUser(row)}>编辑</Button>{row.id !== authUser?.id && <Popconfirm title="删除账号会立即撤销其所有会话，确认继续？" onOk={() => deleteUser.mutate(row.id)}><Button type="text" status="danger" icon={<IconDelete />} /></Popconfirm>}</Space> },
  ];
  const auditColumns: any[] = [
    { title: '时间', dataIndex: 'created_at', render: formatTime, width: 180 },
    { title: '操作者', render: (_: any, row: any) => <div><Typography.Text bold>{row.username || '系统'}</Typography.Text><br /><Typography.Text type="secondary" code>{row.ip || '—'}</Typography.Text></div> },
    { title: '动作', dataIndex: 'action', render: (value: string) => <Tag>{({ POST: '创建/执行', PUT: '更新', DELETE: '删除', READ: '敏感读取' } as Record<string, string>)[value] || value}</Tag> },
    { title: '对象', dataIndex: 'object', render: (value: string) => <Typography.Text code ellipsis={{ showTooltip: true }}>{value}</Typography.Text> },
    { title: '请求 / 变更摘要', dataIndex: 'detail', render: (value: any) => { const detail = typeof value === 'string' ? (() => { try { return JSON.parse(value); } catch { return {}; } })() : value || {}; const change = detail.change || {}; const text = [detail.request_id && `请求 ${detail.request_id}`, detail.status && `HTTP ${detail.status}`, change.object_type && `${change.object_type} · ${change.object_id}`].filter(Boolean).join(' / '); return <Typography.Text ellipsis={{ showTooltip: true }} copyable={Boolean(detail.request_id)}>{text || '已记录操作'}</Typography.Text>; } },
  ];

  return <>
    <PageHeader title="系统管理" description="双引擎运行状态、账号权限、告警规则与外发通道配置" />
    <Card title="数据引擎" className="engine-overview" extra={<Button type="text" icon={<IconRefresh />} loading={system.isFetching || analytics.isFetching} onClick={() => { system.refetch(); analytics.refetch(); }}>刷新状态</Button>}>
      <Alert type="info" content="业务数据引擎只管理节点、规则、告警流程等业务数据；安全分析引擎专门承载蜜罐攻击事件与分析数据，两个引擎职责相互隔离。" />
      <Grid.Row gutter={[16, 16]}>
        <Grid.Col xs={24} lg={12}><EngineCard engine={mysqlEngine} /></Grid.Col>
        <Grid.Col xs={24} lg={12}><EngineCard engine={clickhouseEngine} /></Grid.Col>
      </Grid.Row>
      {clickhouseEngine.state === 'unknown' && <Typography.Text type="secondary" className="engine-fallback-hint">当前 Server 尚未返回分析引擎运行指标，页面会在接口可用后自动显示版本、Schema、最后写入和积压。</Typography.Text>}
    </Card>
    <Grid.Row gutter={[16, 16]}>
      <Grid.Col xs={24} xl={14}><Card title="告警规则" className="panel-card"><Table rowKey="id" pagination={false} scroll={{ x: 920 }} data={rules.data || []} columns={[
        { title: '规则名称', dataIndex: 'name' },
        { title: '事件匹配', dataIndex: 'event_type', render: (value) => <Typography.Text title={value}>{eventTypeLabel(value)}</Typography.Text> },
        { title: '级别', dataIndex: 'level', render: (value) => <RiskTag value={value} /> },
        { title: '阈值 / 时间窗', render: (_, row: any) => `${row.threshold || 1} 次 / ${row.window_minute || 1} 分钟` },
        { title: '静默', dataIndex: 'silence_minute', render: (value) => `${value || 0} 分钟` },
        { title: '投递通道', width: 210, render: (_, row: any) => role === 'admin' ? <Select mode="multiple" allowClear placeholder="空 = 全部启用通道" value={row.channel_ids || []} onChange={(value) => updateRuleChannels(row, value)} options={(channels.data || []).map((channel) => ({ label: channel.name, value: channel.id }))} /> : ((row.channel_ids || []).length ? `${row.channel_ids.length} 个指定通道` : '全部启用通道') },
        { title: '状态', dataIndex: 'enabled', render: (value) => <Tag color={value ? 'green' : 'gray'}>{value ? '启用' : '停用'}</Tag> },
      ]} /></Card></Grid.Col>
      <Grid.Col xs={24} xl={10}><Card title="系统信息" className="panel-card">
        <div className="detail-row"><span>产品版本</span><b>{system.data?.version ? `v${system.data.version}` : '读取中'}</b></div>
        <div className="detail-row"><span>网络协议</span><Tag color="blue">{system.data?.ipv6_capable ? 'IPv4 + IPv6' : 'IPv4'}</Tag></div>
        <div className="detail-row"><span>业务数据</span><Tag color="blue">MySQL</Tag></div>
        <div className="detail-row"><span>安全分析</span><Tag color={stateMeta[clickhouseEngine.state].color}>分析引擎 · {stateMeta[clickhouseEngine.state].label}</Tag></div>
        <div className="detail-row"><span>IP 归属地</span><Tag color={system.data?.geoip_enabled ? 'green' : 'gray'}>{system.data?.geoip_enabled ? 'IPIP 城市库' : '未启用'}</Tag></div>
        <div className="detail-row"><span>实时通道</span><Tag color="green">WebSocket</Tag></div>
        <div className="detail-row"><span>告警投递</span><Tag color="purple">持久化重试</Tag></div>
        <div className="detail-row"><span>当前权限</span><Tag>{role}</Tag></div>
      </Card></Grid.Col>
    </Grid.Row>

    <Card title="告警通道" className="panel-card" extra={role === 'admin' && <Button type="primary" icon={<IconPlus />} onClick={() => openChannel()}>新建通道</Button>}>
      <Table rowKey="id" pagination={false} loading={channels.isLoading} data={channels.data || []} columns={channelColumns} scroll={{ x: 820 }} />
    </Card>
    <Card title="最近投递记录" className="panel-card" extra={<Button type="text" icon={<IconRefresh />} onClick={() => deliveries.refetch()}>刷新</Button>}>
      <Table rowKey="id" loading={deliveries.isLoading} data={deliveries.data?.items || []} columns={deliveryColumns} pagination={false} scroll={{ x: 760 }} />
    </Card>
    {role === 'admin' && <Card title="系统账号" className="panel-card" extra={<Button type="primary" icon={<IconPlus />} onClick={() => openUser()}>新建账号</Button>}><Alert type="info" content="账号禁用、角色或密码变更会立即撤销该账号的 HTTP Token 与已连接 WebSocket。" style={{ marginBottom: 12 }} /><Table rowKey="id" pagination={false} loading={users.isLoading} data={users.data || []} columns={userColumns} scroll={{ x: 860 }} /></Card>}
    {role === 'admin' && <Card title="操作审计" className="panel-card" extra={<Button type="text" icon={<IconRefresh />} loading={auditLogs.isFetching} onClick={() => auditLogs.refetch()}>刷新</Button>}><Table rowKey="id" loading={auditLogs.isLoading} data={auditLogs.data?.items || []} columns={auditColumns} pagination={{ pageSize: 20 }} scroll={{ x: 980 }} /></Card>}

    <Modal title={editingChannel ? '编辑告警通道' : '新建告警通道'} visible={channelModal} onCancel={() => setChannelModal(false)} onOk={() => saveChannel.mutate()} confirmLoading={saveChannel.isPending} style={{ width: 620 }}>
      <Form form={form} layout="vertical">
        <Grid.Row gutter={16}>
          <Grid.Col xs={24} sm={12}><Form.Item label="通道名称" field="name" rules={[{ required: true }]}><Input placeholder="例如：SOC 告警群" /></Form.Item></Grid.Col>
          <Grid.Col xs={24} sm={12}><Form.Item label="通道类型" field="type" rules={[{ required: true }]}><Select options={channelTypes} onChange={setChannelType} /></Form.Item></Grid.Col>
        </Grid.Row>
        <Form.Item label="状态" field="enabled"><Select options={[{ label: '启用', value: 'true' }, { label: '停用', value: 'false' }]} /></Form.Item>
        {channelType === 'webhook' && <>
          <Form.Item label="Webhook URL" field="url" rules={[{ required: true }]}><Input placeholder="https://example.com/hooks/..." /></Form.Item>
          <Form.Item label="HMAC 密钥" field="secret"><Input.Password placeholder="可选；编辑时留空或保持掩码表示不修改" /></Form.Item>
        </>}
        {['wecom', 'dingtalk', 'feishu'].includes(channelType) && <>
          <Form.Item label="机器人 Webhook" field="webhook_url" rules={[{ required: true }]}><Input placeholder="粘贴群机器人 Webhook 地址" /></Form.Item>
          {channelType !== 'wecom' && <Form.Item label="机器人签名密钥" field="secret"><Input.Password placeholder="建议启用签名；编辑时留空表示不修改" /></Form.Item>}
        </>}
        {channelType === 'email' && <>
          <Grid.Row gutter={16}><Grid.Col xs={24} sm={16}><Form.Item label="SMTP 主机" field="host" rules={[{ required: true }]}><Input placeholder="smtp.example.com" /></Form.Item></Grid.Col><Grid.Col xs={24} sm={8}><Form.Item label="端口" field="port" rules={[{ required: true }]}><InputNumber min={1} max={65535} style={{ width: '100%' }} /></Form.Item></Grid.Col></Grid.Row>
          <Grid.Row gutter={16}><Grid.Col xs={24} sm={12}><Form.Item label="用户名" field="username"><Input /></Form.Item></Grid.Col><Grid.Col xs={24} sm={12}><Form.Item label="密码" field="password"><Input.Password placeholder="编辑时留空表示不修改" /></Form.Item></Grid.Col></Grid.Row>
          <Form.Item label="发件人" field="from" rules={[{ required: true }]}><Input placeholder={`${branding.system_name} <alerts@example.com>`} /></Form.Item>
          <Form.Item label="收件人" field="to" rules={[{ required: true }]}><Input placeholder="soc@example.com, admin@example.com" /></Form.Item>
          <Form.Item label="TLS 模式" field="tls_mode"><Select options={[{ label: 'STARTTLS（推荐）', value: 'starttls' }, { label: '隐式 TLS', value: 'implicit' }, { label: '无 TLS', value: 'none' }]} /></Form.Item>
        </>}
        {channelType === 'syslog' && <>
          <Form.Item label="Syslog 地址" field="address" rules={[{ required: true }]}><Input placeholder="192.0.2.30:514" /></Form.Item>
          <Grid.Row gutter={16}><Grid.Col xs={24} sm={12}><Form.Item label="传输协议" field="network"><Select options={[{ label: 'UDP', value: 'udp' }, { label: 'TCP', value: 'tcp' }, { label: 'TLS', value: 'tls' }]} /></Form.Item></Grid.Col><Grid.Col xs={24} sm={12}><Form.Item label="Facility" field="facility"><InputNumber min={0} max={23} style={{ width: '100%' }} /></Form.Item></Grid.Col></Grid.Row>
        </>}
      </Form>
    </Modal>
    <Modal title={editingUser ? '编辑系统账号' : '新建系统账号'} visible={userModal} onCancel={() => { setUserModal(false); userForm.resetFields(); }} onOk={() => saveUser.mutate()} confirmLoading={saveUser.isPending} maskClosable={false} style={{ width: 560 }}>
      <Form form={userForm} layout="vertical" autoComplete="off">
        <Grid.Row gutter={16}><Grid.Col xs={24} sm={12}><Form.Item label="用户名" field="username" rules={[{ required: true, message: '请输入用户名' }]}><Input disabled={Boolean(editingUser)} maxLength={64} /></Form.Item></Grid.Col><Grid.Col xs={24} sm={12}><Form.Item label="显示名称" field="display_name"><Input maxLength={128} /></Form.Item></Grid.Col></Grid.Row>
        <Grid.Row gutter={16}><Grid.Col xs={24} sm={12}><Form.Item label="角色" field="role" rules={[{ required: true }]}><Select options={[{ label: '管理员', value: 'admin' }, { label: '运营人员', value: 'operator' }, { label: '只读用户', value: 'viewer' }]} /></Form.Item></Grid.Col><Grid.Col xs={24} sm={12}><Form.Item label="状态" field="enabled" rules={[{ required: true }]}><Select options={[{ label: '启用', value: 'true' }, { label: '禁用', value: 'false' }]} /></Form.Item></Grid.Col></Grid.Row>
        <Form.Item label={editingUser ? '重置密码' : '初始密码'} field="password" rules={editingUser ? [{ minLength: 8, message: '密码至少 8 个字符' }] : [{ required: true, minLength: 8, message: '密码至少 8 个字符' }]} extra={editingUser ? '留空保持现有密码；填写后会立即撤销该账号的全部会话。' : '长度 8-72 字节。'}><Input.Password maxLength={72} autoComplete="new-password" /></Form.Item>
      </Form>
    </Modal>
  </>;
}
