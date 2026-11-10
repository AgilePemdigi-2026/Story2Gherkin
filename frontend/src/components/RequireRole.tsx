import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import type { UserRole } from '../services/authApi'
import { roleHome } from '../services/authSession'
import { useAuthSession } from '../hooks/useAuthSession'

export default function RequireRole({ role, children }: { role: UserRole; children: ReactNode }) {
  const session = useAuthSession()
  const location = useLocation()
  if (!session) return <Navigate to="/login" replace state={{ from: location.pathname + location.search + location.hash }} />
  if (session.role !== role) return <Navigate to={roleHome(session.role)} replace />
  return children
}
