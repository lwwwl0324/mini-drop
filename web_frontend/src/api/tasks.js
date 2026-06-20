import client from './client';

export const createTask = (data) => {
  return client.post('/tasks', data);
};

export const getTasks = () => {
  return client.get('/tasks');
};

export const getTask = (taskId) => {
  return client.get(`/tasks/${taskId}`);
};
