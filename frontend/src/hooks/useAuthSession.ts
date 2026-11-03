import { useSyncExternalStore } from 'react'
import { getAuthSession, subscribeAuthSession } from '../services/authSession'

export function useAuthSession() {
  return useSyncExternalStore(subscribeAuthSession, getAuthSession)
}
