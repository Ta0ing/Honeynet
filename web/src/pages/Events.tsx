import { useEffect, useMemo, useRef, useState } from 'react';
import { Alert, Button, DatePicker, Drawer, Input, Message, Pagination, Select, Space, Table, Tabs, Tag, Typography } from '@arco-design/web-react';
import { IconDownload, IconEye, IconRefresh, IconRobot, IconSafe, IconSearch } from '@arco-design/web-react/icon';
import { useMutation, useQuery } from '@tanstack/react-query';
import { api, errorMessage, formatTime } from '../api';
import PageHeader from '../components/PageHeader';
import RiskTag from '../components/RiskTag';
import { buildServiceNameMap, eventTypeLabel, nodeName, preferredNodeAddress, riskMeta, serviceName } from '../presentation';
import { useAuth } from '../store';

type EventFilters = { src_ip: string; event_type: string; service: string; node_id: string; from: string; to: string };

const EVENT_TOTAL_REFRESH_INTERVAL_MS = 60_000;

function emptyEventFilters(): EventFilters {
  return { src_ip: '', event_type: '', service: '', node_id: '', from: '', to: '' };
}

function payloadOf(row: any) {
  if (row?.payload && typeof row.payload === 'object') return row.payload;
  try { return JSON.parse(row?.payload || '{}'); } catch { return {}; }
}

function eventColor(value: string) {
  if (value?.includes('credential') || value?.startsWith('decoy.')) return 'red';
  if (value === 'port.scan') return 'orange';
  if (value?.startsWith('web.')) return 'arcoblue';
  return 'purple';
}

function detectionsOf(row: any): any[] {
  if (Array.isArray(row?.detections)) return row.detections;
  try { return JSON.parse(row?.detections || '[]'); } catch { return []; }
}

function detectionSourceLabel(value: string) {
  return value === 'builtin' ? '内置规则' : value === 'custom' ? '自定义规则' : '导入规则';
}

function readableValue(value: any): string {
  if (value == null || value === '') return '—';
  if (typeof value === 'string') return value;
  if (Array.isArray(value)) return value.map(readableValue).join(', ');
  if (typeof value === 'object') return Object.entries(value).map(([key, item]) => `${key}: ${readableValue(item)}`).join('\n');
  return String(value);
}

function isEvidenceSummary(value: any) {
  return value && typeof value === 'object' && value.redacted === true && typeof value.sha256 === 'string';
}

function readableEvidenceValue(value: any): string {
  if (!isEvidenceSummary(value)) return readableValue(value);
  const kind = { text: '文本', object: '对象', array: '数组', number: '数值', boolean: '布尔值' }[value.kind as string] || '数据';
  return `敏感原文已隐藏\n类型：${kind}\n原始长度：${value.byte_length ?? 0} 字节\nSHA-256：${value.sha256}`;
}

function evidenceByteLength(value: any) {
  if (isEvidenceSummary(value)) return Number(value.byte_length || 0);
  return new TextEncoder().encode(typeof value === 'string' ? value : readableValue(value)).length;
}

const fieldLabels: Record<string, string> = {
  method: '请求方法', path: '请求路径', host: 'Host', scheme: '协议', user_agent: 'User-Agent', headers: '请求头', body: '请求正文',
  command: '命令', query: '查询语句', username: '用户名', password: '密码', protocol: '协议', mechanism: '认证方式', template_code: '蜜罐模板',
  template_name: '模板名称', template_source: '模板来源', body_truncated: '正文是否截断', matched: '是否匹配页面', filename: '文件名', content: '内容', name: '名称',
};

function normalizedFields(row: any, serviceNames: Map<string, string>) {
  const payload = payloadOf(row);
  return Object.entries(payload).filter(([key]) => !['raw_request', 'body_base64'].includes(key)).map(([key, value]) => ({
    key,
    label: fieldLabels[key] || key,
    value: key === 'template_code' || key === 'service'
      ? serviceName(value, serviceNames)
      : key === 'template_source' ? '内置 Web 模板资源' : readableEvidenceValue(value),
  }));
}

