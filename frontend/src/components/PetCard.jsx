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
    <div className="bg-black border border-terminal-text rounded-lg p-6">
      <div className="flex items-center justify-between mb-4">
        <div>
          <h2 className="text-xl font-bold text-terminal-accent flex items-center gap-2">
            {getPersonalityIcon(pet.personality)} {pet.name}
          </h2>
          <p className="text-sm text-gray-400">
            主人: {pet.owner} | 性格: {getPersonalityText(pet.personality)}
          </p>
        </div>
        <div className="text-right">
          <div className="text-lg font-bold">Lv.{pet.level}</div>
          <div className={`text-sm ${getStatusColor(pet.status)}`}>
            <Activity className="inline w-4 h-4 mr-1" />
            {pet.status}
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4 mb-4">
        <div className="flex items-center gap-2">
          <Heart className="w-4 h-4 text-red-400" />
          <div className="flex-1">
            <div className="flex justify-between text-sm">
              <span>生命值</span>
              <span>{pet.health}/{pet.max_health}</span>
            </div>
            <div className="w-full bg-gray-700 rounded-full h-2">
              <div 
                className="bg-red-400 h-2 rounded-full transition-all duration-300"
                style={{ width: `${healthPercentage}%` }}
              ></div>
            </div>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Coins className="w-4 h-4 text-yellow-400" />
          <span className="text-yellow-400">{pet.coins}</span>
        </div>

        <div className="flex items-center gap-2">
          <Zap className="w-4 h-4 text-blue-400" />
          <span>攻击: {pet.attack}</span>
        </div>

        <div className="flex items-center gap-2">
          <Shield className="w-4 h-4 text-green-400" />
          <span>防御: {pet.defense}</span>
        </div>
      </div>

      <div className="mb-4 flex items-center gap-2">
        <MapPin className="w-4 h-4 text-purple-400" />
        <span className="text-purple-400">{pet.location}</span>
      </div>

      <div className="mb-4">
        <div className="flex justify-between text-sm mb-1">
          <span>经验值</span>
          <span>{pet.experience}/{pet.level * 100}</span>
        </div>
        <div className="w-full bg-gray-700 rounded-full h-2">
          <div 
            className="bg-blue-400 h-2 rounded-full transition-all duration-300"
            style={{ width: `${(pet.experience / (pet.level * 100)) * 100}%` }}
          ></div>
        </div>
      </div>

      {pet.status === '等待中' && (
        <button
          onClick={() => onStartExploration(pet.id)}
          className="w-full bg-terminal-text text-black py-2 px-4 rounded font-bold hover:bg-terminal-accent transition-colors"
        >
          开始探索
        </button>
      )}
    </div>
  );
};

export default PetCard;