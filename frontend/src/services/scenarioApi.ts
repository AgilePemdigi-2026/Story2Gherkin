
const API_BASE_URL =
  import.meta.env.VITE_API_URL ||
  'http://localhost:8888/api/v1'

// Request dari frontend ke backend
export interface GenerateScenarioPayload {
  user_story: string
  gui_elements: string[]
  free_scenario: string
}

// Struktur skenario dari backend
export interface GeneratedScenario {
  name: string
  kind: string
  gui_element_ids: number[]
  given: string[]
  when: string[]
  then: string[]
}

// Hasil generasi
export interface GenerateScenarioResult {
  feature: {
    feature_name: string
    scenarios: GeneratedScenario[]
    warnings: string[]
  }
  gherkin: string
  duration_ms: number
}

interface ApiResponse<T> {
  status: boolean
  message?: string
  error?: string
  data?: T
}

export async function generateScenario(
  payload: GenerateScenarioPayload
): Promise<GenerateScenarioResult> {

  // Ambil access token dari sesi Login
  const accessToken =
    sessionStorage.getItem('access_token') ||
    localStorage.getItem('access_token')

  if (!accessToken) {
    throw new Error(
      'Please login before generating scenarios.'
    )
  }

  let response: Response

  try {
    response = await fetch(
      `${API_BASE_URL}/scenarios/generate`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${accessToken}`,
        },
        body: JSON.stringify(payload),
      }
    )
  } catch {
    throw new Error(
      'Cannot connect to the backend server.'
    )
  }

  const result: ApiResponse<GenerateScenarioResult> | null =
    await response.json().catch(() => null)

  if (response.status === 401) {
    throw new Error(
      'Your session is invalid or expired. Please login again.'
    )
  }

  if (!response.ok || result?.status !== true) {
    throw new Error(
      result?.error ||
      result?.message ||
      'Failed to generate Gherkin scenarios.'
    )
  }

  if (
    !result.data ||
    typeof result.data.gherkin !== 'string' ||
    !result.data.feature
  ) {
    throw new Error(
      'Invalid scenario response from backend.'
    )
  }

  return result.data
}
