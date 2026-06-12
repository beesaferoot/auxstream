import { Track } from './tracks'

export interface Playlist {
  id: string
  name: string
  description: string
  is_public: boolean
  track_count: number
  created_at: string
}

export interface PlaylistDetail extends Playlist {
  tracks: Track[]
}

export interface GetPlaylistsResponse {
  error?: string
  data: Playlist[]
}

export interface PlaylistInput {
  name: string
  description?: string
  is_public?: boolean
}
