import { useMemo, useState } from 'react';
import { Alert, Button, Card, Descriptions, Drawer, Form, Grid, Input, InputNumber, Message, Modal, Popconfirm, Progress, Select, Space, Steps, Switch, Table, Tabs, Tag, Timeline, Typography } from '@arco-design/web-react';
import { IconCheckCircle, IconExperiment, IconRefresh, IconRobot, IconSettings, IconUndo } from '@arco-design/web-react/icon';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, errorMessage, formatTime } from '../api';
import PageHeader from '../components/PageHeader';
import { useAuth } from '../store';

const labels: Record<string, string> = { summary: '结论摘要', confidence: '置信度', attack_type: '攻击类型', severity: '风险级别', ttps: 'TTP', iocs: 'IOC', attacker_profile: '攻击者画像', recommended_actions: '处置建议', evidence_basis: '证据依据', content: '模型输出' };
const proposalStatus: Record<string, { label: string; color: string }> = {
  pending_review: { label: '待人工审批', color: 'orange' }, approved: { label: '已批准，待发布', color: 'arcoblue' }, rejected: { label: '已拒绝', color: 'gray' },
  evaluation_failed: { label: '评估未通过', color: 'red' }, published: { label: '已发布', color: 'green' }, rolled_back: { label: '已回滚', color: 'purple' },
};
const traceStage: Record<string, string> = { objective: '目标确认', environment_snapshot: '环境与策略快照', read_only_evidence: '只读证据采集', feedback_snapshot: '历史反馈快照', model_proposal: '生成候选提案', static_validation: '静态规则校验', historical_replay: '隐藏集历史回放' };

function objectOf(value: any) { if (value && typeof value === 'object') return value; try { return JSON.parse(value || '{}'); } catch { return {}; } }
function resultOf(row: any) { return objectOf(row?.result) || { content: row?.summary }; }
function readable(value: any): string { if (value == null) return '—'; if (typeof value === 'string') return value; if (Array.isArray(value)) return value.map((item, index) => `${index + 1}. ${readable(item)}`).join('\n'); if (typeof value === 'object') return Object.entries(value).map(([key, item]) => `${key}: ${readable(item)}`).join('\n'); return String(value); }
function percentage(value: any) { const number = Number(value || 0); return Math.max(0, Math.min(100, Math.round(number * 1000) / 10)); }
function eventID(row: any) { return String(row?.event_id || row?.id || ''); }

