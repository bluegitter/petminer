import React, { useState } from 'react';
import { User, Plus } from 'lucide-react';
import Logo from './Logo';

const CreatePetForm = ({ onCreatePet }) => {
  const [ownerName, setOwnerName] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [logoError, setLogoError] = useState(false);

  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!ownerName.trim()) return;

    setIsLoading(true);
    setError('');
    try {
      await onCreatePet(ownerName.trim());
      setOwnerName('');
    } catch (error) {
      console.error('创建宠物失败:', error);
      // 显示服务器返回的错误信息
      const errorMessage = error.response?.data?.error || '创建宠物失败，请稍后再试';
      setError(errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="bg-gradient-to-br from-gray-900 to-black border border-terminal-text rounded-xl p-4 md:p-6 shadow-xl relative overflow-hidden">
      {/* 背景装饰 */}
      <div className="absolute -top-4 -right-4 w-16 h-16 bg-terminal-accent opacity-10 rounded-full animate-pulse"></div>
      <div className="absolute -bottom-2 -left-2 w-12 h-12 bg-blue-400 opacity-10 rounded-full animate-pulse" style={{animationDelay: '1s'}}></div>
      
      <div className="relative z-10">
        <h2 className="text-lg md:text-xl font-bold text-terminal-accent mb-4 md:mb-6 flex items-center gap-2 md:gap-3">
          <div className="flex items-center gap-1 md:gap-2 bg-terminal-accent bg-opacity-20 rounded-lg p-1.5 md:p-2">
            {!logoError ? (
              <img 
                src="/logo.png" 
                alt="Pet Icon" 
                className="w-5 md:w-6 h-5 md:h-6 rounded"
                onError={() => setLogoError(true)}
              />
            ) : (
              <Logo className="w-5 md:w-6 h-5 md:h-6" />
            )}
            <Plus className="w-4 md:w-5 h-4 md:h-5" />
          </div>
          <span className="text-sm md:text-base">创建新宠物</span>
        </h2>
        
        <form onSubmit={handleSubmit} className="space-y-4 md:space-y-6">
          <div className="space-y-2 md:space-y-3">
            <label className="block text-xs md:text-sm font-medium text-gray-300 flex items-center gap-1 md:gap-2">
              <User className="w-3 md:w-4 h-3 md:h-4 text-terminal-accent" />
              <span className="hidden md:inline">训练师名称</span>
              <span className="md:hidden">名称</span>
            </label>
            <div className="relative">
              <input
                type="text"
                value={ownerName}
                onChange={(e) => setOwnerName(e.target.value)}
                placeholder="输入训练师名称..."
                className="w-full bg-gray-800 border-2 border-gray-600 rounded-lg px-3 md:px-4 py-2 md:py-3 text-sm md:text-base text-terminal-text placeholder-gray-500 focus:outline-none focus:border-terminal-accent focus:ring-2 focus:ring-terminal-accent focus:ring-opacity-50 transition-all duration-300"
                disabled={isLoading}
                maxLength={20}
              />
              <div className="absolute right-2 md:right-3 top-1/2 transform -translate-y-1/2 text-xs text-gray-500">
                {ownerName.length}/20
              </div>
            </div>
          </div>
          
          <button
            type="submit"
            disabled={!ownerName.trim() || isLoading}
            className="w-full bg-gradient-to-r from-terminal-text to-terminal-accent text-black py-2 md:py-3 px-4 md:px-6 rounded-lg font-bold hover:from-terminal-accent hover:to-terminal-text transition-all duration-300 transform hover:scale-105 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none shadow-lg relative overflow-hidden group text-sm md:text-base"
          >
            <span className="relative z-10 flex items-center justify-center gap-2">
              {isLoading ? (
                <>
                  <div className="w-3 md:w-4 h-3 md:h-4 border-2 border-black border-t-transparent rounded-full animate-spin"></div>
                  <span className="hidden md:inline">创建中...</span>
                  <span className="md:hidden">创建中</span>
                </>
              ) : (
                <>
                  ✨ <span className="hidden md:inline">召唤宠物</span>
                  <span className="md:hidden">召唤</span>
                </>
              )}
            </span>
            <div className="absolute inset-0 bg-white opacity-0 group-hover:opacity-20 transition-opacity"></div>
          </button>

          {/* 错误提示 */}
          {error && (
            <div className="mt-3 p-3 bg-red-900 bg-opacity-50 border border-red-500 border-opacity-50 rounded-lg">
              <div className="flex items-start gap-2">
                <span className="text-red-400 text-sm">⚠️</span>
                <div className="text-red-300 text-xs md:text-sm">
                  <div className="font-medium mb-1">创建失败</div>
                  <p>{error}</p>
                </div>
              </div>
            </div>
          )}
        </form>
        
        <div className="mt-4 md:mt-6 p-3 md:p-4 bg-blue-900 bg-opacity-30 rounded-lg border border-blue-500 border-opacity-30">
          <div className="flex items-start gap-2 md:gap-3">
            <span className="text-lg md:text-2xl">🎲</span>
            <div className="text-xs md:text-sm text-gray-300">
              <div className="font-medium text-blue-400 mb-1 text-xs md:text-sm">随机生成特性</div>
              <ul className="space-y-1 text-xs">
                <li>• <span className="hidden md:inline">宠物会获得随机名字和独特性格</span><span className="md:hidden">随机名字和性格</span></li>
                <li>• <span className="hidden md:inline">初始属性根据性格特征调整</span><span className="md:hidden">属性根据性格调整</span></li>
                <li>• <span className="hidden md:inline">每只宠物都有独特的探索风格</span><span className="md:hidden">独特探索风格</span></li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CreatePetForm;