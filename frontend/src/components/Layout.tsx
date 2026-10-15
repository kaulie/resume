import { useEffect, useState } from 'react';
import { NavLink, Outlet } from 'react-router-dom';
import {
  LayoutDashboard,
  FileText,
  Briefcase,
  Sparkles,
  Gauge,
  GraduationCap,
  Download,
  BrainCircuit,
  Settings,
} from 'lucide-react';
import { api, Meta } from '../api';

const NAV = [
  { to: '/', label: '工作台', icon: LayoutDashboard },
  { to: '/resumes', label: '简历管理', icon: FileText },
  { to: '/jds', label: 'JD 管理', icon: Briefcase },
  { to: '/optimize', label: '简历优化', icon: Sparkles },
  { to: '/evaluate', label: '简历评价', icon: Gauge },
  { to: '/interview', label: '面试备战', icon: GraduationCap },
  { to: '/export', label: '导出中心', icon: Download },
  { to: '/settings', label: '系统设置', icon: Settings },
];

export default function Layout() {
  const [meta, setMeta] = useState<Meta | null>(null);

  useEffect(() => {
    api.meta().then(setMeta).catch(() => {});
  }, []);

  return (
    <div className="flex min-h-screen">
      <aside className="fixed inset-y-0 left-0 flex w-60 flex-col border-r border-slate-200 bg-white">
        <div className="flex items-center gap-2 border-b border-slate-200 px-5 py-5">
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-brand-600 text-white">
            <BrainCircuit size={20} />
          </div>
          <div>
            <div className="text-sm font-semibold text-slate-800">简历工作台</div>
            <div className="text-xs text-slate-400">个人简历动态更新系统</div>
          </div>
        </div>
        <nav className="flex-1 space-y-1 px-3 py-4">
          {NAV.map(({ to, label, icon: Icon }) => (
            <NavLink
              key={to}
              to={to}
              end={to === '/'}
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors ${
                  isActive ? 'bg-brand-50 text-brand-700' : 'text-slate-600 hover:bg-slate-100'
                }`
              }
            >
              <Icon size={18} />
              {label}
            </NavLink>
          ))}
        </nav>
        <div className="border-t border-slate-200 px-5 py-4">
          <div className="flex items-center gap-2 text-xs text-slate-400">
            <span className="badge bg-slate-100 text-slate-500">{meta?.engine || '…'}</span>
            <span>引擎模式</span>
          </div>
        </div>
      </aside>
      <main className="ml-60 flex-1 p-8">
        <Outlet />
      </main>
    </div>
  );
}
