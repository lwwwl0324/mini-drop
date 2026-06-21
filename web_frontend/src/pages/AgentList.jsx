import React, { useState, useEffect } from 'react';
import { Table, Tag, Card, Loading } from 'tdesign-react';
import axios from 'axios';

const AgentList = () => {
  const [agents, setAgents] = useState([]);
  const [loading, setLoading] = useState(false);

  const fetchAgents = async () => {
    setLoading(true);
    try {
      const res = await axios.get('http://localhost:8191/api/v1/agents');
      setAgents(res.data.data || []);
    } catch (error) {
      console.error('获取 Agent 列表失败:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAgents();
    const interval = setInterval(fetchAgents, 10000);
    return () => clearInterval(interval);
  }, []);

  const columns = [
    { colKey: 'uid', title: 'Agent ID', width: 200 },
    { colKey: 'hostname', title: '主机名', width: 150 },
    { colKey: 'ip_addr', title: 'IP 地址', width: 150 },
    { colKey: 'version', title: '版本', width: 100 },
    {
      colKey: 'status',
      title: '状态',
      width: 100,
      cell: ({ row }) => {
        const statusMap = {
          online: { theme: 'success', label: '🟢 在线' },
          offline: { theme: 'danger', label: '🔴 离线' },
        };
        const status = statusMap[row.status] || statusMap.offline;
        return <Tag theme={status.theme}>{status.label}</Tag>;
      },
    },
    {
      colKey: 'last_heartbeat',
      title: '最后心跳',
      width: 200,
      cell: ({ row }) => new Date(row.last_heartbeat).toLocaleString(),
    },
  ];

  return (
    <div>
      <h2>🤖 Agent 列表</h2>
      <Card bordered style={{ marginTop: '16px' }}>
        <Loading loading={loading}>
          <Table data={agents} columns={columns} rowKey="uid" stripe hover />
        </Loading>
      </Card>
    </div>
  );
};

export default AgentList;