export default function AIAnalysis() {
  const role = useAuth((state) => state.user?.role);
  const client = useQueryClient();
  const [ip, setIP] = useState('');
  const [selected, setSelected] = useState<any>();
  const [agentResult, setAgentResult] = useState<any>();
  const [selectedProposal, setSelectedProposal] = useState<any>();
  const [selectedRun, setSelectedRun] = useState<any>();
  const [configVisible, setConfigVisible] = useState(false);
  const [harnessVisible, setHarnessVisible] = useState(false);
  const [rollbackTarget, setRollbackTarget] = useState<any>();
  const [rollbackReason, setRollbackReason] = useState('');
  const [sampleLabels, setSampleLabels] = useState<Record<string, 'malicious' | 'benign'>>({});
  const [manualSamples, setManualSamples] = useState('');
  const [configForm] = Form.useForm();
  const [harnessForm] = Form.useForm();
  const [feedbackForm] = Form.useForm();

  const status = useQuery<any>({ queryKey: ['ai', 'status'], queryFn: () => api.get('/ai/status') });
  const capabilities = useQuery<any>({ queryKey: ['ai', 'agent', 'capabilities'], queryFn: () => api.get('/ai/agent/capabilities') });
  const config = useQuery<any>({ queryKey: ['ai', 'config'], queryFn: () => api.get('/ai/config'), enabled: role === 'admin' });
  const analyses = useQuery<any>({ queryKey: ['ai', 'analyses'], queryFn: () => api.get('/ai/analyses?page_size=100'), enabled: role !== 'viewer' });
  const runs = useQuery<any>({ queryKey: ['ai', 'harness', 'runs'], queryFn: () => api.get('/ai/harness/runs?page_size=100'), enabled: role !== 'viewer' });
  const proposals = useQuery<any>({ queryKey: ['ai', 'rule-proposals'], queryFn: () => api.get('/ai/rule-proposals?page_size=100'), enabled: role !== 'viewer' });
  const feedback = useQuery<any>({ queryKey: ['ai', 'rule-feedback'], queryFn: () => api.get('/ai/rule-feedback?page_size=200'), enabled: role !== 'viewer' });
  const rules = useQuery<any>({ queryKey: ['detection-rules', 'ai-harness'], queryFn: () => api.get('/detection-rules?page_size=200'), enabled: role !== 'viewer' });
  const recentEvents = useQuery<any>({ queryKey: ['events', 'ai-harness-samples'], queryFn: () => api.get('/events', { params: { page: 1, page_size: 50 } }), enabled: role !== 'viewer' && harnessVisible });

  const refreshHarness = async () => Promise.all([
    client.invalidateQueries({ queryKey: ['ai', 'harness', 'runs'] }), client.invalidateQueries({ queryKey: ['ai', 'rule-proposals'] }),
    client.invalidateQueries({ queryKey: ['ai', 'rule-feedback'] }), client.invalidateQueries({ queryKey: ['detection-rules'] }),
  ]);
  const profile = useMutation({ mutationFn: () => api.post(`/attackers/${encodeURIComponent(ip)}/ai-profile`), onSuccess: (item) => { Message.success('攻击者画像已生成'); client.invalidateQueries({ queryKey: ['ai', 'analyses'] }); setSelected(item); }, onError: (error) => Message.error(errorMessage(error)) });
  const agentRun = useMutation({
    mutationFn: () => {
      const available = new Set((capabilities.data?.tools || []).map((tool: any) => tool.name));
      const toolCalls = [{ name: 'attacker_timeline', input: { source_ip: ip, days: 30 } }, { name: 'recent_security_events', input: { source_ip: ip, hours: 720, limit: 25 } }, { name: 'detection_rule_summary', input: {} }].filter((call) => available.has(call.name));
      return api.post('/ai/agent/runs', { goal: `调查攻击来源 ${ip} 的活动时间线、攻击偏好、规则命中与风险，明确区分事实和推断并给出蓝队处置建议。`, context: { source_ip: ip }, tool_calls: toolCalls, max_steps: Math.min(toolCalls.length, 4) });
    },
    onSuccess: (result) => { Message.success('AI Agent 受控调查已完成'); setAgentResult(result); }, onError: (error) => Message.error(errorMessage(error)),
  });
  const createHarness = useMutation({
    mutationFn: async () => {
      const values = await harnessForm.validate();
      const samples = new Map<string, 'malicious' | 'benign'>(Object.entries(sampleLabels));
      manualSamples.split(/\r?\n/).forEach((line) => {
        const [id, label] = line.split(/[,\s]+/).map((value) => value.trim());
        if (id && (label === 'malicious' || label === 'benign')) samples.set(id, label);
      });
      const evidence = [...samples.entries()].map(([event_id, label]) => ({ event_id, label }));
      const malicious = evidence.filter((item) => item.label === 'malicious').length;
      const benign = evidence.filter((item) => item.label === 'benign').length;
      if (evidence.length < 20 || malicious < 5 || benign < 5) throw new Error('至少选择 20 条样本，其中恶意与正常样本各不少于 5 条');
      return api.post('/ai/harness/runs', { goal: values.goal, target_rule_id: values.target_rule_id || '', evidence });
    },
    onSuccess: async (item: any) => { Message.success(item.status === 'pending_review' ? '隐藏集评估通过，已进入人工审批' : '评估完成，候选未达到发布门槛'); setHarnessVisible(false); harnessForm.resetFields(); setSampleLabels({}); setManualSamples(''); await refreshHarness(); },
    onError: (error) => Message.error(errorMessage(error)),
  });
  const reviewProposal = useMutation({ mutationFn: ({ id, decision }: any) => api.post(`/ai/rule-proposals/${id}/review`, { decision }), onSuccess: async () => { Message.success('审批结果已保存'); await refreshHarness(); }, onError: (error) => Message.error(errorMessage(error)) });
  const publishProposal = useMutation({ mutationFn: (id: string) => api.post(`/ai/rule-proposals/${id}/publish`), onSuccess: async () => { Message.success('规则已发布并进入统一下发流程'); await refreshHarness(); }, onError: (error) => Message.error(errorMessage(error)) });
  const rollbackProposal = useMutation({ mutationFn: ({ id, reason }: any) => api.post(`/ai/rule-proposals/${id}/rollback`, { reason }), onSuccess: async () => { Message.success('规则版本已安全回滚'); setRollbackTarget(undefined); setRollbackReason(''); await refreshHarness(); }, onError: (error) => Message.error(errorMessage(error)) });
  const createFeedback = useMutation({ mutationFn: async () => { const values = await feedbackForm.validate(); return api.post('/ai/rule-feedback', { proposal_id: selectedProposal?.id, rule_id: selectedProposal?.published_rule_id || selectedProposal?.rule_id, ...values }); }, onSuccess: async () => { Message.success('人工反馈已加入后续优化证据'); feedbackForm.resetFields(); await refreshHarness(); }, onError: (error) => Message.error(errorMessage(error)) });

  const refreshConfig = async () => { await Promise.all([client.invalidateQueries({ queryKey: ['ai', 'status'] }), client.invalidateQueries({ queryKey: ['ai', 'config'] })]); };
  const saveConfig = useMutation({ mutationFn: async () => { const values = await configForm.validate(); return api.put('/ai/config', { enabled: !!values.enabled, provider: values.provider, base_url: values.base_url, api_key: String(values.api_key || '').trim(), model: values.model, timeout_seconds: values.timeout_seconds, send_raw_packet: !!values.send_raw_packet }); }, onSuccess: async () => { Message.success('AI 配置已保存并实时生效'); setConfigVisible(false); configForm.resetFields(); await refreshConfig(); }, onError: (error) => Message.error(errorMessage(error)) });
  const testConfig = useMutation({ mutationFn: () => api.post('/ai/config/test'), onSuccess: () => Message.success('模型连接测试成功'), onError: (error) => Message.error(errorMessage(error)) });
  const clearAPIKey = useMutation({ mutationFn: () => api.put('/ai/config', { enabled: false, clear_api_key: true }), onSuccess: async () => { Message.success('API Key 已清除，AI 模块已停用'); configForm.setFieldsValue({ enabled: false, api_key: '' }); await refreshConfig(); }, onError: (error) => Message.error(errorMessage(error)) });
  const openConfig = async () => { const response = config.data ? { data: config.data } : await config.refetch(); const values = response.data; if (!values) { Message.error('读取 AI 配置失败'); return; } configForm.resetFields(); configForm.setFieldsValue({ enabled: !!values.enabled, provider: values.provider || 'openai-compatible', base_url: values.base_url || '', api_key: '', model: values.model || '', timeout_seconds: values.timeout_seconds || 45, send_raw_packet: !!values.send_raw_packet }); setConfigVisible(true); };

  const selectedSamples = useMemo(() => Object.values(sampleLabels).reduce((result, label) => ({ ...result, [label]: (result[label] || 0) + 1 }), {} as Record<string, number>), [sampleLabels]);
  const loadFeedbackSamples = () => {
    const next = { ...sampleLabels };
    (feedback.data?.items || []).forEach((item: any) => { if (item.event_id) next[item.event_id] = item.verdict === 'false_positive' ? 'benign' : 'malicious'; });
    setSampleLabels(next);
    Message.success(`已载入 ${Object.keys(next).length} 条去重样本，可继续补充正常与恶意样本`);
  };
  const openHarness = () => { harnessForm.resetFields(); harnessForm.setFieldsValue({ goal: '基于已标注事件与历史误报反馈，生成一个最小、可解释、低误报的检测规则改进提案。' }); setSampleLabels({}); setManualSamples(''); setHarnessVisible(true); };
  const confirmReview = (row: any, decision: 'approve' | 'reject') => Modal.confirm({ title: decision === 'approve' ? '批准规则提案？' : '拒绝规则提案？', content: decision === 'approve' ? '批准只会改变提案状态，不会直接发布规则。发布仍需管理员再次确认。' : '拒绝后该候选不会进入规则运行时。', okButtonProps: decision === 'reject' ? { status: 'danger' } : undefined, onOk: () => reviewProposal.mutateAsync({ id: row.id, decision }) });
  const confirmPublish = (row: any) => Modal.confirm({ title: '发布已审批规则？', content: '系统会再次核验隐藏集评分与规则版本血缘，通过后才写入检测规则并下发节点。', onOk: () => publishProposal.mutateAsync(row.id) });

  const proposalColumns: any[] = [
    { title: '时间', dataIndex: 'created_at', render: formatTime, width: 170 },
    { title: '提案', dataIndex: 'title', ellipsis: true },
    { title: '动作', dataIndex: 'action', width: 80, render: (value: string) => value === 'update' ? '优化' : '新增' },
    { title: '隐藏集评分', width: 250, render: (_: any, row: any) => { const value = objectOf(row.evaluation); return <Space size="mini" wrap><Tag color={value.status === 'passed' ? 'green' : 'red'}>精准率 {percentage(value.precision)}%</Tag><Tag color="blue">召回率 {percentage(value.recall)}%</Tag><Tag color={Number(value.false_positive_rate) <= .1 ? 'green' : 'red'}>误报率 {percentage(value.false_positive_rate)}%</Tag></Space>; } },
    { title: '状态', dataIndex: 'status', width: 125, render: (value: string) => <Tag color={(proposalStatus[value] || {}).color || 'gray'}>{(proposalStatus[value] || {}).label || value}</Tag> },
    { title: '操作', width: 250, fixed: 'right', render: (_: any, row: any) => <Space size="mini" wrap><Button type="text" size="small" onClick={() => setSelectedProposal(row)}>详情</Button>{role === 'admin' && row.status === 'pending_review' && <><Button type="text" size="small" onClick={() => confirmReview(row, 'approve')}>批准</Button><Button type="text" status="danger" size="small" onClick={() => confirmReview(row, 'reject')}>拒绝</Button></>}{role === 'admin' && row.status === 'approved' && <Button type="text" size="small" onClick={() => confirmPublish(row)}>发布</Button>}{role === 'admin' && row.status === 'published' && <Button type="text" status="warning" size="small" icon={<IconUndo />} onClick={() => setRollbackTarget(row)}>回滚</Button>}</Space> },
  ];
  const runColumns: any[] = [
    { title: '时间', dataIndex: 'created_at', render: formatTime, width: 170 }, { title: '目标', dataIndex: 'goal', ellipsis: true },
    { title: '阶段', dataIndex: 'stage', width: 140, render: (value: string) => traceStage[value] || value },
    { title: '状态', dataIndex: 'status', width: 120, render: (value: string) => <Tag color={value === 'pending_review' ? 'orange' : value === 'evaluation_failed' || value === 'failed' ? 'red' : 'green'}>{value === 'pending_review' ? '待审批' : value === 'evaluation_failed' ? '评估未通过' : value === 'failed' ? '执行失败' : value}</Tag> },
    { title: '操作', width: 80, render: (_: any, row: any) => <Button type="text" onClick={() => setSelectedRun(row)}>轨迹</Button> },
  ];
  const analysisColumns: any[] = [
    { title: '时间', dataIndex: 'created_at', render: formatTime, width: 180 }, { title: '分析类型', render: (_: any, row: any) => <Tag color={row.kind === 'attacker-profile' ? 'purple' : 'arcoblue'}>{row.kind === 'attacker-profile' ? '攻击者画像' : '事件分析'}</Tag> },
    { title: '目标', render: (_: any, row: any) => <Typography.Text code>{row.target_id}</Typography.Text> }, { title: 'AI 结论', dataIndex: 'summary', render: (value: string) => <Typography.Text ellipsis={{ showTooltip: true }}>{value || '—'}</Typography.Text> },
    { title: '模型', render: (_: any, row: any) => `${row.provider || '—'} / ${row.model || '—'}` }, { title: '状态', dataIndex: 'status', render: (value: string) => <Tag color={value === 'completed' ? 'green' : value === 'failed' ? 'red' : 'orange'}>{value}</Tag> },
    { title: '操作', width: 90, render: (_: any, row: any) => <Button type="text" onClick={() => setSelected(row)}>查看</Button> },
  ];
  const eventColumns: any[] = [
    { title: '事件时间', dataIndex: 'ts', width: 170, render: formatTime }, { title: '事件标识', width: 220, render: (_: any, row: any) => <Typography.Text code ellipsis={{ showTooltip: true }}>{eventID(row)}</Typography.Text> },
    { title: '攻击来源', dataIndex: 'src_ip', width: 145 }, { title: '服务', dataIndex: 'service', width: 110 },
    { title: '人工标签', width: 150, fixed: 'right', render: (_: any, row: any) => <Select value={sampleLabels[eventID(row)]} allowClear placeholder="选择标签" style={{ width: 130 }} options={[{ value: 'malicious', label: '恶意样本' }, { value: 'benign', label: '正常样本' }]} onChange={(value) => setSampleLabels((current) => { const next = { ...current }; if (value) next[eventID(row)] = value; else delete next[eventID(row)]; return next; })} /> },
  ];
  const selectedResult = resultOf(selected);
  const proposalEvaluation = objectOf(selectedProposal?.evaluation);
  const proposalCandidate = objectOf(selectedProposal?.candidate);
  const runTrace = objectOf(selectedRun?.trace);

  return <>
    <PageHeader title="AI Agent" description="采用可审计 Harness 架构：目标、只读工具、证据轨迹、隐藏集评估、人工审批、发布与反馈回滚形成持续改进闭环" extra={<Space wrap>{role === 'admin' && <Button type="primary" icon={<IconSettings />} onClick={openConfig}>在线配置</Button>}<Button icon={<IconRefresh />} loading={analyses.isFetching || capabilities.isFetching || proposals.isFetching} onClick={() => { analyses.refetch(); capabilities.refetch(); runs.refetch(); proposals.refetch(); feedback.refetch(); }}>刷新</Button></Space>} />
    {!status.data?.enabled && <Alert type="warning" content={role === 'admin' ? 'AI 模块当前未启用，可通过“在线配置”安全接入模型，保存后无需重启 Server。' : 'AI 模块当前未启用，请联系管理员完成模型配置。'} style={{ marginBottom: 14 }} />}
    {status.data?.enabled && !status.data?.configured && <Alert type="error" content="AI 模块已启用，但提供商配置不完整。" style={{ marginBottom: 14 }} />}
    <Card title="Harness Engineering 安全闭环" className="panel-card" extra={<Tag color="arcoblue">模型无线上写权限</Tag>}>
      <Steps current={3} size="small" className="ai-agent-steps"><Steps.Step title="目标与工具" description="目标固定、白名单只读工具、预算上限" /><Steps.Step title="证据与轨迹" description="环境快照、证据摘要、持久执行轨迹" /><Steps.Step title="离线评估" description="训练集与隐藏集隔离、静态校验、历史回放" /><Steps.Step title="人工发布" description="审批后发布、反馈监测、版本可回滚" /></Steps>
      <Alert type="info" content="每次任务至少需要 20 条人工标注样本。模型只能看到训练集；隐藏评估集不会发送给模型。精准率需 ≥80%、召回率 ≥60%、误报率需 ≤10%，通过后仍必须由管理员审批和再次发布确认。" />
    </Card>
    <Tabs type="card-gutter" defaultActiveTab="harness" className="ai-harness-tabs">
      <Tabs.TabPane key="harness" title="规则持续优化">
        <div className="ai-harness-metrics"><Card><Typography.Text type="secondary">运行任务</Typography.Text><Typography.Title heading={3}>{runs.data?.total || 0}</Typography.Title></Card><Card><Typography.Text type="secondary">待人工审批</Typography.Text><Typography.Title heading={3}>{(proposals.data?.items || []).filter((item: any) => item.status === 'pending_review').length}</Typography.Title></Card><Card><Typography.Text type="secondary">已发布规则提案</Typography.Text><Typography.Title heading={3}>{(proposals.data?.items || []).filter((item: any) => item.status === 'published').length}</Typography.Title></Card><Card><Typography.Text type="secondary">人工反馈</Typography.Text><Typography.Title heading={3}>{feedback.data?.total || 0}</Typography.Title></Card></div>
        <Card title="检测规则改进提案" className="panel-card" extra={role !== 'viewer' && <Button type="primary" icon={<IconExperiment />} disabled={!capabilities.data?.ready} onClick={openHarness}>新建评估任务</Button>}>
          <Table rowKey="id" loading={proposals.isLoading} data={proposals.data?.items || []} columns={proposalColumns} pagination={{ pageSize: 10 }} scroll={{ x: 1150 }} />
        </Card>
        <Card title="持久执行轨迹" className="panel-card"><Table rowKey="id" loading={runs.isLoading} data={runs.data?.items || []} columns={runColumns} pagination={{ pageSize: 10 }} /></Card>
      </Tabs.TabPane>
      <Tabs.TabPane key="investigation" title="受控威胁调查">
        <div className="ai-status-grid"><Card title="模型与智能体状态" className="panel-card"><Descriptions column={1} data={[{ label: '模型状态', value: status.data?.enabled ? '已启用' : '未启用' }, { label: 'Agent 运行模式', value: <Tag color="blue">可审计 Harness</Tag> }, { label: '已注册只读工具', value: `${capabilities.data?.tools?.length || 0} 个` }, { label: '工具预算', value: `${capabilities.data?.harness?.tool_budget || 3} 步/任务` }, { label: 'API Key', value: status.data?.has_api_key ? <Tag color="green">已安全配置</Tag> : <Tag color="gray">未配置</Tag> }, { label: '提供商 / 模型', value: `${status.data?.provider || '—'} / ${status.data?.model || '—'}` }, { label: '原始请求包外发', value: status.data?.send_raw_packet ? '允许（请注意数据合规）' : '关闭（默认）' }]} /></Card><Card title="攻击者研判任务" className="panel-card"><Typography.Paragraph type="secondary">输入 IPv4 / IPv6。AI Agent 会按白名单读取攻击者时间线、最近事件与检测规则，并保留可复核工具轨迹。</Typography.Paragraph><Space wrap style={{ width: '100%' }}><Input value={ip} onChange={setIP} placeholder="攻击来源 IPv4 / IPv6" style={{ flex: 1, minWidth: 220 }} /><Button disabled={!status.data?.enabled || role === 'viewer' || !ip} loading={profile.isPending} onClick={() => profile.mutate()}>快速画像</Button><Button type="primary" icon={<IconRobot />} disabled={!capabilities.data?.ready || role === 'viewer' || !ip} loading={agentRun.isPending} onClick={() => agentRun.mutate()}>受控调查</Button></Space><Typography.Paragraph type="secondary" style={{ marginTop: 16 }}>当前工具：{(capabilities.data?.tools || []).map((tool: any) => tool.description).join('；') || '加载中'}。</Typography.Paragraph></Card></div>
      </Tabs.TabPane>
      <Tabs.TabPane key="analyses" title="分析记录"><div className="table-panel"><Table rowKey="id" loading={analyses.isLoading} data={analyses.data?.items || []} columns={analysisColumns} pagination={{ pageSize: 20 }} /></div></Tabs.TabPane>
    </Tabs>

    <Modal title="新建检测规则评估任务" visible={harnessVisible} onCancel={() => setHarnessVisible(false)} onOk={() => createHarness.mutate()} confirmLoading={createHarness.isPending} maskClosable={false} unmountOnExit style={{ width: 1000, maxWidth: 'calc(100vw - 32px)' }}>
      <Alert type="warning" content="样本标签决定评估基准，请由分析人员确认。系统会按事件标识哈希进行分层切分，模型不会看到隐藏评估集；任务只生成提案，不直接修改检测规则。" style={{ marginBottom: 16 }} />
      <Form form={harnessForm} layout="vertical"><Form.Item field="goal" label="优化目标" rules={[{ required: true, message: '请输入明确的规则优化目标' }]}><Input.TextArea maxLength={2000} autoSize={{ minRows: 2, maxRows: 4 }} /></Form.Item><Form.Item field="target_rule_id" label="目标检测规则（留空表示新增规则）"><Select allowClear showSearch placeholder="选择要优化的检测规则" options={(rules.data?.items || []).map((item: any) => ({ value: item.id, label: `${item.name}（${item.key}）` }))} /></Form.Item></Form>
      <Space wrap style={{ marginBottom: 10 }}><Tag color="red">恶意样本 {selectedSamples.malicious || 0}</Tag><Tag color="green">正常样本 {selectedSamples.benign || 0}</Tag><Tag>已选 {Object.keys(sampleLabels).length}</Tag><Button size="small" onClick={loadFeedbackSamples} disabled={!feedback.data?.items?.length}>载入历史人工反馈</Button></Space>
      <Table rowKey={eventID} loading={recentEvents.isLoading} data={recentEvents.data?.items || []} columns={eventColumns} pagination={{ pageSize: 8 }} scroll={{ x: 850 }} />
      <Typography.Paragraph type="secondary" style={{ margin: '12px 0 6px' }}>也可按行粘贴不在当前列表中的样本：<Typography.Text code>事件标识 malicious</Typography.Text> 或 <Typography.Text code>事件标识 benign</Typography.Text></Typography.Paragraph>
      <Input.TextArea value={manualSamples} onChange={setManualSamples} placeholder={'event-id-1 malicious\nevent-id-2 benign'} autoSize={{ minRows: 2, maxRows: 5 }} />
    </Modal>

    <Modal title="AI 模型在线配置" visible={configVisible} onCancel={() => { setConfigVisible(false); configForm.resetFields(); }} onOk={() => saveConfig.mutate()} confirmLoading={saveConfig.isPending} maskClosable={false} unmountOnExit style={{ width: 680, maxWidth: 'calc(100vw - 32px)' }}>
      <Alert type="info" content="API Key 采用加密存储，保存后不会在页面或接口中回显。再次编辑时留空即可保持现有密钥。" style={{ marginBottom: 16 }} />
      <Form form={configForm} layout="vertical" autoComplete="off"><Grid.Row gutter={[16, 0]}><Grid.Col xs={24} md={12}><Form.Item label="启用 AI 分析" field="enabled" triggerPropName="checked"><Switch checkedText="启用" uncheckedText="停用" /></Form.Item></Grid.Col><Grid.Col xs={24} md={12}><Form.Item label="允许发送原始请求包" field="send_raw_packet" triggerPropName="checked" extra="默认关闭；规则优化 Harness 永不外发原始请求包"><Switch checkedText="允许" uncheckedText="关闭" /></Form.Item></Grid.Col><Grid.Col xs={24} md={12}><Form.Item label="提供商" field="provider" rules={[{ required: true, message: '请输入提供商名称' }]}><Input placeholder="deepseek / glm / openai-compatible" maxLength={64} /></Form.Item></Grid.Col><Grid.Col xs={24} md={12}><Form.Item label="模型" field="model"><Input placeholder="deepseek-chat / glm-4" maxLength={128} /></Form.Item></Grid.Col><Grid.Col span={24}><Form.Item label="Base URL" field="base_url" extra="填写到版本路径，系统会自动追加 /chat/completions"><Input placeholder="https://api.deepseek.com/v1" maxLength={1024} /></Form.Item></Grid.Col><Grid.Col xs={24} md={16}><Form.Item label="API Key" field="api_key" extra={config.data?.has_api_key ? '密钥已配置；留空保持不变' : '密钥尚未配置'}><Input.Password placeholder={config.data?.has_api_key ? '已安全配置，留空保持不变' : '输入 API Key'} autoComplete="new-password" /></Form.Item></Grid.Col><Grid.Col xs={24} md={8}><Form.Item label="请求超时（秒）" field="timeout_seconds" rules={[{ required: true }]}><InputNumber min={1} max={300} precision={0} style={{ width: '100%' }} /></Form.Item></Grid.Col></Grid.Row><Space wrap><Button onClick={() => testConfig.mutate()} loading={testConfig.isPending} disabled={!config.data?.configured}>测试已保存配置</Button>{config.data?.has_api_key && <Popconfirm title="确认清除 API Key？清除后将自动停用 AI 模块。" onOk={() => clearAPIKey.mutate()}><Button status="danger" loading={clearAPIKey.isPending}>清除 API Key</Button></Popconfirm>}</Space></Form>
    </Modal>

    <Drawer width={760} title="规则提案与隐藏集评估" visible={!!selectedProposal} onCancel={() => setSelectedProposal(undefined)} footer={null}>{selectedProposal && <div className="ai-result"><Descriptions border column={{ xs: 1, sm: 2 }} data={[{ label: '提案状态', value: (proposalStatus[selectedProposal.status] || {}).label || selectedProposal.status }, { label: '动作', value: selectedProposal.action === 'update' ? '优化现有规则' : '创建新规则' }, { label: '训练样本', value: proposalEvaluation.training_sample_count || '—' }, { label: '隐藏评估样本', value: proposalEvaluation.sample_count || '—' }, { label: '创建时间', value: formatTime(selectedProposal.created_at) }, { label: '发布版本', value: selectedProposal.published_revision_text || '—' }]} /><Typography.Title heading={6}>隐藏集评分</Typography.Title><div className="ai-eval-grid"><Card><Progress type="circle" percent={percentage(proposalEvaluation.precision)} /><span>精准率</span></Card><Card><Progress type="circle" percent={percentage(proposalEvaluation.recall)} /><span>召回率</span></Card><Card><Progress type="circle" status={Number(proposalEvaluation.false_positive_rate) <= .1 ? 'success' : 'error'} percent={percentage(proposalEvaluation.false_positive_rate)} /><span>误报率（越低越好）</span></Card></div>{proposalEvaluation.baseline_evaluated && <Alert type={proposalEvaluation.status === 'passed' ? 'success' : 'warning'} content={proposalEvaluation.improvement ? `相较现有版本：${proposalEvaluation.improvement}。基线精准率 ${percentage(proposalEvaluation.baseline_precision)}%，召回率 ${percentage(proposalEvaluation.baseline_recall)}%，误报率 ${percentage(proposalEvaluation.baseline_false_positive_rate)}%。` : `候选必须在隐藏集上不退化且至少改善一项指标。现有基线：精准率 ${percentage(proposalEvaluation.baseline_precision)}%，召回率 ${percentage(proposalEvaluation.baseline_recall)}%，误报率 ${percentage(proposalEvaluation.baseline_false_positive_rate)}%。`} style={{ marginTop: 12 }} />}{proposalEvaluation.reason && <Alert type="error" content={proposalEvaluation.reason} style={{ marginTop: 12 }} />}<Typography.Title heading={6}>改进理由</Typography.Title><Typography.Paragraph>{selectedProposal.rationale || '—'}</Typography.Paragraph><Typography.Title heading={6}>候选规则</Typography.Title><pre>{readable(proposalCandidate)}</pre><Typography.Title heading={6}>人工反馈</Typography.Title><Form form={feedbackForm} layout="vertical"><Grid.Row gutter={12}><Grid.Col xs={24} md={12}><Form.Item field="event_id" label="关联事件标识" rules={[{ required: true }]}><Input /></Form.Item></Grid.Col><Grid.Col xs={24} md={12}><Form.Item field="verdict" label="分析结论" rules={[{ required: true }]}><Select options={[{ value: 'true_positive', label: '正确命中' }, { value: 'false_positive', label: '误报' }, { value: 'false_negative', label: '漏报' }]} /></Form.Item></Grid.Col></Grid.Row><Form.Item field="comment" label="说明"><Input.TextArea maxLength={1024} /></Form.Item><Button type="primary" loading={createFeedback.isPending} onClick={() => createFeedback.mutate()}>加入持续改进证据</Button></Form>{selectedProposal.rollback_reason && <Alert type="warning" content={`回滚原因：${selectedProposal.rollback_reason}`} style={{ marginTop: 16 }} />}</div>}</Drawer>
    <Drawer width={720} title="Harness 持久执行轨迹" visible={!!selectedRun} onCancel={() => setSelectedRun(undefined)} footer={null}>{selectedRun && <div className="ai-result"><Descriptions column={1} data={[{ label: '目标', value: selectedRun.goal }, { label: '证据摘要', value: <Typography.Text code copyable>{selectedRun.evidence_digest}</Typography.Text> }, { label: '状态 / 阶段', value: `${selectedRun.status} / ${traceStage[selectedRun.stage] || selectedRun.stage}` }]} /><Timeline>{(Array.isArray(runTrace) ? runTrace : []).map((step: any) => <Timeline.Item key={`${step.index}-${step.stage}`} label={`${step.index}. ${traceStage[step.stage] || step.stage}`} dotColor={step.status === 'completed' ? 'green' : 'red'}><pre>{readable(step.summary)}</pre></Timeline.Item>)}</Timeline>{selectedRun.error && <Alert type="error" content={selectedRun.error} />}</div>}</Drawer>
    <Modal title="回滚已发布规则" visible={!!rollbackTarget} onCancel={() => { setRollbackTarget(undefined); setRollbackReason(''); }} onOk={() => rollbackProposal.mutate({ id: rollbackTarget?.id, reason: rollbackReason })} confirmLoading={rollbackProposal.isPending} okButtonProps={{ status: 'danger', disabled: !rollbackReason.trim() }} maskClosable={false}><Alert type="warning" content="仅当规则发布后未被再次人工编辑时才能回滚，以免覆盖后续变更。新增规则会被停用，优化规则会恢复审批前基线。" style={{ marginBottom: 12 }} /><Input.TextArea value={rollbackReason} onChange={setRollbackReason} maxLength={1024} placeholder="请填写可审计的回滚原因" /></Modal>
    <Drawer width={720} title="AI 分析详情" visible={!!selected} onCancel={() => setSelected(undefined)} footer={null}>{selected && <div className="ai-result"><Descriptions column={1} data={[{ label: '分析目标', value: `${selected.target_type} / ${selected.target_id}` }, { label: '模型', value: `${selected.provider} / ${selected.model}` }, { label: '时间', value: formatTime(selected.created_at) }]} />{Object.entries(selectedResult).map(([key, value]) => <section key={key}><Typography.Title heading={6}>{labels[key] || key}</Typography.Title><pre>{readable(value)}</pre></section>)}{selected.error && <Alert type="error" content={selected.error} />}</div>}</Drawer>
    <Drawer width={760} title="AI Agent 调查轨迹" visible={!!agentResult} onCancel={() => setAgentResult(undefined)} footer={null}>{agentResult && <div className="ai-result"><Alert type="success" content={`已完成 ${agentResult.steps?.length || 0} 个白名单只读工具步骤；证据与模型结论均可复核。`} /><Typography.Title heading={6}>工具执行轨迹</Typography.Title>{(agentResult.steps || []).map((step: any) => <Card key={`${step.index}-${step.tool}`} size="small" title={`${step.index}. ${step.tool}`} extra={<Tag color={step.error ? 'red' : 'green'}>{step.error ? '失败' : '已完成'}</Tag>} style={{ marginBottom: 10 }}><pre>{readable(step.error || step.output)}</pre></Card>)}<Typography.Title heading={6}>模型结论</Typography.Title><pre>{agentResult.response?.json ? readable(agentResult.response.json) : agentResult.response?.content || '模型未返回内容'}</pre></div>}</Drawer>
  </>;
}
