import type { ErrorInfo, ReactNode } from 'react';
import { Component } from 'react';
import { Button, Result, Space } from '@arco-design/web-react';

type Props = { children: ReactNode };
type State = { error?: Error };

export default class AppErrorBoundary extends Component<Props, State> {
  state: State = {};

  static getDerivedStateFromError(error: Error): State { return { error }; }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('Console UI error', error, info.componentStack);
  }

  render() {
    if (!this.state.error) return this.props.children;
    return <Result
      status="500"
      title="控制台页面加载失败"
      subTitle="当前操作没有修改服务端数据。可重新加载页面；若问题持续存在，请携带时间和页面地址查看 Server 日志。"
      extra={<Space><Button type="primary" onClick={() => location.reload()}>重新加载</Button><Button onClick={() => { location.href = '/dashboard'; }}>返回首页</Button></Space>}
    />;
  }
}
