/* eslint-disable react-refresh/only-export-components -- provider + hook co-located by design */
import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { isAuthenticated, logout as apiLogout } from '../utils/api'

interface AuthContextType {
  isAuthenticated: boolean
  userEmail: string | null
  /** Display name derived from the email local-part (e.g. "harper" -> "Harper"). */
  userName: string
  /** Single-letter avatar initial. */
  userInitial: string
  login: (token: string, email?: string) => void
  logout: () => Promise<void>
  checkAuth: () => void
}

/** Derive a friendly display name from an email address. */
function deriveName(email: string | null): string {
  if (!email) return 'Guest'
  const local = email.split('@')[0].replace(/[._-]+/g, ' ').trim()
  if (!local) return 'Guest'
  return local
    .split(' ')
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ')
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

export const useAuth = () => {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}

interface AuthProviderProps {
  children: ReactNode
}

export const AuthProvider = ({ children }: AuthProviderProps) => {
  const [authenticated, setAuthenticated] = useState(isAuthenticated())
  const [userEmail, setUserEmail] = useState<string | null>(
    localStorage.getItem('user_email')
  )

  const checkAuth = () => {
    const authStatus = isAuthenticated()
    setAuthenticated(authStatus)
    if (!authStatus) {
      setUserEmail(null)
      localStorage.removeItem('user_email')
    }
  }

  const login = (_token: string, email?: string) => {
    setAuthenticated(true)
    if (email) {
      setUserEmail(email)
      localStorage.setItem('user_email', email)
    }
  }

  const logout = async () => {
    await apiLogout()
    setAuthenticated(false)
    setUserEmail(null)
    localStorage.removeItem('user_email')
  }

  useEffect(() => {
    checkAuth()

    // Listen for token expiration events
    const handleTokenExpired = () => {
      setAuthenticated(false)
      setUserEmail(null)
      localStorage.removeItem('user_email')
      // Optionally show a notification that the session expired
      console.log('Session expired. Please log in again.')
    }

    window.addEventListener('auth:token-expired', handleTokenExpired)

    return () => {
      window.removeEventListener('auth:token-expired', handleTokenExpired)
    }
  }, [])

  const userName = deriveName(userEmail)
  const userInitial = (userName.charAt(0) || 'G').toUpperCase()

  const value = {
    isAuthenticated: authenticated,
    userEmail,
    userName,
    userInitial,
    login,
    logout,
    checkAuth,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

