import net from 'node:net';

const base = (process.env.HONEYPOT_PUBLIC_URL || 'http://127.0.0.1:8080').replace(/\/$/, '');
const username = process.env.HONEYPOT_ADMIN_USERNAME || 'admin';
const password = process.env.HONEYPOT_ADMIN_PASSWORD || 'Honeynet@123';
const nodeID = process.env.HONEYPOT_SMOKE_NODE_ID || '00000000-0000-4000-8000-000000000001';
let headers;
let potID;
let currentPort;

async function api(path, options = {}) {
  const response = await fetch(`${base}/api/v1${path}`, { ...options, headers: { ...headers, ...options.headers } });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(`${options.method || 'GET'} ${path}: HTTP ${response.status} ${payload.message || ''}`.trim());
  return payload.data;
}

async function freePort() {
  return new Promise((resolve, reject) => {
    const server = net.createServer();
    server.once('error', reject);
    server.listen(0, '127.0.0.1', () => {
      const address = server.address();
      server.close((error) => error ? reject(error) : resolve(address.port));
    });
  });
}

async function tcpOpen(port) {
  return new Promise((resolve) => {
    const socket = net.createConnection({ host: '127.0.0.1', port });
    const finish = (value) => {
      socket.removeAllListeners();
      socket.destroy();
      resolve(value);
    };
    socket.setTimeout(250);
    socket.once('connect', () => finish(true));
    socket.once('timeout', () => finish(false));
    socket.once('error', () => finish(false));
  });
}

async function waitFor(label, predicate, timeout = 10_000) {
  const deadline = Date.now() + timeout;
  let lastError;
  while (Date.now() < deadline) {
    try {
      if (await predicate()) return;
    } catch (error) {
      lastError = error;
    }
    await new Promise((resolve) => setTimeout(resolve, 200));
  }
  throw new Error(`${label} timed out${lastError ? `: ${lastError.message}` : ''}`);
}

async function waitStatus(status) {
  await waitFor(`pot status ${status}`, async () => (await api(`/pots/${potID}`)).actual_status === status);
}

async function waitPort(port, open) {
  await waitFor(`port ${port} ${open ? 'open' : 'closed'}`, async () => (await tcpOpen(port)) === open);
}

try {
  const login = await fetch(`${base}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ username, password }),
  });
  const auth = await login.json();
  if (!login.ok) throw new Error(`login: HTTP ${login.status}`);
  headers = { authorization: `Bearer ${auth.data.token}`, 'content-type': 'application/json' };

  const firstPort = await freePort();
  let secondPort = await freePort();
  while (secondPort === firstPort) secondPort = await freePort();
  currentPort = firstPort;

  const created = await api('/pots', {
    method: 'POST',
    body: JSON.stringify({
      node_id: nodeID,
      service_code: 'http',
      name: 'smoke-port-control-v1',
      port: firstPort,
      config: { bind: '127.0.0.1', title: 'smoke-v1' },
      desired_status: 'running',
    }),
  });
  potID = created.id;
  await waitStatus('running');
  await waitPort(firstPort, true);

  currentPort = secondPort;
  await api(`/pots/${potID}`, {
    method: 'PUT',
    body: JSON.stringify({ name: 'smoke-port-control-v2', port: secondPort, config: { bind: '127.0.0.1', title: 'smoke-v2' } }),
  });
  await waitStatus('running');
  await waitPort(firstPort, false);
  await waitPort(secondPort, true);
  const updated = await api(`/pots/${potID}`);
  if (updated.name !== 'smoke-port-control-v2' || updated.port !== secondPort) throw new Error('updated pot was not returned by GET');

  await api(`/pots/${potID}/stop`, { method: 'POST' });
  await waitStatus('stopped');
  await waitPort(secondPort, false);

  await api(`/pots/${potID}/start`, { method: 'POST' });
  await waitStatus('running');
  await waitPort(secondPort, true);

  await api(`/pots/${potID}`, { method: 'DELETE' });
  potID = undefined;
  await waitPort(secondPort, false);
  const deleted = await fetch(`${base}/api/v1/pots/${created.id}`, { headers });
  if (deleted.status !== 404) throw new Error(`deleted pot GET returned HTTP ${deleted.status}`);

  console.log(JSON.stringify({ create: 'listening', read: 'ok', update: 'port-moved', stop: 'released', start: 'listening', delete: 'released' }));
} finally {
  if (potID && headers) {
    await fetch(`${base}/api/v1/pots/${potID}`, { method: 'DELETE', headers }).catch(() => {});
    if (currentPort) await waitPort(currentPort, false).catch(() => {});
  }
}
