import { useState } from 'react';
import { Button, Form, Input, Message, Modal, Popconfirm, Select, Space, Switch, Table, Tag, Typography } from '@arco-design/web-react';
import { IconDelete, IconPlus } from '@arco-design/web-react/icon';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, errorMessage, formatTime } from '../api';
import PageHeader from '../components/PageHeader';
import { useAuth } from '../store';

const typeName: Record<string, string> = { credential: '凭据蜜饵', file: '文件蜜饵', network: '网络蜜饵' };
const statusName: Record<string, string> = {
  pending: '等待节点同步', monitoring: '监控中', passive: '被动关联', stopped: '已停止', error: '部署失败',
};
const statusColor: Record<string, string> = { pending: 'orange', monitoring: 'green', passive: 'arcoblue', stopped: 'gray', error: 'red' };

function defaultToken() {
  return `honeynet-${crypto.randomUUID().replaceAll('-', '')}`;
}

function buildConfig(value: any) {
  if (value.type === 'network') {
    return { token: value.token?.trim(), description: value.description?.trim() || undefined };
  }
  const config: Record<string, unknown> = {
    path: value.path?.trim(), mode: value.mode,
    marker: value.marker?.trim() || undefined,
    content: value.content || undefined,
    create_parent: Boolean(value.create_parent),
    monitor_existing: Boolean(value.monitor_existing),
  };
  if (value.type === 'credential') {
    config.username = value.username?.trim();
    config.password = value.password?.trim();
  }
  return config;
}

