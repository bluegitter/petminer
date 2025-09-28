import React, { useState, useEffect } from 'react';
import { petAPI } from './services/api';
import { useWebSocket } from './hooks/useWebSocket';
import Terminal from './components/Terminal';
import PetCard from './components/PetCard';
import CreatePetForm from './components/CreatePetForm';
import { Wifi, WifiOff, RefreshCw } from 'lucide-react';
import './index.css';

function App() {
  const [pets, setPets] = useState([]);
  const [selectedPet, setSelectedPet] = useState(null);
  const [loading, setLoading] = useState(false);
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
    <div className="min-h-screen bg-terminal-bg text-terminal-text p-4">
      <header className="mb-6">
        <div className="flex items-center justify-between">
          <h1 className="text-3xl font-bold text-terminal-accent">
            🐾 MiningPet
          </h1>
          <div className="flex items-center gap-2 text-sm">
            {getConnectionStatusIcon()}
            <span>{connectionStatus}</span>
          </div>
        </div>
        <p className="text-gray-400 mt-2">
          命令行挂机游戏 - 让你的AI宠物探索虚拟世界
        </p>
      </header>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-1 space-y-6">
          {pets.length === 0 ? (
            <CreatePetForm onCreatePet={handleCreatePet} />
          ) : (
            <>
              <PetCard 
                pet={selectedPet} 
                onStartExploration={handleStartExploration}
              />
              
              {pets.length > 1 && (
                <div className="bg-black border border-terminal-text rounded-lg p-4">
                  <h3 className="font-bold mb-2">所有宠物</h3>
                  <div className="space-y-2">
                    {pets.map(pet => (
                      <button
                        key={pet.id}
                        onClick={() => setSelectedPet(pet)}
                        className={`w-full text-left p-2 rounded transition-colors ${
                          selectedPet?.id === pet.id 
                            ? 'bg-terminal-text text-black' 
                            : 'hover:bg-gray-800'
                        }`}
                      >
                        {pet.name} (Lv.{pet.level})
                      </button>
                    ))}
                  </div>
                </div>
              )}
              
              <CreatePetForm onCreatePet={handleCreatePet} />
            </>
          )}
        </div>

        <div className="lg:col-span-2">
          <Terminal events={events} title="实时事件日志" />
        </div>
      </div>

      <footer className="mt-8 text-center text-gray-500 text-sm">
        <p>灵感来自早期比特币挖矿 - 界面简陋，但内核强大</p>
        <div className="mt-2">
          活跃事件: {events.length} | 连接状态: {connectionStatus}
        </div>
      </footer>
    </div>
  );
}

export default App;