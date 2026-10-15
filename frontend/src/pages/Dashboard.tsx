import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { FileText, Briefcase, Gauge, GraduationCap, Upload, Plus, Sparkles, Download } from 'lucide-react';
import { api, Resume, JD, Evaluation, InterviewPrep } from '../api';

export default function Dashboard() {
  const [resumes, setResumes] = useState<Resume[]>([]);
  const [jds, setJds] = useState<JD[]>([]);
  const [evaluations, setEvaluations] = useState<Evaluation[]>([]);
  const [preps, setPreps] = useState<InterviewPrep[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      api.listResumes(),
      api.listJDs(),
      api.listEvaluations(),
      api.listInterviewPreps(),
    ])
      .then(([r, j, e, p]) => {
        setResumes(r);
        setJds(j);
        setEvaluations(e);
        setPreps(p);
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const stats = [
    { label: '简历', value: resumes.length, icon: FileText, to: '/resumes', color: 'bg-blue-50 text-blue-600' },
    { label: 'JD', value: jds.length, icon: Briefcase, to: '/jds', color: 'bg-emerald-50 text-emerald-600' },
    { label: '评价', value: evaluations.length, icon: Gauge, to: '/evaluate', color: 'bg-amber-50 text-amber-600' },
    { label: '面试备战', value: preps.length, icon: GraduationCap, to: '/interview', color: 'bg-purple-50 text-purple-600' },
  ];

  const latest = [...evaluations, ...preps]
    .map((x) => ({ ...x, kind: 'overallScore' in x ? '评价' : '备战' }))
    .sort((a, b) => (b.createdAt > a.createdAt ? 1 : -1))
    .slice(0, 5);

  return (
    <div className="mx-auto max-w-5xl">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-slate-800">工作台</h1>
        <p className="mt-1 text-sm text-slate-500">欢迎回来，管理你的简历、JD 与面试准备。</p>
      </div>

      {loading ? (
        <div className="text-sm text-slate-400">加载中…</div>
      ) : (
        <>
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
            {stats.map(({ label, value, icon: Icon, to, color }) => (
              <Link key={label} to={to} className="card p-5 transition-shadow hover:shadow-md">
                <div className={`mb-3 inline-flex h-10 w-10 items-center justify-center rounded-lg ${color}`}>
                  <Icon size={20} />
                </div>
                <div className="text-2xl font-semibold text-slate-800">{value}</div>
                <div className="text-sm text-slate-500">{label}</div>
              </Link>
            ))}
          </div>

          <div className="card mt-6 p-6">
            <h2 className="mb-4 text-base font-semibold text-slate-800">快速开始</h2>
            <div className="flex flex-wrap gap-3">
              <Link to="/resumes" className="btn-primary">
                <Upload size={16} /> 上传简历
              </Link>
              <Link to="/jds" className="btn-secondary">
                <Plus size={16} /> 新建 JD
              </Link>
              <Link to="/optimize" className="btn-secondary">
                <Sparkles size={16} /> 一键优化
              </Link>
              <Link to="/export" className="btn-secondary">
                <Download size={16} /> 导出简历
              </Link>
            </div>
          </div>

          <div className="card mt-6 p-6">
            <h2 className="mb-4 text-base font-semibold text-slate-800">最近活动</h2>
            {latest.length === 0 ? (
              <p className="text-sm text-slate-400">暂无活动，先上传一份简历或新建一个 JD 吧。</p>
            ) : (
              <ul className="space-y-3">
                {latest.map((x: any) => (
                  <li key={x.id} className="flex items-center justify-between border-b border-slate-100 pb-3 text-sm last:border-0">
                    <span className="text-slate-600">
                      <span className="badge bg-slate-100 text-slate-600 mr-2">{x.kind}</span>
                      {new Date(x.createdAt).toLocaleString('zh-CN')}
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </>
      )}
    </div>
  );
}
