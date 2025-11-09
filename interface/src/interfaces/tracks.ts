import { Artist } from './artists'

export interface AddTrackResponse {
  error?: string
  data?: Track
}

export interface GetTrackResponse {
  error?: string
  data: Track[]
}

export interface Track {
  id: string
  title: string
  artist_id: string
  artist?: Artist
  file: string
  duration?: number
  thumbnail?: string
  created_at: string
  updated_at?: string
}

export interface SearchResult {
  id: string
  title: string
  artist: string
  duration: number
  thumbnail: string
  source: 'local' | 'youtube' | 'soundcloud'
  external_id?: string
  stream_url: string
  description?: string
}

export interface SearchResponse {
  query: string
  results: SearchResult[]
  total_count: number
  source: string
  cached_at?: string
  searched_at: string
}
