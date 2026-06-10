import { useState } from 'react';

export default function CarpoolCard({ carpool, onJoin }) {
  const [showJoinModal, setShowJoinModal] = useState(false);
  const [playerName, setPlayerName] = useState('');
  const [playerContact, setPlayerContact] = useState('');
  const [isJoining, setIsJoining] = useState(false);

  const missingPlayers = carpool.required_players - carpool.current_players;
  const isFull = carpool.status === 'full' || missingPlayers <= 0;
  const progress = (carpool.current_players / carpool.required_players) * 100;

  const handleJoin = async () => {
    if (!playerName.trim()) return;
    setIsJoining(true);
    try {
      await onJoin(carpool.id, playerName.trim(), playerContact.trim());
      setShowJoinModal(false);
      setPlayerName('');
      setPlayerContact('');
    } catch (error) {
      alert(error.message);
    } finally {
      setIsJoining(false);
    }
  };

  const formatTime = (dateStr) => {
    const date = new Date(dateStr);
    const now = new Date();
    const diff = now - date;
    const minutes = Math.floor(diff / 60000);
    if (minutes < 1) return '刚刚';
    if (minutes < 60) return `${minutes}分钟前`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours}小时前`;
    return `${Math.floor(hours / 24)}天前`;
  };

  const getTypeColor = (type) => {
    if (type.includes('情感')) return 'bg-pink-500/20 text-pink-300 border-pink-500/30';
    if (type.includes('硬核')) return 'bg-red-500/20 text-red-300 border-red-500/30';
    if (type.includes('惊悚') || type.includes('恐怖')) return 'bg-purple-500/20 text-purple-300 border-purple-500/30';
    if (type.includes('日式')) return 'bg-amber-500/20 text-amber-300 border-amber-500/30';
    return 'bg-blue-500/20 text-blue-300 border-blue-500/30';
  };

  const getDifficultyColor = (difficulty) => {
    switch (difficulty) {
      case '简单': return 'bg-green-500/20 text-green-300';
      case '中等': return 'bg-yellow-500/20 text-yellow-300';
      case '困难': return 'bg-red-500/20 text-red-300';
      default: return 'bg-gray-500/20 text-gray-300';
    }
  };

  return (
    <>
      <div className={`relative bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-sm rounded-2xl border transition-all duration-300 hover:scale-[1.02] hover:shadow-2xl overflow-hidden ${
        isFull 
          ? 'border-emerald-500/40 shadow-emerald-500/10' 
          : 'border-slate-600/50 shadow-xl hover:border-indigo-500/50 hover:shadow-indigo-500/20'
      }`}>
        {!isFull && (
          <div className="absolute top-0 right-0 w-32 h-32 bg-gradient-to-bl from-indigo-500/20 to-transparent rounded-bl-full" />
        )}
        {isFull && (
          <div className="absolute top-0 right-0 w-32 h-32 bg-gradient-to-bl from-emerald-500/20 to-transparent rounded-bl-full" />
        )}

        <div className="p-6 relative z-10">
          <div className="flex items-start justify-between mb-4">
            <div className="flex-1">
              <h3 className="text-2xl font-bold text-white mb-2">
                {carpool.script?.name || '未知剧本'}
              </h3>
              <div className="flex flex-wrap gap-2 mb-3">
                <span className={`px-3 py-1 rounded-full text-xs font-medium border ${getTypeColor(carpool.script?.type || '')}`}>
                  {carpool.script?.type}
                </span>
                <span className={`px-3 py-1 rounded-full text-xs font-medium ${getDifficultyColor(carpool.script?.difficulty)}`}>
                  {carpool.script?.difficulty}
                </span>
                <span className="px-3 py-1 rounded-full text-xs font-medium bg-slate-700/50 text-slate-300">
                  {carpool.script?.duration && `${Math.floor(carpool.script.duration / 60)}小时${carpool.script.duration % 60 > 0 ? `${carpool.script.duration % 60}分钟` : ''}`}
                </span>
              </div>
            </div>

            <div className={`px-4 py-2 rounded-xl font-bold text-sm ${
              isFull 
                ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30' 
                : 'bg-amber-500/20 text-amber-300 border border-amber-500/30 animate-pulse'
            }`}>
              {isFull ? '已满员' : '招募中'}
            </div>
          </div>

          <div className="mb-6">
            <div className="flex items-center justify-between mb-2">
              <span className="text-slate-400 text-sm">拼车进度</span>
              <span className={`font-bold text-lg ${
                isFull ? 'text-emerald-400' : 'text-indigo-400'
              }`}>
                {carpool.current_players}/{carpool.required_players} 人
              </span>
            </div>
            <div className="h-3 bg-slate-700/50 rounded-full overflow-hidden">
              <div 
                className={`h-full rounded-full transition-all duration-500 ${
                  isFull 
                    ? 'bg-gradient-to-r from-emerald-500 to-emerald-400' 
                    : 'bg-gradient-to-r from-indigo-500 to-purple-500'
                }`}
                style={{ width: `${progress}%` }}
              />
            </div>
            {!isFull && (
              <p className="text-amber-400 text-sm mt-2 font-medium">
                还差 {missingPlayers} 人即可发车！
              </p>
            )}
          </div>

          <div className="mb-4">
            <p className="text-slate-400 text-sm mb-2">已加入玩家</p>
            <div className="flex flex-wrap gap-2">
              {carpool.players?.map((player, index) => (
                <div 
                  key={index}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm ${
                    player.is_host 
                      ? 'bg-indigo-500/20 text-indigo-300 border border-indigo-500/30' 
                      : 'bg-slate-700/50 text-slate-300'
                  }`}
                >
                  {player.is_host && (
                    <span className="text-yellow-400">👑</span>
                  )}
                  <span>{player.name}</span>
                </div>
              ))}
            </div>
          </div>

          <div className="flex items-center justify-between pt-4 border-t border-slate-700/50">
            <div className="text-sm text-slate-500">
              发起人：<span className="text-slate-300">{carpool.host_name}</span>
              <span className="mx-2">·</span>
              {formatTime(carpool.created_at)}
            </div>

            <button
              onClick={() => setShowJoinModal(true)}
              disabled={isFull}
              className={`px-6 py-2.5 rounded-xl font-semibold transition-all duration-300 ${
                isFull
                  ? 'bg-slate-700/50 text-slate-500 cursor-not-allowed'
                  : 'bg-gradient-to-r from-indigo-600 to-purple-600 text-white hover:from-indigo-500 hover:to-purple-500 hover:shadow-lg hover:shadow-indigo-500/30 active:scale-95'
              }`}
            >
              {isFull ? '已满员' : '加入拼车'}
            </button>
          </div>
        </div>
      </div>

      {showJoinModal && (
        <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="bg-gradient-to-br from-slate-800 to-slate-900 rounded-2xl border border-slate-600/50 w-full max-w-md p-6 animate-float">
            <h3 className="text-2xl font-bold text-white mb-2">
              加入《{carpool.script?.name}》拼车
            </h3>
            <p className="text-slate-400 mb-6">
              当前还差 {missingPlayers} 人，填写信息即可加入
            </p>

            <div className="space-y-4 mb-6">
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  你的昵称 <span className="text-red-400">*</span>
                </label>
                <input
                  type="text"
                  value={playerName}
                  onChange={(e) => setPlayerName(e.target.value)}
                  placeholder="请输入你的昵称"
                  className="w-full px-4 py-3 bg-slate-700/50 border border-slate-600/50 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 transition-all"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-300 mb-2">
                  联系方式
                </label>
                <input
                  type="text"
                  value={playerContact}
                  onChange={(e) => setPlayerContact(e.target.value)}
                  placeholder="手机号或微信号（选填）"
                  className="w-full px-4 py-3 bg-slate-700/50 border border-slate-600/50 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 transition-all"
                />
              </div>
            </div>

            <div className="flex gap-3">
              <button
                onClick={() => setShowJoinModal(false)}
                className="flex-1 px-6 py-3 bg-slate-700/50 text-slate-300 rounded-xl font-semibold hover:bg-slate-700 transition-all active:scale-95"
              >
                取消
              </button>
              <button
                onClick={handleJoin}
                disabled={!playerName.trim() || isJoining}
                className="flex-1 px-6 py-3 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-xl font-semibold hover:from-indigo-500 hover:to-purple-500 transition-all active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isJoining ? '加入中...' : '确认加入'}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
