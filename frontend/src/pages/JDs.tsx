import { useEffect, useState } from 'react';
import { Plus, Pencil, Trash2, Save, X, Briefcase } from 'lucide-react';
import { api, JD } from '../api';

const EMPTY: Partial<JD> = { title: '', company: '', description: '', responsibilities: '', requirements: '', plusPoints: '' };

export default function JDs() {
  const [jds, setJds] = useState<JD[]>([]);
  const [editing, setEditing] = useState<Partial<JD> | null>(null);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  const load = () => api.listJDs().then(setJds).catch((e) => setError(e.message));

  useEffect(() => {
    load();
  }, []);

  async function save() {
    if (!editing) return;
    setBusy(true);
    setError('');
    try {
      if (editing.id) await api.updateJD(editing.id, editing);
      else await api.createJD(editing);
      setEditing(null);
      load();
    } catch (e: any) {
      setError(e.message);
    } finally {
      setBusy(false);
    }
  }

  async function remove(id: string) {
    if (!confirm('确认删除该 JD？')) return;
    await api.deleteJD(id).catch((e) => setError(e.message));
    load();
  }

  return (
    <div className="mx-auto max-w-5xl">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-slate-800">JD 管理</h1>
          <p className="mt-1 text-sm text-slate-500">维护职位描述，用于简历优化、评价与面试备战。</p>
        </div>
        <button className="btn-primary" onClick={() => setEditing({ ...EMPTY })}>
          <Plus size={16} /> 新建 JD
        </button>
      </div>

      {error && <div className="mb-4 rounded-lg bg-red-50 px-4 py-3 text-sm text-red-600">{error}</div>}

      <div className="space-y-3">
        {jds.length === 0 && (
          <div className="card p-10 text-center text-sm text-slate-400">
            <Briefcase className="mx-auto mb-2 text-slate-300" size={32} />
            还没有 JD，点击右上角新建。
          </div>
        )}
        {jds.map((jd) => (
          <div key={jd.id} className="card flex items-center justify-between p-4">
            <div>
              <div className="font-medium text-slate-800">
                {jd.title}
                {jd.company && <span className="ml-2 text-sm font-normal text-slate-500">@ {jd.company}</span>}
              </div>
              <div className="mt-1 line-clamp-1 max-w-xl text-xs text-slate-400">{jd.requirements || jd.description || '暂无要求'}</div>
            </div>
            <div className="flex items-center gap-1">
              <button className="btn-secondary" onClick={() => setEditing({ ...jd })}>
                <Pencil size={16} /> 编辑
              </button>
              <button className="btn-danger" onClick={() => remove(jd.id)}>
                <Trash2 size={16} /> 删除
              </button>
            </div>
          </div>
        ))}
      </div>

      {editing && (
        <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/40 p-6">
          <div className="card my-8 w-full max-w-2xl p-6">
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-lg font-semibold text-slate-800">{editing.id ? '编辑 JD' : '新建 JD'}</h2>
              <button className="btn-secondary" onClick={() => setEditing(null)}>
                <X size={16} /> 关闭
              </button>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="label">职位标题 *</label>
                <input className="input" value={editing.title || ''} onChange={(e) => setEditing({ ...editing, title: e.target.value })} />
              </div>
              <div>
                <label className="label">公司</label>
                <input className="input" value={editing.company || ''} onChange={(e) => setEditing({ ...editing, company: e.target.value })} />
              </div>
            </div>
            <Field label="职位描述" value={editing.description || ''} onChange={(v) => setEditing({ ...editing, description: v })} />
            <Field label="岗位职责" value={editing.responsibilities || ''} onChange={(v) => setEditing({ ...editing, responsibilities: v })} />
            <Field label="任职要求" value={editing.requirements || ''} onChange={(v) => setEditing({ ...editing, requirements: v })} />
            <Field label="加分项" value={editing.plusPoints || ''} onChange={(v) => setEditing({ ...editing, plusPoints: v })} />

            <div className="mt-6 flex justify-end gap-2">
              <button className="btn-secondary" onClick={() => setEditing(null)}>取消</button>
              <button className="btn-primary" onClick={save} disabled={busy || !editing.title}>
                <Save size={16} /> 保存
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function Field({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  return (
    <div className="mt-4">
      <label className="label">{label}</label>
      <textarea className="input min-h-[80px]" value={value} onChange={(e) => onChange(e.target.value)} />
    </div>
  );
}
