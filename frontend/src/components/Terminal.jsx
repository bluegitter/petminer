import React, { useEffect, useRef } from 'react';

const Terminal = ({ events, title = "事件日志" }) => {
  const terminalRef = useRef(null);

  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
    }
  }, [events]);

  const getEventColor = (eventType) => {
    switch (eventType) {
      case 'battle':
        return 'text-red-400';
      case 'discovery':
        return 'text-yellow-400';
      case 'rare_find':
        return 'text-purple-400';
      case 'level_up':
        return 'text-blue-400';
      case 'social':
        return 'text-pink-400';
      case 'reward':
        return 'text-green-400';
      default:
        return 'text-terminal-text';
    }
  };

  const formatTimestamp = (timestamp) => {
    return new Date(timestamp).toLocaleTimeString('zh-CN', { 
      hour12: false,
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  };

  return (
    <div className="bg-black border border-terminal-text rounded-lg overflow-hidden">
      <div className="bg-terminal-text text-black px-3 md:px-4 py-2 font-bold text-sm md:text-base">
        <span className="hidden md:inline">{title}</span>
        <span className="md:hidden">事件日志</span>
      </div>
      <div 
        ref={terminalRef}
        className="h-64 md:h-80 lg:h-96 overflow-y-auto p-3 md:p-4 terminal-scroll"
      >
        {events.length === 0 ? (
          <div className="text-gray-500 text-sm">等待事件...</div>
        ) : (
          events.map((event, index) => (
            <div key={`${event.id}-${index}`} className="mb-1 md:mb-2">
              <div className="flex flex-col md:flex-row md:items-start gap-1 md:gap-2">
                <span className="text-gray-400 text-xs md:text-sm flex-shrink-0">
                  [{formatTimestamp(event.timestamp)}]
                </span>
                <span className={`text-xs md:text-sm ${getEventColor(event.type)} break-words`}>
                  {event.message}
                </span>
              </div>
            </div>
          ))
        )}
        <div className="typing-cursor opacity-50"></div>
      </div>
    </div>
  );
};

export default Terminal;