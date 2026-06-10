const API_BASE = '/api';

export const api = {
  async getScripts() {
    const response = await fetch(`${API_BASE}/scripts`);
    if (!response.ok) throw new Error('获取剧本列表失败');
    return response.json();
  },

  async getCarpools(status = '') {
    const url = status 
      ? `${API_BASE}/carpools?status=${status}` 
      : `${API_BASE}/carpools`;
    const response = await fetch(url);
    if (!response.ok) throw new Error('获取拼车列表失败');
    return response.json();
  },

  async createCarpool(data) {
    const response = await fetch(`${API_BASE}/carpools`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || '发起拼车失败');
    }
    return response.json();
  },

  async joinCarpool(data) {
    const response = await fetch(`${API_BASE}/carpools/join`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(data),
    });
    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || '加入拼车失败');
    }
    return response.json();
  },
};
