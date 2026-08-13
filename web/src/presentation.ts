export type ServiceCatalogItem = { code?: string; name?: string };
export type NodeDisplayItem = {
  id?: string;
  name?: string;
  ip?: string;
  public_ip?: string;
  private_ips?: string[];
};

export const RISK_LEVELS = ['critical', 'high', 'medium', 'low', 'info'] as const;

const riskMetadata: Record<string, { label: string; color: string }> = {
  critical: { label: '紧急', color: '#7f1d1d' },
  high: { label: '高危', color: '#f53f3f' },
  medium: { label: '中危', color: '#ffb400' },
  low: { label: '低危', color: '#00b42a' },
  info: { label: '信息', color: '#86909c' },
};

export function normalizeRiskLevel(value?: unknown) {
  const level = String(value || '').trim().toLowerCase();
  if (level === 'emergency' || level === 'urgent') return 'critical';
  if (level === 'warning' || level === 'warn') return 'medium';
  return riskMetadata[level] ? level : 'info';
}

export function riskMeta(value?: unknown) {
  return riskMetadata[normalizeRiskLevel(value)];
}

const exactEventLabels: Record<string, string> = {
  '*.credential': '所有凭据尝试',
  'decoy.*': '所有蜜饵命中',
  'web.request': 'Web 访问请求',
  'web.credential': 'Web 凭据尝试',
  'port.scan': '端口扫描',
  'ssh.credential': 'SSH 凭据尝试',
  'ssh.command': 'SSH 命令交互',
  'ssh.session': 'SSH 会话建立',
  'decoy.file': '文件蜜饵命中',
  'decoy.network': '网络蜜饵命中',
};

const protocolLabels: Record<string, string> = {
  web: 'Web', port: '端口', ssh: 'SSH', telnet: 'Telnet', ftp: 'FTP', tftp: 'TFTP',
  smtp: 'SMTP', pop3: 'POP3', imap: 'IMAP', mysql: 'MySQL', postgresql: 'PostgreSQL',
  mssql: 'SQL Server', oracle: 'Oracle', redis: 'Redis', mongodb: 'MongoDB',
  elasticsearch: 'Elasticsearch', memcached: 'Memcached', zookeeper: 'ZooKeeper', kafka: 'Kafka',
  ldap: 'LDAP', dns: 'DNS', smb: 'SMB', vnc: 'VNC', rdp: 'RDP', rtsp: 'RTSP', mqtt: 'MQTT',
  snmp: 'SNMP', coap: 'CoAP', modbus: 'Modbus', s7: 'S7', bacnet: 'BACnet', decoy: '蜜饵',
};

const actionLabels: Record<string, string> = {
  request: '访问请求', credential: '凭据尝试', authentication: '认证尝试',
  username: '用户名捕获', command: '命令交互', query: '查询操作', session: '会话建立',
  connection: '连接建立', connect: '连接建立', message: '消息投递', publish: '消息发布',
  subscribe: '主题订阅', data: '数据传输', item: '数据项操作', write: '写入操作',
  read: '读取操作', discovery: '设备发现', read_property: '属性读取',
  write_property: '属性写入', bind: '目录绑定', search: '目录搜索', unbind: '目录解绑',
  operation: '协议操作', prelogin: '登录前探测', negotiate: '协议协商', scan: '扫描',
  community: '社区字符串尝试', file: '文件蜜饵命中', network: '网络蜜饵命中',
};

export function eventTypeLabel(value?: unknown) {
  const eventType = String(value || '').trim().toLowerCase();
  if (!eventType) return '未知攻击行为';
  if (exactEventLabels[eventType]) return exactEventLabels[eventType];
  const [protocol, ...actionParts] = eventType.split('.');
  const action = actionParts.join('_');
  const actionLabel = actionLabels[action] || actionLabels[actionParts[actionParts.length - 1]];
  if (!actionLabel) return '其他协议行为';
  return `${protocolLabels[protocol] || '协议'} ${actionLabel}`;
}

export function buildServiceNameMap(items?: ServiceCatalogItem[]) {
  const result = new Map<string, string>();
  (items || []).forEach((item) => {
    const code = String(item.code || '').trim();
    const name = String(item.name || '').trim();
    if (code && name) result.set(code, name);
  });
  result.set('decoy', '蜜饵服务');
  result.set('sense', '全端口扫描感知');
  return result;
}

export function serviceName(value: unknown, names: Map<string, string>) {
  const code = String(value || '').trim();
  if (!code) return '未知服务';
  return names.get(code) || '未知服务';
}

function firstPrivateIP(node?: NodeDisplayItem) {
  return Array.isArray(node?.private_ips) ? String(node.private_ips.find(Boolean) || '') : '';
}

export function preferredNodeAddress(node?: NodeDisplayItem, fallback = '') {
  return String(node?.ip || node?.public_ip || firstPrivateIP(node) || fallback || '').trim();
}

export function publicFirstNodeAddress(node?: NodeDisplayItem, fallback = '') {
  return String(node?.public_ip || node?.ip || firstPrivateIP(node) || fallback || '').trim();
}

export function nodeName(node?: NodeDisplayItem, fallbackID = '') {
  return String(node?.name || (fallbackID ? `节点 ${fallbackID.slice(0, 8)}` : '未知节点'));
}

export function formatHostPort(host: string, port?: number) {
  const value = String(host || '').trim();
  if (!value) return port ? `:${port}` : '等待节点上报';
  const displayHost = value.includes(':') && !value.startsWith('[') ? `[${value}]` : value;
  return port ? `${displayHost}:${port}` : displayHost;
}
