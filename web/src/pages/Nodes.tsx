import { useState } from 'react';
import { Alert, Button, Empty, Form, Grid, Input, InputNumber, Message, Modal, Popconfirm, Select, Space, Switch, Table, Tabs, Tag, Tooltip, Typography } from '@arco-design/web-react';
import { IconCopy, IconDelete, IconDownload, IconEdit, IconExperiment, IconPlus, IconRefresh, IconSafe, IconSettings } from '@arco-design/web-react/icon';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { api, errorMessage, formatTime } from '../api';
import PageHeader from '../components/PageHeader';
import { useAuth } from '../store';

const BUILTIN_NODE_ID = '00000000-0000-4000-8000-000000000001';
const statusMap: Record<string, [string, string]> = { online: ['在线', 'green'], offline: ['离线', 'gray'], registering: ['待注册', 'arcoblue'], registered: ['已认证', 'cyan'], degraded: ['异常', 'orange'], revoked: ['已吊销', 'red'] };
const senseStatusMap: Record<string, [string, string]> = { running: ['运行中', 'green'], disabled: ['未启用', 'gray'], pending: ['待同步', 'arcoblue'], error: ['运行错误', 'red'], unsupported: ['平台不支持', 'orange'] };
const potStatusMap: Record<string, string> = { running: '在线', stopped: '离线', pending: '同步中', error: '异常', unsupported: '不支持' };
const addressModeMap: Record<string, string> = { auto: '自动', public: '公网', private: '私网', custom: '自定义' };

function actualPotStatus(value?: string) { return potStatusMap[value || ''] || value || '未知'; }

function NodeAddress({ value }: { value?: string }) {
  const address = String(value || '').trim();
  if (!address) return <Typography.Text type="secondary">等待上报</Typography.Text>;
  return <div className="node-address-line">
    <Tooltip content={address} position="top">
      <Typography.Text className="node-address-value" code ellipsis>{address}</Typography.Text>
    </Tooltip>
    <Tooltip content="复制节点地址" mini>
      <Button
        className="node-address-copy"
        type="text"
        size="mini"
        icon={<IconCopy />}
        aria-label="复制节点地址"
        onClick={(event) => {
          event.stopPropagation();
          navigator.clipboard.writeText(address).then(() => Message.success('节点地址已复制'));
        }}
      />
    </Tooltip>
  </div>;
}

function certificateView(expiresAt?: string, issuedAt?: string) {
  if (!expiresAt) return { label: '未签发', color: 'gray', detail: '等待节点注册', state: 'unsigned' };
  const expiry = new Date(expiresAt);
  if (Number.isNaN(expiry.getTime())) return { label: '时间异常', color: 'red', detail: '证书到期时间无效', state: 'error' };
  const remainingDays = Math.ceil((expiry.getTime() - Date.now()) / 86_400_000);
  const issued = issuedAt ? new Date(issuedAt) : null;
  const validityDays = issued && !Number.isNaN(issued.getTime()) ? Math.round((expiry.getTime() - issued.getTime()) / 86_400_000) : 0;
  const validity = validityDays > 0 ? ` · 签发有效期 ${validityDays} 天` : '';
  if (remainingDays < 0) return { label: '已过期', color: 'red', detail: `${formatTime(expiresAt)} · 已过期 ${Math.abs(remainingDays)} 天${validity}`, state: 'expired' };
  if (remainingDays <= 45) return { label: '即将过期', color: 'orange', detail: `${formatTime(expiresAt)} · 剩余 ${remainingDays} 天${validity}`, state: 'expiring' };
  return { label: '有效', color: 'green', detail: `${formatTime(expiresAt)} · 剩余 ${remainingDays} 天${validity}`, state: 'valid' };
}

function parsePorts(value?: string) {
  if (!value?.trim()) return [];
  const ports = value.split(/[\s,，]+/).filter(Boolean).map(Number);
  if (ports.some((port) => !Number.isInteger(port) || port < 1 || port > 65535)) throw new Error('排除端口必须是 1–65535 的整数');
  return [...new Set(ports)];
}

function parseCIDRs(value?: string) {
  return value?.split(/[\n,，]+/).map((item) => item.trim()).filter(Boolean) || [];
}

function Command({ value }: { value: string }) {
  return <div className="code-block">{value}<Button type="text" icon={<IconCopy />} onClick={() => navigator.clipboard.writeText(value).then(() => Message.success('安装命令已复制'))} /></div>;
}

