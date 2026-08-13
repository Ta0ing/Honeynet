import { useMemo, useState } from 'react';
import { Alert, Button, Card, Descriptions, Grid, Input, InputNumber, Message, Modal, Progress, Select, Space, Switch, Table, Tag, Tooltip, Typography } from '@arco-design/web-react';
import { IconCheckCircle, IconCopy, IconDelete, IconEdit, IconPlus, IconRefresh, IconSend, IconStorage, IconThunderbolt, IconUpload } from '@arco-design/web-react/icon';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import axios from 'axios';
import { api, errorMessage, formatTime } from '../api';
import PageHeader from '../components/PageHeader';
import RiskTag from '../components/RiskTag';
import { RISK_LEVELS, riskMeta } from '../presentation';
import { useAuth } from '../store';

type Pattern = { id: string; field: string; operator: string; value: string; nocase?: boolean; min_count?: number };
const emptyPattern = (): Pattern => ({ id: `feature_${Date.now()}`, field: 'raw', operator: 'contains', value: '', min_count: 1 });
const emptyRule = () => ({ key: '', name: '', description: '', severity: 'medium', enabled: true, agent_enabled: true, server_enabled: true, patterns: [emptyPattern()] as Pattern[] });

const sourceLabel = (value: string) => value === 'builtin' ? '内置规则' : value === 'custom' ? '自定义规则' : '导入规则';

const pipelineStages = [
  { key: 'agent_scan', title: 'Agent 实时扫描', description: '攻击发生时用 YARA 兼容规则就地初筛、立即打标，降低告警延迟并分摊计算。', owner: '边缘节点', icon: <IconThunderbolt /> },
  { key: 'rule_distribution', title: '规则存储与下发', description: 'Server 在 MySQL 统一保存规则和版本，修改后推送在线 Agent，离线节点重连后补齐。', owner: '管理中心', icon: <IconSend /> },
  { key: 'server_review', title: '入库统一复核', description: '事件进入安全分析引擎前，Server 使用中心统一版本兜底重扫，补偿节点规则滞后。', owner: '管理中心', icon: <IconStorage /> },
  { key: 'alert_generation', title: '告警生成', description: '仅基于中心复核结果生成规则告警，保留 Agent 与 Server 两阶段命中证据。', owner: '管理中心', icon: <IconCheckCircle /> },
];

const NO_REVISION = '尚无版本';
const INEXACT_REVISION = '精确版本不可用';

// Revisions are Unix nanoseconds and exceed JavaScript's safe integer range.
// Only the decimal string emitted by the Server is authoritative in the UI.
function pipelineRevision(value: unknown) {
  if (typeof value !== 'string') return value == null ? NO_REVISION : INEXACT_REVISION;
  const revision = value.trim();
  return revision && revision !== '0' ? revision : NO_REVISION;
}

function compactRevision(value: unknown) {
  const revision = pipelineRevision(value);
  return /^\d+$/.test(revision) && revision.length > 14 ? `r${revision.slice(0, 6)}…${revision.slice(-6)}` : revision;
}

async function copyRevision(revision: string) {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(revision);
    return;
  }
  const textarea = document.createElement('textarea');
  textarea.value = revision;
  textarea.style.position = 'fixed';
  textarea.style.opacity = '0';
  document.body.appendChild(textarea);
  textarea.select();
  const copied = document.execCommand('copy');
  textarea.remove();
  if (!copied) throw new Error('copy failed');
}

function RevisionValue({ value, copyable = false }: { value: unknown; copyable?: boolean }) {
  const revision = pipelineRevision(value);
  return <div className="rule-revision-cell">
    <Tooltip content={<div className="rule-revision-tooltip"><span>完整规则版本</span><code>{revision}</code></div>} position="top">
      <span className="rule-revision-preview" aria-label={`规则版本 ${revision}`}><Typography.Text className="rule-revision-value" code>{compactRevision(value)}</Typography.Text></span>
    </Tooltip>
    {copyable && /^\d+$/.test(revision) && <Tooltip content="复制完整规则版本" mini><Button className="rule-revision-copy" type="text" size="mini" icon={<IconCopy />} aria-label="复制完整规则版本" onClick={() => copyRevision(revision).then(() => Message.success('完整规则版本已复制')).catch(() => Message.error('复制失败，请在悬浮提示中手动复制'))} /></Tooltip>}
  </div>;
}

