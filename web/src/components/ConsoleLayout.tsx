import { useEffect, useMemo, useState } from 'react';
import { Avatar, Badge, Breadcrumb, Button, Drawer, Dropdown, Layout, Menu, Message, Space, Tag, Tooltip, Typography } from '@arco-design/web-react';
import {
  IconBug, IconCloud, IconCode, IconDashboard, IconDesktop, IconExclamationCircle,
  IconExperiment, IconFile, IconList, IconMoonFill, IconNotification, IconPoweroff, IconSafe,
  IconMenu, IconRobot, IconSettings, IconStorage, IconSunFill, IconUserGroup,
} from '@arco-design/web-react/icon';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '../api';
import { useAuth, useTheme } from '../store';
import { publicExternalURL, usePlatformBranding } from '../branding';
import BrandLogo from './BrandLogo';

const routeMeta: Record<string, [string, string]> = {
  '/dashboard': ['首页', '运行总览'],
  '/events': ['威胁感知', '攻击列表'],
  '/scanners': ['威胁感知', '扫描感知'],
  '/decoys': ['威胁感知', '失陷感知'],
  '/intel': ['威胁实体', '攻击来源'],
  '/credentials': ['威胁实体', '账号资源'],
  '/assets': ['威胁实体', '资产画像'],
  '/alerts': ['威胁实体', '告警中心'],
  '/detection-rules': ['威胁感知', '检测规则'],
  '/ai': ['威胁感知', 'AI Agent'],
  '/nodes': ['环境管理', '节点管理'],
  '/templates': ['环境管理', '模板管理'],
  '/services': ['环境管理', '服务管理'],
  '/pots': ['环境管理', '蜜罐实例'],
  '/releases': ['平台管理', 'Agent 发布'],
  '/system': ['平台管理', '系统配置'],
  '/platform-config': ['平台管理', '配置管理'],
};

function NavigationMenu({ path, navigate, isAdmin }: { path: string; navigate: (path: string) => void; isAdmin: boolean }) {
  return <Menu
    selectedKeys={[path]}
    defaultOpenKeys={['threat', 'entity', 'environment', 'platform']}
    onClickMenuItem={(key) => navigate(key)}
  >
    <Menu.Item key="/dashboard"><IconDashboard />首页</Menu.Item>
    <Menu.Item key="/screen"><IconDesktop />大屏</Menu.Item>
    <Menu.SubMenu key="threat" title={<><IconBug />威胁感知</>}>
      <Menu.Item key="/events">攻击列表</Menu.Item>
      <Menu.Item key="/scanners">扫描感知</Menu.Item>
      <Menu.Item key="/decoys">失陷感知</Menu.Item>
      <Menu.Item key="/detection-rules">检测规则</Menu.Item>
      <Menu.Item key="/ai"><IconRobot />AI Agent</Menu.Item>
    </Menu.SubMenu>
    <Menu.SubMenu key="entity" title={<><IconSafe />威胁实体</>}>
      <Menu.Item key="/intel">攻击来源</Menu.Item>
      <Menu.Item key="/credentials">账号资源</Menu.Item>
      <Menu.Item key="/assets">资产画像</Menu.Item>
      <Menu.Item key="/alerts">告警中心</Menu.Item>
    </Menu.SubMenu>
    <Menu.SubMenu key="environment" title={<><IconCloud />环境管理</>}>
      <Menu.Item key="/nodes">节点管理</Menu.Item>
      <Menu.Item key="/templates">模板管理</Menu.Item>
      <Menu.Item key="/services">服务管理</Menu.Item>
      <Menu.Item key="/pots">蜜罐实例</Menu.Item>
    </Menu.SubMenu>
    <Menu.SubMenu key="platform" title={<><IconSettings />平台管理</>}>
      <Menu.Item key="/releases">Agent 发布</Menu.Item>
      <Menu.Item key="/system">系统配置</Menu.Item>
      {isAdmin && <Menu.Item key="/platform-config">配置管理</Menu.Item>}
    </Menu.SubMenu>
  </Menu>;
}

