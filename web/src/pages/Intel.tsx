import { useState } from 'react';
import { Alert, Button, Card, Descriptions, Grid, Input, Message, Space, Table, Tag, Typography } from '@arco-design/web-react';
import { IconRefresh, IconSearch } from '@arco-design/web-react/icon';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, errorMessage, formatTime } from '../api';
import PageHeader from '../components/PageHeader';
import { useAuth } from '../store';

export default function Intel() {
  const [address, setAddress] = useState('');
  const [lookup, setLookup] = useState<any>();
  const role = useAuth((state) => state.user?.role);
  const client = useQueryClient();
  const iocs = useQuery<any>({ queryKey: ['iocs'], queryFn: () => api.get('/intel/iocs?page_size=100') });
  const database = useQuery<any>({ queryKey: ['intel', 'database'], queryFn: () => api.get('/intel/database/status') });
  const search = useMutation({
    mutationFn: () => api.get(`/intel/database/query?ip=${encodeURIComponent(address.trim())}`),
    onSuccess: setLookup,
    onError: (error) => Message.error(errorMessage(error)),
  });
  const update = useMutation({
    mutationFn: () => api.post('/intel/database/update'),
    onSuccess: (status) => { Message.success('威胁情报数据库已更新'); client.setQueryData(['intel', 'database'], status); },
    onError: (error) => Message.error(errorMessage(error)),
  });
  const columns: any[] = [
    { title: '类型', dataIndex: 'type', render: (value: string) => <Tag color="purple">{value.toUpperCase()}</Tag> },
    { title: '指标值', dataIndex: 'value', render: (value: string) => <Typography.Text code copyable>{value}</Typography.Text> },
    { title: '来源', dataIndex: 'source' },
    { title: '置信度', dataIndex: 'confidence', render: (value: number) => <Tag color={value >= 80 ? 'red' : value >= 60 ? 'orange' : 'blue'}>{value}%</Tag> },
    { title: '首次发现', dataIndex: 'first_seen', render: formatTime },
    { title: '最近发现', dataIndex: 'last_seen', render: formatTime },
  ];
  const status = database.data || {};
  const statusColor = status.loaded ? 'green' : status.enabled ? 'orange' : 'gray';
  const statusText = status.loaded ? '数据库已就绪' : status.enabled ? '等待首次更新' : '未启用';
  return <>
    <PageHeader
      title="威胁情报"
      description="使用离线 IPv4 / IPv6 情报为攻击来源添加风险标签，同时沉淀企业私有 IOC"
      extra={<Button icon={<IconRefresh />} loading={database.isFetching || iocs.isFetching} onClick={() => { database.refetch(); iocs.refetch(); }}>刷新</Button>}
    />
    <Grid.Row gutter={16} style={{ marginBottom: 16 }}>
      <Grid.Col xs={24} xl={12}>
        <Card title="离线情报数据库" className="panel-card" extra={<Space><Tag color={statusColor}>{statusText}</Tag>{role === 'admin' && status.enabled && <Button size="small" disabled={!status.download_ready} loading={update.isPending || status.updating} onClick={() => update.mutate()}>立即更新</Button>}</Space>}>
          {status.last_error && <Alert type="warning" showIcon content={status.last_error} style={{ marginBottom: 12 }} />}
          <Descriptions column={{ xs: 1, sm: 2 }} layout="vertical" data={[
            { label: '数据来源', value: status.source || '免费社区威胁情报库' },
            { label: '地址支持', value: status.ipv4_ipv6 ? 'IPv4 + IPv6' : 'IPv4' },
            { label: '情报记录', value: Number(status.record_count || 0).toLocaleString('zh-CN') },
            { label: '查询模式', value: status.lookup_mode || '—' },
            { label: '数据导出时间', value: formatTime(status.database_updated_at) },
            { label: '最近更新成功', value: formatTime(status.last_successful_at || status.installed_at) },
            { label: '下次自动更新', value: formatTime(status.next_update_at) },
            { label: '更新状态', value: status.updating ? '正在下载并验证' : '空闲' },
          ]} />
        </Card>
      </Grid.Col>
      <Grid.Col xs={24} xl={12}>
        <Card title="IP 情报查询" className="panel-card">
          <Typography.Paragraph type="secondary">查询只访问 Server 的离线数据库，不会把攻击 IP 发送到外部服务。</Typography.Paragraph>
          <Space style={{ width: '100%' }}>
            <Input value={address} onChange={setAddress} allowClear placeholder="输入 IPv4 或 IPv6 地址" onPressEnter={() => address.trim() && search.mutate()} />
            <Button type="primary" icon={<IconSearch />} disabled={!status.loaded || !address.trim()} loading={search.isPending} onClick={() => search.mutate()}>查询</Button>
          </Space>
          {lookup && <div style={{ marginTop: 16 }}>
            <Space wrap><Typography.Text code copyable>{lookup.ip}</Typography.Text><Tag color={lookup.matched ? 'red' : 'green'}>{lookup.matched ? '命中情报' : '未命中'}</Tag></Space>
            {lookup.matched && <Space wrap style={{ marginTop: 10 }}>{(lookup.tags || []).map((tag: string) => <Tag key={tag} color={tag.includes('高危') ? 'red' : tag.includes('中危') ? 'orange' : 'arcoblue'}>{tag}</Tag>)}</Space>}
          </div>}
        </Card>
      </Grid.Col>
    </Grid.Row>
    <div className="table-panel"><Table rowKey="id" loading={iocs.isLoading} data={iocs.data?.items || []} columns={columns} pagination={{ pageSize: 20 }} scroll={{ x: 920 }} /></div>
  </>;
}
