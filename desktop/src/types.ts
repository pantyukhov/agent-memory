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

declare global {
  interface Window {
    electronAPI: {
      getProjects: () => Promise<Project[]>
      getTasks: (projectId: string) => Promise<Task[]>
      getArtifacts: (projectId: string, taskId: string) => Promise<Artifact[]>
      getBinaryPath: () => Promise<string>
      copyToClipboard: (text: string) => Promise<boolean>
    }
  }
}
