import { useEffect, useMemo, useState } from 'react';
import { Alert, Button, Form, Input, InputNumber, Message, Modal, Popconfirm, Select, Space, Table, Tag, Typography } from '@arco-design/web-react';
import { IconDelete, IconEdit, IconLaunch, IconPlayArrow, IconPlus, IconStop } from '@arco-design/web-react/icon';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useSearchParams } from 'react-router-dom';
import { api, errorMessage } from '../api';
import PageHeader from '../components/PageHeader';
import { availableBuiltinPort, BUILTIN_NODE_ID, webTemplateServices, potAccessURL } from '../potAccess';
import { formatHostPort, preferredNodeAddress } from '../presentation';
import { useAuth } from '../store';

const localConsole = ['localhost', '127.0.0.1', '::1'].includes(window.location.hostname);
const actualStatusColors: Record<string, string> = { running: 'green', pending: 'arcoblue', stopped: 'gray', unsupported: 'orange', error: 'red' };
const actualStatusLabels: Record<string, string> = { running: '运行中', pending: '等待节点同步', stopped: '已停止', unsupported: 'Agent 不支持', error: '运行错误' };

function ListenerEndpoint({ row }: { row: any }) {
  const protocol = row.service?.protocol?.toUpperCase() || 'TCP';
  const address = preferredNodeAddress(row.node);
  const accessURL = row.actual_status === 'running' ? potAccessURL(row) : '';
  let accessEndpoint = '';
  try { accessEndpoint = accessURL ? new URL(accessURL).host : ''; } catch { accessEndpoint = ''; }
  const listenerEndpoint = address ? formatHostPort(address, Number(row.port)) : '';
  const mapped = Boolean(accessEndpoint && listenerEndpoint && accessEndpoint !== listenerEndpoint);
  return <div className="endpoint-cell">
    <div className="endpoint-main"><Tag color={protocol === 'UDP' ? 'purple' : 'arcoblue'}>{protocol}</Tag><Typography.Text bold copyable={{ text: String(row.port) }}>监听 {row.port}</Typography.Text></div>
    <Typography.Text type="secondary" code copyable={Boolean(listenerEndpoint)}>{listenerEndpoint || '等待节点地址'}</Typography.Text>
    {mapped && <Typography.Text type="secondary">访问入口：{accessEndpoint}</Typography.Text>}
    {accessURL && <a className="endpoint-link" href={accessURL} target="_blank" rel="noreferrer" title={accessURL.startsWith('https://') ? '首次访问需要接受节点自签名证书' : undefined}><IconLaunch />访问蜜罐</a>}
  </div>;
}

