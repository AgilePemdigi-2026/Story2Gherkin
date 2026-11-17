import type { LoginResult, UserRole } from './authApi'

export interface AuthSession {
  userId: string
  role: UserRole
  expiresAt: number
}

const AUTH_CHANGED = 'story2gherkin:auth-changed'
let cachedToken: string | null = null
let cachedSession: AuthSession | null = null

function decodeSession(token: string): AuthSession | null {
  try {
    const parts = token.split('.')
    if (parts.length !== 3) return null
    const encoded = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const claims: unknown = JSON.parse(atob(encoded.padEnd(Math.ceil(encoded.length / 4) * 4, '=')))
    if (!claims || typeof claims !== 'object') return null
    const { role, exp, user_id } = claims as Record<string, unknown>
    if ((role !== 'system_analyst' && role !== 'qa') || typeof exp !== 'number' ||
      !Number.isFinite(exp) || typeof user_id !== 'string' || !user_id) return null
    return { userId: user_id, role, expiresAt: exp * 1000 }
  } catch { return null }
}

// Claims only drive browser navigation; decoding does not verify a signature.
// Authentication and authorization of API requests belong to the backend.
export function getAuthSession(): AuthSession | null {
  const token = sessionStorage.getItem('access_token') || localStorage.getItem('access_token')
  if (token !== cachedToken) {
    cachedToken = token
    cachedSession = token ? decodeSession(token) : null
  }
  if (cachedSession && cachedSession.expiresAt <= Date.now()) cachedSession = null
  return cachedSession
}

export function subscribeAuthSession(listener: () => void) {
  let timer: ReturnType<typeof setTimeout> | undefined
  function changed() {
    clearTimeout(timer)
    listener()
    scheduleExpiry()
  }
  function scheduleExpiry() {
    const session = getAuthSession()
    if (session) timer = setTimeout(changed, Math.min(session.expiresAt - Date.now() + 1, 2147483647))
  }
  scheduleExpiry()
  window.addEventListener(AUTH_CHANGED, changed)
  window.addEventListener('storage', changed)
  window.addEventListener('focus', changed)
  return () => {
    clearTimeout(timer)
    window.removeEventListener(AUTH_CHANGED, changed)
    window.removeEventListener('storage', changed)
    window.removeEventListener('focus', changed)
  }
}

function removeTokens() {
  for (const storage of [sessionStorage, localStorage]) {
    storage.removeItem('access_token')
    storage.removeItem('refresh_token')
  }
}

export function saveAuthSession(result: LoginResult, rememberMe: boolean) {
  const session = decodeSession(result.access_token)
  if (!session || session.expiresAt <= Date.now() || session.role !== result.role) {
    throw new Error('Respons sesi tidak valid. Silakan login kembali.')
  }
  removeTokens()
  const storage = rememberMe ? localStorage : sessionStorage
  storage.setItem('access_token', result.access_token)
  if (result.refresh_token) storage.setItem('refresh_token', result.refresh_token)
  window.dispatchEvent(new Event(AUTH_CHANGED))
}

export function clearAuthSession() {
  removeTokens()
  window.dispatchEvent(new Event(AUTH_CHANGED))
}

export function roleHome(role: UserRole) {
  return role === 'system_analyst' ? '/generate' : '/evaluation'
}

export function loginDestination(from: unknown, role: UserRole) {
  if (typeof from !== 'string' || !from.startsWith('/') || from.startsWith('//') || from.includes('\\')) return roleHome(role)
  const path = from.split(/[?#]/)[0]
  const allowed = ['/about', '/documentation', ...(role === 'system_analyst' ? ['/generate', '/review'] : ['/evaluation'])]
  return allowed.includes(path) ? from : roleHome(role)
}
