import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { Card, Descriptions, Tag, Loading, Button, MessagePlugin } from 'tdesign-react';
import { ChevronLeftIcon } from 'tdesign-icons-react';
import { getTask } from '../api/tasks';

const TaskDetail = () => {
  const { taskId } = useParams();
  const [task, setTask] = useState(null);
  const [loading, setLoading] = useState(true);

  const fetchTask = async () => {
    try {
      const res = await getTask(taskId);
      setTask(res.data);
    } catch (error) {
      MessagePlugin.error('获取任务详情失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTask();
    const interval = setInterval(() => {
      if (task?.status !== 'done' && task?.status !== 'failed') {
        fetchTask();
      }
    }, 3000);
    return () => clearInterval(interval);
  }, [taskId]);

  if (loading) return <Loading text="加载中..." />;
  if (!task) return <div>任务不存在</div>;

  const statusMap = {
    pending: { theme: 'default', label: '等待中' },
    running: { theme: 'primary', label: '采集中' },
    done: { theme: 'success', label: '已完成' },
    failed: { theme: 'danger', label: '失败' },
  };
  const status = statusMap[task.status] || statusMap.pending;

  return (
    <div>
      <div style={{ marginBottom: '16px' }}>
        <Link to="/">
          <Button variant="text" icon={<ChevronLeftIcon />}>返回列表</Button>
        </Link>
      </div>

      <Card title={`任务: ${task.task_id}`} bordered>
        <Descriptions layout="vertical" colon>
          <Descriptions.DescriptionsItem label="状态"><Tag theme={status.theme}>{status.label}</Tag></Descriptions.DescriptionsItem>
          <Descriptions.DescriptionsItem label="目标 IP">{task.target_ip}</Descriptions.DescriptionsItem>
          <Descriptions.DescriptionsItem label="进程 PID">{task.pid || '系统全局采样'}</Descriptions.DescriptionsItem>
          <Descriptions.DescriptionsItem label="采样时长">{task.duration} 秒</Descriptions.DescriptionsItem>
          <Descriptions.DescriptionsItem label="采样频率">{task.frequency} Hz</Descriptions.DescriptionsItem>
          <Descriptions.DescriptionsItem label="采集器">{task.profiler_type}</Descriptions.DescriptionsItem>
          <Descriptions.DescriptionsItem label="状态信息" span={2}>{task.status_msg}</Descriptions.DescriptionsItem>
          <Descriptions.DescriptionsItem label="创建时间">{new Date(task.created_at).toLocaleString()}</Descriptions.DescriptionsItem>
          <Descriptions.DescriptionsItem label="更新时间">{new Date(task.updated_at).toLocaleString()}</Descriptions.DescriptionsItem>
        </Descriptions>
      </Card>

      <Card title="🔥 火焰图" style={{ marginTop: '16px' }}>
        {task.flamegraph_url ? (
          <div style={{ width: '100%', height: '600px' }}>
            <iframe src={task.flamegraph_url} style={{ width: '100%', height: '100%', border: 'none' }} title="Flamegraph" />
          </div>
        ) : (
          <div style={{ textAlign: 'center', padding: '40px', color: '#999', background: '#f5f5f5', borderRadius: '8px' }}>
            {task.status === 'running' ? '⏳ 采集进行中，请等待...' : '📭 暂无火焰图'}
          </div>
        )}
      </Card>
    </div>
  );
};

export default TaskDetail;
