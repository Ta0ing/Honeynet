import { useState } from 'react';
import { Button, Form, Input, Message, Space, Tooltip, Typography } from '@arco-design/web-react';
import { IconLock, IconMoonFill, IconSunFill, IconUser } from '@arco-design/web-react/icon';
import { useNavigate } from 'react-router-dom';
import { api, errorMessage } from '../api';
import { useAuth, useTheme } from '../store';
import { publicExternalURL, usePlatformBranding } from '../branding';
import BrandLogo from '../components/BrandLogo';

export default function Login() {
  const [loading, setLoading] = useState(false); const navigate = useNavigate(); const login = useAuth((s) => s.login);
  const { mode, toggle } = useTheme();
  const branding = usePlatformBranding().data!;
  const website = publicExternalURL(branding.official_website_url);
  const documentation = publicExternalURL(branding.product_documentation_url);
  const submit = async (values: { username: string; password: string }) => {
    setLoading(true); try { const result: any = await api.post('/auth/login', values); login(result.token, { ...result.user, display_name: result.display_name }); navigate('/dashboard'); } catch (e) { Message.error(errorMessage(e)); } finally { setLoading(false); }
  };
  return <div className="login-page"><Tooltip content={mode === 'light' ? '切换暗黑主题' : '切换亮色主题'}><Button className="login-theme" aria-label="切换主题" shape="circle" icon={mode === 'light' ? <IconMoonFill /> : <IconSunFill />} onClick={toggle} /></Tooltip><div className="login-aura aura-one" /><div className="login-aura aura-two" />
    <div className="login-intro"><BrandLogo src={branding.system_logo_url} alt={`${branding.system_name} Logo`} className="login-logo" /><Typography.Title>{branding.system_name}</Typography.Title><div className="login-rule" /><Typography.Title heading={2}>让攻击者走进<br /><em>你设计的世界</em></Typography.Title><Typography.Paragraph>面向企业安全运营的蜜网与欺骗防御平台。低误报、深度检测、持续生产私有威胁情报。</Typography.Paragraph><div className="login-metrics"><span><b>90+</b> 蜜罐服务</span><span><b>实时</b> 攻击感知</span><span><b>全域</b> 节点覆盖</span></div></div>
    <div className="login-card"><div className="login-card-brand"><BrandLogo src={branding.company_logo_url} alt="公司 Logo" kind="company" className="login-company-logo" /><div><Typography.Title heading={3}>登录{branding.system_name}</Typography.Title><Typography.Text type="secondary">使用管理员分配的账号继续</Typography.Text></div></div><Form layout="vertical" onSubmit={submit} initialValues={{ username: 'admin' }}><Form.Item label="用户名" field="username" rules={[{ required: true }]}><Input size="large" prefix={<IconUser />} placeholder="请输入用户名" /></Form.Item><Form.Item label="密码" field="password" rules={[{ required: true }]}><Input.Password size="large" prefix={<IconLock />} placeholder="请输入密码" /></Form.Item><Button htmlType="submit" type="primary" size="large" long loading={loading}>进入控制台</Button></Form><div className="login-support"><Typography.Text type="secondary">{branding.copyright}</Typography.Text><Space wrap split={<span>·</span>}>{branding.customer_service_phone && <a href={`tel:${branding.customer_service_phone}`}>客服热线</a>}{branding.customer_service_email && <a href={`mailto:${branding.customer_service_email}`}>客服邮箱</a>}{website && <a href={website} target="_blank" rel="noreferrer">官方网站</a>}{documentation && <a href={documentation} target="_blank" rel="noreferrer">产品文档</a>}</Space></div></div>
  </div>;
}