function showInstaller(result: any) {
  const commands = result.install_commands || { linux: result.install_command };
  const defaultTab = result.node?.os === 'windows' ? 'windows' : 'linux';
  Modal.info({
    title: '安装节点端',
    style: { width: 720 },
    content: <div>
      <Typography.Paragraph type="secondary">节点端是独立 Go 二进制，由 systemd 或 Windows Service 原生运行，不依赖 Docker。注册令牌仅本次显示，30 分钟内有效；注册后事件、心跳与控制通道全部使用 TLS 1.3 双向证书认证。</Typography.Paragraph>
      <Tabs defaultActiveTab={defaultTab}>
        <Tabs.TabPane key="linux" title="Linux">
          <Typography.Paragraph type="secondary">支持 AMD64、ARM64、x86、ARMv7 与 LoongArch，自动注册 systemd 服务。</Typography.Paragraph>
          <Command value={commands.linux} />
        </Tabs.TabPane>
        <Tabs.TabPane key="windows" title="Windows">
          <Typography.Paragraph type="secondary">请在管理员 PowerShell 中执行，Agent 将注册为 Windows Service。</Typography.Paragraph>
          <Command value={commands.windows} />
        </Tabs.TabPane>
      </Tabs>
    </div>,
  });
}

export default function Nodes() {
  const [visible, setVisible] = useState(false);
  const [senseVisible, setSenseVisible] = useState(false);
  const [addressVisible, setAddressVisible] = useState(false);
  const [addressMode, setAddressMode] = useState('auto');
  const [addressNode, setAddressNode] = useState<any>();
  const [senseLoading, setSenseLoading] = useState(false);
  const [senseNode, setSenseNode] = useState<any>();
  const [form] = Form.useForm();
  const [addressForm] = Form.useForm();
  const [senseForm] = Form.useForm();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const role = useAuth((s) => s.user?.role);
  const query = useQuery<any>({ queryKey: ['nodes'], queryFn: () => api.get('/nodes?page_size=100') });
  const pots = useQuery<any>({ queryKey: ['pots', 'node-expand'], queryFn: () => api.get('/pots?page_size=100') });
  const create = useMutation({ mutationFn: (value: any) => api.post('/nodes', value), onSuccess: (result: any) => { setVisible(false); form.resetFields(); queryClient.invalidateQueries({ queryKey: ['nodes'] }); showInstaller(result); }, onError: (e) => Message.error(errorMessage(e)) });
  const remove = useMutation({ mutationFn: (id: string) => api.delete(`/nodes/${id}`), onSuccess: () => queryClient.invalidateQueries({ queryKey: ['nodes'] }), onError: (e) => Message.error(errorMessage(e)) });
  const saveAddress = useMutation({
    mutationFn: ({ id, value }: { id: string; value: any }) => api.put(`/nodes/${id}`, value),
    onSuccess: () => {
      setAddressVisible(false);
      queryClient.invalidateQueries({ queryKey: ['nodes'] });
      queryClient.invalidateQueries({ queryKey: ['pots'] });
      Message.success('节点访问地址已保存，Agent 心跳不会覆盖手动选择');
    },
    onError: (e) => Message.error(errorMessage(e)),
  });
  const saveSense = useMutation({
    mutationFn: ({ id, value }: { id: string; value: any }) => api.put(`/nodes/${id}/sense`, value),
    onSuccess: () => { setSenseVisible(false); queryClient.invalidateQueries({ queryKey: ['nodes'] }); Message.success('全端口扫描感知配置已保存'); },
    onError: (e) => Message.error(errorMessage(e)),
  });
  const issueInstaller = async (id: string) => { try { showInstaller(await api.post(`/nodes/${id}/install`)); queryClient.invalidateQueries({ queryKey: ['nodes'] }); } catch (e) { Message.error(errorMessage(e)); } };
  const rotate = async (id: string) => { try { showInstaller(await api.post(`/nodes/${id}/token`)); queryClient.invalidateQueries({ queryKey: ['nodes'] }); } catch (e) { Message.error(errorMessage(e)); } };
  const openAddress = (node: any) => {
    const mode = node.address_mode || 'auto';
    setAddressNode(node);
    setAddressMode(mode);
    addressForm.setFieldsValue({ address_mode: mode, ip: node.ip || '' });
    setAddressVisible(true);
  };
  const changeAddressMode = (mode: string) => {
    setAddressMode(mode);
    const privateIPs = Array.isArray(addressNode?.private_ips) ? addressNode.private_ips : [];
    const selected = mode === 'public' ? (addressNode?.public_ip || '')
      : mode === 'private' ? (privateIPs.includes(addressNode?.ip) ? addressNode.ip : privateIPs[0] || '')
        : mode === 'custom' ? (addressNode?.ip || '') : '';
    addressForm.setFieldValue('ip', selected);
  };
  const submitAddress = () => addressForm.validate().then((value) => saveAddress.mutate({ id: addressNode.id, value: { address_mode: value.address_mode, ip: value.address_mode === 'auto' ? '' : value.ip } }));
  const openSense = async (node: any) => {
    setSenseNode(node); setSenseVisible(true); setSenseLoading(true);
    try {
      const value: any = await api.get(`/nodes/${node.id}/sense`);
      senseForm.setFieldsValue({ ...value, excluded_ports: (value.excluded_ports || []).join(', '), ignored_cidrs: (value.ignored_cidrs || []).join('\n') });
    } catch (e) { setSenseVisible(false); Message.error(errorMessage(e)); }
    finally { setSenseLoading(false); }
  };
  const submitSense = async () => {
    try {
      const value = await senseForm.validate();
      saveSense.mutate({ id: senseNode.id, value: { ...value, excluded_ports: parsePorts(value.excluded_ports), ignored_cidrs: parseCIDRs(value.ignored_cidrs) } });
    } catch (e) {
      if (e instanceof Error) Message.error(e.message);
    }
  };
  const columns: any[] = [
    { title: '节点名称', dataIndex: 'name', render: (v: string, r: any) => <div><Typography.Text bold>{v}</Typography.Text><br /><Typography.Text type="secondary">{r.group || '默认分组'}</Typography.Text></div> },
    { title: '状态', dataIndex: 'status', render: (v: string) => <Tag color={(statusMap[v] || statusMap.offline)[1]}>{(statusMap[v] || [v])[0]}</Tag> },
    { title: '节点地址', dataIndex: 'ip', width: 270, render: (v: string, r: any) => <div className="node-address-cell"><NodeAddress value={v} /><Tooltip content={`${addressModeMap[r.address_mode || 'auto'] || '自动'} · 公网 ${r.public_ip || '未识别'} · 私网 ${Array.isArray(r.private_ips) ? r.private_ips.length : 0} 个`}><Typography.Text className="node-address-meta" type="secondary" ellipsis>{addressModeMap[r.address_mode || 'auto'] || '自动'} · 公网 {r.public_ip || '未识别'} · 私网 {Array.isArray(r.private_ips) ? r.private_ips.length : 0} 个</Typography.Text></Tooltip></div> },
    { title: '系统 / 架构', render: (_: any, r: any) => `${r.os || '—'} / ${r.arch || '—'}` },
    { title: '版本', dataIndex: 'version', render: (v: string) => v || '—' },
    { title: '扫描感知', render: (_: any, r: any) => <div><Tag color={(senseStatusMap[r.sense?.actual_status] || senseStatusMap.disabled)[1]}>{(senseStatusMap[r.sense?.actual_status] || senseStatusMap.disabled)[0]}</Tag>{r.sense?.detections > 0 && <Typography.Text type="secondary"> {r.sense.detections} 次</Typography.Text>}</div> },
    { title: '节点证书', dataIndex: 'certificate_expires_at', width: 245, render: (v: string, r: any) => {
      const certificate = certificateView(v, r.certificate_issued_at);
      if (r.id === BUILTIN_NODE_ID && !v) return <div className="certificate-cell"><Tag color="arcoblue" icon={<IconSafe />}>等待签发</Tag><Typography.Text type="secondary">本机 Agent 首次注册后签发</Typography.Text></div>;
      return <div className="certificate-cell"><Tag color={certificate.color} icon={<IconSafe />}>{certificate.label}</Tag><Typography.Text type="secondary">{certificate.detail}</Typography.Text></div>;
    } },
    { title: '最后心跳', dataIndex: 'last_heartbeat_at', render: formatTime },
    { title: '操作', width: 430, render: (_: any, r: any) => <Space wrap size={4}>
      <Button type="text" icon={<IconEdit />} disabled={role === 'viewer'} onClick={() => openAddress(r)}>地址</Button>
      <Button type="text" icon={<IconSettings />} disabled={role === 'viewer'} onClick={() => openSense(r)}>感知配置</Button>
      {r.id === BUILTIN_NODE_ID ? <Tag color="arcoblue">本机原生 Agent</Tag> : <>
        <Button type="text" icon={<IconDownload />} disabled={role === 'viewer'} onClick={() => issueInstaller(r.id)}>安装</Button>
        <Button type="text" icon={<IconRefresh />} disabled={role === 'viewer'} onClick={() => rotate(r.id)}>重发令牌</Button>
        {role === 'admin' && <Popconfirm title="确认删除此节点？" onOk={() => remove.mutate(r.id)}><Button type="text" status="danger" icon={<IconDelete />} /></Popconfirm>}
      </>}
    </Space> },
  ];
  const expandedRow = (node: any) => {
    const instances = (pots.data?.items || []).filter((pot: any) => pot.node_id === node.id);
    return <div className="node-expand-panel"><div className="node-expand-head"><div><Typography.Text bold>{node.name} · 蜜罐服务</Typography.Text><Typography.Text type="secondary"> 已部署 {instances.length} 个</Typography.Text></div>{role !== 'viewer' && <Button type="primary" size="small" icon={<IconPlus />} onClick={() => navigate(`/pots?node_id=${node.id}`)}>添加蜜罐服务</Button>}</div>{instances.length ? <div className="node-pot-grid">{instances.map((pot: any) => <div className="node-pot-card" key={pot.id}><div className="node-pot-icon"><IconExperiment /></div><div><Typography.Text bold>{pot.name}</Typography.Text><span>{pot.service?.name || pot.service_code}</span><Typography.Text code>{pot.service?.protocol?.toUpperCase() || 'TCP'}/{pot.port}</Typography.Text></div><Tag color={pot.actual_status === 'running' ? 'green' : pot.actual_status === 'error' ? 'red' : 'gray'}>{actualPotStatus(pot.actual_status)}</Tag></div>)}</div> : <Empty description="该节点尚未部署蜜罐服务" />}</div>;
  };
  return <>
    <PageHeader title="节点管理" description="管理节点安装、在线状态、全端口扫描感知及节点下的蜜罐服务" extra={<Space wrap><Button icon={<IconRefresh />} onClick={() => { query.refetch(); pots.refetch(); }}>刷新</Button>{role !== 'viewer' && <Button type="primary" icon={<IconPlus />} onClick={() => setVisible(true)}>增加节点</Button>}</Space>} />
    <Alert type="info" showIcon content="新签发的节点双向认证证书默认有效期为 400 天；到期前 Agent 会自动续期，剩余 45 天以内在列表中标记为“即将过期”。" style={{ marginBottom: 12 }} />
    <div className="table-panel"><Table rowKey="id" loading={query.isLoading} data={query.data?.items || []} columns={columns} pagination={{ pageSize: 20 }} expandedRowRender={expandedRow} expandProps={{ expandRowByClick: true }} scroll={{ x: 1520 }} /></div>
    <Modal title="新建蜜罐节点" visible={visible} onCancel={() => setVisible(false)} onOk={() => form.validate().then(create.mutate)} confirmLoading={create.isPending}>
      <Form form={form} layout="vertical">
        <Form.Item label="节点名称" field="name" rules={[{ required: true }]}><Input placeholder="例如：上海办公网-01" /></Form.Item>
        <Form.Item label="节点分组" field="group"><Input placeholder="默认分组" /></Form.Item>
        <Form.Item label="预期系统" field="os" initialValue="linux"><Select options={[{ label: 'Linux', value: 'linux' }, { label: 'Windows', value: 'windows' }]} /></Form.Item>
        <Form.Item label="预设访问地址" field="ip" extra="可选；填写后按自定义地址保存，留空则自动优先选择 Server 识别到的公网地址"><Input placeholder="例如：203.0.113.10" /></Form.Item>
      </Form>
    </Modal>
    <Modal title={`${addressNode?.name || '节点'} · 访问地址`} visible={addressVisible} onCancel={() => setAddressVisible(false)} onOk={submitAddress} confirmLoading={saveAddress.isPending} style={{ width: 620 }}>
      <Alert type="info" content="公网地址由 Server 根据 Agent 的 8443 连接来源识别；私网地址由 Agent 上报本机网卡。自动模式优先公网，手动选择后不会再被心跳覆盖。" style={{ marginBottom: 20 }} />
      <Form form={addressForm} layout="vertical">
        <Form.Item label="地址策略" field="address_mode" rules={[{ required: true }]}>
          <Select onChange={changeAddressMode} options={[
            { label: '自动选择（公网优先）', value: 'auto' },
            { label: `公网地址${addressNode?.public_ip ? ` · ${addressNode.public_ip}` : ' · 未识别'}`, value: 'public', disabled: !addressNode?.public_ip },
            { label: `私网地址 · ${(Array.isArray(addressNode?.private_ips) ? addressNode.private_ips.length : 0)} 个候选`, value: 'private', disabled: !Array.isArray(addressNode?.private_ips) || addressNode.private_ips.length === 0 },
            { label: '自定义地址', value: 'custom' },
          ]} />
        </Form.Item>
        {addressMode === 'auto' && <Alert type="success" content={`当前将使用：${addressNode?.public_ip || addressNode?.private_ips?.[0] || '等待 Agent 上报'}`} />}
        {addressMode === 'public' && <Form.Item label="公网地址" field="ip" rules={[{ required: true }]}><Input readOnly /></Form.Item>}
        {addressMode === 'private' && <Form.Item label="私网地址" field="ip" rules={[{ required: true, message: '请选择私网地址' }]}><Select options={(addressNode?.private_ips || []).map((ip: string) => ({ label: ip, value: ip }))} /></Form.Item>}
        {addressMode === 'custom' && <Form.Item label="自定义 IP 地址" field="ip" rules={[{ required: true, message: '请输入 IP 地址' }]} extra="适用于端口映射、专线、VPN 或 Server 无法自动识别的地址"><Input placeholder="IPv4 或 IPv6 地址" /></Form.Item>}
      </Form>
    </Modal>
    <Modal title={`${senseNode?.name || '节点'} · 全端口扫描感知`} visible={senseVisible} onCancel={() => setSenseVisible(false)} onOk={submitSense} confirmLoading={senseLoading || saveSense.isPending} style={{ width: 720 }}>
      <Alert type="info" content="Linux Agent 通过 AF_PACKET 被动观察入站 TCP SYN 与 UDP 探测，不开放端口、不修改防火墙。启用后需要 root 或 CAP_NET_RAW 权限；Windows Agent 暂不支持。" style={{ marginBottom: 20 }} />
      <Form form={senseForm} layout="vertical" disabled={senseLoading || role === 'viewer'}>
        <Grid.Row gutter={20}>
          <Grid.Col xs={24} sm={8}><Form.Item label="启用感知" field="enabled" triggerPropName="checked"><Switch /></Form.Item></Grid.Col>
          <Grid.Col xs={24} sm={8}><Form.Item label="TCP SYN" field="tcp_enabled" triggerPropName="checked" rules={[{ required: true }]}><Switch /></Form.Item></Grid.Col>
          <Grid.Col xs={24} sm={8}><Form.Item label="UDP 探测" field="udp_enabled" triggerPropName="checked" rules={[{ required: true }]}><Switch /></Form.Item></Grid.Col>
        </Grid.Row>
        <Form.Item label="监听网卡" field="interface" extra="留空监听所有网卡；例如 eth0、ens160"><Input allowClear placeholder="全部网卡" maxLength={64} /></Form.Item>
        <Grid.Row gutter={20}>
          <Grid.Col xs={24} sm={8}><Form.Item label="不同端口阈值" field="distinct_ports" rules={[{ required: true }]}><InputNumber min={3} max={1024} style={{ width: '100%' }} /></Form.Item></Grid.Col>
          <Grid.Col xs={24} sm={8}><Form.Item label="统计窗口（秒）" field="window_seconds" rules={[{ required: true }]}><InputNumber min={1} max={300} style={{ width: '100%' }} /></Form.Item></Grid.Col>
          <Grid.Col xs={24} sm={8}><Form.Item label="告警冷却（秒）" field="cooldown_seconds" rules={[{ required: true }]}><InputNumber min={1} max={86400} style={{ width: '100%' }} /></Form.Item></Grid.Col>
        </Grid.Row>
        <Form.Item label="排除端口" field="excluded_ports" extra="逗号或空格分隔；这些目标端口不参与扫描聚合"><Input allowClear placeholder="例如：22, 80, 443" /></Form.Item>
        <Form.Item label="忽略来源 CIDR" field="ignored_cidrs" extra="每行一个 IPv4/IPv6 网段，常用于排除授权扫描器"><Input.TextArea autoSize={{ minRows: 2, maxRows: 5 }} placeholder={'例如：10.10.0.0/16\n2001:db8::/32'} /></Form.Item>
        {senseNode?.sense?.last_error && <Alert type="error" title="节点最近一次运行错误" content={senseNode.sense.last_error} />}
      </Form>
    </Modal>
  </>;
}
