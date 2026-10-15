import { useEffect, useState } from 'react';
import { Download, FileText } from 'lucide-react';
import { api, Resume, ResumeVersion, exportUrl } from '../api';

const FORMATS = [
  { id: 'docx', label: 'Word (.docx)' },
  { id: 'md', label: 'Markdown (.md)' },
  { id: 'pdf', label: 'PDF (.pdf)' },
  { id: 'txt', label: '纯文本 (.txt)' },
];

export default function ExportCenter() {
  const [resumes, setResumes] = useState<Resume[]>([]);
  const [versions, setVersions] = useState<ResumeVersion[]>([]);
  const [resumeId, setResumeId] = useState('');
  const [versionId, setVersionId] = useState('');
  const [format, setFormat] = useState('docx');
  const [error, setError] = useState('');

  useEffect(() => {
    api.listResumes().then(setResumes).catch((e) => setError(e.message));
  }, []);

  useEffect(() => {
    if (resumeId) api.listVersions(resumeId).then(setVersions).catch(() => {});
    else setVersions([]);
    setVersionId('');
  }, [resumeId]);

  const ready = Boolean(resumeId || versionId);

  return (
    <div className="mx-auto max-w-5xl">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-slate-800">导出中心</h1>
        <p className="mt-1 text-sm text-slate-500">将简历或优化版本导出为 docx / md / pdf / txt。</p>
      </div>

      <div className="card p-6">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="label">选择简历</label>
            <select className="input" value={resumeId} onChange={(e) => setResumeId(e.target.value)}>
              <option value="">请选择…</option>
              {resumes.map((r) => (<option key={r.id} value={r.id}>{r.title}</option>))}
            </select>
          </div>
          <div>
            <label className="label">版本（可选）</label>
            <select className="input" value={versionId} onChange={(e) => setVersionId(e.target.value)}>
              <option value="">原始简历</option>
              {versions.map((v) => (<option key={v.id} value={v.id}>{v.label}</option>))}
            </select>
          </div>
        </div>

        <div className="mt-6">
          <label className="label">导出格式</label>
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-4">
            {FORMATS.map((f) => (
              <button
                key={f.id}
                className={`rounded-lg border p-3 text-sm transition-colors ${
                  format === f.id ? 'border-brand-500 bg-brand-50 text-brand-700' : 'border-slate-200 text-slate-600 hover:bg-slate-50'
                }`}
                onClick={() => setFormat(f.id)}
              >
                <FileText size={16} className="mb-1" />
                {f.label}
              </button>
            ))}
          </div>
        </div>

        {error && <div className="mt-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">{error}</div>}

        <div className="mt-6 flex justify-end">
          <a
            className={`btn-primary ${!ready ? 'pointer-events-none opacity-50' : ''}`}
            href={ready ? exportUrl({ resumeId: resumeId || undefined, versionId: versionId || undefined, format }) : undefined}
            onClick={() => !ready && setError('请先选择简历或版本')}
          >
            <Download size={16} /> 导出
          </a>
        </div>
      </div>
    </div>
  );
}
