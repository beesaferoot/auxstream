// A single normalized track shape spoken by every surface (feed, search, library,
// now-playing bar, player overlay) and the player engine. Both DB tracks (`Track`)
// and multi-source search hits (`SearchResult`) map into this.

import { Track, SearchResult } from '../interfaces/tracks'
import { formatDuration, getAudioUrl } from '../utils/api'
import { FruitName, pickFruit } from './covers'

export type TrackSource = 'Local' | 'YouTube' | 'SoundCloud'

export interface PlayableTrack {
  /** Stable key: `${source}:${id}`. */
  key: string
  id: string
  title: string
  artist: string
  source: TrackSource
  durationLabel: string // m:ss ("0:00" when unknown)
  durationSeconds: number
  thumbnail?: string
  streamUrl: string
  fruit: FruitName // deterministic fallback cover palette
}

const SOURCE_LABEL: Record<string, TrackSource> = {
  local: 'Local',
  youtube: 'YouTube',
  soundcloud: 'SoundCloud',
}

/** Map a backend DB track to a PlayableTrack. DB tracks are always local. */
export function fromTrack(t: Track): PlayableTrack {
  const seconds = t.duration ?? 0
  return {
    key: `Local:${t.id}`,
    id: t.id,
    title: t.title,
    artist: t.artist?.name || 'Unknown artist',
    source: 'Local',
    durationSeconds: seconds,
    durationLabel: seconds > 0 ? formatDuration(seconds) : '0:00',
    thumbnail: t.thumbnail || undefined,
    streamUrl: getAudioUrl(t.file),
    fruit: pickFruit(t.id || t.title),
  }
}

/** Map a multi-source search result to a PlayableTrack. */
export function fromSearchResult(r: SearchResult): PlayableTrack {
  const source = SOURCE_LABEL[r.source] ?? 'Local'
  return {
    key: `${source}:${r.id}`,
    id: r.id,
    title: r.title,
    artist: r.artist || 'Unknown artist',
    source,
    durationSeconds: r.duration || 0,
    durationLabel: r.duration > 0 ? formatDuration(r.duration) : '0:00',
    thumbnail: r.thumbnail || undefined,
    streamUrl: getAudioUrl(r.stream_url),
    fruit: pickFruit(`${r.source}:${r.id}` || r.title),
  }
}

/** Format a seconds value to m:ss (defensive against NaN). */
export function fmtSeconds(sec: number): string {
  if (!isFinite(sec) || sec < 0) return '0:00'
  return formatDuration(Math.floor(sec))
}
