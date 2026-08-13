import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ConfigProvider } from '@arco-design/web-react';
import zhCN from '@arco-design/web-react/es/locale/zh-CN';
import '@arco-design/web-react/dist/css/arco.css';
import './styles.css';
import App from './App';
import { applyTheme, useTheme } from './store';
import AppErrorBoundary from './components/AppErrorBoundary';
import PlatformIdentity from './components/PlatformIdentity';

applyTheme(useTheme.getState().mode);
const queryClient = new QueryClient({ defaultOptions: { queries: { retry: 1, staleTime: 10_000 } } });

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <ConfigProvider
        locale={zhCN}
        componentConfig={{
          Table: { border: false },
          Modal: { maskClosable: false },
          Drawer: { maskClosable: false },
          Select: { allowClear: true },
        }}
      >
        <AppErrorBoundary><PlatformIdentity><BrowserRouter><App /></BrowserRouter></PlatformIdentity></AppErrorBoundary>
      </ConfigProvider>
    </QueryClientProvider>
  </React.StrictMode>,
);
