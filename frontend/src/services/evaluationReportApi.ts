const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8888/api/v1'

export type EvaluationRunStatus = 'QUEUED' | 'RUNNING' | 'COMPLETED' | 'FAILED'
export type EvaluationCaseStatus = 'PASS' | 'WARN' | 'FAIL'

export interface EvaluationRun {
  run_id: string
  dataset_name: string
  created_at: string
  total_cases: number
  status: EvaluationRunStatus
}

export interface EvaluationRunList {
  runs: EvaluationRun[]
  total_pages: number
}

interface MethodSummary {
  validation_percent: number | null
  gui_coverage_percent: number | null
}

export interface MethodResult {
  status: EvaluationCaseStatus
  gui_coverage_percent: number | null
}

export interface EvaluationCase {
  case_id: string
  title: string
  expected_gherkin: string
  llm: MethodResult | null
  rule_based: MethodResult | null
}

export interface EvaluationReport extends EvaluationRun {
  metrics: { llm: MethodSummary; rule_based: MethodSummary } | null
  cases: EvaluationCase[]
  error_message: string | null
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object'
}

function isCount(value: unknown): value is number {
  return typeof value === 'number' && Number.isSafeInteger(value) && value >= 0
}

function isPercent(value: unknown) {
  return value === null || (typeof value === 'number' && Number.isFinite(value) && value >= 0 && value <= 100)
}

function isRun(value: unknown): value is EvaluationRun {
  return isRecord(value) && typeof value.run_id === 'string' && value.run_id.length > 0 &&
    typeof value.dataset_name === 'string' && typeof value.created_at === 'string' &&
    isCount(value.total_cases) &&
    (value.status === 'QUEUED' || value.status === 'RUNNING' || value.status === 'COMPLETED' || value.status === 'FAILED')
}

function isMethodSummary(value: unknown) {
  return isRecord(value) && isPercent(value.validation_percent) && isPercent(value.gui_coverage_percent)
}

function isMethodResult(value: unknown) {
  return value === null || (isRecord(value) &&
    (value.status === 'PASS' || value.status === 'WARN' || value.status === 'FAIL') &&
    isPercent(value.gui_coverage_percent))
}

function isCase(value: unknown): value is EvaluationCase {
  return isRecord(value) && typeof value.case_id === 'string' && typeof value.title === 'string' &&
    typeof value.expected_gherkin === 'string' && isMethodResult(value.llm) && isMethodResult(value.rule_based)
}

function isReport(value: unknown): value is EvaluationReport {
  if (!isRecord(value)) return false
  const { metrics, cases, error_message } = value
  return isRun(value) && Array.isArray(cases) && cases.every(isCase) &&
    (error_message === null || typeof error_message === 'string') &&
    (metrics === null || (isRecord(metrics) && isMethodSummary(metrics.llm) && isMethodSummary(metrics.rule_based)))
}

async function request(path: string, signal: AbortSignal) {
  const token = sessionStorage.getItem('access_token') || localStorage.getItem('access_token')
  if (!token) throw new Error('Sesi berakhir. Silakan login kembali.')

  let response: Response
  try {
    response = await fetch(`${API_BASE_URL}/evaluation-runs${path}`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: 'no-store',
      signal: AbortSignal.any([signal, AbortSignal.timeout(30000)]),
    })
  } catch (error) {
    if (signal.aborted) throw error
    throw new Error('Server tidak dapat dihubungi. Silakan coba lagi.', { cause: error })
  }

  if (response.status === 400) throw new Error('ID laporan atau nomor halaman tidak valid.')
  if (response.status === 401) throw new Error('Sesi berakhir. Silakan login kembali.')
  if (response.status === 403) throw new Error('Hanya Quality Assurance yang dapat membuka laporan evaluasi.')
  if (response.status === 404) throw new Error('Laporan evaluasi belum tersedia atau tidak ditemukan.')
  if (response.status === 409) throw new Error('Laporan belum siap diunduh.')
  if (!response.ok) throw new Error('Permintaan laporan evaluasi gagal. Silakan coba lagi.')
  return response
}

async function jsonData(path: string, signal: AbortSignal): Promise<unknown> {
  const response = await request(path, signal)
  const envelope: unknown = await response.json().catch(() => null)
  if (!isRecord(envelope) || envelope.status !== true) {
    throw new Error('Respons laporan evaluasi tidak valid.')
  }
  return envelope.data
}

export async function listEvaluationRuns(page: number, signal: AbortSignal): Promise<EvaluationRunList> {
  const data = await jsonData(`?page=${page}&page_size=10`, signal)
  if (!isRecord(data) || !Array.isArray(data.runs) || !data.runs.every(isRun) || !isCount(data.total_pages)) {
    throw new Error('Daftar hasil evaluasi tidak valid.')
  }
  return { runs: data.runs, total_pages: data.total_pages }
}

export async function getEvaluationReport(runId: string, signal: AbortSignal) {
  const data = await jsonData(`/${encodeURIComponent(runId)}`, signal)
  if (!isReport(data) || data.run_id !== runId) throw new Error('Detail hasil evaluasi tidak valid.')
  return data
}

export async function downloadEvaluationReport(runId: string, signal: AbortSignal) {
  const response = await request(`/${encodeURIComponent(runId)}/download`, signal)
  const contentType = response.headers.get('content-type')?.split(';')[0].trim()
  if (contentType !== 'text/csv') throw new Error('Server tidak mengembalikan file CSV yang valid.')
  const file = await response.blob()
  if (signal.aborted) return

  const url = URL.createObjectURL(file)
  const link = document.createElement('a')
  link.href = url
  link.download = `evaluation-report-${runId}.csv`
  document.body.append(link)
  link.click()
  link.remove()
  window.setTimeout(() => URL.revokeObjectURL(url), 1000)
}
