const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8888/api/v1'

export type BehatStepStatus = 'PASS' | 'FAIL' | 'ERROR' | 'SKIPPED'

export interface BehatStepResult {
  index: number
  text: string
  status: BehatStepStatus
  message: string
}

export interface BehatRun {
  run_id: string
  status: 'PASS' | 'FAIL' | 'ERROR'
  duration_ms: number
  steps: BehatStepResult[]
  message: string
  created_at?: string
}

async function send(path: string, init: RequestInit): Promise<unknown> {
  const token = sessionStorage.getItem('access_token') || localStorage.getItem('access_token')
  if (!token) throw new Error('Your session has expired. Please log in again.')

  let response: Response
  try {
    response = await fetch(`${API_BASE_URL}/scenarios${path}`, {
      ...init,
      cache: 'no-store',
      headers: { ...init.headers, Authorization: `Bearer ${token}` },
      signal: AbortSignal.timeout(90000),
    })
  } catch {
    throw new Error('The server could not be reached or the test timed out.')
  }

  if (response.status === 401) throw new Error('Your session has expired. Please log in again.')
  if (response.status === 403) throw new Error('Only System Analysts may run end-to-end tests.')
  if (response.status === 404) throw new Error('Scenario not found.')
  if (response.status === 400) throw new Error('Provide a base URL and every placeholder value.')
  if (!response.ok) throw new Error('The end-to-end test failed to start. Please try again.')

  const envelope: unknown = await response.json().catch(() => null)
  if (!envelope || typeof envelope !== 'object' || (envelope as { status?: unknown }).status !== true) {
    throw new Error('Invalid end-to-end test response.')
  }
  return (envelope as { data: unknown }).data
}

export async function runBehat(
  scenarioId: string,
  baseUrl: string,
  scenario: string,
  values: Record<string, string>,
): Promise<BehatRun> {
  return (await send(`/${encodeURIComponent(scenarioId)}/behat-runs`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ base_url: baseUrl, scenario, values }),
  })) as BehatRun
}

export async function listBehatRuns(scenarioId: string): Promise<BehatRun[]> {
  return (await send(`/${encodeURIComponent(scenarioId)}/behat-runs`, { method: 'GET' })) as BehatRun[]
}
