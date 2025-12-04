import { useEffect, useState } from 'react'
import { ArrowLeft, ChevronRight, RefreshCw, CheckCircle2, Circle, Clock, Archive } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import type { Project, Task } from '@/types'

interface TasksProps {
  project: Project
  onBack: () => void
  onSelectTask: (task: Task) => void
}

const statusConfig = {
  open: { icon: Circle, label: 'Open', variant: 'outline' as const },
  in_progress: { icon: Clock, label: 'In Progress', variant: 'warning' as const },
  completed: { icon: CheckCircle2, label: 'Completed', variant: 'success' as const },
  archived: { icon: Archive, label: 'Archived', variant: 'secondary' as const },
}

export function Tasks({ project, onBack, onSelectTask }: TasksProps) {
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(true)

  const loadTasks = async () => {
    setLoading(true)
    const data = await window.electronAPI.getTasks(project.id)
    setTasks(data)
    setLoading(false)
  }

  useEffect(() => {
    loadTasks()
  }, [project.id])

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" onClick={onBack}>
          <ArrowLeft className="h-5 w-5" />
        </Button>
        <div className="flex-1">
          <h2 className="text-2xl font-bold tracking-tight">{project.name}</h2>
          <p className="text-muted-foreground">
            {tasks.length} task{tasks.length !== 1 ? 's' : ''}
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={loadTasks} disabled={loading}>
          <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      {tasks.length === 0 && !loading ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Circle className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground text-center">
              No tasks in this project.<br />
              Create a task using MCP tools.
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-3">
          {tasks.map((task) => {
            const status = statusConfig[task.status] || statusConfig.open
            const StatusIcon = status.icon

            return (
              <Card
                key={task.id}
                className="cursor-pointer hover:bg-accent/50 transition-colors"
                onClick={() => onSelectTask(task)}
              >
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <StatusIcon className="h-5 w-5 text-muted-foreground" />
                      <div>
                        <CardTitle className="text-base flex items-center gap-2">
                          {task.name}
                          <Badge variant={status.variant}>{status.label}</Badge>
                        </CardTitle>
                        <CardDescription className="text-xs font-mono">
                          {task.id}
                        </CardDescription>
                      </div>
                    </div>
                    <ChevronRight className="h-5 w-5 text-muted-foreground" />
                  </div>
                </CardHeader>
                {task.description && (
                  <CardContent className="pt-0">
                    <p className="text-sm text-muted-foreground line-clamp-2">
                      {task.description}
                    </p>
                  </CardContent>
                )}
              </Card>
            )
          })}
        </div>
      )}
    </div>
  )
}