export default function Pots() {
  const [visible, setVisible] = useState(false);
  const [editing, setEditing] = useState<any>();
  const [selectedNodeID, setSelectedNodeID] = useState('');
  const [selectedService, setSelectedService] = useState('');
  const [form] = Form.useForm();
  const [searchParams, setSearchParams] = useSearchParams();
  const queryClient = useQueryClient();
  const role = useAuth((state) => state.user?.role);
  const query = useQuery<any>({ queryKey: ['pots'], queryFn: () => api.get('/pots?page_size=100'), refetchInterval: 5_000 });
  const nodes = useQuery<any>({ queryKey: ['nodes', 'options'], queryFn: () => api.get('/nodes?page_size=200') });
  const services = useQuery<any>({ queryKey: ['services'], queryFn: () => api.get('/pot-services?page_size=200') });
  const templates = useQuery<any>({ queryKey: ['templates', 'options'], queryFn: () => api.get('/pot-templates?page_size=200') });
  const refresh = () => queryClient.invalidateQueries({ queryKey: ['pots'] });
  const save = useMutation({
    mutationFn: (value: any) => editing
      ? api.put(`/pots/${editing.id}`, {
        name: value.name,
        port: value.port,
        config: { ...(editing.config || {}), bind: value.bind },
        ...(editing.service_code === 'web-template' ? { template_id: value.template_id } : {}),
      })
      : api.post('/pots', { ...value, config: { bind: value.bind }, desired_status: 'running' }),
    onSuccess: () => {
      Message.success(editing ? '蜜罐配置已下发，Agent 正在重载监听端口' : '蜜罐部署指令已下发，Agent 正在启动服务');
      setVisible(false);
      setEditing(undefined);
      setSelectedNodeID('');
      setSelectedService('');
      form.resetFields();
      refresh();
    },
    onError: (error) => Message.error(errorMessage(error)),
  });
  const action = useMutation({
    mutationFn: ({ id, action: nextAction }: { id: string; action: string }) => api.post(`/pots/${id}/${nextAction}`),
    onSuccess: refresh,
    onError: (error) => Message.error(errorMessage(error)),
  });
  const remove = useMutation({ mutationFn: (id: string) => api.delete(`/pots/${id}`), onSuccess: refresh, onError: (error) => Message.error(errorMessage(error)) });

  const selectedNode = (nodes.data?.items || []).find((node: any) => node.id === selectedNodeID);
  const capabilitiesKnown = Array.isArray(selectedNode?.capabilities);
  const supportedCapabilities = useMemo(() => new Set<string>(Array.isArray(selectedNode?.capabilities) ? selectedNode.capabilities : []), [selectedNode]);
  const serviceItems = services.data?.items || [];
  const supportedCount = serviceItems.filter((service: any) => supportedCapabilities.has(`pot.${service.code}`)).length;
  const serviceOptions = serviceItems.map((service: any) => {
    const supported = supportedCapabilities.has(`pot.${service.code}`);
    const suffix = !selectedNodeID ? ' · 请先选择节点' : !capabilitiesKnown ? ' · 能力未知' : !supported ? ' · 当前 Agent 不支持' : '';
    return {
      label: `${service.name} · ${service.protocol.toUpperCase()}/${service.default_port}${suffix}`,
      value: service.code,
      disabled: !selectedNodeID || !capabilitiesKnown || !supported,
      extra: service,
    };
  });

  const openCreate = () => {
    form.resetFields();
    setEditing(undefined);
    setSelectedNodeID('');
    setSelectedService('');
    setVisible(true);
  };
  const openEdit = (row: any) => {
    setEditing(row);
    setSelectedNodeID(row.node_id);
    setSelectedService(row.service_code);
    form.setFieldsValue({
      node_id: row.node_id,
      service_code: row.service_code,
      template_id: row.template_id,
      name: row.name,
      port: row.port,
      bind: row.config?.bind || '0.0.0.0',
    });
    setVisible(true);
  };
  const closeCreate = () => {
    setVisible(false);
    setEditing(undefined);
    setSelectedNodeID('');
    setSelectedService('');
    form.resetFields();
  };
  useEffect(() => {
    const nodeID = searchParams.get('node_id');
    if (!nodeID || !(nodes.data?.items || []).some((node: any) => node.id === nodeID)) return;
    form.resetFields();
    setEditing(undefined);
    form.setFieldValue('node_id', nodeID);
    setSelectedNodeID(nodeID);
    setSelectedService('');
    setVisible(true);
    form.setFieldValue('bind', '0.0.0.0');
    setSearchParams({}, { replace: true });
  }, [form, nodes.data, searchParams, setSearchParams]);
  const onFormChange = (changed: any) => {
    if (Object.prototype.hasOwnProperty.call(changed, 'node_id')) {
      setSelectedNodeID(changed.node_id || '');
      setSelectedService('');
      form.setFieldsValue({ service_code: undefined, template_id: undefined });
    }
    if (changed.service_code) {
      setSelectedService(changed.service_code);
      const service = serviceItems.find((item: any) => item.code === changed.service_code);
      if (service) {
        const localPort = selectedNodeID === BUILTIN_NODE_ID && localConsole
          ? availableBuiltinPort(service.code, service.default_port, query.data?.items || [])
          : undefined;
        form.setFieldValue('port', localPort || service.default_port);
      }
      if (changed.service_code !== 'web-template') form.setFieldValue('template_id', undefined);
    }
  };

  const columns: any[] = [
    { title: '实例名称', dataIndex: 'name', render: (value: string, row: any) => <div><Typography.Text bold>{value}</Typography.Text><br /><Typography.Text type="secondary">{row.template ? `${row.template.name} · v${row.template.version}` : (row.service?.name || row.service_code)}</Typography.Text></div> },
    { title: '所属节点', render: (_: any, row: any) => <div className="entity-cell"><Typography.Text bold>{row.node?.name || row.node_id}</Typography.Text><Typography.Text type="secondary" code>{preferredNodeAddress(row.node) || '地址未上报'}</Typography.Text></div> },
    { title: '监听端口', width: 230, render: (_: any, row: any) => <ListenerEndpoint row={row} /> },
    { title: '交互深度', render: (_: any, row: any) => <Tag>{row.service?.depth || '—'}</Tag> },
    { title: '目标状态', dataIndex: 'desired_status', render: (value: string) => <Tag color={value === 'running' ? 'green' : 'gray'}>{value === 'running' ? '运行' : '停止'}</Tag> },
    { title: '实际状态', dataIndex: 'actual_status', render: (value: string) => <Tag color={actualStatusColors[value] || 'gray'}>{actualStatusLabels[value] || value}</Tag> },
    { title: '操作', width: 245, fixed: 'right', render: (_: any, row: any) => <Space wrap size={4}>{role !== 'viewer' && <Button size="small" type="text" icon={<IconEdit />} onClick={() => openEdit(row)}>编辑</Button>}{row.desired_status === 'running' ? <Button size="small" type="text" icon={<IconStop />} disabled={role === 'viewer'} onClick={() => action.mutate({ id: row.id, action: 'stop' })}>停止</Button> : <Button size="small" type="text" icon={<IconPlayArrow />} disabled={role === 'viewer'} onClick={() => action.mutate({ id: row.id, action: 'start' })}>启动</Button>}{role !== 'viewer' && <Popconfirm title="确认删除该实例？删除后 Agent 会立即释放监听端口。" onOk={() => remove.mutate(row.id)}><Button size="small" type="text" status="danger" icon={<IconDelete />} /></Popconfirm>}</Space> },
  ];

  return <>
    <PageHeader
      title="蜜罐实例"
      description={`服务目录收录 ${services.data?.total || 0} 种；创建时仅允许选择目标节点 Agent 实际支持的服务`}
      extra={role !== 'viewer' && <Button type="primary" icon={<IconPlus />} onClick={openCreate}>创建蜜罐</Button>}
    />
    <div className="table-panel"><Table rowKey="id" loading={query.isLoading} data={query.data?.items || []} columns={columns} pagination={{ pageSize: 20 }} scroll={{ x: 1280 }} /></div>
    <Modal title={editing ? '编辑蜜罐实例' : '创建蜜罐实例'} visible={visible} onCancel={closeCreate} onOk={() => form.validate().then(save.mutate)} confirmLoading={save.isPending} okText={editing ? '保存并重载' : '创建并启动'} style={{ width: 620 }}>
      <Form form={form} layout="vertical" onValuesChange={onFormChange}>
        <Form.Item label="所属节点" field="node_id" rules={[{ required: true }]}>
          <Select disabled={Boolean(editing)} placeholder="先选择部署节点" options={(nodes.data?.items || []).map((node: any) => ({ label: `${node.name} · ${preferredNodeAddress(node) || '地址未上报'} (${node.status})`, value: node.id }))} />
        </Form.Item>
        {selectedNodeID && !capabilitiesKnown && <Alert type="warning" content="该节点尚未上报能力。请安装或升级 Agent，等待节点上线后再创建蜜罐。" style={{ marginBottom: 16 }} />}
        {selectedNodeID && capabilitiesKnown && <Alert type="info" content={`该节点当前可运行 ${supportedCount} / ${services.data?.total || 0} 种目录服务；不支持的服务已禁用。`} style={{ marginBottom: 16 }} />}
        <Form.Item label="蜜罐服务" field="service_code" rules={[{ required: true }]} extra="服务可用性来自 Agent 最近一次 hello/heartbeat 上报">
          <Select showSearch disabled={!selectedNodeID || Boolean(editing)} placeholder="搜索并选择可运行服务" options={serviceOptions} />
        </Form.Item>
        {selectedService === 'web-template' && <Form.Item label="Web 蜜罐模板" field="template_id" rules={[{ required: true, message: '请选择 Web 蜜罐模板' }]} extra={(templates.data?.items || []).length ? '模板更新后，运行中的实例会自动重载' : '暂无模板，请先在“自定义 Web 蜜罐”页面创建'}><Select showSearch placeholder="选择要部署的模板" options={(templates.data?.items || []).map((item: any) => ({ label: `${item.name} · v${item.version}`, value: item.id }))} /></Form.Item>}
        <Form.Item label="实例名称" field="name" rules={[{ required: true }]}><Input placeholder="例如：办公网邮件入口" /></Form.Item>
        <Form.Item label="监听端口" field="port" rules={[{ required: true }]} extra={selectedNodeID === BUILTIN_NODE_ID && localConsole && webTemplateServices.has(selectedService) ? '本机节点会优先使用 20000–20099 空闲端口，启动后可从列表直接打开页面。' : undefined}><InputNumber min={1} max={65535} style={{ width: '100%' }} /></Form.Item>
        <Form.Item label="监听地址 / 地址族" field="bind" initialValue="0.0.0.0" rules={[{ required: true }, { validator: (value, callback) => { const text = String(value || '').trim(); if (text === '0.0.0.0' || text === '::' || /^\d{1,3}(\.\d{1,3}){3}$/.test(text) || text.includes(':')) callback(); else callback('请输入有效 IPv4 或 IPv6 地址'); } }]} extra="0.0.0.0 监听全部 IPv4；:: 在双栈系统监听 IPv6，并通常同时接受 IPv4。也可填写节点上的具体 IPv4/IPv6。"><Select allowCreate showSearch options={[{ label: '全部 IPv4 · 0.0.0.0', value: '0.0.0.0' }, { label: 'IPv6 / 双栈 · ::', value: '::' }]} /></Form.Item>
      </Form>
    </Modal>
  </>;
}
