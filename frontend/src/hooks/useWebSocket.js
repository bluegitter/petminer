import { useState, useEffect, useRef } from 'react';
import { createWebSocketConnection } from '../services/api';

export const useWebSocket = () => {
  const [events, setEvents] = useState([]);
  const [connectionStatus, setConnectionStatus] = useState('Disconnected');
  const ws = useRef(null);
  const reconnectTimer = useRef(null);

  useEffect(() => {
    const connectWebSocket = () => {
      // 防止重复连接
      if (ws.current && ws.current.readyState === WebSocket.OPEN) {
        console.log('WebSocket already connected, skipping...');
        return;
      }

      try {
        console.log('Creating new WebSocket connection...');
        ws.current = createWebSocketConnection();

        ws.current.onopen = () => {
          console.log('WebSocket Connected');
          setConnectionStatus('Connected');
          // 清除重连计时器
          if (reconnectTimer.current) {
            clearTimeout(reconnectTimer.current);
            reconnectTimer.current = null;
          }
        };

        ws.current.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data);
            if (message.type === 'event') {
              setEvents(prevEvents => {
                // 前端去重检查 - 防止重复事件ID
                const eventExists = prevEvents.some(existingEvent => 
                  existingEvent.id === message.data.id
                );
                
                if (eventExists) {
                  return prevEvents;
                }
                
                const newEvents = [...prevEvents, message.data];
                return newEvents.slice(-100);
              });
            }
          } catch (error) {
            console.error('Error parsing WebSocket message:', error);
          }
        };

        ws.current.onclose = (event) => {
          console.log('WebSocket Disconnected', event.code, event.reason);
          setConnectionStatus('Disconnected');
          
          // 只有在非正常关闭时才重连
          if (event.code !== 1000 && !reconnectTimer.current) {
            reconnectTimer.current = setTimeout(connectWebSocket, 3000);
          }
        };

        ws.current.onerror = (error) => {
          console.error('WebSocket Error:', error);
          setConnectionStatus('Error');
        };
      } catch (error) {
        console.error('Failed to create WebSocket connection:', error);
        setConnectionStatus('Error');
        if (!reconnectTimer.current) {
          reconnectTimer.current = setTimeout(connectWebSocket, 3000);
        }
      }
    };

    connectWebSocket();

    return () => {
      console.log('Cleaning up WebSocket connection...');
      if (reconnectTimer.current) {
        clearTimeout(reconnectTimer.current);
        reconnectTimer.current = null;
      }
      if (ws.current) {
        ws.current.close(1000, 'Component unmounting');
        ws.current = null;
      }
    };
  }, []);

  return { events, connectionStatus };
};