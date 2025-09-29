import React, { useEffect, useRef } from 'react';

const Terminal = ({ events, title = "宠物冒险日记" }) => {
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
    <div className="w-full bg-transparent" style={{ height: '100%' }}>
      <div className="px-3 py-2 text-sm font-bold text-black bg-terminal-text md:px-4 md:text-base">
        <span className="hidden md:inline">{title}</span>
        <span className="md:hidden">冒险日记</span>
      </div>
      <div 
        ref={terminalRef}
        className="p-3 overflow-y-auto font-mono text-xs bg-transparent md:p-4 terminal-scroll text-terminal-text md:text-sm responsive-terminal-content"
      >
        {events.length === 0 ? (
          <div className="text-sm text-gray-500">等待事件...</div>
        ) : (
          events.map((event, index) => (
            <div key={`${event.id}-${index}`} className="mb-1 md:mb-2">
              <div className="flex flex-col gap-1 md:flex-row md:items-start md:gap-2">
                <span className="flex-shrink-0 text-xs text-gray-400 md:text-sm">
                  [{formatTimestamp(event.timestamp)}]
                </span>
                <span className={`text-xs md:text-sm ${getEventColor(event.type)} break-words`}>
                  {event.message}
                </span>
              </div>
            </div>
          ))
        )}
        <div className="opacity-50 typing-cursor"></div>
      </div>
    </div>
  );
};

export default Terminal;