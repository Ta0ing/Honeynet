import { Button, Card, Grid, Progress, Skeleton, Space, Table, Tag, Typography } from '@arco-design/web-react';
import { IconBug, IconCloud, IconDesktop, IconExclamationCircle, IconExperiment, IconRefresh } from '@arco-design/web-react/icon';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { api, formatTime } from '../api';
import PageHeader from '../components/PageHeader';
import { usePlatformBranding } from '../branding';

const colors = ['#165dff', '#00b42a', '#722ed1', '#ff7d00', '#f53f3f', '#14c9c9', '#3491fa'];

export default function Dashboard() {
  const navigate = useNavigate();
  const branding = usePlatformBranding().data!;
  const summary = useQuery<any>({ queryKey: ['dashboard', 'summary'], queryFn: () => api.get('/dashboard/summary'), refetchInterval: 30_000 });
  const trends = useQuery<any[]>({ queryKey: ['dashboard', 'trends'], queryFn: () => api.get('/dashboard/trends?days=7') });
  const top = useQuery<any[]>({ queryKey: ['dashboard', 'top'], queryFn: () => api.get('/dashboard/top-attackers') });
  const data = summary.data || {};
  const max = Math.max(...(trends.data || []).map((item) => item.count), 1);
  const nodeHealth = data.nodes ? Math.round((data.online_nodes || 0) / data.nodes * 100) : 100;
  const potHealth = data.pots ? Math.round((data.running_pots || 0) / data.pots * 100) : 100;
  const cards = [
    ['在线节点', data.online_nodes ?? 0, `/ ${data.nodes ?? 0}`, <IconCloud />, 'blue'],
    ['在线蜜罐', data.running_pots ?? 0, `/ ${data.pots ?? 0}`, <IconExperiment />, 'green'],
    ['24h 攻击', data.events_24h ?? 0, '次', <IconBug />, 'purple'],
    ['未确认告警', data.new_alerts ?? 0, '条', <IconExclamationCircle />, 'orange'],
  ];
  return <>
    <PageHeader title="首页" description="节点、蜜罐、攻击链与威胁实体的统一运行视图" extra={<Space><Button icon={<IconRefresh />} loading={summary.isFetching} onClick={() => summary.refetch()}>刷新</Button><Button type="primary" icon={<IconDesktop />} onClick={() => navigate('/screen')}>进入大屏</Button></Space>} />
    <Card className="runtime-strip" bordered={false}><Space split={<span className="runtime-divider" />} size={24}><Tag color="arcoblue">{branding.system_name} {(branding.system_version || data.version) ? `${/^v/i.test(branding.system_version || data.version) ? '' : 'v'}${branding.system_version || data.version}` : '版本读取中'}</Tag><span><i className={`status-dot ${summary.isError ? 'status-error' : 'status-ok'}`} />Server {summary.isError ? '状态异常' : '运行正常'}</span><span>双引擎 <b>业务数据 + 安全分析</b></span><span>网络栈 <b>{data.ipv6_capable ? 'IPv4 + IPv6' : 'IPv4'}</b></span><span>IP 定位 <b>{data.geoip_enabled ? 'IPIP 城市库' : '未启用'}</b></span><span>实时通道 <b>WebSocket</b></span><span>Agent 通信 <b>mTLS 1.3</b></span></Space></Card>
    <Grid.Row gutter={16} className="dashboard-row">
      <Grid.Col xs={24} xl={9}><Card title="节点与蜜罐状态" className="status-overview panel-card"><div className="status-gauges"><div className="status-gauge"><div className="gauge-ring"><Progress type="circle" showText={false} percent={nodeHealth} color="#165dff" /><b>{data.online_nodes || 0}/{data.nodes || 0}</b></div><span>在线节点</span></div><div className="status-gauge"><div className="gauge-ring"><Progress type="circle" showText={false} percent={potHealth} color="#00b42a" /><b>{data.running_pots || 0}/{data.pots || 0}</b></div><span>在线蜜罐</span></div></div></Card></Grid.Col>
      <Grid.Col xs={24} xl={15}><Card title="攻击链" className="attack-chain-card panel-card"><div className="attack-chain"><button onClick={() => navigate('/intel')}><span>攻击 IP</span><b>{top.data?.length || 0}</b><small>威胁实体</small></button><i /><button onClick={() => navigate('/scanners')}><span>扫描</span><b>{data.events_24h || 0}</b><small>探测行为</small></button><i /><button onClick={() => navigate('/events')}><span>攻击</span><b>{data.events_24h || 0}</b><small>交互事件</small></button><i /><button onClick={() => navigate('/alerts')}><span>告警</span><b>{data.new_alerts || 0}</b><small>待处置</small></button></div></Card></Grid.Col>
    </Grid.Row>
    <Grid.Row gutter={16}>{cards.map(([label, value, suffix, icon, color]) => <Grid.Col xs={24} sm={12} xl={6} key={String(label)}><Card className={`metric-card metric-${color}`}><div className="metric-icon">{icon}</div><Typography.Text type="secondary">{label}</Typography.Text><div className="metric-value">{summary.isLoading ? <Skeleton text={{ rows: 1 }} /> : <>{value}<small>{suffix}</small></>}</div></Card></Grid.Col>)}</Grid.Row>
    <Grid.Row gutter={16} className="dashboard-row"><Grid.Col xs={24} xl={16}><Card title="攻击趋势" extra={<Tag color="green">最近 7 天</Tag>} className="panel-card"><div className="trend-chart">{(trends.data || []).length === 0 ? <div className="empty-chart">暂未捕获攻击事件</div> : trends.data!.map((item, index) => <div className="trend-column" key={item.day}><span className="trend-count">{item.count}</span><div className="trend-bar" style={{ height: `${Math.max(item.count / max * 180, 4)}px`, background: colors[index] }} /><label>{item.day.slice(5, 10)}</label></div>)}</div></Card></Grid.Col>
      <Grid.Col xs={24} xl={8}><Card title="黑客画像" className="panel-card"><div className="attacker-summary"><IconBug /><b>{top.data?.length || 0}</b><span>近七日活跃攻击来源</span></div><div className="mini-attacker-list">{(top.data || []).slice(0, 4).map((item, index) => <div key={item.src_ip}><span className="rank">{index + 1}</span><Typography.Text code>{item.src_ip}</Typography.Text><Tag color="red">{item.count}</Tag></div>)}</div></Card></Grid.Col></Grid.Row>
    <Card title="TOP 攻击源（近 7 天）" className="panel-card"><Table rowKey="src_ip" pagination={false} data={top.data || []} columns={[{ title: '排名', render: (_, __, index) => <span className="rank">{index + 1}</span>, width: 80 }, { title: '来源 IP', dataIndex: 'src_ip', render: (value) => <Typography.Text code>{value}</Typography.Text> }, { title: '归属地', dataIndex: 'geo', render: (value) => value || '未知位置' }, { title: '事件次数', dataIndex: 'count', render: (value) => <Tag color="red">{value}</Tag> }, { title: '威胁判定', render: () => <Tag color="orange">待研判</Tag> }, { title: '最近活动', dataIndex: 'last_seen', render: formatTime }]} /></Card>
  </>;
}
