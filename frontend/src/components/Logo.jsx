import React from 'react';

const Logo = ({ className = "w-10 h-10", showText = false }) => {
  return (
    <div className={`flex items-center gap-2 ${showText ? '' : ''}`}>
      <div className={`${className} bg-gradient-to-br from-yellow-400 to-orange-500 rounded-lg flex items-center justify-center shadow-lg`}>
        <span className="text-black font-bold text-lg">🐾</span>
      </div>
      {showText && (
        <span className="text-terminal-accent font-bold text-xl">MiningPet</span>
      )}
    </div>
  );
};

export default Logo;