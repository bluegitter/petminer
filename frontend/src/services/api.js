import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8081/api/v1';

export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const petAPI = {
  // 创建宠物
  createPet: (ownerName) => api.post('/pets', { owner_name: ownerName }),
  
  // 获取宠物信息
  getPet: (petId) => api.get(`/pets/${petId}`),
  getAllPets: () => api.get('/pets'),
  getPetStatus: (petId) => api.get(`/pets/${petId}/status`),
  
  // 宠物初始化
  rollRace: (petId) => api.post(`/pets/${petId}/roll-race`),
  rollSkill: (petId) => api.post(`/pets/${petId}/roll-skill`),

  // 宠物行为
  startExploration: (petId) => api.post(`/pets/${petId}/explore`),
  restPet: (petId) => api.post(`/pets/${petId}/rest`),
  feedPet: (petId, amount = 20) => api.post(`/pets/${petId}/feed`, { amount }),
  socializePet: (petId) => api.post(`/pets/${petId}/socialize`),
  
  // 通用命令接口
  executeCommand: (petId, command, params = {}) => api.post(`/pets/${petId}/command`, { command, params }),
  
  // 事件
  getEvents: (limit = 50) => api.get(`/events?limit=${limit}`),
};

export const createWebSocketConnection = () => {
  const wsURL = process.env.REACT_APP_WS_URL || 'ws://localhost:8081/ws';
  return new WebSocket(wsURL);
};