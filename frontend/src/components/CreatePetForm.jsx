import React, { useState } from 'react';
import { User, Plus } from 'lucide-react';

const CreatePetForm = ({ onCreatePet }) => {
  const [ownerName, setOwnerName] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!ownerName.trim()) return;

    setIsLoading(true);
    try {
      await onCreatePet(ownerName.trim());
      setOwnerName('');
    } catch (error) {
      console.error('创建宠物失败:', error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="bg-black border border-terminal-text rounded-lg p-6">
      <h2 className="text-xl font-bold text-terminal-accent mb-4 flex items-center gap-2">
        <Plus className="w-5 h-5" />
        创建新宠物
      </h2>
      
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium mb-2 flex items-center gap-2">
            <User className="w-4 h-4" />
            主人姓名
          </label>
          <input
            type="text"
            value={ownerName}
            onChange={(e) => setOwnerName(e.target.value)}
            placeholder="输入你的名字..."
            className="w-full bg-gray-900 border border-gray-600 rounded px-3 py-2 text-terminal-text focus:outline-none focus:border-terminal-accent"
            disabled={isLoading}
          />
        </div>
        
        <button
          type="submit"
          disabled={!ownerName.trim() || isLoading}
          className="w-full bg-terminal-text text-black py-2 px-4 rounded font-bold hover:bg-terminal-accent transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isLoading ? '创建中...' : '创建宠物'}
        </button>
      </form>
      
      <div className="mt-4 text-sm text-gray-400">
        💡 宠物会自动获得随机名字和性格特征
      </div>
    </div>
  );
};

export default CreatePetForm;