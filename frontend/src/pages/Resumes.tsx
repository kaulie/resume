import { useEffect, useRef, useState } from 'react';
import { Upload, Eye, Trash2, Save, X, FileText } from 'lucide-react';
import { api, Resume, ResumeContent } from '../api';

export default function Resumes() {
  const [resumes, setResumes] = useState<Resume[]>([]);
  const [selected, setSelected] = useState<Resume | null>(null);
  const [draft, setDraft] = useState<ResumeContent | null>(null);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [busy, setBusy] = useState(false);
  const fileRef = useRef<HTMLInputElement>(null);

  const load = () => api.listResumes().then(setResumes).catch((e) => setError(e.message));

  useEffect(() => {
    load();
  }, []);

  async function onUpload(file: File) {
    setBusy(true);
    setError('');
    setNotice('');
    try {
      const r = await api.uploadResume(file);
      await load();
      if (!r.sourceText || !r.sourceText.trim()) {
        setNotice('该文件未提取到文字内容：可能是扫描件/图片型 PDF。建议上传文本型 PDF，或改用 docx / md / txt。');
      }
      openEditor(r.id);
    } catch (e: any) {
      setError(e.message || '上传失败');
    } finally {
      setBusy(false);
    }
  }

  async function openEditor(id: string) {
    try {
      const r = await api.getResume(id);
      setSelected(r);
      setDraft(JSON.parse(JSON.stringify(r.content)));
    } catch (e: any) {
      setError(e.message);
    }
  }

  async function save() {
    if (!selected || !draft) return;
    setBusy(true);
    setError('');
    try {
      await api.updateResume(selected.id, { content: draft });
      await load();
      setSelected(null);
      setDraft(null);
    } catch (e: any) {
      setError(e.message);
    } finally {
      setBusy(false);
    }
  }

  async function remove(id: string) {
    if (!confirm('确认删除该简历？')) return;
    await api.deleteResume(id).catch((e) => setError(e.message));
    load();
  }

  return (
    <div className="mx-auto max-w-5xl">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-slate-800">简历管理</h1>
          <p className="mt-1 text-sm text-slate-500">上传并管理你的简历，支持 docx / md / pdf / txt。</p>
        </div>
        <button className="btn-primary" onClick={() => fileRef.current?.click()} disabled={busy}>
          <Upload size={16} /> {busy ? '上传中…' : '上传简历'}
        </button>
        <input
          ref={fileRef}
          type="file"
          accept=".docx,.md,.markdown,.pdf,.txt,application/pdf,text/plain,text/markdown,application/vnd.openxmlformats-officedocument.wordprocessingml.document"
          className="hidden"
          onChange={(e) => {
            const f = e.target.files?.[0];
            if (f) onUpload(f);
            e.target.value = '';
          }}
        />
      </div>

      {error && <div className="mb-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">{error}</div>}
      {notice && <div className="mb-4 rounded-lg bg-amber-50 px-4 py-3 text-sm text-amber-700">{notice}</div>}

      <div className="space-y-3">
        {resumes.length === 0 && (
          <div className="card p-10 text-center text-sm text-slate-400">
            <FileText className="mx-auto mb-2 text-slate-300" size={32} />
            还没有简历，点击右上角上传。
          </div>
        )}
        {resumes.map((r) => (
          <div key={r.id} className="card flex items-center justify-between p-4">
            <div>
              <div className="font-medium text-slate-800">{r.title}</div>
              <div className="text-xs text-slate-400">
                {r.originalName || '手动创建'} · 更新于 {new Date(r.updatedAt).toLocaleString('zh-CN')}
              </div>
            </div>
            <div className="flex items-center gap-1">
              <button className="btn-secondary" onClick={() => openEditor(r.id)}>
                <Eye size={16} /> 预览/编辑
              </button>
              <button className="btn-danger" onClick={() => remove(r.id)}>
                <Trash2 size={16} /> 删除
              </button>
            </div>
          </div>
        ))}
      </div>

      {selected && draft && (
        <EditorModal draft={draft} setDraft={setDraft} busy={busy} onClose={() => { setSelected(null); setDraft(null); }} onSave={save} />
      )}
    </div>
  );
}

function EditorModal({
  draft,
  setDraft,
  busy,
  onClose,
  onSave,
}: {
  draft: ResumeContent;
  setDraft: (c: ResumeContent) => void;
  busy: boolean;
  onClose: () => void;
  onSave: () => void;
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/40 p-6">
      <div className="card my-8 w-full max-w-3xl p-6">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-lg font-semibold text-slate-800">编辑简历</h2>
          <button className="btn-secondary" onClick={onClose}>
            <X size={16} /> 关闭
          </button>
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="label">姓名</label>
            <input className="input" value={draft.name} onChange={(e) => setDraft({ ...draft, name: e.target.value })} />
          </div>
          <div>
            <label className="label">求职方向</label>
            <input className="input" value={draft.title} onChange={(e) => setDraft({ ...draft, title: e.target.value })} />
          </div>
        </div>
        <div className="mt-4">
          <label className="label">联系方式</label>
          <input className="input" value={draft.contact} onChange={(e) => setDraft({ ...draft, contact: e.target.value })} />
        </div>
        <div className="mt-4">
          <label className="label">个人简介 / 自我评价</label>
          <textarea className="input min-h-[80px]" value={draft.summary} onChange={(e) => setDraft({ ...draft, summary: e.target.value })} />
        </div>

        <div className="mt-4 space-y-4">
          {(draft.sections || []).map((sec, si) => (
            <div key={si} className="rounded-lg border border-slate-200 p-3">
              <input
                className="input mb-2 font-medium"
                value={sec.title}
                onChange={(e) => {
                  const sections = draft.sections.map((s, i) => (i === si ? { ...s, title: e.target.value } : s));
                  setDraft({ ...draft, sections });
                }}
              />
              <textarea
                className="input min-h-[90px] font-mono text-xs"
                value={sec.bullets.join('\n')}
                placeholder="每行一条"
                onChange={(e) => {
                  const bullets = e.target.value.split('\n');
                  const sections = draft.sections.map((s, i) => (i === si ? { ...s, bullets } : s));
                  setDraft({ ...draft, sections });
                }}
              />
            </div>
          ))}
        </div>

        <div className="mt-6 flex justify-end gap-2">
          <button className="btn-secondary" onClick={onClose}>取消</button>
          <button className="btn-primary" onClick={onSave} disabled={busy}>
            <Save size={16} /> 保存
          </button>
        </div>
      </div>
    </div>
  );
}
