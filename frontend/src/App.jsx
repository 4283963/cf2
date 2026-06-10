import { useState, useEffect, useCallback } from 'react';
import { api } from './services/api';
import CarpoolCard from './components/CarpoolCard';
import CreateCarpoolModal from './components/CreateCarpoolModal';

function App() {
  const [carpools, setCarpools] = useState([]);
  const [scripts, setScripts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('all');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [lastUpdate, setLastUpdate] = useState(new Date());
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [currentUser, setCurrentUser] = useState(() => localStorage.getItem('current_user') || '');
  const [showUserModal, setShowUserModal] = useState(!localStorage.getItem('current_user'));
  const [tempUserName, setTempUserName] = useState('');
  const [notifications, setNotifications] = useState([]);
  const [showNotifications, setShowNotifications] = useState(false);

  const saveCurrentUser = (name) => {
    setCurrentUser(name);
    localStorage.setItem('current_user', name);
  };

  const fetchCarpools = useCallback(async (showLoading = true) => {
    if (showLoading) setIsRefreshing(true);
    try {
      const status = filter === 'all' ? '' : filter;
      const data = await api.getCarpools(status);
      setCarpools(data);
      setLastUpdate(new Date());
    } catch (error) {
      console.error('获取拼车列表失败:', error);
    } finally {
      if (showLoading) setIsRefreshing(false);
      setLoading(false);
    }
  }, [filter]);

  const fetchScripts = useCallback(async () => {
    try {
      const data = await api.getScripts();
      setScripts(data);
    } catch (error) {
      console.error('获取剧本列表失败:', error);
    }
  }, []);

  const fetchNotifications = useCallback(async () => {
    if (!currentUser) return;
    try {
      const data = await api.getNotifications(currentUser);
      setNotifications(data);
    } catch (e) {}
  }, [currentUser]);

  useEffect(() => {
    fetchCarpools();
    fetchScripts();
  }, [fetchCarpools, fetchScripts]);

  useEffect(() => {
    fetchNotifications();
  }, [fetchNotifications]);

  useEffect(() => {
    const interval = setInterval(() => {
      fetchCarpools(false);
      fetchNotifications();
    }, 3000);
    return () => clearInterval(interval);
  }, [fetchCarpools, fetchNotifications]);

  const handleCreateCarpool = async (data) => {
    await api.createCarpool(data);
    fetchCarpools();
  };

  const filteredCarpools = carpools;
  const recruitingCount = carpools.filter(c => c.status === 'recruiting').length;
  const fullCount = carpools.filter(c => c.status === 'full').length;
  const totalPlayers = carpools.reduce((sum, c) => sum + c.current_players, 0);
  const waitlistTotal = carpools.reduce((sum, c) => sum + (c.waitlist?.length || 0), 0);
  const unreadNotifs = notifications.filter(n => !n.is_read).length;

  return (
    <div className="min-h-screen text-white">
      <header className="sticky top-0 z-40 backdrop-blur-xl bg-slate-900/80 border-b border-slate-700/50">
        <div className="max-w-7xl mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div className="w-12 h-12 rounded-2xl bg-gradient-to-br from-indigo-500 to-purple-600 flex items-center justify-center text-2xl shadow-lg shadow-indigo-500/30">
                🎭
              </div>
              <div>
                <h1 className="text-2xl font-bold bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">
                  剧本杀拼车大厅
                </h1>
                <p className="text-slate-500 text-sm">实时拼单 · 智能候补 · 自动补位</p>
              </div>
            </div>

            <div className="flex items-center gap-4">
              <div className="hidden md:flex items-center gap-6 text-sm">
                <div className="text-center">
                  <div className="text-2xl font-bold text-indigo-400">{recruitingCount}</div>
                  <div className="text-slate-500">招募中</div>
                </div>
                <div className="w-px h-10 bg-slate-700" />
                <div className="text-center">
                  <div className="text-2xl font-bold text-emerald-400">{fullCount}</div>
                  <div className="text-slate-500">已满员</div>
                </div>
                <div className="w-px h-10 bg-slate-700" />
                <div className="text-center">
                  <div className="text-2xl font-bold text-amber-400">{totalPlayers}</div>
                  <div className="text-slate-500">在车玩家</div>
                </div>
                <div className="w-px h-10 bg-slate-700" />
                <div className="text-center">
                  <div className="text-2xl font-bold text-orange-400">{waitlistTotal}</div>
                  <div className="text-slate-500">候补中</div>
                </div>
              </div>

              <div className="relative">
                <button
                  onClick={() => setShowNotifications(!showNotifications)}
                  className="relative w-11 h-11 flex items-center justify-center rounded-xl bg-slate-800/50 hover:bg-slate-700/50 transition-all"
                >
                  <span className="text-xl">🔔</span>
                  {unreadNotifs > 0 && (
                    <span className="absolute -top-1 -right-1 w-5 h-5 bg-red-500 text-white text-xs font-bold rounded-full flex items-center justify-center">
                      {unreadNotifs}
                    </span>
                  )}
                </button>
                {showNotifications && (
                  <div className="absolute right-0 top-14 w-80 bg-gradient-to-br from-slate-800 to-slate-900 border border-slate-600/50 rounded-xl shadow-2xl overflow-hidden z-50 max-h-96 overflow-y-auto">
                    <div className="px-4 py-3 border-b border-slate-700/50 flex items-center justify-between">
                      <span className="font-semibold text-white">通知中心</span>
                      {unreadNotifs > 0 && (
                        <span className="text-xs text-red-400">{unreadNotifs} 条未读</span>
                      )}
                    </div>
                    {notifications.length === 0 ? (
                      <div className="p-8 text-center text-slate-500">暂无通知</div>
                    ) : (
                      <div>
                        {notifications.slice(0, 20).map((n) => (
                          <div key={n.id} className={`px-4 py-3 border-b border-slate-700/30 ${!n.is_read ? 'bg-indigo-500/5' : ''}`}>
                            <div className={`text-sm ${!n.is_read ? 'text-white' : 'text-slate-400'}`}>{n.message}</div>
                            <div className="text-xs text-slate-600 mt-1">
                              {new Date(n.created_at).toLocaleString()}
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                )}
              </div>

              <button
                onClick={() => setShowUserModal(true)}
                className="px-4 py-2 rounded-xl bg-slate-800/50 text-slate-300 hover:bg-slate-700/50 transition-all flex items-center gap-2"
              >
                <span className="text-lg">👤</span>
                <span className="text-sm">{currentUser || '设置昵称'}</span>
              </button>

              <button
                onClick={() => setShowCreateModal(true)}
                className="px-6 py-3 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-xl font-semibold hover:from-indigo-500 hover:to-purple-500 transition-all shadow-lg shadow-indigo-500/30 hover:shadow-indigo-500/50 active:scale-95 flex items-center gap-2"
              >
                <span className="text-xl">+</span>
                发起拼车
              </button>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-8">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8">
          <div className="flex items-center gap-3">
            <div className={`flex items-center gap-2 px-4 py-2 rounded-xl ${
              isRefreshing ? 'bg-indigo-500/20 text-indigo-300' : 'bg-slate-800/50 text-slate-400'
            }`}>
              <div className={`w-2 h-2 rounded-full ${
                isRefreshing ? 'bg-indigo-400 animate-pulse' : 'bg-emerald-400'
              }`} />
              <span className="text-sm">
                {isRefreshing ? '刷新中...' : `最后更新：${lastUpdate.toLocaleTimeString()}`}
              </span>
            </div>
            <button
              onClick={() => fetchCarpools()}
              disabled={isRefreshing}
              className="p-2 rounded-xl bg-slate-800/50 text-slate-400 hover:text-white hover:bg-slate-700/50 transition-all disabled:opacity-50"
            >
              🔄
            </button>
          </div>

          <div className="flex gap-2 bg-slate-800/50 p-1.5 rounded-xl">
            {[
              { key: 'all', label: '全部', count: carpools.length },
              { key: 'recruiting', label: '招募中', count: recruitingCount },
              { key: 'full', label: '已满员', count: fullCount },
            ].map((item) => (
              <button
                key={item.key}
                onClick={() => setFilter(item.key)}
                className={`px-5 py-2 rounded-lg font-medium transition-all flex items-center gap-2 ${
                  filter === item.key
                    ? 'bg-gradient-to-r from-indigo-600 to-purple-600 text-white shadow-lg'
                    : 'text-slate-400 hover:text-white hover:bg-slate-700/50'
                }`}
              >
                {item.label}
                <span className={`px-2 py-0.5 rounded-full text-xs ${
                  filter === item.key ? 'bg-white/20' : 'bg-slate-700'
                }`}>
                  {item.count}
                </span>
              </button>
            ))}
          </div>
        </div>

        {loading ? (
          <div className="flex flex-col items-center justify-center py-20">
            <div className="w-16 h-16 border-4 border-indigo-500/30 border-t-indigo-500 rounded-full animate-spin mb-4" />
            <p className="text-slate-400">正在加载拼车数据...</p>
          </div>
        ) : filteredCarpools.length === 0 ? (
          <div className="text-center py-20">
            <div className="text-6xl mb-4">🎲</div>
            <h3 className="text-xl font-semibold text-slate-300 mb-2">暂无拼车</h3>
            <p className="text-slate-500 mb-6">成为第一个发起拼车的玩家吧！</p>
            <button
              onClick={() => setShowCreateModal(true)}
              className="px-6 py-3 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-xl font-semibold hover:from-indigo-500 hover:to-purple-500 transition-all"
            >
              发起拼车
            </button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {filteredCarpools.map((carpool) => (
              <CarpoolCard
                key={carpool.id}
                carpool={carpool}
                currentUser={currentUser}
                onRefresh={() => fetchCarpools()}
              />
            ))}
          </div>
        )}
      </main>

      <footer className="border-t border-slate-800 py-6 mt-12">
        <div className="max-w-7xl mx-auto px-4 text-center text-slate-500 text-sm">
          <p>剧本杀智能拼单系统 · 每 3 秒自动刷新 · 候补自动补位 · 超时自动解散</p>
        </div>
      </footer>

      <CreateCarpoolModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        scripts={scripts}
        onCreate={handleCreateCarpool}
      />

      {showUserModal && (
        <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4">
          <div className="bg-gradient-to-br from-slate-800 to-slate-900 rounded-2xl border border-slate-600/50 w-full max-w-md p-6 animate-float">
            <h3 className="text-2xl font-bold text-white mb-2">👋 欢迎来到拼车大厅</h3>
            <p className="text-slate-400 mb-6">请先设置你的昵称，这样其他玩家才能认出你</p>
            <input
              type="text"
              value={tempUserName}
              onChange={(e) => setTempUserName(e.target.value)}
              placeholder="输入你的昵称"
              className="w-full px-4 py-3 bg-slate-700/50 border border-slate-600/50 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 transition-all mb-6"
            />
            <div className="flex gap-3">
              <button
                onClick={() => setShowUserModal(false)}
                className="flex-1 px-6 py-3 bg-slate-700/50 text-slate-300 rounded-xl font-semibold hover:bg-slate-700 transition-all"
              >
                稍后设置
              </button>
              <button
                onClick={() => {
                  if (tempUserName.trim()) {
                    saveCurrentUser(tempUserName.trim());
                    setShowUserModal(false);
                  }
                }}
                disabled={!tempUserName.trim()}
                className="flex-1 px-6 py-3 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-xl font-semibold hover:from-indigo-500 hover:to-purple-500 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
              >
                确认昵称
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default App;
