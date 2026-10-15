import { useEffect, useState } from 'react';
import { GraduationCap, BookOpen, Lightbulb } from 'lucide-react';
import { api, JD, InterviewPrep } from '../api';

export default function InterviewPrepPage() {
  const [jds, setJds] = useState<JD[]>([]);
  const [jdId, setJdId] = useState('');
  const [result, setResult] = useState<InterviewPrep | null>(null);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    api.listJDs().then(setJds).catch((e) => setError(e.message));
  }, []);

  async function run() {
    if (!jdId) return setError('请选择 JD');
    setBusy(true);
    setError('');
    try {
      setResult(await api.interviewPrep(jdId));
    } catch (e: any) {
      setError(e.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="mx-auto max-w-5xl">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-slate-800">面试备战</h1>
        <p className="mt-1 text-sm text-slate-500">输入 JD 后，生成针对性的面试备战建议。</p>
      </div>

      <div className="card p-5">
        <div className="flex items-end gap-4">
          <div className="flex-1">
            <label className="label">选择 JD</label>
            <select className="input" value={jdId} onChange={(e) => setJdId(e.target.value)}>
              <option value="">请选择…</option>
              {jds.map((j) => (<option key={j.id} value={j.id}>{j.title}{j.company ? ` @ ${j.company}` : ''}</option>))}
            </select>
          </div>
          <button className="btn-primary" onClick={run} disabled={busy || !jdId}>
            <GraduationCap size={16} /> {busy ? '生成中…' : '生成备战建议'}
          </button>
        </div>
      </div>

      {error && <div className="mt-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">{error}</div>}

      {result && (
        <div className="mt-6 space-y-6">
          <div className="card p-5">
            <p className="text-sm leading-relaxed text-slate-700">{result.content.summary}</p>
          </div>

          {result.content.questions.map((cat, i) => (
            <div key={i} className="card p-5">
              <h2 className="mb-3 font-semibold text-slate-800">{cat.category}</h2>
              <div className="space-y-3">
                {cat.items.map((item, j) => (
                  <div key={j} className="rounded-lg bg-slate-50 p-3">
                    <div className="font-medium text-slate-700">{item.question}</div>
                    <div className="mt-1 text-sm text-slate-500">{item.hint}</div>
                  </div>
                ))}
              </div>
            </div>
          ))}

          <div className="grid grid-cols-2 gap-6">
            <div className="card p-5">
              <h2 className="mb-3 flex items-center gap-2 font-semibold text-slate-800">
                <BookOpen size={16} /> 需要复习的知识点
              </h2>
              <div className="flex flex-wrap gap-2">
                {result.content.studyTopics.map((t, i) => (
                  <span key={i} className="badge bg-purple-50 text-purple-700">{t}</span>
                ))}
              </div>
            </div>
            <div className="card p-5">
              <h2 className="mb-3 flex items-center gap-2 font-semibold text-slate-800">
                <Lightbulb size={16} /> 项目经历准备建议
              </h2>
              <ul className="list-disc space-y-1 pl-5 text-sm text-slate-600">
                {result.content.projectTips.map((t, i) => (<li key={i}>{t}</li>))}
              </ul>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
