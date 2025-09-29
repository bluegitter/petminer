import React, { useState, useRef, useEffect } from 'react';
import { Terminal as TerminalIcon, Send } from 'lucide-react';

const CLITerminal = ({ selectedPet, onCommand }) => {
  const [input, setInput] = useState('');
  const [history, setHistory] = useState([
    { type: 'system', content: '🚀 MiningPet CLI v1.0 已启动' },
    { type: 'system', content: '输入 "help" 查看可用命令' },
  ]);
  const [commandHistory, setCommandHistory] = useState([]);
  const [historyIndex, setHistoryIndex] = useState(-1);
  const terminalRef = useRef(null);
  const inputRef = useRef(null);

  // 可用命令列表
  const commands = {
    help: {
      description: '显示所有可用命令',
      usage: 'help [command]',
      examples: ['help', 'help status']
    },
    status: {
      description: '显示宠物当前状态',
      usage: 'status [pet_name]',
      examples: ['status', 'status Lucky']
    },
    explore: {
      description: '让宠物开始探索',
      usage: 'explore <direction>',
      examples: ['explore north', 'explore forest']
    },
    rest: {
      description: '让宠物休息',
      usage: 'rest [duration]',
      examples: ['rest', 'rest 30']
    },
    feed: {
      description: '给宠物喂食',
      usage: 'feed [amount]',
      examples: ['feed', 'feed 20']
    },
    socialize: {
      description: '让宠物社交',
      usage: 'socialize',
      examples: ['socialize']
    },
    inventory: {
      description: '查看宠物背包',
      usage: 'inventory',
      examples: ['inventory']
    },
    clear: {
      description: '清空终端',
      usage: 'clear',
      examples: ['clear']
    },
    addcoins: {
      description: '给宠物添加金币（调试用）',
      usage: 'addcoins [amount]',
      examples: ['addcoins', 'addcoins 100']
    },
    friends: {
      description: '查看宠物的好友列表',
      usage: 'friends',
      examples: ['friends']
    }
  };

  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
    }
  }, [history]);

  const addToHistory = (type, content) => {
    setHistory(prev => [...prev, { type, content, timestamp: new Date() }]);
  };

  const handleCommand = async (commandLine) => {
    const trimmed = commandLine.trim();
    if (!trimmed) return;

    // 添加命令到历史记录
    setCommandHistory(prev => [...prev, trimmed]);
    setHistoryIndex(-1);

    // 显示用户输入
    addToHistory('user', `> ${trimmed}`);

    const [command, ...args] = trimmed.toLowerCase().split(' ');

    try {
      switch (command) {
        case 'help':
          handleHelpCommand(args[0]);
          break;
        case 'status':
          handleStatusCommand(args[0]);
          break;
        case 'explore':
          handleExploreCommand(args);
          break;
        case 'rest':
          handleRestCommand(args[0]);
          break;
        case 'feed':
          handleFeedCommand(args[0]);
          break;
        case 'socialize':
          handleSocializeCommand();
          break;
        case 'inventory':
          handleInventoryCommand();
          break;
        case 'clear':
          handleClearCommand();
          break;
        case 'addcoins':
          handleAddCoinsCommand(args[0]);
          break;
        case 'friends':
          handleFriendsCommand();
          break;
        default:
          addToHistory('error', `未知命令: ${command}. 输入 "help" 查看可用命令.`);
      }
    } catch (error) {
      addToHistory('error', `命令执行错误: ${error.message}`);
    }
  };

  const handleHelpCommand = (specificCommand) => {
    if (specificCommand && commands[specificCommand]) {
      const cmd = commands[specificCommand];
      addToHistory('system', `📖 ${specificCommand} - ${cmd.description}`);
      addToHistory('system', `用法: ${cmd.usage}`);
      addToHistory('system', `示例: ${cmd.examples.join(', ')}`);
    } else {
      addToHistory('system', '📚 可用命令:');
      Object.entries(commands).forEach(([name, cmd]) => {
        addToHistory('system', `  ${name.padEnd(12)} - ${cmd.description}`);
      });
      addToHistory('system', '输入 "help <command>" 获取具体命令的详细信息');
    }
  };

  const handleStatusCommand = (petName) => {
    if (!selectedPet) {
      addToHistory('error', '未选择宠物');
      return;
    }

    const pet = selectedPet;
    addToHistory('system', `🐾 ${pet.name} 的状态:`);
    addToHistory('system', `  等级: ${pet.level}`);
    addToHistory('system', `  生命值: ${pet.health}/${pet.max_health}`);
    addToHistory('system', `  体力: ${pet.energy || 100}/${pet.max_energy || 100}`);
    addToHistory('system', `  饱食度: ${pet.hunger || 80}/100`);
    addToHistory('system', `  社交度: ${pet.social || 50}/100`);
    addToHistory('system', `  心情: ${pet.mood || '普通'}`);
    addToHistory('system', `  金币: ${pet.coins}`);
    addToHistory('system', `  位置: ${pet.location}`);
    addToHistory('system', `  状态: ${pet.status}`);
    addToHistory('system', `  性格: ${pet.personality}`);
  };

  const handleExploreCommand = async (args) => {
    if (!selectedPet) {
      addToHistory('error', '未选择宠物');
      return;
    }

    if (selectedPet.status !== '等待中') {
      addToHistory('error', `${selectedPet.name} 当前正在 ${selectedPet.status}，无法开始探索`);
      return;
    }

    const direction = args[0] || 'unknown';
    addToHistory('system', `🚀 ${selectedPet.name} 开始向 ${direction} 探索...`);
    
    if (onCommand) {
      try {
        await onCommand('explore', { petId: selectedPet.id, direction });
        addToHistory('system', `✅ ${selectedPet.name} 已开始探索`);
      } catch (error) {
        const errorMessage = error.response?.data?.error || error.message || '探索命令执行失败';
        addToHistory('error', `❌ 探索失败: ${errorMessage}`);
      }
    }
  };

  const handleRestCommand = async (duration) => {
    if (!selectedPet) {
      addToHistory('error', '未选择宠物');
      return;
    }

    const restDuration = duration ? parseInt(duration) : 30;
    addToHistory('system', `😴 ${selectedPet.name} 开始休息 ${restDuration} 秒...`);
    
    if (onCommand) {
      try {
        await onCommand('rest', { petId: selectedPet.id, duration: restDuration });
        addToHistory('system', `✅ ${selectedPet.name} 已开始休息`);
      } catch (error) {
        const errorMessage = error.response?.data?.error || error.message || '休息命令执行失败';
        addToHistory('error', `❌ 休息失败: ${errorMessage}`);
      }
    }
  };

  const handleFeedCommand = async (amount) => {
    if (!selectedPet) {
      addToHistory('error', '未选择宠物');
      return;
    }

    const feedAmount = amount ? parseInt(amount) : 20;
    addToHistory('system', `🍖 给 ${selectedPet.name} 喂食 ${feedAmount} 点...`);
    
    if (onCommand) {
      try {
        await onCommand('feed', { petId: selectedPet.id, amount: feedAmount });
        addToHistory('system', `✅ ${selectedPet.name} 已进食完毕`);
      } catch (error) {
        const errorMessage = error.response?.data?.error || error.message || '喂食命令执行失败';
        addToHistory('error', `❌ 喂食失败: ${errorMessage}`);
      }
    }
  };

  const handleSocializeCommand = async () => {
    if (!selectedPet) {
      addToHistory('error', '未选择宠物');
      return;
    }

    addToHistory('system', `🤝 ${selectedPet.name} 开始寻找朋友进行社交...`);
    
    if (onCommand) {
      try {
        await onCommand('socialize', { petId: selectedPet.id });
        addToHistory('system', `✅ ${selectedPet.name} 已开始社交活动`);
      } catch (error) {
        const errorMessage = error.response?.data?.error || error.message || '社交命令执行失败';
        addToHistory('error', `❌ 社交失败: ${errorMessage}`);
      }
    }
  };

  const handleInventoryCommand = () => {
    if (!selectedPet) {
      addToHistory('error', '未选择宠物');
      return;
    }

    addToHistory('system', `🎒 ${selectedPet.name} 的背包:`);
    addToHistory('system', `  金币: ${selectedPet.coins}`);
    if (selectedPet.friends && selectedPet.friends.length > 0) {
      addToHistory('system', `  朋友: ${selectedPet.friends.join(', ')}`);
    } else {
      addToHistory('system', '  朋友: 暂无');
    }
    if (selectedPet.memory && selectedPet.memory.length > 0) {
      addToHistory('system', '  最近记忆:');
      selectedPet.memory.slice(-3).forEach(memory => {
        addToHistory('system', `    - ${memory}`);
      });
    }
  };

  const handleClearCommand = () => {
    setHistory([
      { type: 'system', content: '🚀 MiningPet CLI v1.0 已启动' },
      { type: 'system', content: '输入 "help" 查看可用命令' },
    ]);
  };

  const handleAddCoinsCommand = async (amount) => {
    if (!selectedPet) {
      addToHistory('error', '未选择宠物');
      return;
    }

    const coinAmount = amount ? parseInt(amount) : 100;
    if (isNaN(coinAmount) || coinAmount <= 0) {
      addToHistory('error', '金币数量必须是正整数');
      return;
    }

    addToHistory('system', `💰 给 ${selectedPet.name} 添加 ${coinAmount} 金币...`);
    
    if (onCommand) {
      try {
        await onCommand('addcoins', { petId: selectedPet.id, amount: coinAmount });
        addToHistory('system', `✅ ${selectedPet.name} 获得了 ${coinAmount} 金币！`);
      } catch (error) {
        const errorMessage = error.response?.data?.error || error.message || '添加金币命令执行失败';
        addToHistory('error', `❌ 添加金币失败: ${errorMessage}`);
      }
    }
  };

  const handleFriendsCommand = () => {
    if (!selectedPet) {
      addToHistory('error', '未选择宠物');
      return;
    }

    addToHistory('system', `👥 ${selectedPet.name} 的好友列表:`);
    
    if (selectedPet.friends && selectedPet.friends.length > 0) {
      addToHistory('system', `  总共 ${selectedPet.friends.length} 位好友:`);
      selectedPet.friends.forEach((friend, index) => {
        addToHistory('system', `  ${index + 1}. ${friend}`);
      });
      
      // 显示社交统计
      addToHistory('system', '');
      addToHistory('system', `📊 社交统计:`);
      addToHistory('system', `  社交度: ${selectedPet.social || 0}/100`);
      addToHistory('system', `  好友数量: ${selectedPet.friends.length}`);
      
      // 根据好友数量给出建议
      if (selectedPet.friends.length >= 5) {
        addToHistory('system', `  🌟 ${selectedPet.name} 是个社交达人！`);
      } else if (selectedPet.friends.length >= 2) {
        addToHistory('system', `  😊 ${selectedPet.name} 有不错的社交圈`);
      } else {
        addToHistory('system', `  💡 建议多使用 'socialize' 命令交朋友`);
      }
    } else {
      addToHistory('system', '  暂无好友');
      addToHistory('system', '');
      addToHistory('system', '💡 使用 "socialize" 命令让宠物主动社交交朋友！');
      addToHistory('system', '🤝 通过社交可以提升宠物的心情和社交度');
    }
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    if (input.trim()) {
      handleCommand(input);
      setInput('');
    }
  };

  const handleKeyDown = (e) => {
    if (e.key === 'ArrowUp') {
      e.preventDefault();
      if (commandHistory.length > 0) {
        const newIndex = historyIndex + 1;
        if (newIndex < commandHistory.length) {
          setHistoryIndex(newIndex);
          setInput(commandHistory[commandHistory.length - 1 - newIndex]);
        }
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault();
      if (historyIndex > 0) {
        const newIndex = historyIndex - 1;
        setHistoryIndex(newIndex);
        setInput(commandHistory[commandHistory.length - 1 - newIndex]);
      } else if (historyIndex === 0) {
        setHistoryIndex(-1);
        setInput('');
      }
    }
  };

  const getMessageStyle = (type) => {
    switch (type) {
      case 'user':
        return 'text-terminal-accent font-medium';
      case 'system':
        return 'text-green-400';
      case 'error':
        return 'text-red-400';
      default:
        return 'text-terminal-text';
    }
  };

  const formatTimestamp = (timestamp) => {
    if (!timestamp) return '';
    return timestamp.toLocaleTimeString('zh-CN', { 
      hour12: false,
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  };

  return (
    <div className="flex flex-col h-full overflow-hidden bg-transparent">
      {/* 终端头部 */}
      <div className="flex items-center flex-shrink-0 gap-2 px-3 py-2 text-sm font-bold text-black bg-terminal-text md:px-4 md:text-base">
        <TerminalIcon className="w-4 h-4" />
        <span className="hidden md:inline">指令面板</span>
        <span className="md:hidden">CLI</span>
        {selectedPet && (
          <span className="px-2 py-1 ml-auto text-xs bg-black rounded text-terminal-accent">
            {selectedPet.name}
          </span>
        )}
      </div>

      {/* 终端内容 */}
      <div 
        ref={terminalRef}
        className="p-3 overflow-y-auto font-mono text-xs bg-transparent md:p-4 terminal-scroll text-terminal-text md:text-sm responsive-cli-content md:responsive-terminal-content"
      >
        {history.map((entry, index) => (
          <div key={index} className={`mb-1 ${getMessageStyle(entry.type)}`}>
            {entry.timestamp && (
              <span className="mr-2 text-xs text-gray-500">
                [{formatTimestamp(entry.timestamp)}]
              </span>
            )}
            <span className="break-words whitespace-pre-wrap">{entry.content}</span>
          </div>
        ))}
        
        {/* 移除重复的光标，输入框会显示光标 */}
      </div>

      {/* 输入区域 */}
      <form onSubmit={handleSubmit} className="flex-shrink-0 p-2 bg-green-900 border-t border-green-500 bg-opacity-30 md:p-3">
        <div className="flex items-center gap-2">
          <span className="flex-shrink-0 font-mono text-sm text-green-400">{'>'}</span>
          <input
            ref={inputRef}
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            className="flex-1 font-mono text-sm placeholder-gray-500 bg-transparent outline-none text-terminal-text caret-green-400"
            placeholder="输入命令..."
            autoComplete="off"
            spellCheck="false"
            autoFocus
          />
          <button
            type="submit"
            className="p-1 text-green-400 transition-colors hover:text-terminal-text"
            disabled={!input.trim()}
          >
            <Send className="w-4 h-4" />
          </button>
        </div>
      </form>
    </div>
  );
};

export default CLITerminal;