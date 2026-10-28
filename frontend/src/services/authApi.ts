
export type UserRole = 'system_analyst' | 'qa'

export interface RegisterPayload {
  name: string
  email: string
  password: string
  role: UserRole
}

interface ApiResponse {
  status?: boolean
  message?: string
  data?: unknown
}

const API_BASE_URL =
  import.meta.env.VITE_API_URL || 'http://localhost:8888/api/v1'

export async function registerUser(payload: RegisterPayload) {
  const response = await fetch(`${API_BASE_URL}/auth/register`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  })

  const result: ApiResponse = await response.json().catch(() => ({}))

  if (!response.ok || result.status === false) {
    throw new Error(
      result.message || 'Registration failed. Please try again.'
    )
  }

  return result
}


export interface LoginPayload {
  email: string
  password: string
}

export interface LoginUser {
  id: string
  name: string
  email: string
  role: string
}

export interface LoginResult {
  access_token: string
  token_type: string
  expires_in: number
  user: LoginUser
}

interface LoginApiResponse {
  status: boolean
  message?: string
  error?: string
  data?: LoginResult
}

export async function loginUser(
  payload: LoginPayload
): Promise<LoginResult> {
  const response = await fetch(
    `${API_BASE_URL}/auth/login`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
    }
  )

  const result: LoginApiResponse = await response.json()

  if (!response.ok || !result.status) {
    throw new Error(
      result.error ||
      result.message ||
      'Login failed. Please try again.'
    )
  }

  if (!result.data?.access_token) {
    throw new Error('Invalid login response from server.')
  }

  return result.data
}


export async function sendPasswordReset(email: string): Promise<void> {
  const response = await fetch(
    `${API_BASE_URL}/auth/password-reset/send`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email }),
    }
  )

  const result: {
    status?: boolean
    message?: string
    error?: string
  } = await response.json().catch(() => ({}))

  if (!response.ok || result.status === false) {
    throw new Error(
      result.error ||
      result.message ||
      'Failed to send password reset email.'
    )
  }
}


export async function resetPassword(
  token: string,
  newPassword: string
): Promise<void> {
  const response = await fetch(
    `${API_BASE_URL}/auth/password-reset/reset`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        token,
        new_password: newPassword,
      }),
    }
  )

  const result: {
    status?: boolean
    message?: string
    error?: string
  } = await response.json().catch(() => ({}))

  if (!response.ok || result.status === false) {
    throw new Error(
      result.error ||
      result.message ||
      'Failed to reset password.'
    )
  }
}
