import { useEffect, useState } from 'react';
import { Save, Settings as SettingsIcon, BrainCircuit } from 'lucide-react';
import { api, Settings, Meta } from '../api';

export default function SettingsPage() {
  const [settings, setSettings] = useState<Settings | null>(null);
  const [meta, setMeta] = useState<Meta | null>(null);
  const [error, setError] = useState('');
  const [saved, setSaved] = useState(false);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    api.getSettings().then(setSettings).catch((e) => setError(e.message));
    api.meta().then(setMeta).catch(() => {});
  }, []);

  async function save() {
    if (!settings) return;
    setBusy(true);
    setError('');
    setSaved(false);
    try {
      await api.updateSettings(settings);
      setSaved(true);
      api.meta().then(setMeta).catch(() => {});
    } catch (e: any) {
      setError(e.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="mx-auto max-w-3xl">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-slate-800">系统设置</h1>
        <p className="mt-1 text-sm text-slate-500">配置大模型接入与 PDF 导出字体。</p>
      </div>

      {error && <div className="mb-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">{error}</div>}
      {saved && <div className="mb-4 rounded-lg bg-green-50 px-4 py-3 text-sm text-green-700">已保存。</div>}

      {!settings ? (
        <div className="text-sm text-slate-400">加载中…</div>
      ) : (
        <div className="space-y-6">
          <div className="card p-6">
            <div className="mb-4 flex items-center gap-2">
              <BrainCircuit size={18} className="text-brand-600" />
              <h2 className="font-semibold text-slate-800">大模型接入</h2>
              <span className="badge bg-slate-100 text-slate-500">当前引擎：{meta?.engine || '…'}</span>
            </div>

            <label className="mb-3 flex cursor-pointer items-center gap-2">
              <input
                type="checkbox"
                checked={settings.llmEnabled}
                onChange={(e) => setSettings({ ...settings, llmEnabled: e.target.checked })}
                className="h-4 w-4 rounded border-slate-300 text-brand-600"
              />
              <span className="text-sm font-medium text-slate-700">启用大模型（关闭时使用内置规则引擎）</span>
            </label>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="label">接口地址（OpenAI 兼容）</label>
                <input
                  className="input"
                  value={settings.llmBaseUrl}
                  placeholder="https://api.deepseek.com"
                  onChange={(e) => setSettings({ ...settings, llmBaseUrl: e.target.value })}
                />
              </div>
              <div>
                <label className="label">模型名称</label>
                <input
                  className="input"
                  value={settings.llmModel}
                  placeholder="deepseek-chat"
                  onChange={(e) => setSettings({ ...settings, llmModel: e.target.value })}
                />
              </div>
            </div>
            <div className="mt-4">
              <label className="label">API Key</label>
              <input
                className="input"
                type="password"
                value={settings.llmApiKey}
                placeholder="sk-..."
                onChange={(e) => setSettings({ ...settings, llmApiKey: e.target.value })}
              />
            </div>
            <p className="mt-2 text-xs text-slate-400">
              支持任何 OpenAI 兼容接口（如 DeepSeek、OpenAI、通义千问、本地 Ollama 等）。API Key 仅保存在本地数据库。
            </p>
          </div>

          <div className="card p-6">
            <div className="mb-4 flex items-center gap-2">
              <SettingsIcon size={18} className="text-brand-600" />
              <h2 className="font-semibold text-slate-800">PDF 导出</h2>
            </div>
            <div>
              <label className="label">中文字体文件路径（.ttf）</label>
              <input
                className="input"
                value={settings.pdfFontPath}
                placeholder="留空则自动检测系统字体"
                onChange={(e) => setSettings({ ...settings, pdfFontPath: e.target.value })}
              />
            </div>
            <p className="mt-2 text-xs text-slate-400">用于 PDF 导出的中文字体，需为单个 .ttf 文件（如 Arial Unicode、Noto Sans SC）。</p>
          </div>

          <div className="flex justify-end">
            <button className="btn-primary" onClick={save} disabled={busy}>
              <Save size={16} /> {busy ? '保存中…' : '保存设置'}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
