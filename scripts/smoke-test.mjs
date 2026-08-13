import { createHash } from 'node:crypto';

const base = (process.env.HONEYPOT_PUBLIC_URL || 'http://127.0.0.1:8080').replace(/\/$/, '');
const username = process.env.HONEYPOT_ADMIN_USERNAME || 'admin';
const password = process.env.HONEYPOT_ADMIN_PASSWORD || 'Honeynet@123';
let nodeID;
let headers;

async function json(response, label) {
  if (!response.ok) throw new Error(`${label}: HTTP ${response.status}`);
  return response.json();
}

try {
  const login = await fetch(`${base}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ username, password }),
  });
  const auth = await json(login, 'login');
  headers = { authorization: `Bearer ${auth.data.token}`, 'content-type': 'application/json' };

  const created = await json(await fetch(`${base}/api/v1/nodes`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ name: `qa-native-${Date.now()}`, os: 'windows' }),
  }), 'create node');
  nodeID = created.data.node.id;
  const commands = created.data.install_commands;
  if (!commands?.linux?.includes('install-agent.sh') || !commands?.windows?.includes('Install-HoneynetAgent')) {
    throw new Error('platform install commands are missing');
  }
  if (!created.data.mtls || !commands.linux.includes('--ca-sha256') || !commands.windows.includes('-CASHA256')) {
    throw new Error('mTLS bootstrap metadata is missing');
  }

  const expected = (await (await fetch(`${base}/download/agent/linux/amd64/sha256`)).text()).trim();
  const binary = Buffer.from(await (await fetch(`${base}/download/agent/linux/amd64`)).arrayBuffer());
  const actual = createHash('sha256').update(binary).digest('hex');
  if (expected !== actual) throw new Error('Agent checksum mismatch');

  await json(await fetch(`${base}/api/v1/nodes/${nodeID}/install`, { method: 'POST', headers }), 'issue installer');
  const nodes = await json(await fetch(`${base}/api/v1/nodes?page_size=100`, { headers }), 'list nodes');
  const builtin = nodes.data.items.find((item) => item.id === '00000000-0000-4000-8000-000000000001');
  if (builtin?.status !== 'online') throw new Error('built-in node is not online');

  console.log(JSON.stringify({ login: 'ok', builtin_node: 'online', platform_installers: 'ok', agent_sha256: 'ok', mtls_bootstrap: 'ok' }));
} finally {
  if (nodeID && headers) {
    await fetch(`${base}/api/v1/nodes/${nodeID}`, { method: 'DELETE', headers });
  }
}
