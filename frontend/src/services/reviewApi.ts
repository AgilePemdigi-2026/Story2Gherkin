const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8888/api/v1'

export interface ValidationIssue { message: string; line: number; column: number }
export interface ScenarioValidation {
  valid: boolean
  feature_name: string
  checks: { feature: boolean; scenario: boolean; given: boolean; when: boolean; then: boolean }
  issues: ValidationIssue[]
  used_elements: string[]
  not_covered_elements: string[]
  coverage_percent: number
  coverage_available?: boolean
}
export interface ScenarioSummary {
  id: string
  scenario_code?: string
  feature_name: string
  source_type: string
  status: string
  updated_at: string
}
export interface ReviewScenario extends ScenarioSummary {
  gherkin: string
  gui_elements: string[]
  validation: ScenarioValidation
}

export class ReviewApiError extends Error {
  status: number
  validation?: ScenarioValidation

  constructor(message: string, status = 0, validation?: ScenarioValidation) {
    super(message)
    this.name = 'ReviewApiError'
    this.status = status
    this.validation = validation
  }
}

async function request(path: string, options: RequestInit = {}) {
  const token = sessionStorage.getItem('access_token') || localStorage.getItem('access_token')
  if (!token) throw new ReviewApiError('Sesi berakhir. Silakan login kembali.', 401)
  let response: Response
  try {
    const timeout = AbortSignal.timeout(15000)
    response = await fetch(`${API_BASE_URL}/scenarios${path}`, {
      ...options,
      signal: options.signal ? AbortSignal.any([options.signal, timeout]) : timeout,
      cache: 'no-store',
      headers: { ...options.headers, Authorization: `Bearer ${token}` },
    })
  } catch (error) {
    if (options.signal?.aborted) throw error
    throw new ReviewApiError('Server tidak dapat dihubungi atau permintaan melewati batas waktu. Coba lagi.')
  }
  if (!response.ok) {
    const result = await response.json().catch(() => null)
    const message = response.status === 401
      ? 'Sesi berakhir. Silakan login kembali.'
      : response.status === 422 && result?.data?.validation
        ? 'Perubahan belum disimpan. Perbaiki kesalahan pada struktur Gherkin terlebih dahulu.'
        : result?.error || result?.message || 'Permintaan skenario gagal.'
    throw new ReviewApiError(message, response.status, result?.data?.validation)
  }
  return response
}

async function jsonRequest<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await request(path, options)
  const result = await response.json()
  if (result.status !== true || result.data === undefined) {
    throw new ReviewApiError('Respons skenario tidak valid.')
  }
  return result.data as T
}

export function listReviewScenarios(signal?: AbortSignal) {
  return jsonRequest<ScenarioSummary[]>('', { signal })
}
export function getReviewScenario(id: string, signal?: AbortSignal) {
  return jsonRequest<ReviewScenario>(`/${encodeURIComponent(id)}`, { signal })
}
export function saveCorrection(scenario: ReviewScenario, gherkin: string) {
  return jsonRequest<ReviewScenario>(`/${encodeURIComponent(scenario.id)}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ gherkin, updated_at: scenario.updated_at }),
  })
}
export async function downloadScenario(scenario: ReviewScenario) {
  // The server loads and revalidates the stored version; never send draft text.
  const response = await request(`/${encodeURIComponent(scenario.id)}/download`)
  const blob = await response.blob()
  const slug = scenario.feature_name.toLowerCase().replace(/[^a-z0-9]/g, '-').slice(0, 70).replace(/^-+|-+$/g, '')
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = `scenario-${slug || 'gherkin'}.feature`
  document.body.appendChild(anchor)
  anchor.click()
  anchor.remove()
  window.setTimeout(() => URL.revokeObjectURL(url), 1000)
}
