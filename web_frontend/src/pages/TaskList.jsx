import React, { useState, useEffect } from 'react';
import { 
  Table, 
  Button, 
  Dialog, 
  Form, 
  Input, 
  InputNumber, 
  Select,
  Tag,
  Loading
} from 'tdesign-react';
import { AddIcon } from 'tdesign-icons-react';
import { getTasks, createTask } from '../api/tasks';

const { FormItem } = Form;

const TaskList = () => {
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(false);
  const [dialogVisible, setDialogVisible] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [form] = Form.useForm();

  const fetchTasks = async () => {
    setLoading(true);
    try {
      const res = await getTasks();
      setTasks(res.data || []);
    } catch (error) {
      console.error('获取任务列表失败:', error);
      alert('获取任务列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTasks();
    const interval = setInterval(fetchTasks, 5000);
    return () => clearInterval(interval);
  }, []);

  const handleSubmit = async () => {
    if (submitting) return;
    
    try {
      const values = form.getFieldsValue(true);
      console.log('📤 提交任务:', values);
      
      setSubmitting(true);
      
      const res = await createTask({
        target_ip: values.target_ip || '127.0.0.1',
        pid: values.pid || 0,
        duration: values.duration || 10,
        frequency: values.frequency || 999,
        profiler_type: values.profiler_type || 'perf',
      });
      
      console.log('✅ 任务创建成功:', res.data);
      alert(`✅ 任务创建成功: ${res.data.task_id}`);
      
      setDialogVisible(false);
      form.reset();
      setSubmitting(false);
      
      await fetchTasks();
      
    } catch (error) {
      console.error('❌ 创建失败:', error);
      const errMsg = error.response?.data?.msg || error.message || '创建任务失败';
      alert('❌ 创建失败: ' + errMsg);
      setSubmitting(false);
    }
  };

  const handleClose = () => {
    setDialogVisible(false);
    form.reset();
    setSubmitting(false);
  };

  const statusMap = {
    pending: { theme: 'default', label: '等待中' },
    running: { theme: 'primary', label: '采集中' },
    done: { theme: 'success', label: '已完成' },
    failed: { theme: 'danger', label: '失败' },
  };

  const columns = [
    { colKey: 'task_id', title: '任务 ID', width: 200, ellipsis: true },
    { colKey: 'target_ip', title: '目标 IP', width: 120 },
    { colKey: 'pid', title: 'PID', width: 80, cell: ({ row }) => row.pid === 0 ? '系统' : row.pid },
    { colKey: 'duration', title: '时长(s)', width: 80 },
    { colKey: 'frequency', title: '频率(Hz)', width: 90 },
    {
      colKey: 'status',
      title: '状态',
      width: 100,
      cell: ({ row }) => {
        const status = statusMap[row.status] || statusMap.pending;
        return <Tag theme={status.theme}>{status.label}</Tag>;
      },
    },
    { colKey: 'status_msg', title: '状态信息', ellipsis: true },
    {
      colKey: 'flamegraph_url',
      title: '结果',
      width: 120,
      cell: ({ row }) => {
        if (row.flamegraph_url) {
          return (
            <Button 
              size="small" 
              theme="primary" 
              variant="text"
              onClick={() => window.open(row.flamegraph_url, '_blank')}
            >
              查看
            </Button>
          );
        }
        return <span style={{ color: '#999' }}>暂无</span>;
      },
    },
    {
      colKey: 'created_at',
      title: '创建时间',
      width: 180,
      cell: ({ row }) => new Date(row.created_at).toLocaleString(),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: '16px', display: 'flex', justifyContent: 'space-between' }}>
        <h2>📊 任务列表</h2>
        <Button 
          theme="primary" 
          icon={<AddIcon />} 
          onClick={() => {
            form.reset();
            form.setFieldsValue({
              target_ip: '127.0.0.1',
              pid: 0,
              duration: 10,
              frequency: 999,
              profiler_type: 'perf',
            });
            setDialogVisible(true);
            setSubmitting(false);
          }}
        >
          新建任务
        </Button>
      </div>

      <Loading loading={loading}>
        <Table data={tasks} columns={columns} rowKey="id" stripe hover size="medium" empty="暂无任务" />
      </Loading>

      <Dialog
        visible={dialogVisible}
        onClose={handleClose}
        header="新建采样任务"
        footer={
          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
            <Button variant="outline" onClick={handleClose}>
              取消
            </Button>
            <Button 
              theme="primary" 
              loading={submitting}
              onClick={handleSubmit}
            >
              创建
            </Button>
          </div>
        }
        width={500}
        destroyOnClose
      >
        <Form form={form} labelWidth={80} colon>
          <FormItem label="目标 IP" name="target_ip" initialData="127.0.0.1">
            <Input placeholder="请输入目标 IP" />
          </FormItem>
          <FormItem 
            label="进程 PID" 
            name="pid" 
            initialData={0}
            help="输入 0 表示采样整个系统（仅 perf 支持）"
          >
            <InputNumber placeholder="输入 PID" min={0} step={1} theme="normal" />
          </FormItem>
          <FormItem label="采样时长" name="duration" initialData={10}>
            <InputNumber placeholder="秒" min={1} max={60} step={1} theme="normal" />
          </FormItem>
          <FormItem label="采样频率" name="frequency" initialData={999}>
            <InputNumber placeholder="Hz" min={1} max={9999} step={1} theme="normal" />
          </FormItem>
          <FormItem label="采集器" name="profiler_type" initialData="perf">
            <Select
              options={[
                { label: 'perf (CPU采样)', value: 'perf' },
                { label: 'eBPF (IO追踪)', value: 'ebpf' },
                { label: 'py-spy (Python)', value: 'pyspy' },
              ]}
            />
          </FormItem>
        </Form>
      </Dialog>
    </div>
  );
};

export default TaskList;
