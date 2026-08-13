import { useState } from 'react';
import { Button, Form, Input, Message, Modal, Popconfirm, Space, Table, Tag, Typography } from '@arco-design/web-react';
import { IconDelete, IconEdit, IconPlus } from '@arco-design/web-react/icon';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api, errorMessage, formatTime } from '../api';
import PageHeader from '../components/PageHeader';
import { useAuth } from '../store';

const example = `name: fake-oa-portal
listen:
  port: 8080
pages:
  - path: /login
    method: GET
    response:
      status: 200
      body: "Honeynet login portal"
  - path: /login
    method: POST
    capture:
      fields: [username, password]
      event_type: web.credential
    response:
      status: 302
      headers:
        Location: /index
`;

export default function Templates() {
  const [visible,setVisible]=useState(false);const [editing,setEditing]=useState<any>();const [form]=Form.useForm();const qc=useQueryClient();const role=useAuth((s)=>s.user?.role);
  const query=useQuery<any>({queryKey:['templates'],queryFn:()=>api.get('/pot-templates?page_size=100')});
  const save=useMutation({mutationFn:(value:any)=>editing?api.put(`/pot-templates/${editing.id}`,value):api.post('/pot-templates',value),onSuccess:()=>{setVisible(false);setEditing(undefined);form.resetFields();qc.invalidateQueries({queryKey:['templates']});qc.invalidateQueries({queryKey:['pots']})},onError:(e)=>Message.error(errorMessage(e))});
  const remove=useMutation({mutationFn:(id:string)=>api.delete(`/pot-templates/${id}`),onSuccess:()=>qc.invalidateQueries({queryKey:['templates']}),onError:(e)=>Message.error(errorMessage(e))});
  const open=(item?:any)=>{setEditing(item);form.setFieldsValue(item?{name:item.name,yaml:item.yaml}:{name:'',yaml:example});setVisible(true)};
  const columns:any[]=[{title:'模板名称',dataIndex:'name',render:(v:string)=><Typography.Text bold>{v}</Typography.Text>},{title:'版本',dataIndex:'version',render:(v:number)=><Tag color="arcoblue">v{v}</Tag>},{title:'部署实例',dataIndex:'instance_count',render:(v:number)=><Tag color={v?'green':'gray'}>{v||0}</Tag>},{title:'最近更新',dataIndex:'updated_at',render:formatTime},{title:'操作',render:(_:any,r:any)=>role!=='viewer'&&<Space><Button type="text" icon={<IconEdit/>} onClick={()=>open(r)}>编辑</Button><Popconfirm title={r.instance_count?'模板仍有部署实例，无法删除':'确认删除该模板？'} disabled={Boolean(r.instance_count)} onOk={()=>remove.mutate(r.id)}><Button type="text" status="danger" disabled={Boolean(r.instance_count)} icon={<IconDelete/>}/></Popconfirm></Space>}];
  return <><PageHeader title="自定义 Web 蜜罐" description="使用受限 YAML 描述路由、静态响应和捕获字段；可在蜜罐编排中部署，更新后自动热重载" extra={role!=='viewer'&&<Button type="primary" icon={<IconPlus/>} onClick={()=>open()}>新建模板</Button>}/><div className="table-panel"><Table rowKey="id" loading={query.isLoading} data={query.data?.items||[]} columns={columns} pagination={false}/></div><Modal title={editing?'编辑 Web 蜜罐模板':'新建 Web 蜜罐模板'} visible={visible} style={{width:760}} onCancel={()=>setVisible(false)} onOk={()=>form.validate().then(save.mutate)} confirmLoading={save.isPending}><Form form={form} layout="vertical"><Form.Item label="模板名称" field="name" rules={[{required:true}]}><Input placeholder="唯一模板名称"/></Form.Item><Form.Item label="YAML 定义" field="yaml" rules={[{required:true}]} extra="listen.port 用作默认说明；实例实际监听端口以蜜罐编排配置为准"><Input.TextArea className="yaml-editor" autoSize={{minRows:18,maxRows:24}} spellCheck={false}/></Form.Item></Form></Modal></>;
}
