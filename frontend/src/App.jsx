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

  useEffect(() => {
    fetchCarpools();
    fetchScripts();
  }, [fetchCarpools, fetchScripts]);

  useEffect(() => {
    const interval = setInterval(() => {
      fetchCarpools(false);
    }, 3000);
    return () => clearInterval(interval);
  }, [fetchCarpools]);

  const handleCreateCarpool = async (data) => {
    await api.createCarpool(data);
    fetchCarpools();
  };

  const handleJoinCarpool = async (carpoolId, name, contact) => {
    await api.joinCarpool({
      carpool_id: carpoolId,
      name,
      contact,
    });
    fetchCarpools();
  };

  const filteredCarpools = carpools;

  const recruitingCount = carpools.filter(c => c.status === 'recruiting').length;
  const fullCount = carpools.filter(c => c.status === 'full').length;
  const totalPlayers = carpools.reduce((sum, c) => sum + c.current_players, 0);

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
                <p className="text-slate-500 text-sm">实时拼单 · 智能组队 · 即刻发车</p>
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
                  <div className="text-slate-500">在线玩家</div>
                </div>
              </div>

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
                onJoin={handleJoinCarpool}
              />
            ))}
          </div>
        )}
      </main>

      <footer className="border-t border-slate-800 py-6 mt-12">
        <div className="max-w-7xl mx-auto px-4 text-center text-slate-500 text-sm">
          <p>剧本杀智能拼单系统 · 每 3 秒自动刷新 · 数据实时同步</p>
        </div>
      </footer>

      <CreateCarpoolModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        scripts={scripts}
        onCreate={handleCreateCarpool}
      />
    </div>
  );
}

export default App;
