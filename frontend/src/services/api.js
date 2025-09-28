import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080/api/v1';

export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const petAPI = {
  createPet: (ownerName) => api.post('/pets', { owner_name: ownerName }),
  getPet: (petId) => api.get(`/pets/${petId}`),
  getAllPets: () => api.get('/pets'),
  startExploration: (petId) => api.post(`/pets/${petId}/explore`),
  getEvents: (limit = 50) => api.get(`/events?limit=${limit}`),
};

export const createWebSocketConnection = () => {
  const wsURL = process.env.REACT_APP_WS_URL || 'ws://localhost:8080/ws';
  return new WebSocket(wsURL);
};