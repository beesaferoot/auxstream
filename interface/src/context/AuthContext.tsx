import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { isAuthenticated, logout as apiLogout } from '../utils/api'

interface AuthContextType {
  isAuthenticated: boolean
  userEmail: string | null
  login: (token: string, email?: string) => void
  logout: () => Promise<void>
  checkAuth: () => void
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

  const value = {
    isAuthenticated: authenticated,
    userEmail,
    login,
    logout,
    checkAuth,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

