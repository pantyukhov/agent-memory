import { contextBridge, ipcRenderer } from 'electron'

export interface Project {
  id: string
  name: string
  description: string
  workspacePath: string
  metadata: Record<string, string>
  createdAt: string
  updatedAt: string
}

export interface Task {
  id: string
  projectId: string
  name: string
  description: string
  status: 'open' | 'in_progress' | 'completed' | 'archived'
  workspacePath: string
  metadata: Record<string, string>
  createdAt: string
  updatedAt: string
}

export interface Artifact {
  id: string
  projectId: string
  taskId: string
  type: string
  content: string
  createdAt: string
}

contextBridge.exposeInMainWorld('electronAPI', {
  getProjects: (): Promise<Project[]> => ipcRenderer.invoke('get-projects'),
  getTasks: (projectId: string): Promise<Task[]> => ipcRenderer.invoke('get-tasks', projectId),
  getArtifacts: (projectId: string, taskId: string): Promise<Artifact[]> =>
    ipcRenderer.invoke('get-artifacts', projectId, taskId),
  getBinaryPath: (): Promise<string> => ipcRenderer.invoke('get-binary-path'),
  copyToClipboard: (text: string): Promise<boolean> => ipcRenderer.invoke('copy-to-clipboard', text),
})
