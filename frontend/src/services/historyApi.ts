
const API_BASE_URL =
  import.meta.env.VITE_API_URL ||
  'http://localhost:8888/api/v1'

// ------------------------------------
// TYPES
// ------------------------------------

export interface HistorySummary {
  id: string
  user_story: string
  request_status: string
  created_at: string
  updated_at: string
}

export interface HistoryListResponse {
  history: HistorySummary[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface HistoryScenario {
  id: string
  scenario_code: string
  feature_name: string
  gherkin: string
  source_type: string
  status: string
  coverage_percent: number
  updated_at: string
}

export interface HistoryDetail extends HistorySummary {
  gui_elements: string[]
  free_scenario: string
  error_message: string
  scenario: HistoryScenario | null
}

interface ApiResponse<T> {
  status: boolean
  message?: string
  error?: string
  data?: T
}

// ------------------------------------
// ERROR HANDLING
// ------------------------------------

export class HistoryApiError extends Error {
  status: number

  constructor(message: string, status = 0) {
    super(message)
    this.name = 'HistoryApiError'
    this.status = status
  }
}

// ------------------------------------
// AUTHENTICATED REQUEST
// ------------------------------------

async function historyRequest<T>(
  path: string,
  signal?: AbortSignal
): Promise<T> {
  const accessToken =
    sessionStorage.getItem('access_token') ||
    localStorage.getItem('access_token')

  if (!accessToken) {
    throw new HistoryApiError(
      'Sesi berakhir. Silakan login kembali.',
      401
    )
  }

  let response: Response

  try {
    response = await fetch(
      `${API_BASE_URL}/history${path}`,
      {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${accessToken}`,
          Accept: 'application/json',
        },
        cache: 'no-store',
        signal,
      }
    )
  } catch (error) {
    if (signal?.aborted) throw error

    throw new HistoryApiError(
      'Tidak dapat terhubung ke server.'
    )
  }

  const result: ApiResponse<T> | null =
    await response.json().catch(() => null)

  if (response.status === 401) {
    throw new HistoryApiError(
      'Sesi berakhir. Silakan login kembali.',
      401
    )
  }

  if (response.status === 403) {
    throw new HistoryApiError(
      'Hanya System Analyst yang dapat mengakses History.',
      403
    )
  }

  if (response.status === 404) {
    throw new HistoryApiError(
      'Riwayat tidak ditemukan.',
      404
    )
  }

  if (!response.ok || result?.status !== true) {
    throw new HistoryApiError(
      result?.error ||
        result?.message ||
        'Gagal mengambil riwayat Generate.',
      response.status
    )
  }

  if (result.data === undefined || result.data === null) {
    throw new HistoryApiError(
      'Respons History dari server tidak valid.'
    )
  }

  return result.data
}

// ------------------------------------
// GET HISTORY LIST
// ------------------------------------

export function listGenerationHistory(
  page = 1,
  pageSize = 20,
  signal?: AbortSignal
): Promise<HistoryListResponse> {
  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize),
  })

  return historyRequest<HistoryListResponse>(
    `?${params.toString()}`,
    signal
  )
}

// ------------------------------------
// GET HISTORY DETAIL
// ------------------------------------

export function getGenerationHistoryDetail(
  id: string,
  signal?: AbortSignal
): Promise<HistoryDetail> {
  return historyRequest<HistoryDetail>(
    `/${encodeURIComponent(id)}`,
    signal
  )
}
