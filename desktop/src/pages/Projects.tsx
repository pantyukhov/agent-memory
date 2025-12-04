import { useEffect, useState } from 'react'
import { Folder, ChevronRight, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import type { Project } from '@/types'

interface ProjectsProps {
  onSelectProject: (project: Project) => void
}

export function Projects({ onSelectProject }: ProjectsProps) {
  const [projects, setProjects] = useState<Project[]>([])
  const [loading, setLoading] = useState(true)

  const loadProjects = async () => {
    setLoading(true)
    const data = await window.electronAPI.getProjects()
    setProjects(data)
    setLoading(false)
  }

  useEffect(() => {
    loadProjects()
  }, [])

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Projects</h2>
          <p className="text-muted-foreground">
            {projects.length} project{projects.length !== 1 ? 's' : ''} found
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={loadProjects} disabled={loading}>
          <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      {projects.length === 0 && !loading ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Folder className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground text-center">
              No projects found.<br />
              Create a project using MCP tools.
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-3">
          {projects.map((project) => (
            <Card
              key={project.id}
              className="cursor-pointer hover:bg-accent/50 transition-colors"
              onClick={() => onSelectProject(project)}
            >
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <Folder className="h-5 w-5 text-muted-foreground" />
                    <div>
                      <CardTitle className="text-base">{project.name}</CardTitle>
                      <CardDescription className="text-xs font-mono">
                        {project.id}
                      </CardDescription>
                    </div>
                  </div>
                  <ChevronRight className="h-5 w-5 text-muted-foreground" />
                </div>
              </CardHeader>
              {(project.description || project.workspacePath) && (
                <CardContent className="pt-0">
                  {project.description && (
                    <p className="text-sm text-muted-foreground mb-1">
                      {project.description}
                    </p>
                  )}
                  {project.workspacePath && (
                    <p className="text-xs font-mono text-muted-foreground">
                      {project.workspacePath}
                    </p>
                  )}
                </CardContent>
              )}
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
