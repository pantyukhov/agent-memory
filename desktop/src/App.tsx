import { useState } from 'react'
import { Settings, Database } from 'lucide-react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Setup } from '@/pages/Setup'
import { Projects } from '@/pages/Projects'
import { Tasks } from '@/pages/Tasks'
import { Artifacts } from '@/pages/Artifacts'
import type { Project, Task } from '@/types'

type DataView = 'projects' | 'tasks' | 'artifacts'

function App() {
  const [tab, setTab] = useState('setup')
  const [dataView, setDataView] = useState<DataView>('projects')
  const [selectedProject, setSelectedProject] = useState<Project | null>(null)
  const [selectedTask, setSelectedTask] = useState<Task | null>(null)

  const handleSelectProject = (project: Project) => {
    setSelectedProject(project)
    setDataView('tasks')
  }

  const handleSelectTask = (task: Task) => {
    setSelectedTask(task)
    setDataView('artifacts')
  }

  const handleBackToProjects = () => {
    setSelectedProject(null)
    setSelectedTask(null)
    setDataView('projects')
  }

  const handleBackToTasks = () => {
    setSelectedTask(null)
    setDataView('tasks')
  }

  const renderDataContent = () => {
    if (dataView === 'artifacts' && selectedProject && selectedTask) {
      return (
        <Artifacts
          project={selectedProject}
          task={selectedTask}
          onBack={handleBackToTasks}
        />
      )
    }

    if (dataView === 'tasks' && selectedProject) {
      return (
        <Tasks
          project={selectedProject}
          onBack={handleBackToProjects}
          onSelectTask={handleSelectTask}
        />
      )
    }

    return <Projects onSelectProject={handleSelectProject} />
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Titlebar drag region */}
      <div className="h-8 drag-region" />

      <div className="container max-w-3xl mx-auto px-4 pb-8">
        <Tabs value={tab} onValueChange={setTab}>
          <TabsList className="grid w-full grid-cols-2 no-drag">
            <TabsTrigger value="setup" className="gap-2">
              <Settings className="h-4 w-4" />
              Setup
            </TabsTrigger>
            <TabsTrigger value="data" className="gap-2">
              <Database className="h-4 w-4" />
              Data
            </TabsTrigger>
          </TabsList>

          <TabsContent value="setup" className="mt-6">
            <Setup />
          </TabsContent>

          <TabsContent value="data" className="mt-6">
            {renderDataContent()}
          </TabsContent>
        </Tabs>
      </div>
    </div>
  )
}

export default App
