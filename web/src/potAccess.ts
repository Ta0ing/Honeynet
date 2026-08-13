export const BUILTIN_NODE_ID = '00000000-0000-4000-8000-000000000001';

const builtinHostPorts: Record<number, number> = {
  80: 8088,
  22: 2222,
  23: 2323,
  6379: 16379,
  3306: 3307,
  8080: 18080,
};

const localWebPortStart = 20000;
const localWebPortEnd = 20099;

// These codes are backed directly by honeypot-templates-server. All static
// Web profiles in that pack declare HTTPS in its authoritative config.json.
export const webTemplateServices = new Set<string>([
  'ac-sangfor', 'baota', 'canal', 'cisco-vpn', 'cloudreve', 'confluence', 'coremail', 'cpanel',
  'edr-sangfor', 'electric', 'esxi', 'exchange', 'filebrowser', 'fw-360', 'fw-haofeng', 'fw-nsfocus',
  'fw-topsec', 'fw-zkww', 'gitlab', 'gophish', 'huorong-zd', 'iis', 'intel-am', 'iot-hikcam',
  'isport', 'jenkins', 'jira', 'joomla', 'jspspy', 'jumpserver', 'kelai-qll', 'kibana', 'mailu',
  'nagios', 'nas-qnap', 'nginx', 'oa', 'oa-gov', 'oa-tongda', 'oa-yy', 'phpadmin', 'portainer',
  'poste', 'printer-dell', 'qzsec', 'router-aruba', 'router-h3c', 'router-ikuai', 'router-openwrt',
  'router-ruijie', 'router-tplink', 'routos', 'ruoyi', 'sangfor-fcg', 'sangfor-vpn', 'synology-nas',
  'tdp', 'thinkphp', 'tomcat', 'uniaccess-lr', 'weblogic', 'webmin', 'websphere', 'wordpress',
  'zabbix', 'zhongke-kongzhi', 'zimbra',
]);

const browserServices = new Set<string>([...webTemplateServices, 'http', 'https', 'web-template']);
const secureBrowserServices = new Set<string>([...webTemplateServices, 'https']);

export function isBrowserService(code: string) {
  return browserServices.has(code);
}

export function browserScheme(code: string) {
  return secureBrowserServices.has(code) ? 'https' : 'http';
}

export function builtinHostPort(port: number) {
  if (port >= localWebPortStart && port <= localWebPortEnd) return port;
  return builtinHostPorts[port];
}

export function availableBuiltinPort(serviceCode: string, defaultPort: number, rows: any[]) {
  if (!isBrowserService(serviceCode)) return defaultPort;
  const occupied = new Set(
    rows
      .filter((row) => row.node_id === BUILTIN_NODE_ID)
      .map((row) => Number(row.port)),
  );
  for (let candidate = localWebPortStart; candidate <= localWebPortEnd; candidate += 1) {
    if (!occupied.has(candidate)) return candidate;
  }
  return defaultPort;
}

function usableHost(value?: string) {
  const host = (value || '').trim();
  if (!host) return '';
  if (host.includes(':') && !host.startsWith('[')) return `[${host}]`;
  return host;
}

export function potAccessURL(row: any) {
  if (!row || !isBrowserService(row.service_code)) return '';
  const builtin = row.node_id === BUILTIN_NODE_ID;
  const port = builtin ? builtinHostPort(Number(row.port)) : Number(row.port);
  const host = builtin
    ? usableHost(window.location.hostname || '127.0.0.1')
    : usableHost(row.node?.ip);
  if (!host || !port) return '';
  return `${browserScheme(row.service_code)}://${host}:${port}/`;
}
