import { useEffect, useState } from 'react'
import { ArrowLeft, RefreshCw, FileText, Code, Lightbulb, MessageSquare, Link, FileSearch, FolderSearch, Package } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import type { Project, Task, Artifact } from '@/types'

interface ArtifactsProps {
  project: Project
  task: Task
  onBack: () => void
}

const typeConfig: Record<string, { icon: React.ElementType; label: string }> = {
  note: { icon: FileText, label: 'Note' },
  code: { icon: Code, label: 'Code' },
  decision: { icon: Lightbulb, label: 'Decision' },
  discussion: { icon: MessageSquare, label: 'Discussion' },
  reference: { icon: Link, label: 'Reference' },
  file_read: { icon: FileSearch, label: 'File Read' },
  file_list: { icon: FolderSearch, label: 'File List' },
  search: { icon: FileSearch, label: 'Search' },
  artifact: { icon: Package, label: 'Artifact' },
}

export function Artifacts({ project, task, onBack }: ArtifactsProps) {
  const [artifacts, setArtifacts] = useState<Artifact[]>([])
  const [loading, setLoading] = useState(true)
  const [expandedId, setExpandedId] = useState<string | null>(null)

  const loadArtifacts = async () => {
    setLoading(true)
    const data = await window.electronAPI.getArtifacts(project.id, task.id)
    setArtifacts(data)
    setLoading(false)
  }

  useEffect(() => {
    loadArtifacts()
  }, [project.id, task.id])

  const formatDate = (dateStr: string) => {
    if (!dateStr) return ''
    const date = new Date(dateStr)
    return date.toLocaleString()
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" onClick={onBack}>
          <ArrowLeft className="h-5 w-5" />
        </Button>
        <div className="flex-1">
          <h2 className="text-2xl font-bold tracking-tight">{task.name}</h2>
          <p className="text-muted-foreground">
            {artifacts.length} artifact{artifacts.length !== 1 ? 's' : ''} in {project.name}
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={loadArtifacts} disabled={loading}>
          <RefreshCw className={`h-4 w-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      {artifacts.length === 0 && !loading ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <FileText className="h-12 w-12 text-muted-foreground mb-4" />
            <p className="text-muted-foreground text-center">
              No artifacts for this task.<br />
              Save artifacts using MCP tools.
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-3">
          {artifacts.map((artifact) => {
            const typeInfo = typeConfig[artifact.type] || typeConfig.note
            const TypeIcon = typeInfo.icon
            const isExpanded = expandedId === artifact.id

            return (
              <Card
                key={artifact.id}
                className="cursor-pointer hover:bg-accent/50 transition-colors"
                onClick={() => setExpandedId(isExpanded ? null : artifact.id)}
              >
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <TypeIcon className="h-5 w-5 text-muted-foreground" />
                      <div>
                        <CardTitle className="text-base flex items-center gap-2">
                          <Badge variant="secondary">{typeInfo.label}</Badge>
                        </CardTitle>
                        <CardDescription className="text-xs">
                          {formatDate(artifact.createdAt)}
                        </CardDescription>
                      </div>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="pt-0">
                  <pre className={`text-sm font-mono whitespace-pre-wrap bg-muted rounded p-3 ${
                    isExpanded ? '' : 'line-clamp-4'
                  }`}>
                    {artifact.content}
                  </pre>
                  {!isExpanded && artifact.content.split('\n').length > 4 && (
                    <p className="text-xs text-muted-foreground mt-2">
                      Click to expand...
                    </p>
                  )}
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}
    </div>
  )
}
