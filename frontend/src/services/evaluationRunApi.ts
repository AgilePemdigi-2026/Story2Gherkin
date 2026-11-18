const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8888/api/v1'

export type EvaluationRunStatus = 'QUEUED' | 'RUNNING' | 'COMPLETED' | 'FAILED'

export interface EvaluationRunCreated {
  run_id: string
  report_code: string
  status: EvaluationRunStatus
  total_cases: number
}

export interface EvaluationRunProgress {
  run_id: string
  status: EvaluationRunStatus
  total_cases: number
  processed_cases: number
  error_message: string | null
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object'
}

async function send(path: string, init: RequestInit, signal?: AbortSignal) {
  const token = sessionStorage.getItem('access_token') || localStorage.getItem('access_token')
  if (!token) throw new Error('Sesi berakhir. Silakan login kembali.')

  let response: Response
  try {
    const timeout = AbortSignal.timeout(30000)
    response = await fetch(`${API_BASE_URL}/evaluation-runs${path}`, {
      ...init,
      signal: signal ? AbortSignal.any([signal, timeout]) : timeout,
      cache: 'no-store',
      headers: { ...init.headers, Authorization: `Bearer ${token}` },
    })
  } catch (error) {
    if (signal?.aborted) throw error
    throw new Error('Server tidak dapat dihubungi. Silakan coba lagi.', { cause: error })
  }

  if (response.status === 401) throw new Error('Sesi berakhir. Silakan login kembali.')
  if (response.status === 403) throw new Error('Hanya Quality Assurance yang dapat menjalankan evaluasi.')
  if (response.status === 409) throw new Error('Masih ada evaluasi yang berjalan. Tunggu hingga selesai.')
  if (response.status === 400) throw new Error('Dataset tidak valid atau tidak memiliki kasus.')
  if (response.status === 404) throw new Error('Dataset atau evaluasi tidak ditemukan.')
  if (!response.ok) throw new Error('Permintaan evaluasi gagal. Silakan coba lagi.')

  const envelope: unknown = await response.json().catch(() => null)
  if (!isRecord(envelope) || envelope.status !== true || !isRecord(envelope.data)) {
    throw new Error('Respons evaluasi tidak valid.')
  }
  return envelope.data
}

export async function createEvaluationRun(
  goldenDatasetId: string,
  signal?: AbortSignal,
): Promise<EvaluationRunCreated> {
  const data = await send('', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ golden_dataset_id: goldenDatasetId }),
  }, signal)

  if (
    typeof data.run_id !== 'string' ||
    typeof data.report_code !== 'string' ||
    typeof data.total_cases !== 'number'
  ) {
    throw new Error('Respons evaluasi tidak valid.')
  }

  return data as unknown as EvaluationRunCreated
}

export async function getEvaluationRunProgress(
  runId: string,
  signal?: AbortSignal,
): Promise<EvaluationRunProgress> {
  const data = await send(`/${encodeURIComponent(runId)}/progress`, { method: 'GET' }, signal)

  if (
    typeof data.run_id !== 'string' ||
    typeof data.status !== 'string' ||
    typeof data.total_cases !== 'number' ||
    typeof data.processed_cases !== 'number'
  ) {
    throw new Error('Respons evaluasi tidak valid.')
  }

  return data as unknown as EvaluationRunProgress
}
