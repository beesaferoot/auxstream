/* eslint-disable react-refresh/only-export-components -- provider + hook co-located by design */
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  ReactNode,
} from 'react'
import { PlayableTrack } from '../lib/track'
import { trackPlay } from '../utils/api'

const PERSIST_KEY = 'auxstream_hifi_state'

interface PersistShape {
  track: PlayableTrack
  time: number
}

interface PlayerContextValue {
  current: PlayableTrack | null
  queue: PlayableTrack[]
  isPlaying: boolean
  currentTime: number
  duration: number
  /** 0–1 playback position. */
  progress: number
  volume: number
  isOpen: boolean
  shuffle: boolean
  repeat: boolean
  /** Play a track, optionally seeding the queue (defaults to [track]). */
  play: (track: PlayableTrack, queue?: PlayableTrack[]) => void
  toggle: () => void
  next: () => void
  prev: () => void
  /** Seek to a 0–1 fraction of the track. */
  seekFraction: (fraction: number) => void
  setVolume: (v: number) => void
  openPlayer: () => void
  closePlayer: () => void
  toggleShuffle: () => void
  toggleRepeat: () => void
  /** Append a track to the queue (no-op if already queued). Returns false if it was already present. */
  enqueue: (track: PlayableTrack) => boolean
}

const PlayerContext = createContext<PlayerContextValue | undefined>(undefined)

export const usePlayer = () => {
  const ctx = useContext(PlayerContext)
  if (!ctx) throw new Error('usePlayer must be used within a PlayerProvider')
  return ctx
}

