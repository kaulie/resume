import { Routes, Route, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Resumes from './pages/Resumes';
import JDs from './pages/JDs';
import Optimize from './pages/Optimize';
import Evaluate from './pages/Evaluate';
import InterviewPrep from './pages/InterviewPrep';
import ExportCenter from './pages/ExportCenter';
import Settings from './pages/Settings';

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Dashboard />} />
        <Route path="/resumes" element={<Resumes />} />
        <Route path="/jds" element={<JDs />} />
        <Route path="/optimize" element={<Optimize />} />
        <Route path="/evaluate" element={<Evaluate />} />
        <Route path="/interview" element={<InterviewPrep />} />
        <Route path="/export" element={<ExportCenter />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  );
}
