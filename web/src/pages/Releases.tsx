import { useMemo, useState } from 'react';
import { Alert, Button, Card, Form, Grid, Input, InputNumber, Message, Modal, Popconfirm, Progress, Select, Space, Table, Tag, Tooltip, Typography } from '@arco-design/web-react';
import { IconCopy, IconPause, IconPlayArrow, IconPlus, IconRefresh, IconStop } from '@arco-design/web-react/icon';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, errorMessage, formatTime } from '../api';
import PageHeader from '../components/PageHeader';
import { useAuth } from '../store';

type StatusMeta = { label: string; color: string };

const statusMeta: Record<string, StatusMeta> = {
  active: { label: '可发布', color: 'green' },
  running: { label: '发布中', color: 'arcoblue' },
  paused: { label: '已暂停', color: 'orange' },
  completed: { label: '已完成', color: 'green' },
  cancelled: { label: '已取消', color: 'gray' },
  pending: { label: '等待下发', color: 'gray' },
  deploying: { label: '下载并验签', color: 'blue' },
  restarting: { label: '重启确认中', color: 'purple' },
  succeeded: { label: '升级成功', color: 'green' },
  failed: { label: '升级失败', color: 'red' },
  rolled_back: { label: '已自动回滚', color: 'orange' },
  rollback_failed: { label: '回滚失败', color: 'red' },
};

const metaFor = (status: string) => statusMeta[status] || { label: status || '未知', color: 'gray' };
const formatBytes = (value: number) => value >= 1024 * 1024 ? `${(value / 1024 / 1024).toFixed(1)} MiB` : `${Math.max(0, Math.round(value / 1024))} KiB`;
const shortDigest = (value: string) => value ? `${value.slice(0, 10)}…${value.slice(-8)}` : '—';

async function copyText(value: string, success: string) {
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(value);
    } else {
      const textarea = document.createElement('textarea');
      textarea.value = value;
      textarea.style.position = 'fixed';
      textarea.style.opacity = '0';
      document.body.appendChild(textarea);
      textarea.select();
      const copied = document.execCommand('copy');
      textarea.remove();
      if (!copied) throw new Error('copy failed');
    }
    Message.success(success);
  } catch {
    Message.error('复制失败，请手动选择复制');
  }
}

