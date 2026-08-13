import net from 'node:net';

const base = (process.env.HONEYPOT_PUBLIC_URL || 'http://127.0.0.1:8080').replace(/\/$/, '');
const username = process.env.HONEYPOT_ADMIN_USERNAME || 'admin';
const password = process.env.HONEYPOT_ADMIN_PASSWORD || 'Honeynet@123';
const builtinNodeID = '00000000-0000-4000-8000-000000000001';
const targetPorts = [2222, 2323, 8088];
const sleep = (milliseconds) => new Promise((resolve) => setTimeout(resolve, milliseconds));
let headers;
let previous;

async function json(response, label) {
  if (!response.ok) throw new Error(`${label}: HTTP ${response.status} ${(await response.text()).slice(0, 500)}`);
  return response.json();
}

function desiredConfig(value) {
  return {
    enabled: value.enabled,
    interface: value.interface || '',
    tcp_enabled: value.tcp_enabled,
    udp_enabled: value.udp_enabled,
    distinct_ports: value.distinct_ports,
    window_seconds: value.window_seconds,
    cooldown_seconds: value.cooldown_seconds,
    excluded_ports: value.excluded_ports || [],
    ignored_cidrs: value.ignored_cidrs || [],
  };
}

function connect(port) {
  return new Promise((resolve) => {
    const socket = net.createConnection({ host: '127.0.0.1', port });
    const done = () => { socket.destroy(); resolve(); };
    socket.setTimeout(1000, done);
    socket.once('connect', done);
    socket.once('error', done);
  });
}

try {
  const auth = await json(await fetch(`${base}/api/v1/auth/login`, {
    method: 'POST', headers: { 'content-type': 'application/json' }, body: JSON.stringify({ username, password }),
  }), 'login');
  headers = { authorization: `Bearer ${auth.data.token}`, 'content-type': 'application/json' };
  previous = (await json(await fetch(`${base}/api/v1/nodes/${builtinNodeID}/sense`, { headers }), 'get sense config')).data;

  const startedAt = Math.floor(Date.now() / 1000);
  await json(await fetch(`${base}/api/v1/nodes/${builtinNodeID}/sense`, {
    method: 'PUT', headers, body: JSON.stringify({
      enabled: true, interface: '', tcp_enabled: true, udp_enabled: true,
      distinct_ports: 3, window_seconds: 10, cooldown_seconds: 30,
      excluded_ports: [], ignored_cidrs: [],
    }),
  }), 'enable sense');

  let status;
  for (let attempt = 0; attempt < 20; attempt++) {
    status = (await json(await fetch(`${base}/api/v1/nodes/${builtinNodeID}/sense`, { headers }), 'poll sense status')).data;
    if (status.actual_status === 'running') break;
    if (status.actual_status === 'error' || status.actual_status === 'unsupported') throw new Error(`sense did not start: ${status.last_error || status.actual_status}`);
    await sleep(500);
  }
  if (status?.actual_status !== 'running') throw new Error(`sense start timed out: ${JSON.stringify(status)}`);

  for (const port of targetPorts) await connect(port);

  let event;
  for (let attempt = 0; attempt < 30; attempt++) {
    const records = (await json(await fetch(`${base}/api/v1/events?event_type=port.scan&page_size=100`, { headers }), 'list scan events')).data.items;
    event = records.find((item) => {
      const timestamp = typeof item.ts === 'number' ? item.ts : Date.parse(item.ts) / 1000;
      return item.node_id === builtinNodeID && item.service === 'sense' && timestamp >= startedAt;
    });
    if (event) break;
    await sleep(500);
  }
  if (!event) throw new Error('no passive port.scan event was ingested');
  const ports = Array.isArray(event.payload?.ports) ? event.payload.ports : JSON.parse(event.payload || '{}').ports;
  if (!targetPorts.every((_, index) => ports?.includes([22, 23, 80][index]))) throw new Error(`unexpected detected ports: ${JSON.stringify(ports)}`);

  let finalStatus;
  for (let attempt = 0; attempt < 70; attempt++) {
    finalStatus = (await json(await fetch(`${base}/api/v1/nodes/${builtinNodeID}/sense`, { headers }), 'poll detection status')).data;
    if (finalStatus.detections > 0) break;
    await sleep(500);
  }
  if (!finalStatus || finalStatus.detections < 1) throw new Error(`sense detection counter was not reported: ${JSON.stringify(finalStatus)}`);
  const alerts = (await json(await fetch(`${base}/api/v1/alerts?page_size=100`, { headers }), 'list alerts')).data.items;
  console.log(JSON.stringify({ passive_capture: 'running', event_type: event.event_type, source_ip: event.src_ip, detected_ports: ports, alert_created: alerts.some((item) => item.event_id === event.event_id), detections: finalStatus.detections }));
} finally {
  if (headers && previous) {
    await fetch(`${base}/api/v1/nodes/${builtinNodeID}/sense`, { method: 'PUT', headers, body: JSON.stringify(desiredConfig(previous)) });
  }
}
