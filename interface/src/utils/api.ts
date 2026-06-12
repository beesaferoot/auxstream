import {
  GetTrackResponse,
  SearchResponse,
  AddTrackResponse,
  Track,
} from '../interfaces/tracks.ts'
import {
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
} from '../interfaces/auth.ts'
import {
  Playlist,
  PlaylistDetail,
  PlaylistInput,
} from '../interfaces/playlists.ts'
import { BASE_URL as BASE_URL_IMPORT } from './constants.ts'

export const BASE_URL = BASE_URL_IMPORT

type RequestParams = {
  path: string
  method?: HTTPMethods
  body?: Record<string, unknown>
  headers?: Record<string, string>
}

type UploadProgressCallback = (progress: number) => void

export enum HTTPMethods {
  get = 'GET',
  post = 'POST',
  put = 'PUT',
  patch = 'PATCH',
  delete = 'DELETE',
}

// Auth token management
export function getAuthToken(): string | null {
  return localStorage.getItem('access_token')
}

export function setAuthToken(token: string): void {
  localStorage.setItem('access_token', token)
}

export function removeAuthToken(): void {
  localStorage.removeItem('access_token')
}

export function getRefreshToken(): string | null {
  return localStorage.getItem('refresh_token')
}

export function setRefreshToken(token: string): void {
  localStorage.setItem('refresh_token', token)
}

export function removeRefreshToken(): void {
  localStorage.removeItem('refresh_token')
}

export function isAuthenticated(): boolean {
  return !!getAuthToken()
}

// Flag to prevent multiple simultaneous refresh attempts
let isRefreshing = false
let refreshSubscribers: ((token: string) => void)[] = []

function subscribeTokenRefresh(callback: (token: string) => void) {
  refreshSubscribers.push(callback)
}

function onTokenRefreshed(token: string) {
  refreshSubscribers.forEach((callback) => callback(token))
  refreshSubscribers = []
}

