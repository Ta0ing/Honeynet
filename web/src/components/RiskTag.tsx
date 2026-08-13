import { Tag } from '@arco-design/web-react';
import { riskMeta } from '../presentation';

export default function RiskTag({ value }: { value?: unknown }) {
  const meta = riskMeta(value);
  return <Tag className="risk-tag" color={meta.color}>{meta.label}</Tag>;
}
