import { useMemo, useState } from 'react';
import {
  Alert, Button, Drawer, Form, Input, InputNumber, Message, Modal, Radio, Select, Space, Table, Tag, Typography,
} from '@arco-design/web-react';
import { IconEye, IconPlayArrow, IconPlus, IconSearch } from '@arco-design/web-react/icon';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { api, errorMessage } from '../api';
import PageHeader from '../components/PageHeader';
import { availableBuiltinPort, BUILTIN_NODE_ID, isBrowserService, potAccessURL } from '../potAccess';
import { useAuth } from '../store';

async function loadServices(category: string, keyword: string) {
  const params = { category: category || undefined, q: keyword || undefined, page_size: 100 };
  const first: any = await api.get('/pot-services', { params: { ...params, page: 1 } });
  if ((first.items || []).length >= first.total) return first;
  const second: any = await api.get('/pot-services', { params: { ...params, page: 2 } });
  return { ...first, items: [...(first.items || []), ...(second.items || [])] };
}

function supports(node: any, serviceCode: string) {
  return Array.isArray(node?.capabilities) && node.capabilities.includes(`pot.${serviceCode}`);
}

export default function Services() {
  const [category, setCategory] = useState('');
  const [keyword, setKeyword] = useState('');
  const [selected, setSelected] = useState<any>();
  const [deployVisible, setDeployVisible] = useState(false);
  const [selectedNodeID, setSelectedNodeID] = useState('');
  const [selectedServiceCode, setSelectedServiceCode] = useState('');
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const role = useAuth((state) => state.user?.role);
  const query = useQuery<any>({ queryKey: ['services', 'catalog', category, keyword], queryFn: () => loadServices(category, keyword) });
  const allServices = useQuery<any>({ queryKey: ['services', 'deploy-catalog'], queryFn: () => loadServices('', '') });
  const pots = useQuery<any>({ queryKey: ['pots', 'service-counts'], queryFn: () => api.get('/pots?page_size=100'), refetchInterval: deployVisible ? 2_000 : 10_000 });
  const nodes = useQuery<any>({ queryKey: ['nodes', 'deploy-options'], queryFn: () => api.get('/nodes?page_size=200'), refetchInterval: 15_000 });
  const templates = useQuery<any>({ queryKey: ['templates', 'deploy-options'], queryFn: () => api.get('/pot-templates?page_size=200') });

  const serviceItems = allServices.data?.items || [];
  const selectedNode = (nodes.data?.items || []).find((node: any) => node.id === selectedNodeID);
  const selectedService = serviceItems.find((service: any) => service.code === selectedServiceCode);
  const capabilitiesKnown = Array.isArray(selectedNode?.capabilities);
  const supportedCount = serviceItems.filter((service: any) => supports(selectedNode, service.code)).length;
  const counts = useMemo(() => {
    const result = new Map<string, number>();
    (pots.data?.items || []).forEach((pot: any) => result.set(pot.service_code, (result.get(pot.service_code) || 0) + 1));
    return result;
  }, [pots.data]);
  const runningByService = useMemo(() => {
    const result = new Map<string, any>();
    (pots.data?.items || []).forEach((pot: any) => {
      if (pot.actual_status === 'running' && isBrowserService(pot.service_code) && !result.has(pot.service_code)) result.set(pot.service_code, pot);
    });
    return result;
  }, [pots.data]);

  const closeDeploy = () => {
    setDeployVisible(false);
    setSelectedNodeID('');
    setSelectedServiceCode('');
    form.resetFields();
  };

  const setServiceDefaults = (service: any, nodeID: string) => {
    if (!service) return;
    const port = nodeID === BUILTIN_NODE_ID
      ? availableBuiltinPort(service.code, service.default_port, pots.data?.items || [])
      : service.default_port;
    form.setFieldsValue({ service_code: service.code, name: `${service.name} 蜜罐`, port, template_id: undefined });
    setSelectedServiceCode(service.code);
  };

  const openDeploy = (service?: any) => {
    form.resetFields();
    const candidates = nodes.data?.items || [];
    const node = candidates.find((item: any) => item.status === 'online' && (!service || supports(item, service.code)))
      || candidates.find((item: any) => item.status === 'online');
    const nodeID = node?.id || '';
    setSelectedNodeID(nodeID);
    setSelectedServiceCode(service?.code || '');
    setDeployVisible(true);
    form.setFieldsValue({ node_id: nodeID || undefined });
    if (service) setServiceDefaults(service, nodeID);
  };

  const deploy = useMutation({
    mutationFn: (values: any) => api.post('/pots', { ...values, config: {}, desired_status: 'running' }),
    onSuccess: (item: any) => {
      Message.success(`${item.name} 已下发到节点，正在启动`);
      closeDeploy();
      queryClient.invalidateQueries({ queryKey: ['pots'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
    onError: (error) => Message.error(errorMessage(error)),
  });

  const onDeployChange = (changed: any) => {
    if (Object.prototype.hasOwnProperty.call(changed, 'node_id')) {
      const nodeID = changed.node_id || '';
      setSelectedNodeID(nodeID);
      const serviceCode = form.getFieldValue('service_code');
      const service = serviceItems.find((item: any) => item.code === serviceCode);
      const node = (nodes.data?.items || []).find((item: any) => item.id === nodeID);
      if (service && !supports(node, service.code)) {
        form.setFieldsValue({ service_code: undefined, name: undefined, port: undefined, template_id: undefined });
        setSelectedServiceCode('');
      } else if (service) {
        setServiceDefaults(service, nodeID);
      }
    }
    if (Object.prototype.hasOwnProperty.call(changed, 'service_code')) {
      const service = serviceItems.find((item: any) => item.code === changed.service_code);
      if (service) setServiceDefaults(service, selectedNodeID);
      else setSelectedServiceCode('');
    }
  };

  const openRunning = (serviceCode: string) => {
    const url = potAccessURL(runningByService.get(serviceCode));
    if (!url) return;
    if (url.startsWith('https://')) Message.info('蜜罐使用节点自签名证书，浏览器首次访问时请选择“高级 → 继续访问”。');
    window.open(url, '_blank', 'noopener,noreferrer');
  };

  const columns: any[] = [
    { title: '服务名称', dataIndex: 'name', render: (value: string, row: any) => <div><Typography.Text bold>{value}</Typography.Text><br /><Typography.Text type="secondary">{row.code}</Typography.Text></div> },
    { title: '大类 / 协议', render: (_: any, row: any) => <Space direction="vertical" size={2}><Tag color="arcoblue">{row.category}</Tag><Typography.Text type="secondary">{row.protocol?.toUpperCase()}</Typography.Text></Space> },
    { title: '交互类型', dataIndex: 'depth', render: (value: string) => <Tag color={value === '高交互' || value === 'high' ? 'red' : value === '中交互' || value === 'medium' ? 'orange' : 'green'}>{value || '低交互'}</Tag> },
    { title: '已部署', render: (_: any, row: any) => `${counts.get(row.code) || 0} 个` },
    { title: '默认端口', render: (_: any, row: any) => <Typography.Text code>{row.protocol?.toUpperCase()}/{row.default_port}</Typography.Text> },
    { title: '描述', dataIndex: 'description', render: (value: string) => <Typography.Text type="secondary" ellipsis={{ rows: 2, showTooltip: true }}>{value}</Typography.Text> },
    { title: '操作', width: 210, render: (_: any, row: any) => <Space size={2}><Button type="text" icon={<IconEye />} onClick={() => setSelected(row)}>详情</Button><Button type="text" icon={<IconPlus />} disabled={role === 'viewer'} onClick={() => openDeploy(row)}>部署</Button>{runningByService.has(row.code) && <Button type="text" icon={<IconPlayArrow />} onClick={() => openRunning(row.code)}>打开</Button>}</Space> },
  ];
  const categories = query.data?.categories || [];
  const serviceOptions = serviceItems.map((service: any) => ({
    label: `${service.name} · ${service.protocol?.toUpperCase()}/${service.default_port}${selectedNodeID && !supports(selectedNode, service.code) ? ' · 当前节点不支持' : ''}`,
    value: service.code,
    disabled: !selectedNodeID || !supports(selectedNode, service.code),
  }));

  return <>
    <PageHeader title="服务管理" description={`统一管理 Agent 原生服务与 honeypot-templates-server 资源，当前共 ${query.data?.total || 0} 种服务`} extra={<Space><Button onClick={() => navigate('/pots')}>蜜罐实例</Button>{role !== 'viewer' && <Button type="primary" icon={<IconPlus />} onClick={() => openDeploy()}>部署服务</Button>}</Space>} />
    <div className="service-category-bar"><Radio.Group type="button" value={category} onChange={setCategory}><Radio value="">全部服务</Radio>{categories.map((item: string) => <Radio key={item} value={item}>{item}</Radio>)}</Radio.Group></div>
    <div className="filter-bar"><Input allowClear prefix={<IconSearch />} placeholder="搜索服务名称或代码" value={keyword} onChange={setKeyword} style={{ width: 320 }} /><Typography.Text type="secondary">{query.data?.total || 0} 项服务</Typography.Text></div>
    <div className="table-panel"><Table rowKey="code" loading={query.isLoading} data={query.data?.items || []} columns={columns} pagination={{ pageSize: 20 }} scroll={{ x: 1260 }} /></div>

    <Drawer width={520} title="蜜罐服务详情" visible={!!selected} onCancel={() => setSelected(undefined)} footer={<Space style={{ width: '100%', justifyContent: 'flex-end' }}><Button onClick={() => setSelected(undefined)}>关闭</Button><Button type="primary" disabled={role === 'viewer'} onClick={() => { const service = selected; setSelected(undefined); openDeploy(service); }}>部署到节点</Button>{selected && runningByService.has(selected.code) && <Button icon={<IconPlayArrow />} onClick={() => openRunning(selected.code)}>打开运行实例</Button>}</Space>}>
      {selected && <div className="service-detail"><div className="service-hero"><Tag color="arcoblue">{selected.category}</Tag><Typography.Title heading={4}>{selected.name}</Typography.Title><Typography.Text code>{selected.code}</Typography.Text></div><Detail label="协议 / 默认端口" value={`${selected.protocol?.toUpperCase()} / ${selected.default_port}`} /><Detail label="交互类型" value={selected.depth} /><Detail label="Agent 能力标识" value={selected.capability} /><Detail label="已部署实例" value={`${counts.get(selected.code) || 0} 个`} /><Typography.Title heading={6}>服务说明</Typography.Title><Typography.Paragraph>{selected.description || '暂无说明'}</Typography.Paragraph></div>}
    </Drawer>

    <Modal title={selectedService ? `部署 ${selectedService.name}` : '部署蜜罐服务'} visible={deployVisible} onCancel={closeDeploy} onOk={() => form.validate().then(deploy.mutate)} confirmLoading={deploy.isPending} okText="部署并启动" style={{ width: 620, maxWidth: 'calc(100vw - 24px)' }}>
      <Form form={form} layout="vertical" onValuesChange={onDeployChange}>
        <Form.Item label="目标节点" field="node_id" rules={[{ required: true, message: '请选择部署节点' }]}><Select placeholder="选择在线节点" options={(nodes.data?.items || []).map((node: any) => ({ label: `${node.name} · ${node.status === 'online' ? '在线' : '离线'}`, value: node.id, disabled: node.status !== 'online' }))} /></Form.Item>
        {selectedNodeID && !capabilitiesKnown && <Alert type="warning" content="该节点尚未上报服务能力，请升级 Agent 并等待节点上线。" style={{ marginBottom: 16 }} />}
        {selectedNodeID && capabilitiesKnown && <Alert type="info" content={`该节点当前支持 ${supportedCount} / ${allServices.data?.total || 0} 种目录服务。`} style={{ marginBottom: 16 }} />}
        <Form.Item label="蜜罐服务" field="service_code" rules={[{ required: true, message: '请选择蜜罐服务' }]} extra="只显示当前节点 Agent 确认支持的服务"><Select showSearch disabled={!selectedNodeID} placeholder="搜索服务名称或代码" options={serviceOptions} /></Form.Item>
        {selectedServiceCode === 'web-template' && <Form.Item label="Web 蜜罐模板" field="template_id" rules={[{ required: true, message: '请选择 Web 蜜罐模板' }]}><Select showSearch placeholder="选择模板" options={(templates.data?.items || []).map((item: any) => ({ label: `${item.name} · v${item.version}`, value: item.id }))} /></Form.Item>}
        <Form.Item label="实例名称" field="name" rules={[{ required: true, message: '请输入实例名称' }]}><Input placeholder="例如：办公网 Tomcat 入口" /></Form.Item>
        <Form.Item label="监听端口" field="port" rules={[{ required: true, message: '请输入监听端口' }]} extra={selectedNodeID === BUILTIN_NODE_ID && selectedService && isBrowserService(selectedService.code) ? '内置节点自动选择 20000–20099 可访问空闲端口。' : undefined}><InputNumber min={1} max={65535} style={{ width: '100%' }} /></Form.Item>
      </Form>
    </Modal>
  </>;
}

function Detail({ label, value }: { label: string; value: string }) {
  return <div className="detail-row"><Typography.Text type="secondary">{label}</Typography.Text><Typography.Text>{value || '—'}</Typography.Text></div>;
}
