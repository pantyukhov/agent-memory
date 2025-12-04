import { app, BrowserWindow, Tray, Menu, nativeImage, ipcMain } from 'electron'
import path from 'path'
import fs from 'fs'
import os from 'os'

let mainWindow: BrowserWindow | null = null
let tray: Tray | null = null

const VITE_DEV_SERVER_URL = process.env.VITE_DEV_SERVER_URL

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 900,
    height: 700,
    minWidth: 600,
    minHeight: 400,
    show: false,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
    },
    titleBarStyle: 'hiddenInset',
    trafficLightPosition: { x: 15, y: 15 },
  })

  if (VITE_DEV_SERVER_URL) {
    mainWindow.loadURL(VITE_DEV_SERVER_URL)
    mainWindow.webContents.openDevTools()
  } else {
    mainWindow.loadFile(path.join(__dirname, '../dist/index.html'))
  }

  mainWindow.on('close', (event) => {
    if (app.isQuitting !== true) {
      event.preventDefault()
      mainWindow?.hide()
    }
  })

  mainWindow.on('ready-to-show', () => {
    mainWindow?.show()
  })
}

function createTray() {
  // Create a simple icon (16x16 PNG would be ideal)
  const iconPath = path.join(__dirname, '../resources/icon.png')
  let icon: nativeImage

  if (fs.existsSync(iconPath)) {
    icon = nativeImage.createFromPath(iconPath)
  } else {
    // Create a simple colored icon if file doesn't exist
    icon = nativeImage.createEmpty()
  }

  // Resize for tray (macOS uses 16x16 or 22x22)
  if (!icon.isEmpty()) {
    icon = icon.resize({ width: 16, height: 16 })
  }

  tray = new Tray(icon.isEmpty() ? nativeImage.createFromDataURL('data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAAAdgAAAHYBTnsmCAAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAAEeSURBVDiNpZMxTsNAEEX/7NpJQUGDhDgCJ+AedNyBI9ByAo5AR8cRKOkoKCgQEgUSNrveocgmjrNrJzBSsdL8+fP/zqwFfrkUQE+SYUkvks5IktQFlgGukubATdJI0pIkSTPgi6SapDFwDEySVJJmkq6BoyTqum4OXALHSeq6LoArgWm3jYCjJHVdd0BSBS5IrpN0f/W7BDpJHiRtJU9IAhwm2QAek9y9dvsHnCXpAEtg0HXdOsmjpC2w1XXdHXCR5H6SpZuVNPiLgX9gBfSBdZLHJO+S/gGb3daIWgFn+yJqtZIWwBK4T7JJsgFWwKKu66+SrIHlFwPboiTPkm6SPCWpJdmBzyS7rwb+ZWoF9JNsk7xKWgN/AOXCVQePqDdoAAAAAElFTkSuQmCC') : icon)

  tray.setToolTip('Agent Memory')

  const contextMenu = Menu.buildFromTemplate([
    {
      label: 'Open',
      click: () => {
        mainWindow?.show()
      }
    },
    { type: 'separator' },
    {
      label: 'Quit',
      click: () => {
        (app as any).isQuitting = true
        app.quit()
      }
    }
  ])

  tray.setContextMenu(contextMenu)

  tray.on('click', () => {
    if (mainWindow?.isVisible()) {
      mainWindow.hide()
    } else {
      mainWindow?.show()
    }
  })
}

// IPC Handlers
function setupIPC() {
  const tasksPath = path.join(os.homedir(), '.agent-memory', 'tasks')

  ipcMain.handle('get-projects', async () => {
    try {
      if (!fs.existsSync(tasksPath)) {
        return []
      }

      const entries = fs.readdirSync(tasksPath, { withFileTypes: true })
      const projects = []

      for (const entry of entries) {
        if (entry.isDirectory()) {
          const projectPath = path.join(tasksPath, entry.name, 'project.json')
          if (fs.existsSync(projectPath)) {
            const data = JSON.parse(fs.readFileSync(projectPath, 'utf-8'))
            projects.push(data)
          }
        }
      }

      return projects
    } catch (error) {
      console.error('Error reading projects:', error)
      return []
    }
  })

  ipcMain.handle('get-tasks', async (_, projectId: string) => {
    try {
      const projectPath = path.join(tasksPath, projectId)
      if (!fs.existsSync(projectPath)) {
        return []
      }

      const entries = fs.readdirSync(projectPath, { withFileTypes: true })
      const tasks = []

      for (const entry of entries) {
        if (entry.isDirectory() && entry.name !== 'artifacts') {
          const taskPath = path.join(projectPath, entry.name, 'task.json')
          if (fs.existsSync(taskPath)) {
            const data = JSON.parse(fs.readFileSync(taskPath, 'utf-8'))
            tasks.push(data)
          }
        }
      }

      return tasks
    } catch (error) {
      console.error('Error reading tasks:', error)
      return []
    }
  })

  ipcMain.handle('get-artifacts', async (_, projectId: string, taskId: string) => {
    try {
      const projectPath = path.join(tasksPath, projectId)
      const entries = fs.readdirSync(projectPath, { withFileTypes: true })

      // Find task directory (may have status prefix like [open]-task-id)
      let taskDir = ''
      for (const entry of entries) {
        if (entry.isDirectory() && entry.name.includes(taskId)) {
          taskDir = entry.name
          break
        }
      }

      if (!taskDir) return []

      const artifactsPath = path.join(projectPath, taskDir, 'artifacts')
      if (!fs.existsSync(artifactsPath)) {
        return []
      }

      const files = fs.readdirSync(artifactsPath)
      const artifacts = []

      for (const file of files) {
        if (file.endsWith('.md')) {
          const filePath = path.join(artifactsPath, file)
          const content = fs.readFileSync(filePath, 'utf-8')

          // Parse frontmatter
          const frontmatterMatch = content.match(/^---\n([\s\S]*?)\n---\n([\s\S]*)$/)
          if (frontmatterMatch) {
            const frontmatter = frontmatterMatch[1]
            const body = frontmatterMatch[2]

            // Simple YAML parsing
            const metadata: Record<string, string> = {}
            frontmatter.split('\n').forEach(line => {
              const [key, ...valueParts] = line.split(':')
              if (key && valueParts.length) {
                metadata[key.trim()] = valueParts.join(':').trim()
              }
            })

            artifacts.push({
              id: metadata.id || file,
              type: metadata.type || 'note',
              content: body.trim(),
              createdAt: metadata.created_at || '',
              projectId,
              taskId
            })
          }
        }
      }

      return artifacts.sort((a, b) =>
        new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
      )
    } catch (error) {
      console.error('Error reading artifacts:', error)
      return []
    }
  })

  ipcMain.handle('get-binary-path', async () => {
    // Try common locations
    const locations = [
      '/usr/local/bin/agent-memory',
      path.join(os.homedir(), '.local', 'bin', 'agent-memory'),
      path.join(process.cwd(), 'build', 'agent-memory'),
    ]

    for (const loc of locations) {
      if (fs.existsSync(loc)) {
        return loc
      }
    }

    return '/usr/local/bin/agent-memory'
  })

  ipcMain.handle('copy-to-clipboard', async (_, text: string) => {
    const { clipboard } = await import('electron')
    clipboard.writeText(text)
    return true
  })
}

app.whenReady().then(() => {
  createWindow()
  createTray()
  setupIPC()

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow()
    } else {
      mainWindow?.show()
    }
  })
})

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit()
  }
})

app.on('before-quit', () => {
  (app as any).isQuitting = true
})
