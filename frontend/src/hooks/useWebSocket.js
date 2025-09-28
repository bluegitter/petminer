import { useState, useEffect, useRef } from 'react';
import { createWebSocketConnection } from '../services/api';

export const useWebSocket = () => {
  const [events, setEvents] = useState([]);
  const [connectionStatus, setConnectionStatus] = useState('Disconnected');
  const ws = useRef(null);

  useEffect(() => {
    const connectWebSocket = () => {
      try {
        ws.current = createWebSocketConnection();

        ws.current.onopen = () => {
          console.log('WebSocket Connected');
          setConnectionStatus('Connected');
        };

        ws.current.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data);
            if (message.type === 'event') {
              setEvents(prevEvents => {
                const newEvents = [...prevEvents, message.data];
                return newEvents.slice(-100);
              });
            }
          } catch (error) {
            console.error('Error parsing WebSocket message:', error);
          }
        };

        ws.current.onclose = () => {
          console.log('WebSocket Disconnected');
          setConnectionStatus('Disconnected');
          setTimeout(connectWebSocket, 3000);
        };

        ws.current.onerror = (error) => {
          console.error('WebSocket Error:', error);
          setConnectionStatus('Error');
        };
      } catch (error) {
        console.error('Failed to create WebSocket connection:', error);
        setConnectionStatus('Error');
        setTimeout(connectWebSocket, 3000);
      }
    };

    connectWebSocket();

    return () => {
      if (ws.current) {
        ws.current.close();
      }
    };
  }, []);

  return { events, connectionStatus };
};