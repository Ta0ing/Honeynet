import { useMemo } from 'react';
import { Button, Table, Typography } from '@arco-design/web-react';
import { IconCheckCircle } from '@arco-design/web-react/icon';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, formatTime } from '../api';
import PageHeader from '../components/PageHeader';
import RiskTag from '../components/RiskTag';
import { buildServiceNameMap, serviceName } from '../presentation';
import { useAuth } from '../store';

export default function Alerts() {
  const queryClient = useQueryClient();
  const role = useAuth((state) => state.user?.role);
  const query = useQuery<any>({ queryKey: ['alerts'], queryFn: () => api.get('/alerts?page_size=100') });
  const services = useQuery<any>({ queryKey: ['services', 'display-catalog'], queryFn: () => api.get('/pot-services?page_size=200&include_retired=1') });
  const serviceNames = useMemo(() => buildServiceNameMap(services.data?.items), [services.data]);
  const ack = useMutation({
    mutationFn: (id: string) => api.put(`/alerts/${id}/ack`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['alerts'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard'] });
    },
  });
  const columns: any[] = [
    { title: '触发时间', dataIndex: 'created_at', render: formatTime, width: 190 },
    { title: '级别', dataIndex: 'level', width: 90, render: (value: string) => <RiskTag value={value} /> },
    { title: '告警内容', render: (_: any, row: any) => <div><Typography.Text bold>{row.title}</Typography.Text><br /><Typography.Text type="secondary">{row.description}</Typography.Text></div> },
    { title: '来源 IP', dataIndex: 'source_ip', render: (value: string) => <Typography.Text code copyable>{value}</Typography.Text> },
    { title: '服务', dataIndex: 'service', render: (value: string) => <Typography.Text title={value}>{serviceName(value, serviceNames)}</Typography.Text> },
    { title: '状态', dataIndex: 'status', render: (value: string) => <span className={`alert-status alert-status-${value === 'new' ? 'new' : 'acked'}`}>{value === 'new' ? '待确认' : '已确认'}</span> },
    { title: '操作', width: 90, render: (_: any, row: any) => row.status === 'new' && <Button type="text" icon={<IconCheckCircle />} disabled={role === 'viewer'} onClick={() => ack.mutate(row.id)}>确认</Button> },
  ];
  return <>
    <PageHeader title="告警中心" description="高价值威胁告警汇聚、研判与处置确认" />
    <div className="table-panel"><Table rowKey="id" loading={query.isLoading} data={query.data?.items || []} columns={columns} pagination={{ pageSize: 20 }} scroll={{ x: 1080 }} /></div>
  </>;
}