async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) {
    return null
  }

  try {
    const response = await fetch(`${BASE_URL}/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ refresh_token: refreshToken }),
    })

    if (!response.ok) {
      // Refresh token is invalid or expired
      removeAuthToken()
      removeRefreshToken()
      return null
    }

    const data = await response.json()
    if (data.data?.access_token) {
      setAuthToken(data.data.access_token)
      if (data.data.refresh_token) {
        setRefreshToken(data.data.refresh_token)
      }
      return data.data.access_token
    }

    return null
  } catch (error) {
    console.error('Token refresh failed:', error)
    removeAuthToken()
    removeRefreshToken()
    return null
  }
}

export async function request<T>({
  path,
  method = HTTPMethods.get,
  body,
  headers = {},
}: RequestParams): Promise<T> {
  const makeRequest = async (token: string | null): Promise<Response> => {
    const defaultHeaders: Record<string, string> = {
      'Content-Type': 'application/json',
    }

    if (token) {
      defaultHeaders['Authorization'] = `Bearer ${token}`
    }

    return fetch(`${BASE_URL}${path}`, {
      method: method,
      headers: { ...defaultHeaders, ...headers },
      body: body ? JSON.stringify(body) : undefined,
    })
  }

  const authToken = getAuthToken()
  let response = await makeRequest(authToken)

  // If we get a 401, try to refresh the token
  if (response.status === 401 && authToken) {
    if (!isRefreshing) {
      isRefreshing = true
      const newToken = await refreshAccessToken()
      isRefreshing = false

      if (newToken) {
        onTokenRefreshed(newToken)
        // Retry the request with the new token
        response = await makeRequest(newToken)
      } else {
        // Refresh failed, user needs to log in again
        window.dispatchEvent(new CustomEvent('auth:token-expired'))
        throw new Error('Session expired. Please log in again.')
      }
    } else {
      // Wait for the ongoing refresh to complete
      const newToken = await new Promise<string>((resolve) => {
        subscribeTokenRefresh(resolve)
      })
      response = await makeRequest(newToken)
    }
  }

  if (!response.ok) {
    const errorText = await response.text()
    throw new Error(`API Error: ${response.status} - ${errorText}`)
  }

  return (await response.json()) as T
}

export async function getTrendingTracks(
  pagesize: number,
  pagenumber: number,
  sort: 'trending' | 'recent' | 'default' = 'trending',
  days?: number
): Promise<GetTrackResponse | undefined> {
  let path = `/tracks?pagesize=${pagesize}&pagenumber=${pagenumber}&sort=${sort}`

  if (days && sort === 'trending') {
    path += `&days=${days}`
  }

  const res = await request<GetTrackResponse | undefined>({
    path,
  })
  return res
}

// Track a play event (increments play count and records playback history)
export async function trackPlay(
  trackId: string,
  durationPlayed?: number
): Promise<void> {
  try {
    await request({
      path: `/tracks/play`,
      method: HTTPMethods.post,
      body: {
        track_id: trackId,
        duration_played: durationPlayed || 0,
      },
    })
  } catch (error) {
    console.error('Failed to track play:', error)
    // Don't throw - this is a non-critical operation
  }
}

export async function searchTracks(
  query: string,
  source?: 'local' | 'youtube' | 'soundcloud',
  maxResults = 20
): Promise<SearchResponse> {
  let path = `/search?q=${encodeURIComponent(query)}&max_results=${maxResults}`

  if (source) {
    path += `&source=${source}`
  }

  const response = await request<{ data: SearchResponse }>({ path })
  return response.data
}

// Upload track with file and metadata
export async function uploadTrack(
  file: File,
  metadata: {
    title: string
    artistId: string
    duration?: number
    thumbnail?: string
  },
  onProgress?: UploadProgressCallback
): Promise<AddTrackResponse> {
  const uploadWithToken = (token: string | null): Promise<AddTrackResponse> => {
    const formData = new FormData()

    formData.append('audio', file)
    formData.append('title', metadata.title)
    formData.append('artist_id', metadata.artistId)

    if (metadata.duration) {
      formData.append('duration', metadata.duration.toString())
    }

    if (metadata.thumbnail) {
      formData.append('thumbnail', metadata.thumbnail)
    }

    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest()

      // Track upload progress
      if (onProgress) {
        xhr.upload.addEventListener('progress', (e) => {
          if (e.lengthComputable) {
            const percentComplete = (e.loaded / e.total) * 100
            onProgress(percentComplete)
          }
        })
      }

      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const response = JSON.parse(xhr.responseText)
            resolve(response)
          } catch (error) {
            reject(new Error('Failed to parse response'))
          }
        } else if (xhr.status === 401) {
          // Token expired, signal for retry
          reject(new Error('TOKEN_EXPIRED'))
        } else {
          try {
            const errorResponse = JSON.parse(xhr.responseText)
            reject(
              new Error(
                `Upload failed: ${xhr.status} - ${
                  errorResponse.error || xhr.statusText
                }`
              )
            )
          } catch {
            reject(
              new Error(`Upload failed: ${xhr.status} - ${xhr.statusText}`)
            )
          }
        }
      })

      xhr.addEventListener('error', () => {
        reject(new Error('Upload failed: Network error'))
      })

      xhr.addEventListener('abort', () => {
        reject(new Error('Upload cancelled'))
      })

      xhr.open('POST', `${BASE_URL}/upload_track`)

      if (token) {
        xhr.setRequestHeader('Authorization', `Bearer ${token}`)
      }

      xhr.send(formData)
    })
  }

  try {
    const authToken = getAuthToken()
    return await uploadWithToken(authToken)
  } catch (error) {
    // If token expired, try to refresh and retry
    if (error instanceof Error && error.message === 'TOKEN_EXPIRED') {
      const newToken = await refreshAccessToken()
      if (newToken) {
        return await uploadWithToken(newToken)
      } else {
        window.dispatchEvent(new CustomEvent('auth:token-expired'))
        throw new Error('Session expired. Please log in again.')
      }
    }
    throw error
  }
}

// Bulk upload multiple tracks in one request
export async function uploadTracksBulk(
  files: File[],
  titles: string[],
  artistId: string,
  onProgress?: UploadProgressCallback
): Promise<{ data: { saved: string[]; rows: number } }> {
  const uploadWithToken = (
    token: string | null
  ): Promise<{ data: { saved: string[]; rows: number } }> => {
    const formData = new FormData()

    // Append all files
    files.forEach((file) => {
      formData.append('track_files', file)
    })

    // Append all titles
    titles.forEach((title) => {
      formData.append('track_titles', title)
    })

    // Append artist ID
    formData.append('artist_id', artistId)

    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest()

      // Track upload progress
      if (onProgress) {
        xhr.upload.addEventListener('progress', (e) => {
          if (e.lengthComputable) {
            const percentComplete = (e.loaded / e.total) * 100
            onProgress(percentComplete)
          }
        })
      }

      xhr.addEventListener('load', () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            const response = JSON.parse(xhr.responseText)
            resolve(response)
          } catch (error) {
            reject(new Error('Failed to parse response'))
          }
        } else if (xhr.status === 401) {
          // Token expired, signal for retry
          reject(new Error('TOKEN_EXPIRED'))
        } else {
          try {
            const errorResponse = JSON.parse(xhr.responseText)
            reject(
              new Error(
                `Upload failed: ${xhr.status} - ${
                  errorResponse.error || xhr.statusText
                }`
              )
            )
          } catch {
            reject(
              new Error(`Upload failed: ${xhr.status} - ${xhr.statusText}`)
            )
          }
        }
      })

      xhr.addEventListener('error', () => {
        reject(new Error('Upload failed: Network error'))
      })

      xhr.addEventListener('abort', () => {
        reject(new Error('Upload cancelled'))
      })

      xhr.open('POST', `${BASE_URL}/upload_batch_track`)

      if (token) {
        xhr.setRequestHeader('Authorization', `Bearer ${token}`)
      }

      xhr.send(formData)
    })
  }

  try {
    const authToken = getAuthToken()
    return await uploadWithToken(authToken)
  } catch (error) {
    // If token expired, try to refresh and retry
    if (error instanceof Error && error.message === 'TOKEN_EXPIRED') {
      const newToken = await refreshAccessToken()
      if (newToken) {
        return await uploadWithToken(newToken)
      } else {
        window.dispatchEvent(new CustomEvent('auth:token-expired'))
        throw new Error('Session expired. Please log in again.')
      }
    }
    throw error
  }
}

// Create or get artist by name
export async function createArtist(
  artistName: string
): Promise<{ id: string; name: string }> {
  try {
    const response = await request<{ data: { id: string; name: string } }>({
      path: `/artists`,
      method: HTTPMethods.post,
      body: { name: artistName },
    })
    return response.data
  } catch (error) {
    // If artist already exists or other error, generate a temporary UUID
    // In a real app, you'd want to search for the artist first
    console.warn('Failed to create artist, using temporary ID:', error)
    return {
      id: crypto.randomUUID(),
      name: artistName,
    }
  }
}

// Get artist by ID
export async function getArtistById(
  artistId: string
): Promise<{ id: string; name: string }> {
  const response = await request<{ data: { id: string; name: string } }>({
    path: `/artists/${artistId}`,
  })
  return response.data
}

// Get tracks by artist ID
export async function getArtistTracks(
  artistId: string,
  pageSize = 20,
  pageNumber = 1
): Promise<GetTrackResponse> {
  const response = await request<GetTrackResponse>({
    path: `/artists/${artistId}/tracks?pagesize=${pageSize}&pagenumber=${pageNumber}`,
  })
  return response
}

// Get single track by ID
export async function getTrackById(trackId: string): Promise<Track> {
  const response = await request<{ data: Track }>({
    path: `/tracks/${trackId}`,
  })
  return response.data
}

// Get the authenticated user's playlists (with track counts)
export async function getPlaylists(): Promise<Playlist[]> {
  const response = await request<{ data: Playlist[] }>({ path: `/playlists` })
  return response.data ?? []
}

// Get a single playlist with its ordered tracks (public playlists work unauthenticated)
export async function getPlaylist(id: string): Promise<PlaylistDetail> {
  const response = await request<{ data: PlaylistDetail }>({
    path: `/playlists/${id}`,
  })
  return response.data
}

export async function createPlaylist(input: PlaylistInput): Promise<Playlist> {
  const response = await request<{ data: Playlist }>({
    path: `/playlists`,
    method: HTTPMethods.post,
    body: input as unknown as Record<string, unknown>,
  })
  return response.data
}

export async function updatePlaylist(
  id: string,
  input: PlaylistInput
): Promise<Playlist> {
  const response = await request<{ data: Playlist }>({
    path: `/playlists/${id}`,
    method: HTTPMethods.patch,
    body: input as unknown as Record<string, unknown>,
  })
  return response.data
}

export async function deletePlaylist(id: string): Promise<void> {
  await request({ path: `/playlists/${id}`, method: HTTPMethods.delete })
}

export async function addTrackToPlaylist(
  id: string,
  trackId: string
): Promise<void> {
  await request({
    path: `/playlists/${id}/tracks`,
    method: HTTPMethods.post,
    body: { track_id: trackId },
  })
}

export async function removeTrackFromPlaylist(
  id: string,
  trackId: string
): Promise<void> {
  await request({
    path: `/playlists/${id}/tracks/${trackId}`,
    method: HTTPMethods.delete,
  })
}

export async function reorderPlaylistTracks(
  id: string,
  trackIds: string[]
): Promise<void> {
  await request({
    path: `/playlists/${id}/tracks/order`,
    method: HTTPMethods.put,
    body: { track_ids: trackIds },
  })
}

// Authentication API functions
export async function login(credentials: LoginRequest): Promise<LoginResponse> {
  const response = await request<LoginResponse>({
    path: `/login`,
    method: HTTPMethods.post,
    body: credentials as unknown as Record<string, unknown>,
  })

  // Store tokens
  if (response.data) {
    setAuthToken(response.data.access_token)
    setRefreshToken(response.data.refresh_token)
  }

  return response
}

export async function register(
  credentials: RegisterRequest
): Promise<RegisterResponse> {
  return await request<RegisterResponse>({
    path: `/register`,
    method: HTTPMethods.post,
    body: credentials as unknown as Record<string, unknown>,
  })
}

export async function logout(): Promise<void> {
  const refreshToken = getRefreshToken()

  if (refreshToken) {
    try {
      await request<{ message: string }>({
        path: `/logout`,
        method: HTTPMethods.post,
        body: { refresh_token: refreshToken },
      })
    } catch (error) {
      console.error('Logout error:', error)
    }
  }

  // Clear tokens regardless of API response
  removeAuthToken()
  removeRefreshToken()
}

// Helper to format duration from seconds to MM:SS
export function formatDuration(seconds: number): string {
  const minutes = Math.floor(seconds / 60)
  const remainingSeconds = seconds % 60
  return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`
}

// Helper to construct proper audio URL
export function getAudioUrl(filePathOrUrl: string): string {
  // Defensive: never throw during render on missing data.
  if (!filePathOrUrl) return ''

  // If it's already a full URL, return it
  if (
    filePathOrUrl.startsWith('http://') ||
    filePathOrUrl.startsWith('https://')
  ) {
    return filePathOrUrl
  }

  // Same-origin API path. When BASE_URL is absolute (e.g. http://localhost:5009/api/v1)
  // resolve the path against its origin so audio is fetched directly from the API.
  // When BASE_URL is relative (e.g. "/api/v1", the same-origin proxy setup), the path
  // already works as-is — and `new URL("/api/v1")` would throw, so we must not call it.
  if (filePathOrUrl.startsWith('/api/v1/')) {
    if (/^https?:\/\//i.test(BASE_URL)) {
      return `${new URL(BASE_URL).origin}${filePathOrUrl}`
    }
    return filePathOrUrl
  }

  // Otherwise, assume it's just a filename and construct the full URL
  return `${BASE_URL}/serve/${filePathOrUrl}`
}
