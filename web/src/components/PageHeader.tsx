import type { ReactNode } from 'react';
import { Typography } from '@arco-design/web-react';

export default function PageHeader({ title, description, extra }: { title: string; description: string; extra?: ReactNode }) {
  return <div className="page-header"><div><Typography.Title heading={4}>{title}</Typography.Title><Typography.Text type="secondary">{description}</Typography.Text></div>{extra}</div>;
}
