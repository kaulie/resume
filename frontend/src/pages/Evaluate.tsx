import { useEffect, useState } from 'react';
import { Gauge } from 'lucide-react';
import { api, Resume, JD, Evaluation } from '../api';

export default function Evaluate() {
  const [resumes, setResumes] = useState<Resume[]>([]);
  const [jds, setJds] = useState<JD[]>([]);
  const [resumeId, setResumeId] = useState('');
  const [jdId, setJdId] = useState('');
  const [result, setResult] = useState<Evaluation | null>(null);
  const [history, setHistory] = useState<Evaluation[]>([]);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    api.listResumes().then(setResumes).catch((e) => setError(e.message));
    api.listJDs().then(setJds).catch((e) => setError(e.message));
    api.listEvaluations().then(setHistory).catch(() => {});
  }, []);

  async function run() {
    if (!resumeId || !jdId) return setError('请选择简历和 JD');
    setBusy(true);
    setError('');
    try {
      const r = await api.evaluate(resumeId, jdId);
      setResult(r);
      api.listEvaluations().then(setHistory).catch(() => {});
    } catch (e: any) {
      setError(e.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="mx-auto max-w-5xl">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-slate-800">简历评价</h1>
        <p className="mt-1 text-sm text-slate-500">针对不同 JD，对简历进行打分并给出改进意见。</p>
      </div>

      <div className="card p-5">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="label">选择简历</label>
            <select className="input" value={resumeId} onChange={(e) => setResumeId(e.target.value)}>
              <option value="">请选择…</option>
              {resumes.map((r) => (<option key={r.id} value={r.id}>{r.title}</option>))}
            </select>
          </div>
          <div>
            <label className="label">选择 JD</label>
            <select className="input" value={jdId} onChange={(e) => setJdId(e.target.value)}>
              <option value="">请选择…</option>
              {jds.map((j) => (<option key={j.id} value={j.id}>{j.title}{j.company ? ` @ ${j.company}` : ''}</option>))}
            </select>
          </div>
        </div>
        <div className="mt-4 flex justify-end">
          <button className="btn-primary" onClick={run} disabled={busy || !resumeId || !jdId}>
            <Gauge size={16} /> {busy ? '评价中…' : '开始评价'}
          </button>
        </div>
      </div>

      {error && <div className="mt-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">{error}</div>}

      {result && (
        <div className="card mt-6 p-6">
          <div className="flex items-center gap-4">
            <div className="flex h-20 w-20 items-center justify-center rounded-full border-4 border-brand-500 text-2xl font-bold text-brand-600">
              {result.overallScore}
            </div>
            <div>
              <div className="text-lg font-semibold text-slate-800">综合评分 {result.overallScore} / 100</div>
              <div className="text-sm text-slate-400">{new Date(result.createdAt).toLocaleString('zh-CN')}</div>
            </div>
          </div>

          <div className="mt-6 space-y-4">
            {result.dimensions.map((d) => (
              <div key={d.name}>
                <div className="mb-1 flex justify-between text-sm">
                  <span className="font-medium text-slate-700">{d.name}</span>
                  <span className="text-slate-500">{d.score}</span>
                </div>
                <div className="h-2.5 overflow-hidden rounded-full bg-slate-100">
                  <div className="h-full rounded-full bg-brand-500" style={{ width: `${d.score}%` }} />
                </div>
                <div className="mt-1 text-xs text-slate-400">{d.comment}</div>
              </div>
            ))}
          </div>

          {result.suggestions.length > 0 && (
            <div className="mt-6">
              <h3 className="mb-2 font-semibold text-slate-800">改进建议</h3>
              <ul className="list-decimal space-y-1 pl-5 text-sm text-slate-600">
                {result.suggestions.map((s, i) => (<li key={i}>{s}</li>))}
              </ul>
            </div>
          )}
        </div>
      )}

      {history.length > 0 && (
        <div className="card mt-6 p-5">
          <h2 className="mb-3 font-semibold text-slate-800">历史评价</h2>
          <ul className="space-y-2">
            {history.map((h) => (
              <li key={h.id} className="flex items-center justify-between rounded-lg border border-slate-100 px-3 py-2 text-sm">
                <span className="text-slate-600">综合 {h.overallScore} 分 · {new Date(h.createdAt).toLocaleString('zh-CN')}</span>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