function eventTagLabel(value: string, serviceNames: Map<string, string>) {
  const tag = String(value || '').trim();
  const service = serviceNames.get(tag);
  if (service) return service;
  const normalized = tag.toLowerCase();
  if (normalized === 'template' || normalized.endsWith('-template')) return 'Web 模板';
  if (normalized === 'credential') return '凭据尝试';
  if (normalized === 'detected') return '已命中检测';
  if (normalized === 'server-reviewed') return 'Server 已复核';
  return tag;
}

function packetOf(row: any) {
  if (row?.raw_packet) return row.raw_packet;
  const payload = payloadOf(row);
  if (payload.raw_request) return readableEvidenceValue(payload.raw_request);
  if (!payload.method) return '';
  const headers = readableEvidenceValue(payload.headers);
  const body = payload.body ? readableEvidenceValue(payload.body) : '';
  return `${payload.method} ${payload.path || '/'} HTTP/1.1\r\nHost: ${payload.host || row.dst_ip || ''}\r\n${headers === '—' ? '' : `${headers}\r\n`}\r\n${body}`;
}

function displayNodeName(row: any, nodes: Map<string, any>) {
  return String(row?.node_name || nodeName(nodes.get(row?.node_id), row?.node_id));
}

function displayNodeAddress(row: any, nodes: Map<string, any>) {
  return String(row?.display_dst_ip || row?.node_address || preferredNodeAddress(nodes.get(row?.node_id), row?.observed_dst_ip || row?.dst_ip) || '');
}

function observedNodeAddress(row: any) {
  return String(row?.observed_dst_ip || row?.dst_ip || '');
}

function threatIntelOf(row: any) {
  const value = row?.threat_intelligence;
  return value && value.matched ? value : undefined;
}

function threatIntelColor(level: number) {
  if (level >= 4) return 'red';
  if (level >= 2) return 'orange';
  return 'arcoblue';
}

function ThreatIntelTags({ row, compact = false }: { row: any; compact?: boolean }) {
  const intelligence = threatIntelOf(row);
  if (!intelligence) return null;
  const values = Array.isArray(intelligence.tags) ? intelligence.tags : [...(intelligence.labels || []), ...(intelligence.behaviors || [])];
  return <Space size={4} wrap={!compact} className="threat-intel-tags"><Tag color={threatIntelColor(Number(intelligence.level || 0))}>情报命中</Tag>{values.slice(0, compact ? 1 : 6).map((tag: string) => <Tag key={tag}>{tag}</Tag>)}</Space>;
}

function exportEvents(items: any[], serviceNames: Map<string, string>, nodes: Map<string, any>) {
  const rows = [['时间', '攻击行为', '来源IP', '来源端口', '威胁情报标签', '被攻击节点', '节点地址', '目标端口', '蜜罐服务', '地理位置']];
  items.forEach((row) => {
    rows.push([row.ts, eventTypeLabel(row.event_type), row.src_ip, row.src_port, (threatIntelOf(row)?.tags || []).join('；'), displayNodeName(row, nodes), displayNodeAddress(row, nodes), row.dst_port, row.service_name || serviceName(row.service, serviceNames), row.geo]);
  });
  const csv = rows.map((row) => row.map((value) => `"${String(value ?? '').replaceAll('"', '""')}"`).join(',')).join('\n');
  const anchor = document.createElement('a');
  anchor.href = URL.createObjectURL(new Blob([`\uFEFF${csv}`], { type: 'text/csv;charset=utf-8' }));
  anchor.download = `honeynet-events-${Date.now()}.csv`;
  anchor.click();
  URL.revokeObjectURL(anchor.href);
}