export default function Decoys() {
  const [visible, setVisible] = useState(false);
  const [selectedType, setSelectedType] = useState('file');
  const [form] = Form.useForm();
  const qc = useQueryClient();
  const role = useAuth((state) => state.user?.role);
  const query = useQuery<any>({ queryKey: ['decoys'], queryFn: () => api.get('/decoys?page_size=100') });
  const nodes = useQuery<any>({ queryKey: ['nodes', 'options'], queryFn: () => api.get('/nodes?page_size=200') });
  const refresh = () => qc.invalidateQueries({ queryKey: ['decoys'] });
  const create = useMutation({
    mutationFn: (value: any) => api.post('/decoys', {
      node_id: value.node_id, name: value.name, type: value.type, status: 'enabled', config: buildConfig(value),
    }),
    onSuccess: () => { setVisible(false); form.resetFields(); refresh(); },
    onError: (error) => Message.error(errorMessage(error)),
  });
  const toggle = async (item: any, enabled: boolean) => {
    try {
      await api.put(`/decoys/${item.id}`, { status: enabled ? 'enabled' : 'disabled' });
      refresh();
    } catch (error) {
      Message.error(errorMessage(error));
    }
  };
  const remove = useMutation({
    mutationFn: (id: string) => api.delete(`/decoys/${id}`),
    onSuccess: refresh,
    onError: (error) => Message.error(errorMessage(error)),
  });
  const open = () => {
    setSelectedType('file');
    form.setFieldsValue({
      name: '', node_id: undefined, type: 'file', path: '/var/lib/honeynet/decoys/quarterly-report.txt',
      marker: 'honeynet-decoy', content: '', mode: '0644', create_parent: true, monitor_existing: false,
      username: 'backup_service', password: `Backup-${Date.now()}`, token: defaultToken(), description: '',
    });
    setVisible(true);
  };
  const columns: any[] = [
    { title: '蜜饵名称', dataIndex: 'name', render: (value: string, record: any) => <div><Typography.Text bold>{value}</Typography.Text><br/><Tag color="purple">{typeName[record.type] || record.type}</Tag></div> },
    { title: '部署节点', render: (_: unknown, record: any) => record.node?.name || record.node_id },
    { title: '部署目标', render: (_: unknown, record: any) => <Typography.Text code>{record.managed_path || record.config?.path || record.config?.token || '—'}</Typography.Text> },
    { title: '运行状态', render: (_: unknown, record: any) => <div><Tag color={statusColor[record.actual_status] || 'gray'}>{statusName[record.actual_status] || record.actual_status || '未知'}</Tag>{record.last_error && <><br/><Typography.Text type="error">{record.last_error}</Typography.Text></>}</div> },
    { title: '命中', render: (_: unknown, record: any) => <div><Typography.Text bold>{record.hit_count || 0}</Typography.Text><br/><Typography.Text type="secondary">{record.last_hit_at ? formatTime(record.last_hit_at) : '尚未命中'}</Typography.Text></div> },
    { title: '创建时间', dataIndex: 'created_at', render: formatTime },
    { title: '启用', render: (_: unknown, record: any) => <Switch size="small" checked={record.status === 'enabled'} disabled={role === 'viewer'} onChange={(value) => toggle(record, value)} /> },
    { title: '操作', render: (_: unknown, record: any) => role !== 'viewer' && <Popconfirm title="确认删除该蜜饵？Agent 创建且未被修改的文件会被清理。" onOk={() => remove.mutate(record.id)}><Button type="text" status="danger" icon={<IconDelete/>}/></Popconfirm> },
  ];
  return <>
    <PageHeader title="蜜饵管理" description="文件与凭据蜜饵由 Agent 安全投放并监控；网络蜜饵通过唯一 Token 与攻击事件被动关联" extra={role !== 'viewer' && <Button type="primary" icon={<IconPlus/>} onClick={open}>创建蜜饵</Button>} />
    <div className="table-panel"><Table rowKey="id" loading={query.isLoading} data={query.data?.items || []} columns={columns} pagination={false} /></div>
    <Modal title="创建蜜饵" visible={visible} style={{ width: 680 }} onCancel={() => setVisible(false)} onOk={() => form.validate().then(create.mutate)} confirmLoading={create.isPending}>
      <Form form={form} layout="vertical">
        <Form.Item label="蜜饵名称" field="name" rules={[{ required: true }]}><Input placeholder="例如：生产数据库备份凭据" /></Form.Item>
        <Form.Item label="部署节点" field="node_id" rules={[{ required: true }]}><Select onChange={(nodeID) => { const node = (nodes.data?.items || []).find((item: any) => item.id === nodeID); form.setFieldValue('path', node?.os === 'windows' ? 'C:\\ProgramData\\Honeynet\\decoys\\quarterly-report.txt' : '/var/lib/honeynet/decoys/quarterly-report.txt'); }} options={(nodes.data?.items || []).map((node: any) => ({ label: `${node.name} (${node.status})`, value: node.id }))} /></Form.Item>
        <Form.Item label="蜜饵类型" field="type" rules={[{ required: true }]}><Select onChange={(value) => { setSelectedType(value); form.setFieldValue('mode', value === 'credential' ? '0600' : '0644'); }} options={Object.entries(typeName).map(([value, label]) => ({ value, label }))} /></Form.Item>
        {selectedType === 'network' ? <>
          <Form.Item label="唯一关联 Token" field="token" rules={[{ required: true }]} extra="将该 Token 嵌入虚假 URL、连接串或文档；它出现在节点攻击事件中时自动产生 decoy.network 高危事件"><Input /></Form.Item>
          <Form.Item label="投放说明" field="description"><Input.TextArea placeholder="例如：嵌入共享目录中的数据库连接说明" autoSize={{ minRows: 3, maxRows: 5 }} /></Form.Item>
        </> : <>
          <Form.Item label="节点绝对路径" field="path" rules={[{ required: true }]} extra="默认绝不覆盖已有文件；Agent 创建的文件会记录所有权"><Input /></Form.Item>
          {selectedType === 'credential' && <Space size={16} style={{ width: '100%' }}>
            <Form.Item label="伪用户名" field="username" rules={[{ required: true }]} style={{ flex: 1 }}><Input /></Form.Item>
            <Form.Item label="伪密码" field="password" rules={[{ required: true }]} style={{ flex: 1 }}><Input /></Form.Item>
          </Space>}
          <Form.Item label="自定义内容（可选）" field="content" extra="留空时根据蜜饵类型生成带唯一标记的安全内容"><Input.TextArea autoSize={{ minRows: 4, maxRows: 8 }} /></Form.Item>
          <Space size={16} style={{ width: '100%' }}>
            <Form.Item label="文件权限" field="mode" rules={[{ required: true }]} style={{ width: 180 }}><Select options={['0400', '0440', '0600', '0640', '0644'].map((value) => ({ value, label: value }))} /></Form.Item>
            <Form.Item label="自动创建父目录" field="create_parent" triggerPropName="checked"><Switch /></Form.Item>
            <Form.Item label="仅监控已有文件" field="monitor_existing" triggerPropName="checked" extra="不会修改或删除已有文件"><Switch /></Form.Item>
          </Space>
        </>}
      </Form>
    </Modal>
  </>;
}
