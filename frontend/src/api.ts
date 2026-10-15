export interface Section {
  title: string;
  bullets: string[];
}

export interface ResumeContent {
  name: string;
  title: string;
  contact: string;
  summary: string;
  sections: Section[];
}

export interface Resume {
  id: string;
  title: string;
  originalName: string;
  mime: string;
  sourceText: string;
  content: ResumeContent;
  createdAt: string;
  updatedAt: string;
}

export interface JD {
  id: string;
  title: string;
  company: string;
  description: string;
  responsibilities: string;
  requirements: string;
  plusPoints: string;
  createdAt: string;
  updatedAt: string;
}

export interface DiffItem {
  section: string;
  original: string;
  adjusted: string;
  matched: string[];
}

export interface ResumeVersion {
  id: string;
  resumeId: string;
  jdId: string | null;
  label: string;
  content: ResumeContent;
  diff: DiffItem[];
  createdAt: string;
}

export interface Dimension {
  name: string;
  score: number;
  comment: string;
}

export interface Evaluation {
  id: string;
  resumeId: string;
  jdId: string;
  overallScore: number;
  dimensions: Dimension[];
  suggestions: string[];
  createdAt: string;
}

export interface QuestionItem {
  question: string;
  hint: string;
}

export interface QuestionCategory {
  category: string;
  items: QuestionItem[];
}

export interface InterviewPrepContent {
  summary: string;
  questions: QuestionCategory[];
  studyTopics: string[];
  projectTips: string[];
}

export interface InterviewPrep {
  id: string;
  jdId: string;
  content: InterviewPrepContent;
  createdAt: string;
}

export interface Settings {
  llmEnabled: boolean;
  llmBaseUrl: string;
  llmModel: string;
  llmApiKey: string;
  pdfFontPath: string;
}

export interface Meta {
  app: string;
  engine: string;
  version: string;
}

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(body.error || res.statusText);
  }
  return res.json() as Promise<T>;
}

export const api = {
  meta: () => request<Meta>('/api/meta'),

  listResumes: () => request<Resume[]>('/api/resumes'),
  getResume: (id: string) => request<Resume>(`/api/resumes/${id}`),
  updateResume: (id: string, data: Partial<Resume>) => request<Resume>(`/api/resumes/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteResume: (id: string) => request<{ ok: boolean }>(`/api/resumes/${id}`, { method: 'DELETE' }),
  uploadResume: async (file: File) => {
    const form = new FormData();
    form.append('file', file);
    const res = await fetch('/api/resumes', { method: 'POST', body: form });
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: res.statusText }));
      throw new Error(body.error || res.statusText);
    }
    return res.json() as Promise<Resume>;
  },

  listJDs: () => request<JD[]>('/api/jds'),
  createJD: (data: Partial<JD>) => request<JD>('/api/jds', { method: 'POST', body: JSON.stringify(data) }),
  updateJD: (id: string, data: Partial<JD>) => request<JD>(`/api/jds/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteJD: (id: string) => request<{ ok: boolean }>(`/api/jds/${id}`, { method: 'DELETE' }),

  optimize: (resumeId: string, jdId: string) => request<ResumeVersion>('/api/optimize', { method: 'POST', body: JSON.stringify({ resumeId, jdId }) }),
  listVersions: (resumeId?: string) => request<ResumeVersion[]>(`/api/versions${resumeId ? `?resumeId=${resumeId}` : ''}`),
  deleteVersion: (id: string) => request<{ ok: boolean }>(`/api/versions/${id}`, { method: 'DELETE' }),

  evaluate: (resumeId: string, jdId: string) => request<Evaluation>('/api/evaluations', { method: 'POST', body: JSON.stringify({ resumeId, jdId }) }),
  listEvaluations: (resumeId?: string, jdId?: string) => {
    const q = new URLSearchParams();
    if (resumeId) q.set('resumeId', resumeId);
    if (jdId) q.set('jdId', jdId);
    return request<Evaluation[]>(`/api/evaluations?${q.toString()}`);
  },

  interviewPrep: (jdId: string) => request<InterviewPrep>('/api/interview-prep', { method: 'POST', body: JSON.stringify({ jdId }) }),
  listInterviewPreps: (jdId?: string) => request<InterviewPrep[]>(`/api/interview-prep${jdId ? `?jdId=${jdId}` : ''}`),

  getSettings: () => request<Settings>('/api/settings'),
  updateSettings: (data: Settings) => request<Settings>('/api/settings', { method: 'PUT', body: JSON.stringify(data) }),
};

export function exportUrl(params: { resumeId?: string; versionId?: string; format: string }): string {
  const q = new URLSearchParams({ format: params.format });
  if (params.versionId) q.set('versionId', params.versionId);
  if (params.resumeId) q.set('resumeId', params.resumeId);
  return `/api/export?${q.toString()}`;
}

export function emptyContent(): ResumeContent {
  return { name: '', title: '', contact: '', summary: '', sections: [] };
}