const syncStatusMeta: Record<string, { label: string; color: string }> = {
  synced: { label: '已同步', color: 'green' },
  stale: { label: '版本滞后', color: 'orange' },
  pending: { label: '下发中', color: 'arcoblue' },
  error: { label: '同步失败', color: 'red' },
};

export default function DetectionRules() {
  const role = useAuth((state) => state.user?.role);
  const queryClient = useQueryClient();
  const [search, setSearch] = useState('');
  const [visible, setVisible] = useState(false);
  const [editing, setEditing] = useState<any>();
  const [draft, setDraft] = useState<any>(emptyRule());
  const query = useQuery<any>({ queryKey: ['detection-rules', search], queryFn: () => api.get('/detection-rules', { params: { page_size: 200, search } }) });
  const pipeline = useQuery<any>({
    queryKey: ['detection', 'pipeline', 'status'],
    queryFn: async () => {
      try { return await api.get('/detection/pipeline/status'); }
      catch (error) {
        if (axios.isAxiosError(error) && [404, 405].includes(error.response?.status || 0)) return null;
        throw error;
      }
    },
    retry: false,
    refetchInterval: 10_000,
  });
  const refresh = () => queryClient.invalidateQueries({ queryKey: ['detection-rules'] });
  const save = useMutation({ mutationFn: () => editing ? api.put(`/detection-rules/${editing.id}`, draft) : api.post('/detection-rules', draft), onSuccess: () => { Message.success(editing ? '检测规则已更新并下发' : '检测规则已创建并下发'); setVisible(false); refresh(); }, onError: (error) => Message.error(errorMessage(error)) });
  const update = useMutation({ mutationFn: ({ id, values }: any) => api.put(`/detection-rules/${id}`, values), onSuccess: refresh, onError: (error) => Message.error(errorMessage(error)) });
  const remove = useMutation({ mutationFn: (id: string) => api.delete(`/detection-rules/${id}`), onSuccess: () => { Message.success('规则已删除'); refresh(); }, onError: (error) => Message.error(errorMessage(error)) });
  const importer = useMutation({ mutationFn: () => api.post('/detection-rules/import-builtin'), onSuccess: (report: any) => { Message.success(`内置规则同步完成：${report.total} 条，${report.pending_review} 条待审核`); refresh(); }, onError: (error) => Message.error(errorMessage(error)) });
  const items = query.data?.items || [];
  const stats = useMemo(() => ({ enabled: items.filter((item: any) => item.enabled).length, pending: items.filter((item: any) => item.validation_error).length }), [items]);
  const agentRules = pipeline.data?.agent_rules || {};
  const serverRules = pipeline.data?.server_rules || {};
  const nodeState = pipeline.data?.nodes || {};
  const totalNodes = Number(nodeState.total || 0);
  const syncedNodes = Number(nodeState.synced || 0);
  const staleNodes = Number(nodeState.stale || 0);
  const errorNodes = Number(nodeState.error || 0);
  const pendingNodes = Number(nodeState.pending || 0);
  const syncPercent = totalNodes ? Math.round(syncedNodes / totalNodes * 100) : 0;
  const nodeItems = Array.isArray(nodeState.items) ? nodeState.items : [];
  const agentRevisionText = pipelineRevision(agentRules.revision_text);
  const serverRevisionText = pipelineRevision(serverRules.revision_text);
  const ruleSetsSame = /^\d+$/.test(agentRevisionText) && agentRevisionText === serverRevisionText && Number(agentRules.rule_count || 0) === Number(serverRules.rule_count || 0);

  const open = (item?: any) => {
    setEditing(item);
    setDraft(item ? { key: item.key, name: item.name, description: item.description, severity: item.severity, enabled: item.enabled, agent_enabled: item.agent_enabled, server_enabled: item.server_enabled, patterns: Array.isArray(item.patterns) ? item.patterns : JSON.parse(item.patterns || '[]') } : emptyRule());
    setVisible(true);
  };
  const patchPattern = (index: number, values: Partial<Pattern>) => setDraft((current: any) => ({ ...current, patterns: current.patterns.map((pattern: Pattern, position: number) => position === index ? { ...pattern, ...values } : pattern) }));
  const columns: any[] = [
    { title: '规则', render: (_: any, row: any) => <div><Typography.Text bold>{row.name}</Typography.Text><br /><Typography.Text type="secondary">{row.key}</Typography.Text></div> },
    { title: '版本', dataIndex: 'revision_text', width: 180, render: (value: unknown) => <RevisionValue value={value} copyable /> },
    { title: '级别', dataIndex: 'severity', width: 90, render: (value: string) => <RiskTag value={value} /> },
    { title: '来源', width: 110, render: (_: any, row: any) => <div><Tag>{sourceLabel(row.source)}</Tag>{row.external_id && <Typography.Text type="secondary" style={{ display: 'block', marginTop: 4 }}>{row.external_id}</Typography.Text>}</div> },
    { title: '特征', width: 90, render: (_: any, row: any) => `${row.patterns?.length || 0} 项` },
    { title: '规则状态', width: 160, render: (_: any, row: any) => row.validation_error ? <Tag color="red">待审核</Tag> : <Space direction="vertical" size={2}><Tag color={row.agent_enabled ? 'green' : 'gray'}>Agent {row.agent_enabled ? '启用' : '关闭'}</Tag><Tag color={row.server_enabled ? 'arcoblue' : 'gray'}>Server {row.server_enabled ? '启用' : '关闭'}</Tag></Space> },
    { title: '总开关', width: 90, render: (_: any, row: any) => <Switch size="small" checked={row.enabled} disabled={role !== 'admin' || !!row.validation_error} onChange={(enabled) => update.mutate({ id: row.id, values: { enabled } })} /> },
    { title: '操作', width: 135, render: (_: any, row: any) => role === 'admin' && <Space><Button type="text" icon={<IconEdit />} onClick={() => open(row)} /><Button type="text" status="danger" icon={<IconDelete />} onClick={() => Modal.confirm({ title: '删除检测规则？', content: row.name, onOk: () => remove.mutate(row.id) })} /></Space> },
  ];

  return <>
    <PageHeader title="检测规则" description="统一管理 Agent 边缘初筛与 Server 中心复核规则；修改后实时推送到所有在线节点" extra={<Space wrap><Button icon={<IconRefresh />} loading={query.isFetching || pipeline.isFetching} onClick={() => { query.refetch(); pipeline.refetch(); }}>刷新</Button>{role === 'admin' && <Button icon={<IconUpload />} loading={importer.isPending} onClick={() => importer.mutate()}>同步内置规则</Button>}{role === 'admin' && <Button type="primary" icon={<IconPlus />} onClick={() => open()}>自定义规则</Button>}</Space>} />
    <Card className="panel-card detection-pipeline" title="边缘初筛 + 中心复核" extra={<Tag color="arcoblue">YARA 兼容规则</Tag>}>
      <Alert type="info" content="边缘初筛负责低延迟打标，中心复核使用统一版本兜底，避免离线或规则滞后的节点漏检；最终告警只由 Server 中心复核结果生成。" />
      <div className="pipeline-stage-grid">
        {pipelineStages.map((stage, index) => {
          const live = Array.isArray(pipeline.data?.stages) ? pipeline.data.stages.find((item: any) => item.key === stage.key) : null;
          return <div className="pipeline-stage" key={stage.key}>
            <div className="pipeline-stage-top"><span className="pipeline-index">{index + 1}</span><span className="pipeline-stage-icon">{stage.icon}</span><Tag size="small">{stage.owner}</Tag></div>
            <Typography.Title heading={6}>{live?.name || stage.title}</Typography.Title>
            <Typography.Paragraph type="secondary">{live?.purpose || stage.description}</Typography.Paragraph>
          </div>;
        })}
      </div>
      <Grid.Row gutter={[16, 16]}>
        <Grid.Col xs={24} xl={12}><div className="pipeline-status-box rule-version-panel"><Descriptions title="规则版本" border size="small" column={1} data={[
          { label: 'Agent 边缘规则', value: <RevisionValue value={agentRules.revision_text} copyable /> },
          { label: 'Agent 启用规则', value: `${agentRules.rule_count ?? stats.enabled} 条` },
          { label: 'Server 复核规则', value: <RevisionValue value={serverRules.revision_text} copyable /> },
          { label: 'Server 启用规则', value: `${serverRules.rule_count ?? stats.enabled} 条` },
          { label: '规则集关系', value: <Tooltip content="Agent 与 Server 可按规则分别启用，因此版本或数量不同不等于同步失败。节点同步状态只与 Agent 目标规则集比较。"><Tag color={ruleSetsSame ? 'green' : 'purple'}>{ruleSetsSame ? '完全一致' : '按职责存在差异'}</Tag></Tooltip> },
        ]} /></div></Grid.Col>
        <Grid.Col xs={24} xl={12}><div className="pipeline-status-box">
          <div className="node-sync-head"><div><Typography.Text bold>节点规则同步</Typography.Text><Typography.Text type="secondary"> {syncedNodes}/{totalNodes || 0} 个节点已同步</Typography.Text></div><Tag color={errorNodes ? 'red' : staleNodes || pendingNodes ? 'orange' : totalNodes ? 'green' : 'gray'}>{errorNodes ? '存在错误' : staleNodes || pendingNodes ? '待同步' : totalNodes ? '全部同步' : '暂无节点'}</Tag></div>
          <Progress percent={syncPercent} status={errorNodes ? 'error' : staleNodes || pendingNodes ? 'warning' : totalNodes ? 'success' : 'normal'} />
          <Space wrap size={[8, 8]}><Tag color="green">已同步 {syncedNodes}</Tag><Tag color="orange">滞后 {staleNodes}</Tag><Tag color="arcoblue">下发中 {pendingNodes}</Tag><Tag color="red">错误 {errorNodes}</Tag><Tag color="gray">待上传事件 {Number(nodeState.queued_events || 0)}</Tag></Space>
        </div></Grid.Col>
      </Grid.Row>
      {nodeItems.length > 0 && <Table
        style={{ marginTop: 16 }}
        rowKey="id"
        size="small"
        data={nodeItems}
        pagination={nodeItems.length > 5 ? { pageSize: 5, sizeCanChange: false } : false}
        scroll={{ x: 1040 }}
        columns={[
          { title: '节点', width: 220, render: (_: unknown, row: any) => <div><Space size={4}><Typography.Text bold>{row.name || row.id}</Typography.Text><Tag size="small" color={row.node_status === 'online' ? 'green' : row.node_status === 'degraded' ? 'orange' : 'gray'}>{row.node_status === 'online' ? '在线' : row.node_status === 'degraded' ? '降级' : '离线'}</Tag></Space><br /><Typography.Text type="secondary">{row.group || '未分组'} · Agent {row.agent_version || '未知'}</Typography.Text></div> },
          { title: '当前版本', dataIndex: 'revision_text', width: 180, render: (value: unknown) => <RevisionValue value={value} copyable /> },
          { title: '目标版本', dataIndex: 'target_revision_text', width: 180, render: (value: unknown) => <RevisionValue value={value} copyable /> },
          { title: '规则数量', dataIndex: 'rule_count', width: 95, render: (value: number) => `${Number(value || 0)} 条` },
          { title: '同步状态', dataIndex: 'sync_status', width: 110, render: (value: string) => { const meta = syncStatusMeta[value] || syncStatusMeta.pending; return <Tag color={meta.color}>{meta.label}</Tag>; } },
          { title: '最近同步', dataIndex: 'synced_at', width: 170, render: (value: string) => formatTime(value) },
          { title: '详情', render: (_: unknown, row: any) => row.error ? <Tooltip content={row.error}><Typography.Text type="error" ellipsis>{row.error}</Typography.Text></Tooltip> : <Typography.Text type="secondary">待上传 {Number(row.queued_events || 0)} 条事件</Typography.Text> },
        ]}
      />}
      {!pipeline.data && !pipeline.isLoading && <Typography.Text type="secondary" className="pipeline-fallback-hint">Server 暂未返回流水线状态，规则编辑与下发功能仍可正常使用；状态接口上线后此处会自动显示版本和节点同步率。</Typography.Text>}
    </Card>
    <Alert type="info" content={`当前加载 ${items.length} 条，启用 ${stats.enabled} 条，待人工审核 ${stats.pending} 条。响应模板与检测规则相互隔离，不会作为检测逻辑执行。`} style={{ marginBottom: 12 }} />
    <div className="filter-bar"><Input allowClear value={search} onChange={setSearch} placeholder="搜索规则名称、CVE/XVE 或规则标识" style={{ width: 360 }} /><Typography.Text type="secondary">规则匹配采用受限 contains / RE2 regexp，不执行任意脚本</Typography.Text></div>
    <div className="table-panel"><Table rowKey="id" loading={query.isLoading} data={items} columns={columns} pagination={{ pageSize: 20 }} scroll={{ x: 1120 }} /></div>
    <Modal title={editing ? '编辑检测规则' : '创建检测规则'} visible={visible} style={{ width: 860 }} confirmLoading={save.isPending} onOk={() => save.mutate()} onCancel={() => setVisible(false)}>
      {editing?.validation_error && <Alert type="warning" content={`原始规则未自动启用：${editing.validation_error}。保存后将按当前可视化特征转换为“全部命中”。`} style={{ marginBottom: 14 }} />}
      <div className="rule-form-grid"><label>规则标识<Input value={draft.key} disabled={!!editing} onChange={(key) => setDraft({ ...draft, key })} placeholder="custom:web-example" /></label><label>规则名称<Input value={draft.name} onChange={(name) => setDraft({ ...draft, name })} /></label><label>风险级别<Select value={draft.severity} onChange={(severity) => setDraft({ ...draft, severity })} options={RISK_LEVELS.map((value) => ({ label: riskMeta(value).label, value }))} /></label><label className="rule-description">描述<Input.TextArea value={draft.description} onChange={(description) => setDraft({ ...draft, description })} autoSize={{ minRows: 2, maxRows: 4 }} /></label></div>
      <div className="rule-scope"><Space><Switch checked={draft.enabled} onChange={(enabled) => setDraft({ ...draft, enabled })} />总开关<Switch checked={draft.agent_enabled} onChange={(agent_enabled) => setDraft({ ...draft, agent_enabled })} />Agent 本地匹配<Switch checked={draft.server_enabled} onChange={(server_enabled) => setDraft({ ...draft, server_enabled })} />Server 二次确认</Space></div>
      <div className="pattern-head"><Typography.Title heading={6}>匹配特征（全部命中）</Typography.Title><Button size="small" icon={<IconPlus />} onClick={() => setDraft({ ...draft, patterns: [...draft.patterns, emptyPattern()] })}>增加特征</Button></div>
      <div className="pattern-labels"><span>特征标识</span><span>数据范围</span><span>运算</span><span>匹配内容</span><span>次数</span><span>忽略大小写</span><span /></div>
      <div className="pattern-list">{draft.patterns.map((pattern: Pattern, index: number) => <div className="pattern-row" key={`${pattern.id}-${index}`}><Input value={pattern.id} onChange={(id) => patchPattern(index, { id })} placeholder="特征标识" /><Select value={pattern.field} onChange={(field) => patchPattern(index, { field })} options={[['raw', '完整请求包'], ['method', '请求方法'], ['path', '请求路径'], ['headers', '请求头'], ['body', '请求正文'], ['event_type', '事件类型'], ['service', '蜜罐服务']].map(([value, label]) => ({ value, label }))} /><Select value={pattern.operator} onChange={(operator) => patchPattern(index, { operator })} options={[{ value: 'contains', label: '包含' }, { value: 'regexp', label: '正则' }]} /><Input value={pattern.value} onChange={(value) => patchPattern(index, { value })} placeholder="匹配值 / RE2 表达式" /><InputNumber min={1} max={100} value={pattern.min_count || 1} onChange={(value) => patchPattern(index, { min_count: Number(value) || 1 })} style={{ width: 82 }} /><Switch checked={!!pattern.nocase} onChange={(nocase) => patchPattern(index, { nocase })} /><Button type="text" status="danger" icon={<IconDelete />} disabled={draft.patterns.length === 1} onClick={() => setDraft({ ...draft, patterns: draft.patterns.filter((_: Pattern, position: number) => position !== index) })} /></div>)}</div>
    </Modal>
  </>;
}