export const PlayerProvider = ({ children }: { children: ReactNode }) => {
  const audioRef = useRef<HTMLAudioElement>(null)

  const [current, setCurrent] = useState<PlayableTrack | null>(null)
  const [queue, setQueue] = useState<PlayableTrack[]>([])
  const [isPlaying, setIsPlaying] = useState(false)
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(0)
  const [volume, setVolumeState] = useState(1)
  const [isOpen, setIsOpen] = useState(false)
  const [shuffle, setShuffle] = useState(false)
  const [repeat, setRepeat] = useState(false)

  // Keep the latest values available to event handlers without re-binding listeners.
  const stateRef = useRef({ current, queue, shuffle, repeat })
  stateRef.current = { current, queue, shuffle, repeat }

  // Load a track into the audio element and (optionally) start playback.
  const load = useCallback((track: PlayableTrack, autoplay: boolean, startAt = 0) => {
    const audio = audioRef.current
    if (!audio) return
    audio.src = track.streamUrl
    audio.load()
    if (startAt > 0) {
      const seek = () => {
        audio.currentTime = startAt
        audio.removeEventListener('loadedmetadata', seek)
      }
      audio.addEventListener('loadedmetadata', seek)
    }
    if (autoplay) {
      audio.play().catch(() => setIsPlaying(false))
    }
  }, [])

  const play = useCallback(
    (track: PlayableTrack, nextQueue?: PlayableTrack[]) => {
      const q = nextQueue && nextQueue.length ? nextQueue : [track]
      setQueue(q)
      setCurrent(track)
      setCurrentTime(0)
      setDuration(track.durationSeconds || 0)
      load(track, true)
      if (track.source === 'Local') trackPlay(track.id)
    },
    [load]
  )

  const goTo = useCallback(
    (dir: 1 | -1) => {
      const { current: cur, queue: q, shuffle: sh } = stateRef.current
      if (!q.length) return
      const idx = cur ? q.findIndex((t) => t.key === cur.key) : -1
      let nextIdx: number
      if (sh && q.length > 1) {
        do {
          nextIdx = Math.floor(Math.random() * q.length)
        } while (nextIdx === idx)
      } else {
        nextIdx = ((idx === -1 ? 0 : idx) + dir + q.length) % q.length
      }
      const track = q[nextIdx]
      setCurrent(track)
      setCurrentTime(0)
      setDuration(track.durationSeconds || 0)
      load(track, true)
      if (track.source === 'Local') trackPlay(track.id)
    },
    [load]
  )

  const next = useCallback(() => goTo(1), [goTo])
  const prev = useCallback(() => goTo(-1), [goTo])

  const toggle = useCallback(() => {
    const audio = audioRef.current
    if (!audio || !stateRef.current.current) return
    if (audio.paused) audio.play().catch(() => setIsPlaying(false))
    else audio.pause()
  }, [])

  const seekFraction = useCallback((fraction: number) => {
    const audio = audioRef.current
    if (!audio) return
    const f = Math.min(1, Math.max(0, fraction))
    const dur = audio.duration || stateRef.current.current?.durationSeconds || 0
    if (dur > 0) {
      audio.currentTime = f * dur
      setCurrentTime(f * dur)
    }
  }, [])

  const setVolume = useCallback((v: number) => {
    const audio = audioRef.current
    const vol = Math.min(1, Math.max(0, v))
    if (audio) audio.volume = vol
    setVolumeState(vol)
  }, [])

  const enqueue = useCallback((track: PlayableTrack) => {
    const already = stateRef.current.queue.some((t) => t.key === track.key)
    if (already) return false
    setQueue((q) => (q.some((t) => t.key === track.key) ? q : [...q, track]))
    // If nothing is loaded yet, make the enqueued track the current one (paused-ready).
    setCurrent((c) => c ?? track)
    return true
  }, [])

  const openPlayer = useCallback(() => setIsOpen(true), [])
  const closePlayer = useCallback(() => setIsOpen(false), [])
  const toggleShuffle = useCallback(() => setShuffle((s) => !s), [])
  const toggleRepeat = useCallback(() => setRepeat((r) => !r), [])

  // Bind audio element events once.
  useEffect(() => {
    const audio = audioRef.current
    if (!audio) return

    const onTime = () => setCurrentTime(audio.currentTime)
    const onDuration = () => {
      if (isFinite(audio.duration)) setDuration(audio.duration)
    }
    const onPlay = () => setIsPlaying(true)
    const onPause = () => setIsPlaying(false)
    const onEnded = () => {
      if (stateRef.current.repeat) {
        audio.currentTime = 0
        audio.play().catch(() => setIsPlaying(false))
      } else {
        goTo(1)
      }
    }

    audio.addEventListener('timeupdate', onTime)
    audio.addEventListener('durationchange', onDuration)
    audio.addEventListener('loadedmetadata', onDuration)
    audio.addEventListener('play', onPlay)
    audio.addEventListener('pause', onPause)
    audio.addEventListener('ended', onEnded)

    return () => {
      audio.removeEventListener('timeupdate', onTime)
      audio.removeEventListener('durationchange', onDuration)
      audio.removeEventListener('loadedmetadata', onDuration)
      audio.removeEventListener('play', onPlay)
      audio.removeEventListener('pause', onPause)
      audio.removeEventListener('ended', onEnded)
    }
  }, [goTo])

  // Restore last track + position on mount (paused).
  useEffect(() => {
    try {
      const raw = localStorage.getItem(PERSIST_KEY)
      if (!raw) return
      const saved = JSON.parse(raw) as PersistShape
      if (saved?.track?.streamUrl) {
        setCurrent(saved.track)
        setQueue([saved.track])
        setCurrentTime(saved.time || 0)
        setDuration(saved.track.durationSeconds || 0)
        load(saved.track, false, saved.time || 0)
      }
    } catch {
      /* ignore */
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Persist current track + position.
  useEffect(() => {
    if (!current) return
    try {
      localStorage.setItem(
        PERSIST_KEY,
        JSON.stringify({ track: current, time: currentTime } as PersistShape)
      )
    } catch {
      /* ignore */
    }
  }, [current, currentTime])

  const value = useMemo<PlayerContextValue>(
    () => ({
      current,
      queue,
      isPlaying,
      currentTime,
      duration,
      progress: duration > 0 ? Math.min(1, currentTime / duration) : 0,
      volume,
      isOpen,
      shuffle,
      repeat,
      play,
      toggle,
      next,
      prev,
      seekFraction,
      setVolume,
      openPlayer,
      closePlayer,
      toggleShuffle,
      toggleRepeat,
      enqueue,
    }),
    [
      current,
      queue,
      isPlaying,
      currentTime,
      duration,
      volume,
      isOpen,
      shuffle,
      repeat,
      play,
      toggle,
      next,
      prev,
      seekFraction,
      setVolume,
      openPlayer,
      closePlayer,
      toggleShuffle,
      toggleRepeat,
      enqueue,
    ]
  )

  return (
    <PlayerContext.Provider value={value}>
      {children}
      {/* Single shared audio engine for the whole app. */}
      <audio ref={audioRef} preload="metadata" />
    </PlayerContext.Provider>
  )
}
