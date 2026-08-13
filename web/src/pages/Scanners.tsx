import { useMemo, useState } from 'react';
import { Alert, Button, Input, Space, Table, Tag, Typography } from '@arco-design/web-react';
import { IconDownload, IconRefresh, IconSearch, IconSettings } from '@arco-design/web-react/icon';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { api, formatTime } from '../api';
import PageHeader from '../components/PageHeader';
import { nodeName, publicFirstNodeAddress } from '../presentation';

function payloadOf(row: any) {
  if (row?.payload && typeof row.payload === 'object') return row.payload;
  try { return JSON.parse(row?.payload || '{}'); } catch { return {}; }
}

function scanNodeName(row: any, nodes: Map<string, any>) {
  return String(row?.node_name || nodeName(nodes.get(row?.node_id), row?.node_id));
}

function scanDisplayIP(row: any, nodes: Map<string, any>) {
  return String(row?.display_dst_ip || row?.node_public_ip || row?.node_address || publicFirstNodeAddress(nodes.get(row?.node_id), row?.observed_dst_ip || row?.dst_ip) || '');
}

function scanObservedIP(row: any) {
  return String(row?.observed_dst_ip || row?.dst_ip || '');
}

function exportScans(items: any[], nodes: Map<string, any>) {
  const lines = [['扫描 IP', '被扫描节点', '优先分析 IP', '原始目标 IP', '扫描类型', '扫描总次数', '被扫描端口', '扫描开始时间']];
  items.forEach((row) => {
    const payload = payloadOf(row);
    lines.push([row.src_ip, scanNodeName(row, nodes), scanDisplayIP(row, nodes), scanObservedIP(row), String(payload.protocol || 'TCP').toUpperCase(), payload.attempts || 1, (payload.ports || []).join('|'), row.ts]);
  });
  const csv = lines.map((line) => line.map((value) => `"${String(value ?? '').replaceAll('"', '""')}"`).join(',')).join('\n');
  const anchor = document.createElement('a');
  anchor.href = URL.createObjectURL(new Blob([`\uFEFF${csv}`], { type: 'text/csv;charset=utf-8' }));
  anchor.download = `honeynet-scans-${Date.now()}.csv`;
  anchor.click();
  URL.revokeObjectURL(anchor.href);
}

export default function Scanners() {
  const [source, setSource] = useState('');
  const navigate = useNavigate();
  const query = useQuery<any>({ queryKey: ['events', 'scans', source], queryFn: () => api.get('/events', { params: { event_type: 'port.scan', src_ip: source || undefined, page_size: 100 } }) });
  const nodes = useQuery<any>({ queryKey: ['nodes', 'scan-names'], queryFn: () => api.get('/nodes?page_size=100') });
  const nodeMap = useMemo(() => new Map<string, any>((nodes.data?.items || []).map((node: any) => [node.id, node])), [nodes.data]);
  const items = query.data?.items || [];
  const columns: any[] = [
    { title: '扫描 IP', dataIndex: 'src_ip', render: (value: string, row: any) => <div><Typography.Text code>{value}</Typography.Text><br /><Typography.Text type="secondary">{row.geo || '未知位置'}</Typography.Text></div> },
    { title: '被扫描节点', render: (_: any, row: any) => { const address = String(row.node_address || scanDisplayIP(row, nodeMap)); return <div className="entity-cell"><Typography.Text bold>{scanNodeName(row, nodeMap)}</Typography.Text><Typography.Text type="secondary" code copyable={Boolean(address)}>{address || '地址未上报'}</Typography.Text></div>; } },
    { title: '被扫描 IP', dataIndex: 'dst_ip', render: (_value: string, row: any) => { const preferred = scanDisplayIP(row, nodeMap); const observed = scanObservedIP(row); return <div className="entity-cell"><Typography.Text code copyable={Boolean(preferred)}>{preferred || '节点地址'}</Typography.Text>{observed && observed !== preferred && <Typography.Text type="secondary">Agent 观测地址：{observed}</Typography.Text>}</div>; } },
    { title: '扫描类型', render: (_: any, row: any) => <Tag color="orange">{String(payloadOf(row).protocol || 'TCP').toUpperCase()}</Tag> },
    { title: '扫描总次数', render: (_: any, row: any) => payloadOf(row).attempts || 1 },
    { title: '被扫描端口', width: 260, render: (_: any, row: any) => <Space wrap>{(payloadOf(row).ports || [row.dst_port]).slice(0, 12).map((port: number) => <Tag key={port}>{port}</Tag>)}</Space> },
    { title: '不同端口', render: (_: any, row: any) => payloadOf(row).distinct_ports || (payloadOf(row).ports || []).length },
    { title: '扫描开始时间', dataIndex: 'ts', render: formatTime, width: 180 },
  ];
  return <>
    <PageHeader title="扫描感知" description="被动观察节点收到的 TCP SYN、UDP 与 ICMP 探测，不开放额外端口、不修改防火墙" extra={<Space><Button icon={<IconDownload />} disabled={!items.length} onClick={() => exportScans(items, nodeMap)}>导出</Button><Button type="primary" icon={<IconSettings />} onClick={() => navigate('/nodes')}>高级配置</Button></Space>} />
    <Alert type="info" content="扫描事件由 Agent 全端口感知能力产生。可在节点管理中按节点启停、设置端口阈值、时间窗口、冷却时间和忽略网段。" style={{ marginBottom: 16 }} />
    <div className="filter-bar"><Space><Input allowClear prefix={<IconSearch />} placeholder="搜索扫描来源 IP" value={source} onChange={setSource} style={{ width: 260 }} /><Button onClick={() => setSource('')}>重置</Button></Space><Space><Typography.Text type="secondary">共 {query.data?.total || 0} 条扫描记录</Typography.Text><Button type="text" icon={<IconRefresh />} loading={query.isFetching} onClick={() => query.refetch()}>刷新</Button></Space></div>
    <div className="table-panel"><Table rowKey="event_id" loading={query.isLoading} data={items} columns={columns} pagination={{ pageSize: 20 }} scroll={{ x: 1280 }} /></div>
  </>;
}
