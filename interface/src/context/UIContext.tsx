/* eslint-disable react-refresh/only-export-components -- provider + hook co-located by design */
import { createContext, useCallback, useContext, useState, ReactNode } from 'react'

export type AuthMode = 'signin' | 'signup'

interface UIContextValue {
  authOpen: boolean
  authMode: AuthMode
  openAuth: (mode?: AuthMode) => void
  closeAuth: () => void
  setAuthMode: (mode: AuthMode) => void
}

const UIContext = createContext<UIContextValue | undefined>(undefined)

export const useUI = () => {
  const ctx = useContext(UIContext)
  if (!ctx) throw new Error('useUI must be used within a UIProvider')
  return ctx
}

export const UIProvider = ({ children }: { children: ReactNode }) => {
  const [authOpen, setAuthOpen] = useState(false)
  const [authMode, setAuthMode] = useState<AuthMode>('signin')

  const openAuth = useCallback((mode: AuthMode = 'signin') => {
    setAuthMode(mode)
    setAuthOpen(true)
  }, [])
  const closeAuth = useCallback(() => setAuthOpen(false), [])

  return (
    <UIContext.Provider value={{ authOpen, authMode, openAuth, closeAuth, setAuthMode }}>
      {children}
    </UIContext.Provider>
  )
}
