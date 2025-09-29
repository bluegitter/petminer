import React, { useState, useEffect } from 'react';
import { petAPI } from './services/api';
import { useWebSocket } from './hooks/useWebSocket';
import Terminal from './components/Terminal';
import PetCard from './components/PetCard';
import CreatePetForm from './components/CreatePetForm';
import Logo from './components/Logo';
import { Wifi, WifiOff, RefreshCw } from 'lucide-react';
import './index.css';

function App() {
  const [pets, setPets] = useState([]);
  const [selectedPet, setSelectedPet] = useState(null);
  const [loading, setLoading] = useState(false);
  const [logoError, setLogoError] = useState(false);
  const { events, connectionStatus } = useWebSocket();

  useEffect(() => {
    loadPets();
    loadInitialEvents();
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
      setPets(prev => [...prev, newPet]);
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
      <header className="sticky top-0 z-50 bg-terminal-bg border-b border-terminal-text bg-opacity-95 backdrop-blur-sm">
        <div className="container mx-auto px-4 py-2 md:py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 md:gap-3">
              {!logoError ? (
                <img 
                  src="/logo.png" 
                  alt="MiningPet Logo" 
                  className="w-8 md:w-12 h-8 md:h-12 rounded-lg shadow-lg border-2 border-terminal-accent"
                  onError={() => setLogoError(true)}
                />
              ) : (
                <Logo className="w-8 md:w-12 h-8 md:h-12" />
              )}
              <div>
                <h1 className="text-xl md:text-3xl font-bold text-terminal-accent typing-cursor">
                  MiningPet
                </h1>
                <p className="text-gray-400 text-xs md:text-sm hidden md:block">
                  AI宠物挖矿世界
                </p>
              </div>
            </div>
            
            {/* 连接状态和统计 - 手机端简化 */}
            <div className="flex items-center gap-2 md:gap-6">
              <div className="hidden lg:block text-sm text-gray-400">
                活跃事件: <span className="text-terminal-accent font-bold">{events.length}</span>
              </div>
              <div className="flex items-center gap-1 md:gap-2 px-2 md:px-3 py-1 md:py-2 rounded-full border border-terminal-text bg-black bg-opacity-50">
                {getConnectionStatusIcon()}
                <span className="text-xs md:text-sm font-medium hidden md:inline">{connectionStatus}</span>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* 主要内容区域 */}
      <main className="container mx-auto px-4 py-3 md:py-6">
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-3 md:gap-6">
          {/* 左侧宠物面板 - 手机端更紧凑 */}
          <div className="lg:col-span-5 xl:col-span-4 space-y-3 md:space-y-6 animate-slide-in-left">
            {pets.length === 0 ? (
              <div className="animate-slide-in-up">
                <CreatePetForm onCreatePet={handleCreatePet} />
              </div>
            ) : (
              <>
                {/* 当前宠物卡片 */}
                <div className="relative animate-slide-in-up">
                  <div className="absolute -inset-1 bg-gradient-to-r from-terminal-accent to-blue-400 rounded-lg blur opacity-25 animate-glow"></div>
                  <div className="relative card-hover">
                    <PetCard 
                      pet={selectedPet} 
                      onStartExploration={handleStartExploration}
                    />
                  </div>
                </div>
                
                {/* 宠物列表 - 手机端压缩高度 */}
                {pets.length > 1 && (
                  <div className="bg-black border border-terminal-text rounded-lg p-3 md:p-4 shadow-lg animate-slide-in-up card-hover" style={{animationDelay: '0.2s'}}>
                    <h3 className="font-bold mb-2 md:mb-3 text-sm md:text-base text-terminal-accent flex items-center gap-2 text-glow">
                      <span className="animate-float">🏠</span> 
                      <span className="hidden md:inline">宠物仓库</span>
                      <span className="md:hidden">仓库</span>
                    </h3>
                    <div className="grid grid-cols-1 gap-1 md:gap-2 max-h-32 md:max-h-64 overflow-y-auto terminal-scroll">
                      {pets.map((pet, index) => (
                        <button
                          key={pet.id}
                          onClick={() => setSelectedPet(pet)}
                          className={`w-full text-left p-2 md:p-3 rounded-lg transition-all duration-200 flex items-center justify-between animate-fade-in ${
                            selectedPet?.id === pet.id 
                              ? 'bg-terminal-text text-black shadow-lg transform scale-105 animate-glow' 
                              : 'hover:bg-gray-800 hover:border-terminal-accent border border-transparent hover:scale-102'
                          }`}
                          style={{animationDelay: `${index * 0.1}s`}}
                        >
                          <div>
                            <div className="font-medium text-sm md:text-base">{pet.name}</div>
                            <div className="text-xs opacity-75">Lv.{pet.level} • {pet.location}</div>
                          </div>
                          <div className="text-right text-xs">
                            <div className={selectedPet?.id === pet.id ? 'text-black' : 'text-terminal-accent'}>
                              💰 {pet.coins.toLocaleString()}
                            </div>
                          </div>
                        </button>
                      ))}
                    </div>
                  </div>
                )}
                
                {/* 创建新宠物按钮 */}
                <div className="animate-slide-in-up" style={{animationDelay: '0.4s'}}>
                  <CreatePetForm onCreatePet={handleCreatePet} />
                </div>
              </>
            )}
          </div>

          {/* 右侧终端日志 - 手机端优先显示 */}
          <div className="lg:col-span-7 xl:col-span-8 animate-slide-in-right">
            <div className="relative h-full particle-bg">
              <div className="absolute -inset-1 bg-gradient-to-r from-green-400 to-terminal-accent rounded-lg blur opacity-20 animate-glow"></div>
              <div className="relative h-full terminal-enhanced">
                <Terminal events={events} title="实时事件日志" />
              </div>
              {/* 数据流效果 */}
              <div className="data-stream" style={{animationDelay: '0s'}}></div>
              <div className="data-stream" style={{animationDelay: '2s', top: '40%'}}></div>
              <div className="data-stream" style={{animationDelay: '4s', top: '60%'}}></div>
            </div>
          </div>
        </div>
      </main>

      {/* 底部状态栏 - 手机端简化 */}
      <footer className="border-t border-terminal-text bg-black bg-opacity-50 mt-4 md:mt-8">
        <div className="container mx-auto px-4 py-3 md:py-4">
          <div className="flex flex-col sm:flex-row items-center justify-between gap-2 md:gap-4">
            <div className="text-center sm:text-left">
              <p className="text-gray-400 text-xs md:text-sm">
                <span className="hidden md:inline">灵感来自早期比特币挖矿 - 界面简陋，但内核强大</span>
                <span className="md:hidden">AI宠物挖矿世界</span>
              </p>
            </div>
            <div className="flex items-center gap-3 md:gap-4 text-xs md:text-sm">
              <span className="text-gray-400">
                <span className="hidden md:inline">在线宠物: </span>
                <span className="md:hidden">宠物: </span>
                <span className="text-terminal-accent font-bold">{pets.length}</span>
              </span>
              <span className="text-gray-400 hidden md:inline">
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