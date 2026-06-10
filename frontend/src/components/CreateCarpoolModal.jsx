import { useState } from 'react';

export default function CreateCarpoolModal({ isOpen, onClose, scripts, onCreate }) {
  const [selectedScript, setSelectedScript] = useState('');
  const [hostName, setHostName] = useState('');
  const [hostContact, setHostContact] = useState('');
  const [isCreating, setIsCreating] = useState(false);

  if (!isOpen) return null;

  const selectedScriptData = scripts.find(s => s.id === Number(selectedScript));

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!selectedScript || !hostName.trim()) return;

    setIsCreating(true);
    try {
      await onCreate({
        script_id: Number(selectedScript),
        host_name: hostName.trim(),
        host_contact: hostContact.trim(),
      });
      setSelectedScript('');
      setHostName('');
      setHostContact('');
      onClose();
    } catch (error) {
      alert(error.message);
    } finally {
      setIsCreating(false);
    }
  };

  const getTypeColor = (type) => {
    if (!type) return 'bg-slate-500/20 text-slate-300';
    if (type.includes('情感')) return 'bg-pink-500/20 text-pink-300';
    if (type.includes('硬核')) return 'bg-red-500/20 text-red-300';
    if (type.includes('惊悚')) return 'bg-purple-500/20 text-purple-300';
    if (type.includes('日式')) return 'bg-amber-500/20 text-amber-300';
    return 'bg-blue-500/20 text-blue-300';
  };

  return (
    <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-4">
      <div className="bg-gradient-to-br from-slate-800 to-slate-900 rounded-2xl border border-slate-600/50 w-full max-w-lg p-6 animate-float max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h3 className="text-2xl font-bold text-white">发起拼车</h3>
            <p className="text-slate-400 text-sm mt-1">选择剧本，成为车主发起拼车</p>
          </div>
          <button
            onClick={onClose}
            className="w-10 h-10 flex items-center justify-center rounded-xl bg-slate-700/50 text-slate-400 hover:text-white hover:bg-slate-700 transition-all"
          >
            ✕
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="block text-sm font-medium text-slate-300 mb-3">
              选择剧本 <span className="text-red-400">*</span>
            </label>
            <div className="grid grid-cols-2 gap-3 max-h-60 overflow-y-auto pr-2">
              {scripts.map((script) => (
                <div
                  key={script.id}
                  onClick={() => setSelectedScript(String(script.id))}
                  className={`p-3 rounded-xl border cursor-pointer transition-all ${
                    selectedScript === String(script.id)
                      ? 'border-indigo-500 bg-indigo-500/20 shadow-lg shadow-indigo-500/20'
                      : 'border-slate-600/50 bg-slate-700/30 hover:border-slate-500 hover:bg-slate-700/50'
                  }`}
                >
                  <div className="font-semibold text-white text-sm mb-1">{script.name}</div>
                  <div className="flex items-center gap-2 text-xs">
                    <span className="text-slate-400">{script.player_count}人本</span>
                    <span className={`px-2 py-0.5 rounded-full ${getTypeColor(script.type)}`}>
                      {script.type.split('/')[0]}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {selectedScriptData && (
            <div className="p-4 bg-indigo-500/10 border border-indigo-500/30 rounded-xl">
              <div className="font-semibold text-indigo-300 mb-2">
                {selectedScriptData.name}
              </div>
              <div className="text-sm text-slate-400 space-y-1">
                <p>📋 类型：{selectedScriptData.type}</p>
                <p>👥 人数：{selectedScriptData.player_count}人</p>
                <p>⏱️ 时长：{Math.floor(selectedScriptData.duration / 60)}小时{selectedScriptData.duration % 60 > 0 ? `${selectedScriptData.duration % 60}分钟` : ''}</p>
                <p>🎯 难度：{selectedScriptData.difficulty}</p>
              </div>
              <p className="text-xs text-slate-500 mt-2">{selectedScriptData.description}</p>
            </div>
          )}

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">
              你的昵称 <span className="text-red-400">*</span>
            </label>
            <input
              type="text"
              value={hostName}
              onChange={(e) => setHostName(e.target.value)}
              placeholder="作为车主，请输入你的昵称"
              className="w-full px-4 py-3 bg-slate-700/50 border border-slate-600/50 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 transition-all"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-slate-300 mb-2">
              联系方式
            </label>
            <input
              type="text"
              value={hostContact}
              onChange={(e) => setHostContact(e.target.value)}
              placeholder="手机号或微信号（选填，方便联系）"
              className="w-full px-4 py-3 bg-slate-700/50 border border-slate-600/50 rounded-xl text-white placeholder-slate-500 focus:outline-none focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 transition-all"
            />
          </div>

          <div className="flex gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-6 py-3 bg-slate-700/50 text-slate-300 rounded-xl font-semibold hover:bg-slate-700 transition-all active:scale-95"
            >
              取消
            </button>
            <button
              type="submit"
              disabled={!selectedScript || !hostName.trim() || isCreating}
              className="flex-1 px-6 py-3 bg-gradient-to-r from-indigo-600 to-purple-600 text-white rounded-xl font-semibold hover:from-indigo-500 hover:to-purple-500 transition-all active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isCreating ? '创建中...' : '发起拼车'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