export default function ConsoleLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { user, logout } = useAuth();
  const { mode, toggle } = useTheme();
  const branding = usePlatformBranding().data!;
  const [siderCollapsed, setSiderCollapsed] = useState(() => window.matchMedia('(max-width: 1199px)').matches);
  const [mobileNav, setMobileNav] = useState(false);
  const [realtime, setRealtime] = useState<'connecting' | 'online' | 'offline'>('connecting');
  const summary = useQuery<any>({ queryKey: ['dashboard', 'summary'], queryFn: () => api.get('/dashboard/summary'), refetchInterval: 30_000 });
  const role = useMemo(() => ({ admin: '管理员', operator: '运营', viewer: '只读' }[user?.role || 'viewer']), [user]);
  const page = routeMeta[location.pathname] || [branding.system_name, '管理中心'];
  const isAdmin = user?.role === 'admin';
  const website = publicExternalURL(branding.official_website_url);
  const documentation = publicExternalURL(branding.product_documentation_url);
  const displayedVersion = branding.system_version || summary.data?.version || '';

  useEffect(() => {
    const token = localStorage.getItem('honeynet_token');
    if (!token) return;
    const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    let ws: WebSocket | undefined;
    let retry: number | undefined;
    let stopped = false;
    const connect = () => {
      setRealtime('connecting');
      ws = new WebSocket(`${proto}//${window.location.host}/api/v1/ws`, token);
      ws.onopen = () => setRealtime('online');
      ws.onmessage = (event) => {
        let message: any;
        try { message = JSON.parse(event.data); } catch { return; }
        if (message.type === 'alert.new') Message.warning(`新告警：${message.payload.title}`);
        if (message.type === 'event.new') queryClient.invalidateQueries({ queryKey: ['events'] });
        if (message.type === 'node.status' || message.type === 'node.capabilities') queryClient.invalidateQueries({ queryKey: ['nodes'] });
        if (message.type === 'sense.status') queryClient.invalidateQueries({ queryKey: ['nodes'] });
        if (message.type === 'pot.status') queryClient.invalidateQueries({ queryKey: ['pots'] });
        if (message.type === 'decoy.status' || message.type === 'decoy.hit') queryClient.invalidateQueries({ queryKey: ['decoys'] });
        if (message.type === 'upgrade.status') queryClient.invalidateQueries({ queryKey: ['upgrade-rollouts'] });
        queryClient.invalidateQueries({ queryKey: ['dashboard'] });
      };
      ws.onclose = () => { setRealtime('offline'); if (!stopped) retry = window.setTimeout(connect, 3000); };
    };
    connect();
    return () => { stopped = true; if (retry) clearTimeout(retry); ws?.close(); };
  }, [queryClient]);

  const accountMenu = <Menu onClickMenuItem={(key) => {
    if (key === 'logout') { void api.post('/auth/logout').catch(() => undefined).finally(() => { logout(); navigate('/login'); }); }
    if (key === 'system') navigate('/system');
    if (key === 'platform-config') navigate('/platform-config');
  }}>
    <Menu.Item key="system"><IconSettings />账号与系统</Menu.Item>
    {isAdmin && <Menu.Item key="platform-config"><IconSettings />配置管理</Menu.Item>}
    <Menu.Item key="logout"><IconPoweroff />退出登录</Menu.Item>
  </Menu>;

  return <Layout className="app-layout">
    <Layout.Sider
      className="app-sider"
      width={210}
      collapsedWidth={64}
      collapsible
      collapsed={siderCollapsed}
      breakpoint="lg"
      trigger={null}
      onCollapse={(next) => setSiderCollapsed(next)}
    >
      <div className="brand">
        <BrandLogo src={branding.system_logo_url} alt={`${branding.system_name} Logo`} className="brand-mark" />
        <div className="brand-copy"><strong>{branding.system_name}</strong><span>威胁捕捉与诱骗系统</span></div>
      </div>
      <div className="menu-scroll"><NavigationMenu path={location.pathname} navigate={navigate} isAdmin={isAdmin} /></div>
      <div className="sider-foot"><Badge status={summary.isError ? 'warning' : 'success'} text={summary.isError ? 'Server 状态异常' : 'Server 运行中'} /><Typography.Text type="secondary">{displayedVersion ? `${/^v/i.test(displayedVersion) ? '' : 'v'}${displayedVersion}` : '版本读取中'}</Typography.Text></div>
    </Layout.Sider>
    <Layout className="app-main">
      <Layout.Header className="app-header">
        <Button className="mobile-nav-trigger" aria-label="打开导航" type="text" shape="circle" icon={<IconMenu />} onClick={() => setMobileNav(true)} />
        <div className="header-location">
          <Breadcrumb><Breadcrumb.Item>{page[0]}</Breadcrumb.Item><Breadcrumb.Item>{page[1]}</Breadcrumb.Item></Breadcrumb>
        </div>
        <Space size={10} className="header-actions">
          <div className="runtime-indicator"><IconCloud /><span>节点</span><b>{summary.data?.online_nodes ?? 0}/{summary.data?.nodes ?? 0}</b></div>
          <div className="runtime-indicator"><IconExperiment /><span>蜜罐</span><b>{summary.data?.running_pots ?? 0}/{summary.data?.pots ?? 0}</b></div>
          <Tooltip content={mode === 'light' ? '切换暗黑主题' : '切换亮色主题'}><Button aria-label="切换主题" type="text" shape="circle" icon={mode === 'light' ? <IconMoonFill /> : <IconSunFill />} onClick={toggle} /></Tooltip>
          <Tooltip content={`实时通道：${realtime === 'online' ? '已连接' : realtime === 'connecting' ? '连接中' : '正在重连'}`}><Badge status={realtime === 'online' ? 'success' : realtime === 'connecting' ? 'processing' : 'warning'} /></Tooltip>
          <Tooltip content="告警中心"><Badge count={summary.data?.new_alerts || 0} maxCount={99}><Button aria-label="告警中心" type="text" shape="circle" icon={<IconNotification />} onClick={() => navigate('/alerts')} /></Badge></Tooltip>
          <Tag color="arcoblue">{role}</Tag>
          <Dropdown droplist={accountMenu} position="br" trigger="click">
            <button className="account-trigger"><Avatar size={30}>{user?.username.slice(0, 1).toUpperCase()}</Avatar><span>{user?.display_name || user?.username}</span></button>
          </Dropdown>
        </Space>
      </Layout.Header>
      <Layout.Content className="app-content"><Outlet /></Layout.Content>
      <Layout.Footer className="app-footer">
        <div className="app-footer-brand"><BrandLogo src={branding.company_logo_url} alt="公司 Logo" kind="company" className="footer-company-logo" /><span>{branding.copyright}</span></div>
        <Space wrap size={14}>
          {branding.customer_service_phone && <a href={`tel:${branding.customer_service_phone}`}>{branding.customer_service_phone}</a>}
          {branding.customer_service_email && <a href={`mailto:${branding.customer_service_email}`}>{branding.customer_service_email}</a>}
          {website && <a href={website} target="_blank" rel="noreferrer">官方网站</a>}
          {documentation && <a href={documentation} target="_blank" rel="noreferrer">产品文档</a>}
        </Space>
      </Layout.Footer>
      <Drawer className="mobile-nav-drawer" width={280} placement="left" title={`${branding.system_name} 导航`} visible={mobileNav} onCancel={() => setMobileNav(false)} footer={null}>
        <NavigationMenu path={location.pathname} navigate={(path) => { setMobileNav(false); navigate(path); }} isAdmin={isAdmin} />
      </Drawer>
    </Layout>
  </Layout>;
}
