import { useState } from 'react';
import { Card, Grid, Input, Space, Table, Tag, Typography } from '@arco-design/web-react';
import { IconLock, IconSearch, IconUser } from '@arco-design/web-react/icon';
import { useQuery } from '@tanstack/react-query';
import { api, formatTime } from '../api';
import PageHeader from '../components/PageHeader';

type TopValue = { value: string; count: number };

export default function Credentials() {
  const [keyword, setKeyword] = useState('');
  const query = useQuery<any>({
    queryKey: ['credential-resources', keyword],
    queryFn: () => api.get('/credential-resources', { params: { q: keyword.trim() || undefined, page_size: 100 } }),
  });
  const items = query.data?.items || [];
  const usernames: TopValue[] = query.data?.top_usernames || [];
  const passwords: TopValue[] = query.data?.top_passwords || [];
  const columns: any[] = [
    { title: '用户名', dataIndex: 'username', render: (value: string) => <Typography.Text copyable={Boolean(value)} code>{value || '—'}</Typography.Text> },
    { title: '密码 / 认证响应', width: 260, render: (_: any, row: any) => <Typography.Text copyable={Boolean(row.password || row.auth_response)} code ellipsis={{ showTooltip: true }}>{row.password || row.auth_response || '—'}</Typography.Text> },
    { title: '被攻击服务', dataIndex: 'service_name', render: (value: string, row: any) => <Tag color="arcoblue" title={row.service_code}>{value || '未知服务'}</Tag> },
    { title: '攻击来源 IP', dataIndex: 'src_ip', render: (value: string, row: any) => <div className="entity-cell"><Typography.Text code copyable>{value}</Typography.Text><Typography.Text type="secondary">{row.geo || '未知位置'}</Typography.Text></div> },
    { title: '认证方式', dataIndex: 'mechanism', render: (value: string) => value || '未知认证方式' },
    { title: '捕获时间', dataIndex: 'ts', render: formatTime, width: 180 },
  ];
  return <>
    <PageHeader title="账号资源" description="由 Server 统一清洗各协议认证数据，汇聚 Redis、MySQL、RDP、SMB 与 Web 等服务的有效账号资源" />
    <Grid.Row gutter={[16, 16]} className="dashboard-row">
      <Grid.Col xs={24} md={12}><Card title={<Space><IconUser />Top 用户名</Space>} className="panel-card"><div className="credential-rank">{usernames.length ? usernames.map((item, index) => <div key={item.value}><span className="rank">{index + 1}</span><Typography.Text code>{item.value}</Typography.Text><Tag>{item.count} 次</Tag></div>) : <Typography.Text type="secondary">暂无有效用户名记录</Typography.Text>}</div></Card></Grid.Col>
      <Grid.Col xs={24} md={12}><Card title={<Space><IconLock />Top 密码</Space>} className="panel-card"><div className="credential-rank">{passwords.length ? passwords.map((item, index) => <div key={item.value}><span className="rank">{index + 1}</span><Typography.Text code>{item.value}</Typography.Text><Tag>{item.count} 次</Tag></div>) : <Typography.Text type="secondary">暂无有效密码记录</Typography.Text>}</div></Card></Grid.Col>
    </Grid.Row>
    <div className="filter-bar"><Input allowClear prefix={<IconSearch />} placeholder="攻击 IP、用户名、密码、认证方式或服务" value={keyword} onChange={setKeyword} style={{ width: 400 }} /><Typography.Text type="secondary">共 {query.data?.total || 0} 条有效资源 · 脏数据已在 Server 端统一过滤</Typography.Text></div>
    <div className="table-panel"><Table rowKey="event_id" loading={query.isLoading} data={items} columns={columns} pagination={{ pageSize: 20 }} scroll={{ x: 1080 }} /></div>
  </>;
}
