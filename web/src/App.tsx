import type { ReactNode } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { Button, Result } from '@arco-design/web-react';
import { useAuth } from './store';
import ConsoleLayout from './components/ConsoleLayout';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Nodes from './pages/Nodes';
import Pots from './pages/Pots';
import Templates from './pages/Templates';
import Decoys from './pages/Decoys';
import Events from './pages/Events';
import Alerts from './pages/Alerts';
import Assets from './pages/Assets';
import Intel from './pages/Intel';
import System from './pages/System';
import Releases from './pages/Releases';
import Scanners from './pages/Scanners';
import Credentials from './pages/Credentials';
import Services from './pages/Services';
import SecurityScreen from './pages/SecurityScreen';
import DetectionRules from './pages/DetectionRules';
import AIAnalysis from './pages/AIAnalysis';
import PlatformConfig from './pages/PlatformConfig';

function AdminOnly({ children }: { children: ReactNode }) {
  const role = useAuth((state) => state.user?.role);
  return role === 'admin' ? children : <Navigate to="/forbidden" replace />;
}

export default function App() {
  const token = useAuth((s) => s.token);
  if (!token) return <Routes><Route path="/login" element={<Login />} /><Route path="*" element={<Navigate to="/login" replace />} /></Routes>;
  return (
    <Routes>
      <Route path="/screen" element={<SecurityScreen />} />
      <Route element={<ConsoleLayout />}>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/nodes" element={<Nodes />} />
        <Route path="/releases" element={<Releases />} />
        <Route path="/pots" element={<Pots />} />
        <Route path="/templates" element={<Templates />} />
        <Route path="/decoys" element={<Decoys />} />
        <Route path="/events" element={<Events />} />
        <Route path="/scanners" element={<Scanners />} />
        <Route path="/credentials" element={<Credentials />} />
        <Route path="/services" element={<Services />} />
        <Route path="/alerts" element={<Alerts />} />
        <Route path="/detection-rules" element={<DetectionRules />} />
        <Route path="/ai" element={<AIAnalysis />} />
        <Route path="/intel" element={<Intel />} />
        <Route path="/assets" element={<Assets />} />
        <Route path="/system" element={<System />} />
        <Route path="/platform-config" element={<AdminOnly><PlatformConfig /></AdminOnly>} />
      </Route>
      <Route path="/forbidden" element={<Result status="403" title="没有访问权限" subTitle="当前账号没有执行此操作的权限。"><Button type="primary" onClick={() => { location.href = '/dashboard'; }}>返回首页</Button></Result>} />
      <Route path="*" element={<Result status="404" title="页面不存在" subTitle="地址可能已变更或页面已经移除。"><Button type="primary" onClick={() => { location.href = '/dashboard'; }}>返回首页</Button></Result>} />
    </Routes>
  );
}
