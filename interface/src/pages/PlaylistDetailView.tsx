import { useEffect, useMemo, useRef, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  getPlaylist,
  getPlaylists,
  deletePlaylist,
  updatePlaylist,
  removeTrackFromPlaylist,
  reorderPlaylistTracks,
} from '../utils/api'
import { fromTrack, PlayableTrack } from '../lib/track'
import { pickFruit } from '../lib/covers'
import { usePlayer } from '../context/PlayerContext'
import { useAuth } from '../context/AuthContext'
import { useToast } from '../components/ui/Toast'
import Cover from '../components/Cover'
import PlaylistFormModal from '../components/PlaylistFormModal'
import { Playlist } from '../interfaces/playlists'
import {
  PlayIcon,
  PlusIcon,
  TrashIcon,
  GripIcon,
  ShareIcon,
  ChevronDownIcon,
} from '../components/Icons'

const PlaylistDetailView = () => {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { play } = usePlayer()
  const { isAuthenticated } = useAuth()
  const { toast } = useToast()

  const [editing, setEditing] = useState(false)
  const [order, setOrder] = useState<PlayableTrack[]>([])
  const dragIndex = useRef<number | null>(null)

  const { data, isLoading, isError } = useQuery({
    queryKey: ['playlist', id],
    queryFn: () => getPlaylist(id),
    enabled: !!id,
  })

  // Determine ownership from the user's own playlist list (the detail payload is
  // ownership-agnostic so public viewers can read it).
  const { data: mine = [] } = useQuery({
    queryKey: ['library', 'playlists'],
    queryFn: getPlaylists,
    enabled: isAuthenticated,
  })
  const isOwner = useMemo(() => mine.some((p) => p.id === id), [mine, id])

  const tracks: PlayableTrack[] = useMemo(
    () => (data?.tracks ?? []).map(fromTrack),
    [data]
  )

  // Keep the local (drag-reorderable) order in sync with fetched tracks.
  useEffect(() => {
    setOrder(tracks)
  }, [tracks])

  if (isLoading) {
    return (
      <div className="px-[44px] pt-[34px]">
        <div className="h-[200px] w-full animate-pulse rounded-hero bg-[#e7ead9]" />
      </div>
    )
  }

  if (isError || !data) {
    return (
      <div className="flex h-full flex-col items-center justify-center px-[44px] text-center">
        <div className="font-display text-[40px] font-extrabold tracking-[-1.2px]">Playlist not found</div>
        <div className="mt-2 text-[15px] text-muted-2">It may be private, or it no longer exists.</div>
        <button
          onClick={() => navigate('/library')}
          className="mt-6 rounded-pill bg-lime px-6 py-3 text-[15px] font-extrabold text-ink shadow-lime hover:-translate-y-px hover:shadow-lime-hover"
        >
          Back to library
        </button>
      </div>
    )
  }

  const playlistMeta: Playlist = {
    id: data.id,
    name: data.name,
    description: data.description,
    is_public: data.is_public,
    track_count: data.track_count,
    created_at: data.created_at,
  }

  const playAll = () => {
    if (order.length) play(order[0], order)
  }

  const handleRemove = async (track: PlayableTrack) => {
    setOrder((o) => o.filter((t) => t.id !== track.id))
    try {
      await removeTrackFromPlaylist(id, track.id)
      qc.invalidateQueries({ queryKey: ['playlist', id] })
      qc.invalidateQueries({ queryKey: ['library', 'playlists'] })
    } catch {
      toast({ title: 'Could not remove track', status: 'error' })
      qc.invalidateQueries({ queryKey: ['playlist', id] })
    }
  }

  const handleDelete = async () => {
    if (!window.confirm(`Delete "${data.name}"? This can't be undone.`)) return
    try {
      await deletePlaylist(id)
      qc.invalidateQueries({ queryKey: ['library', 'playlists'] })
      toast({ title: 'Playlist deleted', status: 'info' })
      navigate('/library')
    } catch (e) {
      toast({ title: 'Could not delete playlist', description: e instanceof Error ? e.message : undefined, status: 'error' })
    }
  }

  const togglePublic = async () => {
    try {
      await updatePlaylist(id, {
        name: data.name,
        description: data.description,
        is_public: !data.is_public,
      })
      qc.invalidateQueries({ queryKey: ['playlist', id] })
      qc.invalidateQueries({ queryKey: ['library', 'playlists'] })
      toast({ title: data.is_public ? 'Playlist is now private' : 'Playlist is now public', status: 'success' })
    } catch {
      toast({ title: 'Could not update sharing', status: 'error' })
    }
  }

  const copyLink = async () => {
    const url = `${window.location.origin}/library/playlists/${id}`
    try {
      await navigator.clipboard.writeText(url)
      toast({ title: 'Link copied', description: url, status: 'success' })
    } catch {
      toast({ title: 'Copy failed', description: url, status: 'error' })
    }
  }

  const onDrop = (toIndex: number) => {
    const from = dragIndex.current
    dragIndex.current = null
    if (from === null || from === toIndex) return
    const next = [...order]
    const [moved] = next.splice(from, 1)
    next.splice(toIndex, 0, moved)
    setOrder(next)
    reorderPlaylistTracks(id, next.map((t) => t.id))
      .then(() => qc.invalidateQueries({ queryKey: ['playlist', id] }))
      .catch(() => {
        toast({ title: 'Could not reorder', status: 'error' })
        qc.invalidateQueries({ queryKey: ['playlist', id] })
      })
  }

  return (
    <div className="px-[44px] pt-[34px]">
      <button
        onClick={() => navigate('/library')}
        className="mb-5 flex items-center gap-1.5 text-[14px] font-semibold text-muted-2 transition-colors hover:text-ink"
      >
        <span className="rotate-90">
          <ChevronDownIcon size={16} />
        </span>
        Library
      </button>

      {/* header */}
      <div className="mb-8 flex items-end gap-6">
        <Cover
          fruit={pickFruit(id)}
          className="h-[180px] w-[180px] flex-none rounded-hero shadow-hero"
        >
          <div className="absolute bottom-3 left-3 rounded-full bg-black/30 px-[9px] py-1 font-mono text-[11px] text-white backdrop-blur-sm">
            {order.length} tracks
          </div>
        </Cover>
        <div className="min-w-0 flex-1">
          <div className="font-mono text-[12px] uppercase tracking-[2px] text-muted-2">
            {data.is_public ? 'Public playlist' : 'Playlist'}
          </div>
          <div className="mt-1.5 truncate font-display text-[46px] font-extrabold leading-none tracking-[-1.4px]">
            {data.name}
          </div>
          {data.description && (
            <div className="mt-2 text-[16px] text-muted-2">{data.description}</div>
          )}

          <div className="mt-5 flex flex-wrap items-center gap-3">
            <button
              onClick={playAll}
              disabled={!order.length}
              className="flex items-center gap-2.5 rounded-pill bg-lime px-6 py-3 text-[16px] font-extrabold text-ink shadow-lime transition-all hover:-translate-y-px hover:shadow-lime-hover disabled:opacity-50"
            >
              <PlayIcon size={20} />
              Play all
            </button>
            <button
              onClick={() => navigate('/search')}
              className="flex items-center gap-2 rounded-pill border-[1.5px] border-line bg-white px-5 py-3 text-[15px] font-bold text-muted transition-colors hover:border-lime"
            >
              <PlusIcon size={18} />
              Add tracks
            </button>

            {isOwner && (
              <>
                <button
                  onClick={togglePublic}
                  className="flex items-center gap-2 rounded-pill border-[1.5px] border-line bg-white px-5 py-3 text-[15px] font-bold text-muted transition-colors hover:border-lime"
                >
                  <ShareIcon size={18} />
                  {data.is_public ? 'Make private' : 'Make public'}
                </button>
                {data.is_public && (
                  <button
                    onClick={copyLink}
                    className="rounded-pill border-[1.5px] border-line bg-white px-5 py-3 text-[15px] font-bold text-muted transition-colors hover:border-lime"
                  >
                    Copy link
                  </button>
                )}
                <button
                  onClick={() => setEditing(true)}
                  className="rounded-pill border-[1.5px] border-line bg-white px-5 py-3 text-[15px] font-bold text-muted transition-colors hover:border-lime"
                >
                  Edit
                </button>
                <button
                  onClick={handleDelete}
                  className="flex items-center gap-2 rounded-pill border-[1.5px] border-line bg-white px-5 py-3 text-[15px] font-bold text-danger transition-colors hover:border-danger"
                >
                  <TrashIcon size={16} />
                  Delete
                </button>
              </>
            )}
          </div>
        </div>
      </div>

      {/* tracks */}
      <div className="overflow-hidden rounded-[20px] border-[1.5px] border-line-2 bg-white">
        {order.length === 0 ? (
          <div className="px-[18px] py-12 text-center">
            <div className="text-[16px] font-semibold text-muted">No tracks yet</div>
            <div className="mt-1 text-[14px] text-faint">
              Use “Add tracks” or the ⋮ menu on any track to fill this playlist.
            </div>
          </div>
        ) : (
          order.map((t, i) => (
            <div
              key={t.key}
              draggable={isOwner}
              onDragStart={() => (dragIndex.current = i)}
              onDragOver={(e) => e.preventDefault()}
              onDrop={() => onDrop(i)}
              className="group flex items-center gap-4 border-b border-line-sep px-[18px] py-[13px] transition-colors last:border-b-0 hover:bg-[#f7f9ef]"
            >
              {isOwner && (
                <span className="cursor-grab text-faint-2 opacity-0 transition-opacity group-hover:opacity-100">
                  <GripIcon size={18} />
                </span>
              )}
              <button onClick={() => play(t, order)} className="flex min-w-0 flex-1 items-center gap-4 text-left">
                <Cover
                  fruit={t.fruit}
                  thumbnail={t.thumbnail}
                  alt={t.title}
                  className="h-[50px] w-[50px] flex-none rounded-[11px] shadow-cover"
                />
                <div className="min-w-0 flex-1">
                  <div className="truncate text-[16px] font-bold">{t.title}</div>
                  <div className="truncate text-[14px] text-muted-2">{t.artist}</div>
                </div>
              </button>
              <span className="font-mono text-[13px] text-faint">{t.durationLabel}</span>
              {isOwner && (
                <button
                  onClick={() => handleRemove(t)}
                  title="Remove from playlist"
                  className="flex h-[34px] w-[34px] flex-none items-center justify-center rounded-full border-[1.5px] border-line bg-white text-muted transition-colors hover:border-danger hover:text-danger"
                >
                  <TrashIcon size={15} />
                </button>
              )}
            </div>
          ))
        )}
      </div>

      <PlaylistFormModal open={editing} onClose={() => setEditing(false)} playlist={playlistMeta} />
    </div>
  )
}

export default PlaylistDetailView