function eventDisclosureNotice(row: any) {
  if (row?.evidence_redacted) {
    return {
      type: 'info' as const,
      title: '敏感事件脱敏已启用',
      content: '凭据已掩码，请求包和正文仅展示长度与 SHA-256 摘要。持久化证据未被修改，有权限人员可按单条事件显式查看原文。',
    };
  }
  if (row?.sensitive_reveal_audited) {
    return {
      type: 'warning' as const,
      title: '正在查看已审计的敏感原文',
      content: '该事件的请求包、正文与凭据已原样加载，请按敏感数据规范使用；本次显式查看已记录 READ 审计。',
    };
  }
  return {
    type: 'info' as const,
    title: '敏感事件脱敏未启用',
    content: '当前 Server 策略默认展示原始请求包、正文与凭据；这不是一次显式解密查看，不会产生敏感原文查看审计。',
  };
}

export default function Events() {
	const { user } = useAuth();
  const [draftFilters, setDraftFilters] = useState<EventFilters>(emptyEventFilters);
  const [filters, setFilters] = useState<EventFilters>(emptyEventFilters);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [knownTotal, setKnownTotal] = useState<number | null>(null);
  const [searchRevision, setSearchRevision] = useState(0);
  const lastExactTotalAt = useRef(0);
  const [selected, setSelected] = useState<any>();
	const [sensitiveLoading, setSensitiveLoading] = useState(false);
  const [view, setView] = useState('grouped');
  const query = useQuery<any>({
    queryKey: ['events', filters, page, pageSize, searchRevision],
    queryFn: () => api.get('/events', {
      params: {
        ...filters,
        from: filters.from ? new Date(filters.from.replace(' ', 'T')).toISOString() : undefined,
        to: filters.to ? new Date(filters.to.replace(' ', 'T')).toISOString() : undefined,
        pagination: 'page',
        page,
        page_size: pageSize,
        // The first request obtains the exact total. Realtime invalidations
        // refresh rows immediately, while exact recounts are throttled to one
        // per minute on the first page. A manual refresh clears knownTotal and
        // therefore forces an immediate recount from any page.
        include_total: knownTotal == null || (page === 1 && Date.now() - lastExactTotalAt.current >= EVENT_TOTAL_REFRESH_INTERVAL_MS),
      },
    }),
  });
  const services = useQuery<any>({ queryKey: ['services', 'display-catalog'], queryFn: () => api.get('/pot-services?page_size=200&include_retired=1') });
  const nodes = useQuery<any>({ queryKey: ['nodes', 'event-display'], queryFn: () => api.get('/nodes?page_size=200') });
  const aiStatus = useQuery<any>({ queryKey: ['ai', 'status'], queryFn: () => api.get('/ai/status') });
  const analyze = useMutation({ mutationFn: (eventID: string) => api.post(`/events/${eventID}/ai-analysis`), onSuccess: () => Message.success('AI 分析已完成，可在 AI 分析模块查看结果'), onError: (error) => Message.error(errorMessage(error)) });
  const items = query.data?.items || [];
	const maxPage = Number(query.data?.max_page || 100);
	const browsableTotal = Math.min(knownTotal ?? 0, pageSize * maxPage);
	const canRevealSensitive = user?.role === 'admin' || user?.role === 'operator';
  const serviceNames = useMemo(() => buildServiceNameMap(services.data?.items), [services.data]);
  const nodeMap = useMemo(() => new Map<string, any>((nodes.data?.items || []).map((node: any) => [node.id, node])), [nodes.data]);
  const groups = useMemo(() => {
    const result = new Map<string, any>();
    items.forEach((event: any) => {
      const key = `${event.service || 'unknown'}::${event.src_ip || 'unknown'}::${event.pot_id || event.node_id}`;
      const group = result.get(key) || { id: key, service: event.service || 'unknown', src_ip: event.src_ip || 'unknown', geo: event.geo, node_id: event.node_id, pot_id: event.pot_id, events: [], types: new Set<string>(), last_seen: event.ts };
      group.events.push(event);
      group.types.add(event.event_type);
      if (new Date(event.ts) > new Date(group.last_seen)) group.last_seen = event.ts;
      result.set(key, group);
    });
    return [...result.values()].sort((a, b) => +new Date(b.last_seen) - +new Date(a.last_seen));
  }, [items]);
  useEffect(() => {
    if (query.data?.total_known && Number.isFinite(Number(query.data.total))) {
      setKnownTotal(Number(query.data.total));
      lastExactTotalAt.current = Date.now();
    }
  }, [query.data]);
  const rawColumns: any[] = [
    { title: '时间', dataIndex: 'ts', render: formatTime, width: 180 },
    { title: '攻击行为', render: (_: any, row: any) => { const hit = detectionsOf(row)[0]; return hit ? <Space size={6}><RiskTag value={hit.severity} /><Typography.Text>{hit.name}</Typography.Text></Space> : <Tag color={eventColor(row.event_type)} title={row.event_type}>{eventTypeLabel(row.event_type)}</Tag>; } },
    { title: '攻击来源', width: 210, render: (_: any, row: any) => <div className="entity-cell"><Typography.Text code>{row.src_ip}:{row.src_port}</Typography.Text><Typography.Text type="secondary">{row.geo || '未知位置'}</Typography.Text><ThreatIntelTags row={row} compact /></div> },
    { title: '被攻击节点', width: 210, render: (_: any, row: any) => { const address = displayNodeAddress(row, nodeMap); return <div className="entity-cell"><Typography.Text bold>{displayNodeName(row, nodeMap)}</Typography.Text><Typography.Text type="secondary" code copyable={Boolean(address)}>{address || '地址未上报'}</Typography.Text></div>; } },
    { title: '目标端口', dataIndex: 'dst_port', render: (value: number) => value || '—' },
    { title: '服务', dataIndex: 'service', render: (value: string, row: any) => <span className="service-name" title={value}>{row.service_name || serviceName(value, serviceNames)}</span> },
    { title: '请求 / 行为', render: (_: any, row: any) => {
      const payload = payloadOf(row);
      return <Typography.Text ellipsis={{ showTooltip: true }}>{payload.method ? `${payload.method} ${payload.path || '/'}` : payload.command || payload.protocol || eventTypeLabel(row.event_type)}</Typography.Text>;
    } },
    { title: '操作', width: 90, render: (_: any, row: any) => <Button type="text" icon={<IconEye />} onClick={() => void openDetail(row)}>详情</Button> },
  ];
  const groupColumns: any[] = [
    { title: '蜜罐服务', dataIndex: 'service', render: (value: string, row: any) => <div className="entity-cell"><Typography.Text bold title={value}>{row.events?.[0]?.service_name || serviceName(value, serviceNames)}</Typography.Text><Typography.Text type="secondary">{row.pot_id ? `实例 ${row.pot_id.slice(0, 8)}` : '节点感知'}</Typography.Text></div> },
    { title: '被攻击数量', render: (_: any, row: any) => <Tag color="red">{row.events.length}</Tag> },
    { title: '被攻击节点', render: (_: any, row: any) => { const event = row.events?.[0] || row; const address = displayNodeAddress(event, nodeMap); return <div className="entity-cell"><Typography.Text bold>{displayNodeName(event, nodeMap)}</Typography.Text><Typography.Text type="secondary" code copyable={Boolean(address)}>{address || '地址未上报'}</Typography.Text></div>; } },
    { title: '攻击来源', width: 210, render: (_: any, row: any) => <div className="entity-cell"><Typography.Text code>{row.src_ip}</Typography.Text><Typography.Text type="secondary">{row.geo || '未知位置'}</Typography.Text><ThreatIntelTags row={row.events?.find((event: any) => threatIntelOf(event)) || row.events?.[0]} compact /></div> },
    { title: '攻击行为', render: (_: any, row: any) => <Space wrap>{row.events.flatMap((event: any) => detectionsOf(event)).slice(0, 3).map((hit: any, index: number) => <Tag key={`${hit.rule_key}-${index}`} color={riskMeta(hit.severity).color}>{hit.name}</Tag>)}{!row.events.some((event: any) => detectionsOf(event).length) && [...row.types].slice(0, 4).map((type: any) => <Tag key={type} color={eventColor(type)} title={type}>{eventTypeLabel(type)}</Tag>)}</Space> },
    { title: '最近一次被攻击时间', dataIndex: 'last_seen', render: formatTime, width: 190 },
  ];
  const innerColumns: any[] = [
    { title: '时间', dataIndex: 'ts', render: formatTime, width: 180 },
    { title: '请求类型', dataIndex: 'event_type', render: (value: string) => <Tag color={eventColor(value)} title={value}>{eventTypeLabel(value)}</Tag> },
    { title: '请求详情 / 行为', render: (_: any, row: any) => { const payload = payloadOf(row); return payload.method ? `${payload.method} ${payload.path || '/'}` : payload.command || payload.protocol || eventTypeLabel(row.event_type); } },
    { title: '数据长度', render: (_: any, row: any) => `${evidenceByteLength(payloadOf(row).body || '')} 字节` },
    { title: '攻击详情', width: 90, render: (_: any, row: any) => <Button type="text" icon={<IconEye />} onClick={() => void openDetail(row)}>查看</Button> },
  ];
  const applyFilters = (next = draftFilters) => {
    setFilters({ ...next });
    setPage(1);
    setKnownTotal(null);
    lastExactTotalAt.current = 0;
    setSearchRevision((current) => current + 1);
  };
  const updatePageSize = (value: number) => {
    setPageSize(value);
    setPage(1);
    setKnownTotal(null);
    lastExactTotalAt.current = 0;
  };
  const changeView = (nextView: string) => {
    setView(nextView);
  };
  const reset = () => {
    const next = emptyEventFilters();
    setDraftFilters(next);
    applyFilters(next);
  };
  const refreshEvents = () => {
    setKnownTotal(null);
    lastExactTotalAt.current = 0;
    setSearchRevision((current) => current + 1);
  };
  const openDetail = async (row: any) => {
    setSelected(row);
    try {
      // Always refresh a single event from the detail endpoint. This avoids
      // treating a paginated/list projection as authoritative forensic data.
      setSelected(await api.get(`/events/${row.event_id}`));
    } catch (error) {
      Message.error(errorMessage(error));
    }
  };
  const revealSensitiveEvidence = async () => {
    if (!selected?.event_id || !canRevealSensitive) return;
    setSensitiveLoading(true);
    try {
      setSelected(await api.get(`/events/${selected.event_id}`, { params: { include_sensitive: true } }));
      Message.warning('已加载敏感原文，本次查看已写入审计日志');
    } catch (error) {
      Message.error(errorMessage(error));
    } finally {
      setSensitiveLoading(false);
    }
  };
  const pageStart = items.length ? (page - 1) * pageSize + 1 : 0;
  const pageEnd = items.length ? pageStart + items.length - 1 : 0;
  const totalExceedsWindow = (knownTotal ?? 0) > pageSize * maxPage;
  return <>
    <PageHeader title="攻击列表" description="按时间和条件分页查询攻击事件；聚合视图对当前页进行会话归并，可展开查看原始请求和协议载荷" extra={<Space><Button icon={<IconDownload />} disabled={!items.length} onClick={() => exportEvents(items, serviceNames, nodeMap)}>导出当前页</Button><Button icon={<IconRefresh />} loading={query.isFetching} onClick={refreshEvents}>刷新</Button></Space>} />
    <div className="filter-bar"><Space wrap><Input allowClear prefix={<IconSearch />} placeholder="攻击者 IP" value={draftFilters.src_ip} onPressEnter={() => applyFilters()} onChange={(value) => setDraftFilters({ ...draftFilters, src_ip: value })} /><Select allowClear showSearch placeholder="被攻击节点" style={{ width: 190 }} value={draftFilters.node_id || undefined} onChange={(value) => setDraftFilters({ ...draftFilters, node_id: value || '' })} options={(nodes.data?.items || []).map((item: any) => ({ label: `${item.name} · ${preferredNodeAddress(item) || '地址未上报'}`, value: item.id }))} /><Select allowClear showSearch placeholder="蜜罐服务" style={{ width: 190 }} value={draftFilters.service || undefined} onChange={(value) => setDraftFilters({ ...draftFilters, service: value || '' })} options={(services.data?.items || []).map((item: any) => ({ label: item.name, value: item.code }))} /><Select allowClear placeholder="攻击行为" style={{ width: 180 }} value={draftFilters.event_type || undefined} onChange={(value) => setDraftFilters({ ...draftFilters, event_type: value || '' })} options={[{ label: '凭据捕获', value: 'web.credential' }, { label: 'Web 请求', value: 'web.request' }, { label: '端口扫描', value: 'port.scan' }, { label: 'SSH 登录', value: 'ssh.credential' }, { label: '文件蜜饵', value: 'decoy.file' }]} /><DatePicker.RangePicker showTime format="YYYY-MM-DD HH:mm:ss" value={draftFilters.from && draftFilters.to ? [draftFilters.from, draftFilters.to] : undefined} onChange={(value) => { const range = (value || []) as string[]; setDraftFilters({ ...draftFilters, from: range[0] || '', to: range[1] || '' }); }} /><Button type="primary" icon={<IconSearch />} loading={query.isFetching} onClick={() => applyFilters()}>查询</Button><Button onClick={reset}>重置</Button></Space><Typography.Text type="secondary">{query.isLoading && knownTotal == null ? '正在查询事件…' : `共 ${knownTotal ?? 0} 条事件 · 当前页归并为 ${groups.length} 个攻击会话`}</Typography.Text></div>
    <Tabs activeTab={view} onChange={changeView} className="view-tabs"><Tabs.TabPane key="grouped" title="聚合视图" /><Tabs.TabPane key="raw" title="原始事件" /></Tabs>
    <div className="table-panel">{view === 'grouped' ? <Table rowKey="id" loading={query.isFetching} data={groups} columns={groupColumns} pagination={false} scroll={{ x: 1180 }} expandedRowRender={(row) => <div className="event-group-detail"><div className="request-type-strip"><Tag>全部 {row.events.length}</Tag>{[...row.types].map((type: any) => <Tag key={type} color={eventColor(type)} title={type}>{eventTypeLabel(type)}</Tag>)}</div><Table rowKey="event_id" data={row.events} columns={innerColumns} pagination={false} size="small" scroll={{ x: 820 }} /></div>} expandProps={{ expandRowByClick: true }} /> : <Table rowKey="event_id" loading={query.isFetching} data={items} columns={rawColumns} pagination={false} scroll={{ x: 1360 }} />}<div className="table-pagination event-list-pagination"><Typography.Text type="secondary">{items.length ? `第 ${page} 页 · 显示第 ${pageStart}-${pageEnd} 条` : `第 ${page} 页 · 暂无事件`}{totalExceedsWindow ? ` · 为保证查询性能，最多浏览前 ${maxPage} 页` : ''}</Typography.Text><Pagination current={page} pageSize={pageSize} total={browsableTotal} disabled={query.isFetching} showTotal={false} showJumper sizeCanChange pageSizeChangeResetCurrent sizeOptions={[10, 20, 50, 100, 200]} onChange={(nextPage) => setPage(Math.min(nextPage, maxPage))} onPageSizeChange={updatePageSize} /></div></div>
    <Drawer width={760} title="攻击事件详情" visible={!!selected} onCancel={() => setSelected(undefined)} footer={null}>{selected && (() => { const displayAddress = displayNodeAddress(selected, nodeMap); const observedAddress = observedNodeAddress(selected); const notice = eventDisclosureNotice(selected); const intelligence = threatIntelOf(selected); return <div className="event-detail"><Alert type={notice.type} title={notice.title} content={notice.content} /><div className="event-detail-actions"><Space wrap>{aiStatus.data?.enabled && <Button type="primary" icon={<IconRobot />} loading={analyze.isPending} onClick={() => analyze.mutate(selected.event_id)}>AI 自动分析</Button>}{canRevealSensitive && selected.evidence_redacted && <Button status="danger" icon={<IconSafe />} loading={sensitiveLoading} onClick={() => void revealSensitiveEvidence()}>查看敏感原文</Button>}{selected.sensitive_reveal_audited && <Tag color="red">敏感原文已显示 · 已审计</Tag>}{selected.sensitive_visible && !selected.sensitive_reveal_audited && <Tag color="arcoblue">脱敏未启用 · 原文展示</Tag>}</Space></div><Detail label="事件 ID" value={selected.event_id} />{selected.decoy_id && <Detail label="关联蜜饵" value={selected.decoy_id} />}<Detail label="时间" value={formatTime(selected.ts)} /><Detail label="来源" value={`${selected.src_ip}:${selected.src_port}`} />{intelligence && <section className="detection-section"><Typography.Title heading={6}>离线威胁情报</Typography.Title><ThreatIntelTags row={selected} /><Typography.Paragraph type="secondary" style={{ marginTop: 8 }}>该标签来自 Server 本地 IPv4 / IPv6 情报库，历史事件会随当前情报库动态补充，不改写原始取证数据。</Typography.Paragraph></section>}<Detail label="被攻击节点" value={`${displayNodeName(selected, nodeMap)} / ${displayAddress || '节点地址'}:${selected.dst_port}`} />{observedAddress && observedAddress !== displayAddress && <Detail label="Agent 观测地址" value={`${observedAddress}:${selected.dst_port}`} />}<Detail label="类型 / 服务" value={`${eventTypeLabel(selected.event_type)} / ${selected.service_name || serviceName(selected.service, serviceNames)}`} />{detectionsOf(selected).length > 0 && <section className="detection-section"><Typography.Title heading={6}>检测结论</Typography.Title>{detectionsOf(selected).map((hit: any, index: number) => <div className="detection-hit" key={`${hit.rule_key}-${index}`}><div><RiskTag value={hit.severity} /><Typography.Text bold>{hit.name}</Typography.Text></div><Typography.Paragraph>{hit.description || '请求特征命中检测规则'}</Typography.Paragraph><Space wrap><Tag>{detectionSourceLabel(hit.source)}</Tag><Tag color={hit.confirmed ? 'green' : hit.stage === 'agent' ? 'gold' : 'arcoblue'}>{hit.confirmed ? 'Agent + Server 双重确认' : hit.stage === 'agent' ? 'Agent 命中，Server 未确认' : 'Server 二次确认'}</Tag>{hit.external_id && <Tag>{hit.external_id}</Tag>}</Space></div>)}</section>}<Typography.Title heading={6}>{selected.sensitive_visible ? '原始请求包' : '请求包安全摘要'}</Typography.Title>{packetOf(selected) ? <pre className="request-packet">{packetOf(selected)}</pre> : <div className="packet-empty">该协议事件没有 HTTP 请求包，以下展示 Agent 归一化后的协议字段。</div>}<Typography.Title heading={6}>归一化字段</Typography.Title><div className="normalized-fields">{normalizedFields(selected, serviceNames).map((field) => <div key={field.key}><span>{field.label}</span><pre>{field.value}</pre></div>)}</div>{payloadOf(selected).body_base64 && <><Typography.Title heading={6}>{selected.sensitive_visible ? '二进制正文（Base64 原样保留）' : '二进制正文安全摘要'}</Typography.Title><pre className="request-packet">{readableEvidenceValue(payloadOf(selected).body_base64)}</pre></>}<Typography.Title heading={6}>标签</Typography.Title><Space wrap>{(selected.tags || []).map((tag: string) => <Tag key={tag}>{eventTagLabel(tag, serviceNames)}</Tag>)}</Space></div>; })()}</Drawer>
  </>;
}

function Detail({ label, value }: { label: string; value: string }) {
  return <div className="detail-row"><Typography.Text type="secondary">{label}</Typography.Text><Typography.Text copyable>{value || '—'}</Typography.Text></div>;
}
