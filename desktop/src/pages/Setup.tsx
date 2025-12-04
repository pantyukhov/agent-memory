import { useEffect, useState } from 'react'
import { Copy, Check, Terminal } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

interface ConfigItem {
  name: string
  description: string
  config: object
}

export function Setup() {
  const [binaryPath, setBinaryPath] = useState('/usr/local/bin/agent-memory')
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null)

  useEffect(() => {
    window.electronAPI.getBinaryPath().then(setBinaryPath)
  }, [])

  const configs: ConfigItem[] = [
    {
      name: 'Claude Desktop',
      description: 'Add to ~/Library/Application Support/Claude/claude_desktop_config.json',
      config: {
        mcpServers: {
          'agent-memory': {
            command: binaryPath,
            args: ['-tasks-path', '~/.agent-memory/tasks']
          }
        }
      }
    },
    {
      name: 'Claude Code',
      description: 'Add to your Claude Code MCP settings',
      config: {
        mcpServers: {
          'agent-memory': {
            command: binaryPath,
            args: []
          }
        }
      }
    },
    {
      name: 'Cursor',
      description: 'Add to Cursor MCP configuration',
      config: {
        mcpServers: {
          'agent-memory': {
            command: binaryPath,
            args: []
          }
        }
      }
    }
  ]

  const handleCopy = async (index: number, config: object) => {
    const text = JSON.stringify(config, null, 2)
    await window.electronAPI.copyToClipboard(text)
    setCopiedIndex(index)
    setTimeout(() => setCopiedIndex(null), 2000)
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Setup</h2>
        <p className="text-muted-foreground">
          Copy MCP configuration for your preferred client
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <Terminal className="h-4 w-4" />
            Binary Path
          </CardTitle>
        </CardHeader>
        <CardContent>
          <code className="block rounded bg-muted px-3 py-2 text-sm font-mono">
            {binaryPath}
          </code>
        </CardContent>
      </Card>

      <div className="grid gap-4">
        {configs.map((item, index) => (
          <Card key={item.name}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle className="text-base">{item.name}</CardTitle>
                  <CardDescription>{item.description}</CardDescription>
                </div>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => handleCopy(index, item.config)}
                  className="gap-2"
                >
                  {copiedIndex === index ? (
                    <>
                      <Check className="h-4 w-4" />
                      Copied!
                    </>
                  ) : (
                    <>
                      <Copy className="h-4 w-4" />
                      Copy
                    </>
                  )}
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              <pre className="rounded bg-muted p-3 text-xs font-mono overflow-x-auto">
                {JSON.stringify(item.config, null, 2)}
              </pre>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}
