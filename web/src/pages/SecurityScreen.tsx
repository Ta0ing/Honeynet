import { useEffect, useMemo, useState } from 'react';
import { Button, Space, Tag } from '@arco-design/web-react';
import { IconClose, IconCloud, IconExperiment, IconSafe } from '@arco-design/web-react/icon';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { api, formatTime } from '../api';
import { usePlatformBranding } from '../branding';
import BrandLogo from '../components/BrandLogo';

export default function SecurityScreen() {
  const navigate = useNavigate();
  const branding = usePlatformBranding().data!;
  const [now, setNow] = useState(new Date());
  const summary = useQuery<any>({ queryKey: ['dashboard', 'summary'], queryFn: () => api.get('/dashboard/summary'), refetchInterval: 15_000 });
  const trends = useQuery<any[]>({ queryKey: ['dashboard', 'trends'], queryFn: () => api.get('/dashboard/trends?days=7'), refetchInterval: 30_000 });
  const attackers = useQuery<any[]>({ queryKey: ['dashboard', 'top'], queryFn: () => api.get('/dashboard/top-attackers'), refetchInterval: 30_000 });
  const events = useQuery<any>({ queryKey: ['events', 'screen'], queryFn: () => api.get('/events?page_size=12'), refetchInterval: 5000 });
  useEffect(() => { const timer = window.setInterval(() => setNow(new Date()), 1000); return () => window.clearInterval(timer); }, []);
  const maxTrend = Math.max(...(trends.data || []).map((item) => item.count), 1);
  const data = summary.data || {};
  const metrics = useMemo(() => [
    ['在线节点', data.online_nodes ?? 0, data.nodes ?? 0, <IconCloud />],
    ['在线蜜罐', data.running_pots ?? 0, data.pots ?? 0, <IconExperiment />],
    ['日攻击感知', data.events_24h ?? 0, undefined, <IconSafe />],
    ['未确认告警', data.new_alerts ?? 0, undefined, <IconSafe />],
  ], [data]);
  return <div className="security-screen">
    <header className="screen-header"><div className="screen-brand"><BrandLogo src={branding.system_logo_url} alt={`${branding.system_name} Logo`} className="screen-brand-logo" /><b>{branding.system_name}</b></div><h1>企业蜜网攻击态势</h1><div className="screen-clock"><span>{now.toLocaleDateString('zh-CN')}</span><b>{now.toLocaleTimeString('zh-CN', { hour12: false })}</b><Button aria-label="关闭大屏" type="text" shape="circle" icon={<IconClose />} onClick={() => navigate('/dashboard')} /></div></header>
    <div className="screen-grid">
      <section className="screen-panel screen-metrics"><div className="screen-panel-title">本地感知数据</div><div className="screen-metric-grid">{metrics.map(([label, value, total, icon]) => <div className="screen-metric" key={String(label)}><i>{icon}</i><span>{label}</span><b>{String(value)}</b>{total !== undefined && <small>/ {String(total)}</small>}</div>)}</div></section>
      <section className="screen-panel screen-map"><div className="screen-panel-title">威胁攻击链</div><div className="screen-chain"><div className="chain-origin"><span>攻击来源</span><b>{attackers.data?.length || 0}</b><small>活跃 IP</small></div><i /><div className="chain-stage scan"><span>扫描</span><b>{data.events_24h || 0}</b></div><i /><div className="chain-stage attack"><span>攻击</span><b>{data.events_24h || 0}</b></div><i /><div className="chain-stage alarm"><span>告警</span><b>{data.new_alerts || 0}</b></div></div></section>
      <section className="screen-panel screen-trend"><div className="screen-panel-title">近七日攻击趋势</div><div className="screen-chart">{(trends.data || []).map((item) => <div key={item.day}><span>{item.count}</span><i style={{ height: `${Math.max(8, item.count / maxTrend * 150)}px` }} /><small>{item.day.slice(5)}</small></div>)}</div></section>
      <section className="screen-panel screen-attackers"><div className="screen-panel-title">攻击来源排行</div><div className="screen-list">{(attackers.data || []).slice(0, 7).map((item, index) => <div key={item.src_ip}><span className="screen-rank">{index + 1}</span><code>{item.src_ip}</code><b>{item.count}</b></div>)}</div></section>
      <section className="screen-panel screen-events"><div className="screen-panel-title">近期攻击事件 <Tag color="green">实时</Tag></div><div className="screen-event-list">{(events.data?.items || []).map((event: any) => <div key={event.event_id}><time>{formatTime(event.ts)}</time><code>{event.src_ip || 'unknown'}</code><span>→</span><b>{event.service || event.event_type}</b><small>{event.dst_port ? `TCP/${event.dst_port}` : event.event_type}</small></div>)}</div></section>
    </div>
    <footer><Space><span className="screen-live-dot" />WebSocket 实时数据</Space><span>{branding.copyright} · {(branding.system_version || data.version) ? `${/^v/i.test(branding.system_version || data.version) ? '' : 'v'}${branding.system_version || data.version}` : '运行中'}</span></footer>
  </div>;
}
