import { useEffect, useState } from 'react';
import { Sparkles, Download } from 'lucide-react';
import { api, Resume, JD, ResumeVersion, exportUrl } from '../api';

export default function Optimize() {
  const [resumes, setResumes] = useState<Resume[]>([]);
  const [jds, setJds] = useState<JD[]>([]);
  const [resumeId, setResumeId] = useState('');
  const [jdId, setJdId] = useState('');
  const [version, setVersion] = useState<ResumeVersion | null>(null);
  const [versions, setVersions] = useState<ResumeVersion[]>([]);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    api.listResumes().then(setResumes).catch((e) => setError(e.message));
    api.listJDs().then(setJds).catch((e) => setError(e.message));
  }, []);

  useEffect(() => {
    if (resumeId) api.listVersions(resumeId).then(setVersions).catch(() => {});
    else setVersions([]);
  }, [resumeId]);

  async function run() {
    if (!resumeId || !jdId) return setError('请选择简历和 JD');
    setBusy(true);
    setError('');
    try {
      const v = await api.optimize(resumeId, jdId);
      setVersion(v);
      api.listVersions(resumeId).then(setVersions).catch(() => {});
    } catch (e: any) {
      setError(e.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="mx-auto max-w-5xl">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-slate-800">简历优化</h1>
        <p className="mt-1 text-sm text-slate-500">选中简历与 JD，动态调整简历以匹配岗位要求。</p>
      </div>

      <div className="card p-5">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="label">选择简历</label>
            <select className="input" value={resumeId} onChange={(e) => setResumeId(e.target.value)}>
              <option value="">请选择…</option>
              {resumes.map((r) => (
                <option key={r.id} value={r.id}>{r.title}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="label">选择 JD</label>
            <select className="input" value={jdId} onChange={(e) => setJdId(e.target.value)}>
              <option value="">请选择…</option>
              {jds.map((j) => (
                <option key={j.id} value={j.id}>{j.title}{j.company ? ` @ ${j.company}` : ''}</option>
              ))}
            </select>
          </div>
        </div>
        <div className="mt-4 flex justify-end">
          <button className="btn-primary" onClick={run} disabled={busy || !resumeId || !jdId}>
            <Sparkles size={16} /> {busy ? '优化中…' : '开始优化'}
          </button>
        </div>
      </div>

      {error && <div className="mt-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">{error}</div>}

      {version && (
        <div className="card mt-6 p-5">
          <div className="mb-4 flex items-center justify-between">
            <h2 className="font-semibold text-slate-800">{version.label}</h2>
            <div className="flex gap-2">
              {(['md', 'docx', 'pdf', 'txt'] as const).map((f) => (
                <a key={f} className="btn-secondary" href={exportUrl({ versionId: version.id, format: f })} target="_blank" rel="noreferrer">
                  <Download size={14} /> {f.toUpperCase()}
                </a>
              ))}
            </div>
          </div>

          <div className="mb-3 text-sm text-slate-500">
            调整后的简介：<span className="text-slate-700">{version.content.summary}</span>
          </div>

          {version.diff.length > 0 ? (
            <div className="overflow-hidden rounded-lg border border-slate-200">
              <table className="w-full text-sm">
                <thead className="bg-slate-50 text-left text-xs text-slate-500">
                  <tr>
                    <th className="px-3 py-2">模块</th>
                    <th className="px-3 py-2">原文</th>
                    <th className="px-3 py-2">命中关键词</th>
                  </tr>
                </thead>
                <tbody>
                  {version.diff.map((d, i) => (
                    <tr key={i} className="border-t border-slate-100 align-top">
                      <td className="px-3 py-2 text-slate-500">{d.section}</td>
                      <td className="px-3 py-2 text-slate-700">{d.original}</td>
                      <td className="px-3 py-2">
                        {d.matched.map((m) => (
                          <span key={m} className="badge mr-1 bg-brand-50 text-brand-700">{m}</span>
                        ))}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="text-sm text-slate-400">未识别到逐条命中，请查看调整版中的「JD 匹配说明」区块。</p>
          )}

          <h3 className="mb-2 mt-5 font-semibold text-slate-800">调整后内容</h3>
          <Preview content={version.content} />
        </div>
      )}

      {versions.length > 0 && (
        <div className="card mt-6 p-5">
          <h2 className="mb-3 font-semibold text-slate-800">历史优化版本</h2>
          <ul className="space-y-2">
            {versions.map((v) => (
              <li key={v.id} className="flex items-center justify-between rounded-lg border border-slate-100 px-3 py-2 text-sm">
                <span className="text-slate-600">{v.label} · {new Date(v.createdAt).toLocaleString('zh-CN')}</span>
                <button className="text-brand-600 hover:underline" onClick={() => setVersion(v)}>查看</button>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}

function Preview({ content }: { content: ResumeVersion['content'] }) {
  return (
    <div className="space-y-4">
      {content.sections.map((sec, i) => (
        <div key={i}>
          <div className="mb-1 text-sm font-semibold text-slate-700">{sec.title}</div>
          <ul className="list-disc space-y-1 pl-5 text-sm text-slate-600">
            {sec.bullets.map((b, j) => (
              <li key={j}>{b}</li>
            ))}
          </ul>
        </div>
      ))}
    </div>
  );
}
