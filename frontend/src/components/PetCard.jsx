import React from 'react';
import { Heart, Zap, Shield, Coins, MapPin, Activity } from 'lucide-react';

const PetCard = ({ pet, onStartExploration }) => {
  if (!pet) {
    return (
      <div className="bg-black border border-terminal-text rounded-lg p-6">
        <div className="text-center text-gray-500">
          没有宠物数据
        </div>
      </div>
    );
  }

  const getPersonalityIcon = (personality) => {
    switch (personality) {
      case 'brave': return '⚔️';
      case 'greedy': return '💰';
      case 'friendly': return '🤝';
      case 'cautious': return '🛡️';
      case 'curious': return '🔍';
      default: return '🐾';
    }
  };

  const getPersonalityText = (personality) => {
    switch (personality) {
      case 'brave': return '勇敢';
      case 'greedy': return '贪婪';
      case 'friendly': return '友好';
      case 'cautious': return '谨慎';
      case 'curious': return '好奇';
      default: return '未知';
    }
  };

  const getStatusColor = (status) => {
    switch (status) {
      case '探索中': return 'text-yellow-400';
      case '战斗中': return 'text-red-400';
      case '等待中': return 'text-green-400';
      default: return 'text-gray-400';
    }
  };

  const healthPercentage = (pet.health / pet.max_health) * 100;

  return (
    <div className="bg-black border border-terminal-text rounded-lg shadow-xl relative overflow-hidden 
                    p-3 md:p-4
                    lg:p-4">
      {/* 背景装饰 */}
      <div className="absolute top-0 right-0 w-32 h-32 bg-terminal-accent opacity-5 rounded-full -translate-y-16 translate-x-16"></div>
      
      {/* 宠物头部信息 - 紧凑模式 */}
      <div className="relative z-10 mb-1">
        <div className="flex items-center justify-between mb-1">
          <div className="flex items-center gap-2 md:gap-3">
            <div className="text-2xl md:text-3xl animate-bounce">{getPersonalityIcon(pet.personality)}</div>
            <div>
              <h2 className="text-lg md:text-xl font-bold text-terminal-accent">
                {pet.name}
              </h2>
              <p className="text-xs md:text-sm text-gray-400">
                主人: <span className="text-terminal-text font-medium">{pet.owner}</span>
              </p>
            </div>
          </div>
          <div className="text-right">
            <div className="text-xl md:text-2xl font-bold text-blue-400">Lv.{pet.level}</div>
          </div>
        </div>
        
        {/* 状态、性格和心情 - 手机端简化 */}
        <div className="flex items-center justify-between flex-wrap gap-1 md:gap-2">
          <div className="flex items-center gap-1 px-2 py-1 bg-gray-800 rounded-full">
            <span className="text-xs text-gray-400 hidden md:inline">性格:</span>
            <span className="text-xs font-medium text-terminal-accent">{getPersonalityText(pet.personality)}</span>
          </div>
          
          {pet.mood && (
            <div className="flex items-center gap-1 px-2 py-1 bg-purple-900 bg-opacity-50 rounded-full">
              <span className="text-xs">😊</span>
              <span className="text-xs font-medium text-purple-400">{pet.mood}</span>
            </div>
          )}
          
          <div className={`flex items-center gap-1 px-2 py-1 rounded-full ${
            pet.status === '探索中' ? 'bg-yellow-900 bg-opacity-50' :
            pet.status === '战斗中' ? 'bg-red-900 bg-opacity-50' :
            'bg-green-900 bg-opacity-50'
          }`}>
            <Activity className="w-3 h-3" />
            <span className={`text-xs font-medium ${getStatusColor(pet.status)}`}>
              {pet.status}
            </span>
          </div>
        </div>
      </div>

      {/* 状态条显示 - 紧凑显示 */}
      <div className="mb-1 space-y-1">
        {/* 生命值条 */}
        <div>
          <div className="flex justify-between text-xs mb-0.5">
            <div className="flex items-center gap-1">
              <Heart className="w-3 h-3 text-red-400" />
              <span>生命值</span>
            </div>
            <span className="font-mono text-xs">{pet.health}/{pet.max_health}</span>
          </div>
          <div className="w-full bg-gray-700 rounded-full h-1.5 md:h-2 overflow-hidden">
            <div 
              className={`h-1.5 md:h-2 rounded-full transition-all duration-500 ${
                healthPercentage > 70 ? 'bg-gradient-to-r from-green-400 to-green-500' :
                healthPercentage > 30 ? 'bg-gradient-to-r from-yellow-400 to-orange-500' :
                'bg-gradient-to-r from-red-400 to-red-600'
              }`}
              style={{ width: `${healthPercentage}%` }}
            ></div>
          </div>
        </div>

        {/* 体力值条 */}
        <div>
          <div className="flex justify-between text-xs mb-0.5">
            <div className="flex items-center gap-1">
              <Zap className="w-3 h-3 text-yellow-400" />
              <span>体力</span>
            </div>
            <span className="font-mono text-xs">{pet.energy || 100}/{pet.max_energy || 100}</span>
          </div>
          <div className="w-full bg-gray-700 rounded-full h-1.5 md:h-2 overflow-hidden">
            <div 
              className="bg-gradient-to-r from-yellow-400 to-orange-500 h-1.5 md:h-2 rounded-full transition-all duration-500"
              style={{ width: `${((pet.energy || 100) / (pet.max_energy || 100)) * 100}%` }}
            ></div>
          </div>
        </div>

        {/* 饱食度条 */}
        <div>
          <div className="flex justify-between text-xs mb-0.5">
            <div className="flex items-center gap-1">
              <span className="text-xs">🍖</span>
              <span>饱食度</span>
            </div>
            <span className="font-mono text-xs">{pet.hunger || 80}/100</span>
          </div>
          <div className="w-full bg-gray-700 rounded-full h-1.5 md:h-2 overflow-hidden">
            <div 
              className="bg-gradient-to-r from-green-400 to-green-600 h-1.5 md:h-2 rounded-full transition-all duration-500"
              style={{ width: `${(pet.hunger || 80)}%` }}
            ></div>
          </div>
        </div>

        {/* 经验值条 */}
        <div>
          <div className="flex justify-between text-xs mb-0.5">
            <span>经验值</span>
            <span className="font-mono text-xs">{pet.experience}/{pet.level * 100}</span>
          </div>
          <div className="w-full bg-gray-700 rounded-full h-1 md:h-1.5 overflow-hidden">
            <div 
              className="bg-gradient-to-r from-blue-400 to-purple-500 h-1 md:h-1.5 rounded-full transition-all duration-500"
              style={{ width: `${(pet.experience / (pet.level * 100)) * 100}%` }}
            ></div>
          </div>
        </div>
      </div>

      {/* 属性一行显示 - 攻击力、防御力、金币 */}
      <div className="grid grid-cols-3 gap-2 md:gap-3 mb-1">
        <div className="bg-gray-900 rounded-lg p-1.5 md:p-2 border border-gray-700 hover:border-blue-400 transition-colors">
          <div className="flex items-center gap-1 md:gap-2 mb-0.5">
            <Zap className="w-3 md:w-4 h-3 md:h-4 text-blue-400" />
            <span className="text-xs text-gray-400 hidden md:inline">攻击力</span>
            <span className="text-xs text-gray-400 md:hidden">攻击</span>
          </div>
          <div className="text-sm md:text-lg font-bold text-blue-400">{pet.attack}</div>
        </div>

        <div className="bg-gray-900 rounded-lg p-1.5 md:p-2 border border-gray-700 hover:border-green-400 transition-colors">
          <div className="flex items-center gap-1 md:gap-2 mb-0.5">
            <Shield className="w-3 md:w-4 h-3 md:h-4 text-green-400" />
            <span className="text-xs text-gray-400 hidden md:inline">防御力</span>
            <span className="text-xs text-gray-400 md:hidden">防御</span>
          </div>
          <div className="text-sm md:text-lg font-bold text-green-400">{pet.defense}</div>
        </div>

        <div className="bg-gray-900 rounded-lg p-1.5 md:p-2 border border-gray-700 hover:border-yellow-400 transition-colors">
          <div className="flex items-center gap-1 md:gap-2 mb-0.5">
            <Coins className="w-3 md:w-4 h-3 md:h-4 text-yellow-400" />
            <span className="text-xs text-gray-400 hidden md:inline">金币</span>
            <span className="text-xs text-gray-400 md:hidden">金币</span>
          </div>
          <div className="text-sm md:text-lg font-bold text-yellow-400 font-mono">{pet.coins.toLocaleString()}</div>
        </div>
      </div>

      {/* 位置信息 - 移动端简化 */}
      <div className="mb-1 p-2 bg-purple-900 bg-opacity-30 rounded-lg border border-purple-500 border-opacity-30">
        <div className="flex items-center gap-2">
          <MapPin className="w-3 md:w-4 h-3 md:h-4 text-purple-400" />
          <span className="text-xs md:text-sm text-gray-400 hidden md:inline">当前位置:</span>
          <span className="text-xs md:text-sm text-purple-400 font-medium">{pet.location}</span>
        </div>
      </div>

      {/* 行动按钮 - 紧凑高度 */}
      {pet.status === '等待中' && (
        <button
          onClick={() => onStartExploration(pet.id)}
          className="w-full bg-gradient-to-r from-terminal-text to-terminal-accent text-black 
                     py-2 md:py-3 px-3 md:px-4 rounded-lg font-bold 
                     hover:from-terminal-accent hover:to-terminal-text 
                     transition-all duration-300 transform hover:scale-105 hover:shadow-lg 
                     relative overflow-hidden group text-sm md:text-base"
        >
          <span className="relative z-10">🚀 开始探索</span>
          <div className="absolute inset-0 bg-white opacity-0 group-hover:opacity-20 transition-opacity"></div>
        </button>
      )}
      
      {pet.status !== '等待中' && (
        <div className="w-full py-2 md:py-3 px-3 md:px-4 rounded-lg bg-gray-800 text-center border-2 border-dashed border-gray-600">
          <span className="text-gray-400 text-sm md:text-base">🎮 {pet.name} 正在 {pet.status}...</span>
        </div>
      )}
    </div>
  );
};

export default PetCard;