export default function Releases() {
  const role = useAuth((state) => state.user?.role);
  const queryClient = useQueryClient();
  const [visible, setVisible] = useState(false);
  const [scanVisible, setScanVisible] = useState(false);
  const [selectedNodeIDs, setSelectedNodeIDs] = useState<string[]>([]);
  const [selectedGroups, setSelectedGroups] = useState<string[]>([]);
  const [form] = Form.useForm();
  const [scanForm] = Form.useForm();
  const releases = useQuery<any[]>({ queryKey: ['agent-releases'], queryFn: () => api.get('/agent-releases') });
  const rollouts = useQuery<any[]>({ queryKey: ['upgrade-rollouts'], queryFn: () => api.get('/upgrade-rollouts'), refetchInterval: 3000 });
  const nodes = useQuery<any>({ queryKey: ['nodes', 'upgrade-options'], queryFn: () => api.get('/nodes?page_size=200') });
  const allNodeItems = useMemo(() => (nodes.data?.items || []).filter((node: any) => node.status !== 'revoked'), [nodes.data]);
  const nodeItems = useMemo(() => allNodeItems.filter((node: any) => node.os === 'linux'), [allNodeItems]);
  const groupOptions = useMemo(() => Array.from(new Set<string>(nodeItems.map((node: any) => String(node.group || '').trim()).filter(Boolean))).filter((group) => allNodeItems.filter((node: any) => node.group === group).every((node: any) => node.os === 'linux')).sort().map((group) => ({ label: group, value: group })), [allNodeItems, nodeItems]);
  const targetCount = useMemo(() => new Set(nodeItems.filter((node: any) => selectedNodeIDs.includes(node.id) || selectedGroups.includes(node.group)).map((node: any) => node.id)).size, [nodeItems, selectedGroups, selectedNodeIDs]);

  const scan = useMutation({
    mutationFn: async () => {
      const values = await scanForm.validate();
      return api.post('/agent-releases/scan', values);
    },
    onSuccess: () => {
      Message.success('Agent 构建已完成 SHA-256 计算和 Ed25519 签名');
      setScanVisible(false);
      scanForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['agent-releases'] });
    },
    onError: (error) => Message.error(errorMessage(error)),
  });
  const create = useMutation({
    mutationFn: async () => {
      const values = await form.validate();
      if (!(values.node_ids?.length || values.group_names?.length)) throw new Error('请至少选择一个节点或节点组');
      return api.post('/upgrade-rollouts', { ...values, node_ids: values.node_ids || [], group_names: values.group_names || [], strategy: 'canary' });
    },
    onSuccess: () => {
      Message.success('Agent 灰度发布已启动');
      setVisible(false);
      setSelectedNodeIDs([]);
      setSelectedGroups([]);
      form.resetFields();
      queryClient.invalidateQueries({ queryKey: ['upgrade-rollouts'] });
    },
    onError: (error) => Message.error(errorMessage(error)),
  });
  const action = async (id: string, value: 'resume' | 'pause' | 'cancel') => {
    try {
      await api.post(`/upgrade-rollouts/${id}/${value}`);
      queryClient.invalidateQueries({ queryKey: ['upgrade-rollouts'] });
    } catch (error) { Message.error(errorMessage(error)); }
  };

  const buildColumns: any[] = [
    { title: '平台', width: 130, render: (_: unknown, row: any) => <Tag>{row.os}/{row.arch}</Tag> },
    { title: '制品', dataIndex: 'filename' },
    { title: '大小', dataIndex: 'size', width: 100, render: (value: number) => formatBytes(Number(value || 0)) },
    { title: 'SHA-256', dataIndex: 'sha256', width: 250, render: (value: string) => <Space size={4}><Tooltip content={value}><Typography.Text code>{shortDigest(value)}</Typography.Text></Tooltip><Button type="text" size="mini" icon={<IconCopy />} aria-label="复制 SHA-256" onClick={() => copyText(value, 'SHA-256 已复制')} /></Space> },
    { title: '签名', dataIndex: 'signature', width: 130, render: (value: string) => <Tag color="green">Ed25519 已签名</Tag> },
  ];
  const releaseColumns: any[] = [
    { title: '版本', width: 150, render: (_: unknown, row: any) => <div><Typography.Text bold>{row.release.version}</Typography.Text><br /><Tag color={row.available === false ? 'red' : metaFor(row.release.status).color}>{row.available === false ? '制品不可用' : metaFor(row.release.status).label}</Tag></div> },
    { title: '发布说明', render: (_: unknown, row: any) => row.release.notes || <Typography.Text type="secondary">未填写</Typography.Text> },
    { title: '签名密钥', width: 190, render: (_: unknown, row: any) => <Tooltip content={row.release.key_id}><Typography.Text code>{row.release.key_id}</Typography.Text></Tooltip> },
    { title: '平台构建', width: 170, render: (_: unknown, row: any) => <Tag color={row.available === false ? 'red' : 'arcoblue'}>{row.available_builds ?? row.builds.length}/{row.builds.length} 个可用</Tag> },
    { title: '完整性', width: 210, render: (_: unknown, row: any) => <Tag color="green">SHA-256 + Ed25519</Tag> },
    { title: '创建时间', width: 180, render: (_: unknown, row: any) => formatTime(row.release.created_at) },
  ];

  const taskColumns: any[] = [
    { title: '节点', width: 210, render: (_: unknown, task: any) => <div><Typography.Text bold>{task.node?.name || task.node_id}</Typography.Text><br /><Typography.Text type="secondary">{task.node?.group || '未分组'} · {task.node?.os || '未知'}/{task.node?.arch || '未知'}</Typography.Text></div> },
    { title: '版本变化', width: 180, render: (_: unknown, task: any) => <Typography.Text code>{task.from_version || '未知'} → {task.target_version}</Typography.Text> },
    { title: '波次', dataIndex: 'wave', width: 80, render: (value: number) => `第 ${Number(value) + 1} 波` },
    { title: '状态', dataIndex: 'status', width: 130, render: (value: string) => { const meta = metaFor(value); return <Tag color={meta.color}>{meta.label}</Tag>; } },
    { title: '尝试', dataIndex: 'attempt', width: 70, render: (value: number) => `${Number(value || 0)} 次` },
    { title: '确认版本', dataIndex: 'confirmed_version', width: 120, render: (value: string) => value || '—' },
    { title: '完成时间', dataIndex: 'completed_at', width: 180, render: (value: string) => formatTime(value) },
    { title: '结果', render: (_: unknown, task: any) => task.last_error ? <Tooltip content={task.last_error}><Typography.Text type="error" ellipsis>{task.last_error}</Typography.Text></Tooltip> : <Typography.Text type="secondary">—</Typography.Text> },
  ];
  const rolloutColumns: any[] = [
    { title: '发布任务', width: 210, render: (_: unknown, row: any) => <div><Typography.Text bold>{row.rollout.name}</Typography.Text><br /><Typography.Text type="secondary">目标 Agent {row.rollout.version}</Typography.Text></div> },
    { title: '状态', width: 110, render: (_: unknown, row: any) => { const meta = metaFor(row.rollout.status); return <Tag color={meta.color}>{meta.label}</Tag>; } },
    { title: '当前波次', width: 95, render: (_: unknown, row: any) => `第 ${Number(row.rollout.current_wave || 0) + 1} 波` },
    { title: '总体进度', width: 250, render: (_: unknown, row: any) => {
      const total = Number(row.progress?.total ?? row.tasks.length);
      const succeeded = Number(row.counts.succeeded || 0);
      const failures = Number(row.counts.failed || 0) + Number(row.counts.rolled_back || 0) + Number(row.counts.rollback_failed || 0);
      const status = failures ? 'error' : row.rollout.status === 'completed' ? 'success' : 'normal';
      return <div><Progress size="small" percent={Number(row.progress?.percent || 0)} status={status} /><Typography.Text type="secondary">成功 {succeeded}/{total}{failures ? ` · 异常 ${failures}` : ''}</Typography.Text></div>;
    } },
    { title: '开始时间', width: 180, render: (_: unknown, row: any) => formatTime(row.rollout.started_at) },
    { title: '操作', width: 180, fixed: 'right', render: (_: unknown, row: any) => role === 'admin' && <Space>
      {row.rollout.status === 'running' && <Button type="text" icon={<IconPause />} onClick={() => action(row.rollout.id, 'pause')}>暂停</Button>}
      {row.rollout.status === 'paused' && <Button type="text" icon={<IconPlayArrow />} onClick={() => action(row.rollout.id, 'resume')}>重试并继续</Button>}
      {['running', 'paused'].includes(row.rollout.status) && <Popconfirm title="确认取消尚未开始的升级任务？" onOk={() => action(row.rollout.id, 'cancel')}><Button type="text" status="danger" icon={<IconStop />} aria-label="取消发布" /></Popconfirm>}
    </Space> },
  ];

  return <>
    <PageHeader title="Agent 远程升级" description="由 Server 统一发布签名制品；Linux 节点支持灰度下发、健康确认和进程外自动回滚" extra={role === 'admin' && <Space>
      <Button icon={<IconRefresh />} loading={scan.isPending} onClick={() => setScanVisible(true)}>扫描并签名构建</Button>
      <Button type="primary" icon={<IconPlus />} disabled={!releases.data?.some((item) => item.available !== false)} onClick={() => setVisible(true)}>新建灰度发布</Button>
    </Space>} />
    <Alert type="info" content="升级包由 Server 使用 Ed25519 签名。Linux Agent 由不可随升级替换的稳定守护器监督；新版本无法启动、提前退出或 2 分钟内未建立健康控制连接时，会恢复原二进制并回报结果。Windows 构建继续提供签名离线安装包，本版本不允许进入远程灰度任务。" style={{ marginBottom: 16 }} />
    <Card title="已签名 Agent 版本" className="panel-card"><Table rowKey={(row) => row.release.id} loading={releases.isLoading} data={releases.data || []} columns={releaseColumns} pagination={false} scroll={{ x: 1080 }} expandedRowRender={(row) => <Table rowKey="id" size="small" data={row.builds || []} columns={buildColumns} pagination={false} scroll={{ x: 900 }} />} /></Card>
    <Card title="灰度发布与回滚记录" className="panel-card"><Table rowKey={(row) => row.rollout.id} loading={rollouts.isLoading} data={rollouts.data || []} columns={rolloutColumns} pagination={{ pageSize: 10 }} scroll={{ x: 1080 }} expandedRowRender={(row) => <Table rowKey="id" size="small" data={row.tasks || []} columns={taskColumns} pagination={false} scroll={{ x: 1250 }} />} /></Card>

    <Modal title="扫描并签名 Agent 构建" visible={scanVisible} onCancel={() => setScanVisible(false)} onOk={() => scan.mutate()} confirmLoading={scan.isPending} style={{ width: 560 }}>
      <Alert type="warning" content="Server 将扫描 downloads_dir 中现有的各平台 Agent 二进制。若同版本存在进行中或暂停的发布任务，将拒绝重新签名。" style={{ marginBottom: 16 }} />
      <Form form={scanForm} layout="vertical">
        <Form.Item label="版本号" field="version" extra="留空时使用当前 Server 发布版本"><Input placeholder="例如 0.22.1" maxLength={64} /></Form.Item>
        <Form.Item label="发布说明" field="notes"><Input.TextArea placeholder="本版本变更与升级注意事项" maxLength={1024} showWordLimit autoSize={{ minRows: 3, maxRows: 6 }} /></Form.Item>
      </Form>
    </Modal>

    <Modal title="新建 Agent 灰度发布" visible={visible} onCancel={() => setVisible(false)} onOk={() => create.mutate()} confirmLoading={create.isPending} style={{ width: 720 }}>
      <Form form={form} layout="vertical" initialValues={{ node_ids: [], group_names: [], canary_count: 1, batch_size: 5, pause_seconds: 30 }}>
        <Form.Item label="任务名称" field="name" rules={[{ required: true, message: '请输入任务名称' }]}><Input placeholder="例如：0.22.1 办公网节点升级" maxLength={128} showWordLimit /></Form.Item>
        <Form.Item label="目标版本" field="release_id" rules={[{ required: true, message: '请选择目标版本' }]}><Select options={(releases.data || []).filter((item) => item.available !== false).map((item) => ({ label: `${item.release.version} · ${item.builds.length} 个平台`, value: item.release.id }))} /></Form.Item>
        <Grid.Row gutter={16}>
          <Grid.Col xs={24} md={12}><Form.Item label="目标节点组" field="group_names"><Select mode="multiple" showSearch allowClear placeholder="可选择一个或多个节点组" options={groupOptions} onChange={(values) => setSelectedGroups(values || [])} /></Form.Item></Grid.Col>
          <Grid.Col xs={24} md={12}><Form.Item label="补充指定节点" field="node_ids"><Select mode="multiple" showSearch allowClear placeholder="可与节点组组合" options={nodeItems.map((node: any) => ({ label: `${node.name} · ${node.group || '未分组'} · ${node.os}/${node.arch} · ${node.version || '未知版本'} · ${node.status === 'online' ? '在线' : '离线'}`, value: node.id }))} onChange={(values) => setSelectedNodeIDs(values || [])} /></Form.Item></Grid.Col>
        </Grid.Row>
        <Alert type={targetCount ? 'info' : 'warning'} content={targetCount ? `去重后将升级 ${targetCount} 个节点；在线节点优先进入金丝雀波次，离线节点上线后自动继续。` : '请至少选择一个节点组或指定节点。'} style={{ marginBottom: 16 }} />
        <Grid.Row gutter={16}>
          <Grid.Col xs={24} md={8}><Form.Item label="金丝雀数量" field="canary_count"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item></Grid.Col>
          <Grid.Col xs={24} md={8}><Form.Item label="后续批次大小" field="batch_size"><InputNumber min={1} style={{ width: '100%' }} /></Form.Item></Grid.Col>
          <Grid.Col xs={24} md={8}><Form.Item label="波次观察时间（秒）" field="pause_seconds"><InputNumber min={0} max={86400} style={{ width: '100%' }} /></Form.Item></Grid.Col>
        </Grid.Row>
        <Typography.Paragraph type="secondary">任一节点验签失败、安装失败、自动回滚或 10 分钟内未确认目标版本时，发布会自动暂停。修复原因后可点击“重试并继续”，失败节点从等待状态重新执行。</Typography.Paragraph>
      </Form>
    </Modal>
  </>;
}
