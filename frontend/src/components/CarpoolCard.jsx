import { useState } from 'react';
import { api } from '../services/api';

export default function CarpoolCard({ carpool, currentUser, onRefresh }) {
  const [showJoinModal, setShowJoinModal] = useState(false);
  const [showWaitlistModal, setShowWaitlistModal] = useState(false);
  const [showLeaveConfirm, setShowLeaveConfirm] = useState(false);
  const [showWaitlistList, setShowWaitlistList] = useState(false);
  const [playerName, setPlayerName] = useState(currentUser || '');
  const [playerContact, setPlayerContact] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const missingPlayers = carpool.required_players - carpool.current_players;
  const isFull = carpool.status === 'full' || missingPlayers <= 0;
  const isCancelled = carpool.status === 'cancelled';
  const progress = (carpool.current_players / carpool.required_players) * 100;
  const waitlistCount = carpool.waitlist?.length || 0;

  const userInCar = carpool.players?.some(p => p.name === playerName && playerName);
  const userIsHost = carpool.host_name === playerName && playerName;
  const userInWaitlist = carpool.waitlist?.some(w => w.name === playerName && playerName);

  const handleJoin = async () => {
    if (!playerName.trim()) return;
    setIsLoading(true);
    try {
      await api.joinCarpool({
        carpool_id: carpool.id,
        name: playerName.trim(),
        contact: playerContact.trim(),
      });
      setShowJoinModal(false);
      setPlayerContact('');
      onRefresh?.();
    } catch (error) {
      alert(error.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleWaitlist = async () => {
    if (!playerName.trim()) return;
    setIsLoading(true);
    try {
      const result = await api.joinWaitlist({
        carpool_id: carpool.id,
        name: playerName.trim(),
        contact: playerContact.trim(),
      });
      alert(result.message);
      setShowWaitlistModal(false);
      setPlayerContact('');
      onRefresh?.();
    } catch (error) {
      alert(error.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleLeave = async () => {
    if (!playerName.trim()) return;
    setIsLoading(true);
    try {
      const result = await api.leaveCarpool(carpool.id, playerName.trim());
      let msg = '退车成功！';
      if (result.promoted) {
        msg += `\n候补玩家「${result.promoted.name}」已自动补位加入。`;
      }
      alert(msg);
      setShowLeaveConfirm(false);
      onRefresh?.();
    } catch (error) {
      alert(error.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleCancel = async () => {
    if (!window.confirm(`确定要取消《${carpool.script?.name}》拼车吗？所有玩家押金将自动退回。`)) return;
    setIsLoading(true);
    try {
      await api.cancelCarpool(carpool.id, playerName.trim());
      alert('拼车已取消，押金已自动退回');
      onRefresh?.();
    } catch (error) {
      alert(error.message);
    } finally {
      setIsLoading(false);
    }
  };

  const formatTime = (dateStr) => {
    if (!dateStr) return '';
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

  const formatStartTime = (startTime) => {
    if (!startTime) return null;
    const d = new Date(startTime);
    const now = new Date();
    const diff = d - now;
    const hours = Math.floor(diff / 3600000);
    const mins = Math.floor((diff % 3600000) / 60000);
    let info = `${d.getMonth() + 1}/${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
    if (diff > 0) {
      info += `（还有 ${hours > 0 ? `${hours}小时` : ''}${mins}分钟开局）`;
    } else if (Math.abs(diff) < 3600000) {
      info += '（即将开局）';
    }
    return info;
  };

  const getTypeColor = (type) => {
    if (!type) return 'bg-slate-500/20 text-slate-300 border-slate-500/30';
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

  const statusBadge = () => {
    if (isCancelled) {
      return <span className="px-4 py-2 rounded-xl font-bold text-sm bg-slate-700/50 text-slate-400 border border-slate-600/50">已取消</span>;
    }
    if (isFull) {
      return <span className="px-4 py-2 rounded-xl font-bold text-sm bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">已满员</span>;
    }
    return <span className="px-4 py-2 rounded-xl font-bold text-sm bg-amber-500/20 text-amber-300 border border-amber-500/30 animate-pulse">招募中</span>;
  };

  return (
    <>
      <div className={`relative bg-gradient-to-br from-slate-800/80 to-slate-900/80 backdrop-blur-sm rounded-2xl border transition-all duration-300 hover:scale-[1.01] hover:shadow-2xl overflow-hidden ${
        isCancelled
          ? 'border-slate-700/50 opacity-60'
          : isFull
          ? 'border-emerald-500/40 shadow-emerald-500/10'
          : 'border-slate-600/50 shadow-xl hover:border-indigo-500/50 hover:shadow-indigo-500/20'
      }`}>
        {!isCancelled && !isFull && (
          <div className="absolute top-0 right-0 w-32 h-32 bg-gradient-to-bl from-indigo-500/20 to-transparent rounded-bl-full" />
        )}
        {isFull && !isCancelled && (
          <div className="absolute top-0 right-0 w-32 h-32 bg-gradient-to-bl from-emerald-500/20 to-transparent rounded-bl-full" />
        )}

        <div className="p-6 relative z-10">
          <div className="flex items-start justify-between mb-4">
            <div className="flex-1">
              <div className="flex items-center gap-2 mb-1">
                <h3 className="text-2xl font-bold text-white">
                  {carpool.script?.name || '未知剧本'}
                </h3>
                {carpool.deposit_amount > 0 && (
                  <span className="px-2 py-0.5 rounded-md text-xs bg-yellow-500/20 text-yellow-300 border border-yellow-500/30">
                    押金 ¥{carpool.deposit_amount}
                  </span>
                )}
              </div>
              <div className="flex flex-wrap gap-2 mb-3">
                <span className={`px-3 py-1 rounded-full text-xs font-medium border ${getTypeColor(carpool.script?.type || '')}`}>
                  {carpool.script?.type}
                </span>
                <span className={`px-3 py-1 rounded-full text-xs font-medium ${getDifficultyColor(carpool.script?.difficulty)}`}>
                  {carpool.script?.difficulty}
                </span>
                <span className="px-3 py-1 rounded-full text-xs font-medium bg-slate-700/50 text-slate-300">
                  {carpool.script?.duration && `${Math.floor(carpool.script.duration / 60)}h${carpool.script.duration % 60 > 0 ? `${carpool.script.duration % 60}m` : ''}`}
                </span>
                {carpool.start_time && (
                  <span className="px-3 py-1 rounded-full text-xs font-medium bg-cyan-500/20 text-cyan-300 border border-cyan-500/30">
                    🕐 {formatStartTime(carpool.start_time)}
                  </span>
                )}
              </div>
            </div>
            {statusBadge()}
          </div>

          <div className="mb-6">
            <div className="flex items-center justify-between mb-2">
              <span className="text-slate-400 text-sm">拼车进度</span>
              <span className={`font-bold text-lg ${
                isCancelled ? 'text-slate-500' : isFull ? 'text-emerald-400' : 'text-indigo-400'
              }`}>
                {carpool.current_players}/{carpool.required_players} 人
              </span>
            </div>
            <div className="h-3 bg-slate-700/50 rounded-full overflow-hidden">
              <div 
                className={`h-full rounded-full transition-all duration-500 ${
                  isCancelled ? 'bg-slate-600'
                  : isFull ? 'bg-gradient-to-r from-emerald-500 to-emerald-400'
                  : 'bg-gradient-to-r from-indigo-500 to-purple-500'
                }`}
                style={{ width: `${Math.min(progress, 100)}%` }}
              />
            </div>
            {!isCancelled && !isFull && (
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
                  {player.is_host && <span className="text-yellow-400">👑</span>}
                  <span>{player.name}</span>
                  {player.deposit_paid && <span className="text-yellow-400 text-xs">💳</span>}
                </div>
              ))}
            </div>
          </div>

          {waitlistCount > 0 && (
            <div className="mb-4 p-3 bg-orange-500/10 border border-orange-500/30 rounded-xl">
              <div 
                className="flex items-center justify-between cursor-pointer"
                onClick={() => setShowWaitlistList(!showWaitlistList)}
              >
                <span className="text-sm font-medium text-orange-300">
                  🕐 候补队列 ({waitlistCount} 人)
                </span>
                <span className="text-orange-400 text-xs">{showWaitlistList ? '收起' : '展开'}</span>
              </div>
              {showWaitlistList && (
                <div className="mt-2 space-y-1">
                  {carpool.waitlist?.map((w, i) => (
                    <div key={w.id} className="flex items-center gap-2 text-sm text-slate-300 px-2 py-1 bg-slate-800/50 rounded">
                      <span className="w-6 h-6 flex items-center justify-center rounded-full bg-orange-500/30 text-orange-300 text-xs font-bold">
                        {i + 1}
                      </span>
                      <span>{w.name}</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          <div className="flex items-center justify-between pt-4 border-t border-slate-700/50">
            <div className="text-sm text-slate-500">
              发起人：<span className="text-slate-300">{carpool.host_name}</span>
              {carpool.created_at && (
                <>
                  <span className="mx-2">·</span>
                  {formatTime(carpool.created_at)}
                </>
              )}
            </div>

            <div className="flex gap-2">
              {userInCar && !userIsHost && !isCancelled && (
                <button
                  onClick={() => setShowLeaveConfirm(true)}
                  disabled={isLoading}
                  className="px-4 py-2.5 bg-red-500/20 text-red-300 border border-red-500/30 rounded-xl font-medium hover:bg-red-500/30 transition-all active:scale-95 disabled:opacity-50"
                >
                  退车
                </button>
              )}
              {userIsHost && !isCancelled && (
                <button
                  onClick={handleCancel}
                  disabled={isLoading}
                  className="px-4 py-2.5 bg-red-500/20 text-red-300 border border-red-500/30 rounded-xl font-medium hover:bg-red-500/30 transition-all active:scale-95 disabled:opacity-50"
                >
                  取消拼车
                </button>
              )}
              {!userInCar && !isCancelled && (
                <>
                  {isFull && (
                    <button
                      onClick={() => setShowWaitlistModal(true)}
                      disabled={isLoading || userInWaitlist}
                      className="px-4 py-2.5 bg-gradient-to-r from-orange-500 to-amber-500 text-white rounded-xl font-semibold hover:from-orange-400 hover:to-amber-400 transition-all shadow-lg shadow-orange-500/20 active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      {userInWaitlist ? '已在候补' : '候补排队'}
                    </button>
                  )}
                  {!isFull && (
                    <button
                      onClick={() => setShowJoinModal(true)}
                      disabled={isLoading}
                      className="px-6 py-2.5 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-xl font-semibold hover:from-indigo-500 hover:to-purple-500 hover:shadow-lg hover:shadow-indigo-500/30 transition-all active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      加入拼车
                    </button>
                  )}
                </>
              )}
            </div>
          </div>
        </div>
      </div>

      {showJoinModal && (
        <Modal onClose={() => setShowJoinModal(false)} title={`加入《${carpool.script?.name}》拼车`}>
          <p className="text-slate-400 mb-6">当前还差 {missingPlayers} 人，填写信息即可加入</p>
          <PlayerForm
            name={playerName}
            setName={setPlayerName}
            contact={playerContact}
            setContact={setPlayerContact}
          />
          <ModalActions
            onCancel={() => setShowJoinModal(false)}
            onConfirm={handleJoin}
            confirmText={isLoading ? '加入中...' : '确认加入'}
            disabled={!playerName.trim() || isLoading}
          />
        </Modal>
      )}

      {showWaitlistModal && (
        <Modal onClose={() => setShowWaitlistModal(false)} title={`候补《${carpool.script?.name}》`}>
          <p className="text-slate-400 mb-2">拼车已满员，加入候补队列</p>
          <p className="text-orange-400 text-sm mb-6">当前已有 {waitlistCount} 人在排队，一旦有人退车，将按候补顺序自动补位</p>
          <PlayerForm
            name={playerName}
            setName={setPlayerName}
            contact={playerContact}
            setContact={setPlayerContact}
          />
          <ModalActions
            onCancel={() => setShowWaitlistModal(false)}
            onConfirm={handleWaitlist}
            confirmText={isLoading ? '申请中...' : '确认候补'}
            confirmClass="from-orange-500 to-amber-500 hover:from-orange-400 hover:to-amber-400"
            disabled={!playerName.trim() || isLoading}
          />
        </Modal>
      )}

      {showLeaveConfirm && (
        <Modal onClose={() => setShowLeaveConfirm(false)} title="确认退车？">
          <div className="mb-6 space-y-3">
            <p className="text-slate-300">确定要退出《{carpool.script?.name}》拼车吗？</p>
            {carpool.deposit_amount > 0 && (
              <p className="text-yellow-400 text-sm bg-yellow-500/10 p-3 rounded-lg border border-yellow-500/30">
                💳 你的押金 ¥{carpool.deposit_amount} 将自动退回
              </p>
            )}
            {waitlistCount > 0 && (
              <p className="text-orange-400 text-sm bg-orange-500/10 p-3 rounded-lg border border-orange-500/30">
                🕐 候补队列第 1 位玩家将自动补位你的位置
              </p>
            )}
            <div className="space-y-2">
              <label className="block text-sm font-medium text-slate-300">确认你的昵称</label>
              <input
                type="text"
                value={playerName}
                onChange={(e) => setPlayerName(e.target.value)}
                placeholder="请输入你在拼车中的昵称"
                className="w-full px-4 py-3 bg-slate-700/50 border border-slate-600/50 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:border-red-500 focus:ring-2 focus:ring-red-500/20 transition-all"
              />
            </div>
          </div>
          <ModalActions
            onCancel={() => setShowLeaveConfirm(false)}
            onConfirm={handleLeave}
            confirmText={isLoading ? '处理中...' : '确认退车'}
            confirmClass="from-red-600 to-rose-600 hover:from-red-500 hover:to-rose-500"
            disabled={!playerName.trim() || isLoading}
          />
        </Modal>
      )}
    </>
  );
}

function Modal({ onClose, title, children }) {
  return (
    <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-gradient-to-br from-slate-800 to-slate-900 rounded-2xl border border-slate-600/50 w-full max-w-md p-6 animate-float max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-xl font-bold text-white">{title}</h3>
          <button onClick={onClose} className="w-10 h-10 flex items-center justify-center rounded-xl bg-slate-700/50 text-slate-400 hover:text-white hover:bg-slate-700 transition-all">
            ✕
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}

function PlayerForm({ name, setName, contact, setContact }) {
  return (
    <div className="space-y-4 mb-6">
      <div>
        <label className="block text-sm font-medium text-slate-300 mb-2">
          你的昵称 <span className="text-red-400">*</span>
        </label>
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="请输入你的昵称"
          className="w-full px-4 py-3 bg-slate-700/50 border border-slate-600/50 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 transition-all"
        />
      </div>
      <div>
        <label className="block text-sm font-medium text-slate-300 mb-2">联系方式</label>
        <input
          type="text"
          value={contact}
          onChange={(e) => setContact(e.target.value)}
          placeholder="手机号或微信号（选填）"
          className="w-full px-4 py-3 bg-slate-700/50 border border-slate-600/50 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 transition-all"
        />
      </div>
    </div>
  );
}

function ModalActions({ onCancel, onConfirm, confirmText, disabled, confirmClass }) {
  return (
    <div className="flex gap-3">
      <button
        onClick={onCancel}
        className="flex-1 px-6 py-3 bg-slate-700/50 text-slate-300 rounded-xl font-semibold hover:bg-slate-700 transition-all active:scale-95"
      >
        取消
      </button>
      <button
        onClick={onConfirm}
        disabled={disabled}
        className={`flex-1 px-6 py-3 bg-gradient-to-r ${confirmClass || 'from-indigo-600 to-purple-600'} text-white rounded-xl font-semibold transition-all active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed`}
      >
        {confirmText}
      </button>
    </div>
  );
}
