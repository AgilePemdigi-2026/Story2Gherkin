const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8888/api/v1'

export interface GoldenDatasetSummary {
  id: string
  name: string
  description: string | null
  created_by: string
  item_count: number
  used_at: string | null
  created_at: string
  updated_at: string
}

export interface GoldenDatasetItem {
  id: string
  item_title: string
  user_story_text: string
  gui_elements: string[]
  free_scenario_text: string
  expected_gherkin: string
  created_at: string
}

export interface GoldenDatasetDetail extends GoldenDatasetSummary {
  items: GoldenDatasetItem[]
}

export interface GoldenDatasetList {
  datasets: GoldenDatasetSummary[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface GoldenDatasetListParams {
  search?: string
  page?: number
  pageSize?: number
  signal?: AbortSignal
}

export interface GoldenDatasetUpload {
  name: string
  description?: string
  file: File
}

export interface GoldenDatasetUpdate {
  name: string
  description?: string
  file?: File | null
}

export class GoldenDatasetApiError extends Error {
  status: number

  constructor(message: string, status = 0) {
    super(message)
    this.name = 'GoldenDatasetApiError'
    this.status = status
  }
}

function failureMessage(status: number, result: { message?: string; error?: unknown } | null) {
  if (status === 401) return 'Sesi berakhir. Silakan login kembali.'
  if (status === 403) return 'Hanya pengguna Quality Assurance yang dapat mengelola Golden Dataset.'
  if (status === 404) return 'Golden Dataset tidak ditemukan.'
  if (status === 409) return 'Golden Dataset sudah dipakai evaluasi sehingga tidak dapat diubah atau dihapus.'
  if (status === 413) return 'Ukuran unggahan melebihi batas 5 MB.'
  if (status === 415) return 'Unggahan harus dikirim sebagai multipart/form-data.'
  if (status === 502 || status === 503 || status === 504) return 'Validator Gherkin sedang tidak tersedia. Coba lagi.'
  if (typeof result?.error === 'string' && result.error) return result.error
  if (typeof result?.message === 'string' && result.message) return result.message
  return 'Permintaan Golden Dataset gagal.'
}

async function request(path: string, options: RequestInit = {}) {
  const token = sessionStorage.getItem('access_token') || localStorage.getItem('access_token')
  if (!token) throw new GoldenDatasetApiError('Sesi berakhir. Silakan login kembali.', 401)

  let response: Response
  try {
    const timeout = AbortSignal.timeout(30000)
    response = await fetch(`${API_BASE_URL}/golden-datasets${path}`, {
      ...options,
      signal: options.signal ? AbortSignal.any([options.signal, timeout]) : timeout,
      cache: 'no-store',
      headers: { ...options.headers, Authorization: `Bearer ${token}` },
    })
  } catch (error) {
    if (options.signal?.aborted) throw error
    throw new GoldenDatasetApiError('Server tidak dapat dihubungi atau permintaan melewati batas waktu. Coba lagi.')
  }

  if (!response.ok) {
    const result = await response.json().catch(() => null)
    throw new GoldenDatasetApiError(failureMessage(response.status, result), response.status)
  }

  return response
}

async function jsonRequest<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await request(path, options)
  const result = await response.json()
  if (result.status !== true || result.data === undefined) {
    throw new GoldenDatasetApiError('Respons Golden Dataset tidak valid.')
  }
  return result.data as T
}

export function listGoldenDatasets({ search, page = 1, pageSize = 10, signal }: GoldenDatasetListParams = {}) {
  const query = new URLSearchParams({ page: String(page), page_size: String(pageSize) })
  if (search?.trim()) query.set('search', search.trim())
  return jsonRequest<GoldenDatasetList>(`?${query}`, { signal })
}

export function getGoldenDataset(id: string, signal?: AbortSignal) {
  return jsonRequest<GoldenDatasetDetail>(`/${encodeURIComponent(id)}`, { signal })
}

export async function uploadGoldenDataset({ name, description, file }: GoldenDatasetUpload) {
  const form = new FormData()
  form.set('name', name)
  if (description?.trim()) form.set('description', description.trim())
  form.set('file', file)

  // No Content-Type header: the browser sets the multipart boundary.
  const response = await request('', { method: 'POST', body: form })
  const result = await response.json()
  if (result.status !== true || result.data === undefined) {
    throw new GoldenDatasetApiError('Respons Golden Dataset tidak valid.')
  }
  return result.data as GoldenDatasetSummary
}

export async function updateGoldenDataset(id: string, { name, description, file }: GoldenDatasetUpdate) {
  const form = new FormData()
  form.set('name', name)
  form.set('description', description ?? '')
  if (file) form.set('file', file)

  const response = await request(`/${encodeURIComponent(id)}`, { method: 'PUT', body: form })
  const result = await response.json()
  if (result.status !== true || result.data === undefined) {
    throw new GoldenDatasetApiError('Respons Golden Dataset tidak valid.')
  }
  return result.data as GoldenDatasetSummary
}

export async function deleteGoldenDataset(id: string) {
  const response = await request(`/${encodeURIComponent(id)}`, { method: 'DELETE' })
  const result = await response.json()
  if (result.status !== true) {
    throw new GoldenDatasetApiError('Respons Golden Dataset tidak valid.')
  }
}
