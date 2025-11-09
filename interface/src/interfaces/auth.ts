export interface User {
  id: string
  email: string
  created_at: string
}

export interface TokenResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  token_type: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
}

export interface LoginResponse {
  message: string
  data: TokenResponse
}

export interface RegisterResponse {
  message: string
  data: {
    user_id: string
  }
}

export interface AuthResponse {
  error?: string
  message?: string
}
