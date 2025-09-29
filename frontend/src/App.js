import React, { useState, useEffect } from 'react';
import { petAPI } from './services/api';
import { useWebSocket } from './hooks/useWebSocket';
import Terminal from './components/Terminal';
import PetCard from './components/PetCard';
import CreatePetForm from './components/CreatePetForm';
import CLITerminal from './components/CLITerminal';
import Logo from './components/Logo';
import { Wifi, WifiOff, RefreshCw } from 'lucide-react';
import './index.css';

function App() {
  const [pets, setPets] = useState([]);
  const [selectedPet, setSelectedPet] = useState(null);
  const [loading, setLoading] = useState(false);
  const [logoError, setLogoError] = useState(false);
  const [isTablet, setIsTablet] = useState(false);
  const { events, connectionStatus } = useWebSocket();

  useEffect(() => {
    loadPets();
    loadInitialEvents();
    
    // 检测屏幕尺寸
    const checkTabletSize = () => {
      const width = window.innerWidth;
      setIsTablet(width >= 768 && width <= 1279);
    };
    
    checkTabletSize();
    window.addEventListener('resize', checkTabletSize);
    
    return () => window.removeEventListener('resize', checkTabletSize);
  }, []);

  const loadPets = async () => {
    try {
      const response = await petAPI.getAllPets();
      setPets(response.data.pets || []);
      if (response.data.pets && response.data.pets.length > 0) {
        setSelectedPet(response.data.pets[0]);
      }
    } catch (error) {
      console.error('加载宠物失败:', error);
    }
  };

  const loadInitialEvents = async () => {
    try {
      const response = await petAPI.getEvents(50);
      console.log('初始事件:', response.data.events);
    } catch (error) {
      console.error('加载事件失败:', error);
    }
  };

  const handleCreatePet = async (ownerName) => {
    setLoading(true);
    try {
      const response = await petAPI.createPet(ownerName);
      const newPet = response.data;
      setPets([newPet]); // 只设置这一只宠物，因为每个用户只能有一只
      setSelectedPet(newPet);
    } catch (error) {
      console.error('创建宠物失败:', error);
      throw error;
    } finally {
      setLoading(false);
    }
  };

  const handleStartExploration = async (petId) => {
    try {
      await petAPI.startExploration(petId);
      await loadPets();
    } catch (error) {
      console.error('开始探索失败:', error);
    }
  };

  const handleCLICommand = async (command, params) => {
    try {
      switch (command) {
        case 'explore':
          await handleStartExploration(params.petId);
          break;
        case 'rest':
          await petAPI.restPet(params.petId);
          await loadPets();
          break;
        case 'feed':
          await petAPI.feedPet(params.petId, params.amount);
          await loadPets();
          break;
        case 'socialize':
          await petAPI.socializePet(params.petId);
          await loadPets();
          break;
        default:
          console.log('未实现的命令:', command, params);
      }
    } catch (error) {
      console.error('CLI命令执行失败:', error);
      throw error; // 重新抛出错误，让CLI组件处理
    }
  };

  const getConnectionStatusIcon = () => {
    switch (connectionStatus) {
      case 'Connected':
        return <Wifi className="w-4 h-4 text-green-400" />;
      case 'Disconnected':
        return <WifiOff className="w-4 h-4 text-red-400" />;
      default:
        return <RefreshCw className="w-4 h-4 text-yellow-400 animate-spin" />;
    }
  };

  return (
    <div className="min-h-screen bg-terminal-bg text-terminal-text">
      {/* 顶部导航栏 - 手机端紧凑 */}
      <header className="sticky top-0 z-50 border-b bg-terminal-bg border-terminal-text bg-opacity-95 backdrop-blur-sm">
        <div className="container px-4 py-2 mx-auto md:py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 md:gap-3">
              {!logoError ? (
                <img 
                  src="/logo.png" 
                  alt="MiningPet Logo" 
                  className="w-8 h-8 border-2 rounded-lg shadow-lg md:w-12 md:h-12 border-terminal-accent"
                  onError={() => setLogoError(true)}
                />
              ) : (
                <Logo className="w-8 h-8 md:w-12 md:h-12" />
              )}
              <div>
                <h1 className="text-xl font-bold md:text-3xl text-terminal-accent typing-cursor">
                  MiningPet
                </h1>
                <p className="hidden text-xs text-gray-400 md:text-sm md:block">
                  AI宠物挖矿世界
                </p>
              </div>
            </div>
            
            {/* 连接状态和统计 - 手机端简化 */}
            <div className="flex items-center gap-2 md:gap-6">
              <div className="hidden text-sm text-gray-400 lg:block">
                活跃事件: <span className="font-bold text-terminal-accent">{events.length}</span>
              </div>
              <div className="flex items-center gap-1 px-2 py-1 bg-black bg-opacity-50 border rounded-full md:gap-2 md:px-3 md:py-2 border-terminal-text">
                {getConnectionStatusIcon()}
                <span className="hidden text-xs font-medium md:text-sm md:inline">{connectionStatus}</span>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* 主要内容区域 */}
      <main className="container px-4 py-3 mx-auto md:py-6">
        <div className="grid grid-cols-1 gap-3 lg:grid-cols-12 md:gap-6">
          {/* 左侧宠物面板 - 手机端更紧凑 */}
          <div className="space-y-2 lg:col-span-5 xl:col-span-4 md:space-y-4 animate-slide-in-left">
            {pets.length === 0 ? (
              <div className="animate-slide-in-up">
                <CreatePetForm onCreatePet={handleCreatePet} />
              </div>
            ) : (
              <>
                {/* 当前宠物卡片 */}
                <div className="relative animate-slide-in-up">
                  <div className="absolute rounded-lg opacity-25 -inset-1 bg-gradient-to-r from-terminal-accent to-blue-400 blur animate-glow"></div>
                  <div className="relative card-hover">
                    <PetCard 
                      pet={selectedPet} 
                      onStartExploration={handleStartExploration}
                    />
                  </div>
                </div>
                
                {/* 宠物信息面板 */}
                <div className="p-1 bg-black border rounded-lg shadow-lg border-terminal-text md:p-2 animate-slide-in-up card-hover" style={{animationDelay: '0.2s'}}>
                  <h3 className="flex items-center gap-2 mb-1 text-sm font-bold md:text-base text-terminal-accent text-glow">
                    <span className="animate-float">🏠</span> 
                    <span className="hidden md:inline">我的宠物</span>
                    <span className="md:hidden">宠物</span>
                  </h3>
                  
                  {/* 宠物基本信息 */}
                  <div className="space-y-0.5 text-xs md:text-sm">
                    <div className="flex justify-between">
                      <span className="text-gray-400">训练师:</span>
                      <span className="font-medium text-terminal-accent">{selectedPet?.owner}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-400">创建时间:</span>
                      <span className="text-gray-300">
                        {selectedPet?.created_at ? new Date(selectedPet.created_at).toLocaleDateString('zh-CN') : '未知'}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-400">最后活动:</span>
                      <span className="text-gray-300">
                        {selectedPet?.last_activity ? new Date(selectedPet.last_activity).toLocaleTimeString('zh-CN') : '未知'}
                      </span>
                    </div>
                    {selectedPet?.friends && selectedPet.friends.length > 0 && (
                      <div>
                        <span className="text-gray-400">朋友列表:</span>
                        <div className="flex flex-wrap gap-1 mt-1">
                          {selectedPet.friends.slice(0, 3).map((friend, index) => (
                            <span key={index} className="px-2 py-1 text-xs text-purple-300 bg-purple-900 bg-opacity-50 rounded">
                              {friend}
                            </span>
                          ))}
                          {selectedPet.friends.length > 3 && (
                            <span className="text-xs text-gray-500">+{selectedPet.friends.length - 3}...</span>
                          )}
                        </div>
                      </div>
                    )}
                  </div>
                  
                  {/* 单宠物限制提示 */}
                  <div className="p-1 mt-1 bg-blue-900 border border-blue-500 rounded-lg md:p-2 bg-opacity-30 border-opacity-30">
                    <div className="flex items-start gap-2">
                      <span className="text-lg">🔒</span>
                      <div className="text-xs text-gray-300">
                        <div className="mb-1 font-medium text-blue-400">专一伙伴</div>
                        <p>每位训练师只能拥有一只宠物，请珍惜与 {selectedPet?.name} 的冒险时光！</p>
                      </div>
                    </div>
                  </div>
                </div>
              </>
            )}
          </div>

          {/* 右侧终端区域 - 事件日志和CLI */}
          <div className="lg:col-span-7 xl:col-span-8 animate-slide-in-right">
            <div className="grid grid-cols-1 gap-3 xl:grid-cols-2 md:gap-6 responsive-terminal-grid">
              {/* 事件日志 */}
              <div className="relative particle-bg responsive-terminal-panel" 
                   style={{ 
                     height: isTablet ? '350px' : undefined
                   }}>
                <div className="absolute rounded-lg -inset-1 bg-gradient-to-r from-green-400 to-terminal-accent blur opacity-20 animate-glow"></div>
                <div className="relative terminal-enhanced" style={{ height: '100%' }}>
                  <Terminal events={events} title="宠物冒险日记" />
                </div>
                {/* 数据流效果 */}
                <div className="data-stream" style={{animationDelay: '0s'}}></div>
                <div className="data-stream" style={{animationDelay: '2s', top: '40%'}}></div>
                <div className="data-stream" style={{animationDelay: '4s', top: '60%'}}></div>
              </div>
              
              {/* CLI终端 */}
              <div className="relative responsive-cli-panel"
                   style={{ 
                     height: isTablet ? '350px' : undefined
                   }}>
                <div className="absolute rounded-lg -inset-1 bg-gradient-to-r from-blue-400 to-purple-500 blur opacity-20 animate-glow"></div>
                <div className="relative cli-enhanced" style={{ height: '100%' }}>
                  <CLITerminal 
                    selectedPet={selectedPet} 
                    onCommand={handleCLICommand}
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </main>

      {/* 底部状态栏 - 手机端简化 */}
      <footer className="mt-4 bg-black bg-opacity-50 border-t border-terminal-text md:mt-8">
        <div className="container px-4 py-3 mx-auto md:py-4">
          <div className="flex flex-col items-center justify-between gap-2 sm:flex-row md:gap-4">
            <div className="text-center sm:text-left">
              <p className="text-xs text-gray-400 md:text-sm">
                <span className="hidden md:inline">灵感来自早期比特币挖矿 - 界面简陋，但内核强大</span>
                <span className="md:hidden">AI宠物挖矿世界</span>
              </p>
            </div>
            <div className="flex items-center gap-3 text-xs md:gap-4 md:text-sm">
              {pets.length > 0 && (
                <span className="text-gray-400">
                  <span className="hidden md:inline">我的宠物: </span>
                  <span className="md:hidden">宠物: </span>
                  <span className="font-bold text-terminal-accent">{selectedPet?.name || '未知'}</span>
                </span>
              )}
              <span className="hidden text-gray-400 md:inline">
                连接状态: <span className={connectionStatus === 'Connected' ? 'text-green-400' : 'text-red-400'}>{connectionStatus}</span>
              </span>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}

export default App;