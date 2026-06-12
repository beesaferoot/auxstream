import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { getTrendingTracks, getPlaylists } from '../utils/api'
import { fromTrack, PlayableTrack } from '../lib/track'
import { cover, pickFruit } from '../lib/covers'
import { usePlayer } from '../context/PlayerContext'
import { useAuth } from '../context/AuthContext'
import { useUI } from '../context/UIContext'
import UploadDropzone from '../components/UploadDropzone'
import TrackRow from '../components/TrackRow'
import PlaylistFormModal from '../components/PlaylistFormModal'
import { PlusIcon } from '../components/Icons'

const LibraryView = () => {
  const { play } = usePlayer()
  const { isAuthenticated } = useAuth()
  const { openAuth } = useUI()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  // Track ids uploaded this session so we can flag them "fresh".
  const [freshIds, setFreshIds] = useState<Set<string>>(new Set())
  const [creating, setCreating] = useState(false)

  const uploadsQ = useQuery({
    queryKey: ['library', 'uploads'],
    queryFn: () => getTrendingTracks(40, 1, 'recent'),
    staleTime: 60_000,
  })

  const playlistsQ = useQuery({
    queryKey: ['library', 'playlists'],
    queryFn: getPlaylists,
    enabled: isAuthenticated,
    staleTime: 60_000,
  })

  const uploads: PlayableTrack[] = useMemo(
    () => (uploadsQ.data?.data ?? []).map(fromTrack),
    [uploadsQ.data]
  )
  const playlists = playlistsQ.data ?? []

  const handleUploaded = (savedIds: string[]) => {
    setFreshIds((prev) => new Set([...prev, ...savedIds]))
    queryClient.invalidateQueries({ queryKey: ['library', 'uploads'] })
    queryClient.invalidateQueries({ queryKey: ['feed'] })
  }

  return (
    <div className="animate-aux-pop px-[44px] pt-[34px]">
      {/* header */}
      <div className="mb-[26px] flex items-end justify-between gap-5">
        <div>
          <div className="font-mono text-[12px] uppercase tracking-[2px] text-muted-2">Your library</div>
          <div className="mt-1.5 font-display text-[40px] font-extrabold leading-none tracking-[-1.2px]">
            Library
          </div>
        </div>
        <div className="flex items-center gap-[18px]">
          <div className="text-right">
            <div className="font-display text-[24px] font-bold leading-none">{uploads.length}</div>
            <div className="font-mono text-[11px] tracking-[.5px] text-faint">UPLOADS</div>
          </div>
          <div className="h-[34px] w-px bg-line-sep2" />
          <div className="text-right">
            <div className="font-display text-[24px] font-bold leading-none">{playlists.length}</div>
            <div className="font-mono text-[11px] tracking-[.5px] text-faint">PLAYLISTS</div>
          </div>
        </div>
      </div>

      <UploadDropzone onUploaded={handleUploaded} />

      {/* playlists */}
      <div className="mb-4 mt-[34px] flex items-center justify-between">
        <div className="font-display text-[25px] font-bold tracking-[-.5px]">Your playlists</div>
        <button
          onClick={() => (isAuthenticated ? setCreating(true) : openAuth('signin'))}
          className="flex items-center gap-2 rounded-pill border-[1.5px] border-line bg-white px-4 py-2 text-[14px] font-bold text-muted transition-colors hover:border-lime"
        >
          <PlusIcon size={16} />
          New playlist
        </button>
      </div>
      {!isAuthenticated ? (
        <PlaylistsNotice text="Sign in to see your playlists." />
      ) : playlistsQ.isLoading ? (
        <div className="mb-[38px] grid grid-cols-[repeat(auto-fill,minmax(186px,1fr))] gap-[18px]">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="p-3">
              <div className="aspect-square w-full animate-pulse rounded-[14px] bg-[#e7ead9]" />
              <div className="mt-3 h-4 w-2/3 animate-pulse rounded bg-[#e7ead9]" />
              <div className="mt-2 h-3 w-1/3 animate-pulse rounded bg-[#eef0e2]" />
            </div>
          ))}
        </div>
      ) : playlists.length === 0 ? (
        <PlaylistsNotice text="No playlists yet. Saved playlists will show up here." />
      ) : (
        <div className="mb-[38px] grid grid-cols-[repeat(auto-fill,minmax(186px,1fr))] gap-[18px]">
          {playlists.map((p) => (
            <div
              key={p.id}
              onClick={() => navigate(`/library/playlists/${p.id}`)}
              className="cursor-pointer rounded-[18px] p-3 transition-all hover:bg-white hover:shadow-card"
            >
              <div
                className="relative aspect-square w-full overflow-hidden rounded-[14px] shadow-cover"
                style={{ background: cover(pickFruit(p.id)) }}
              >
                <div
                  className="absolute inset-0"
                  style={{
                    background:
                      'radial-gradient(60% 55% at 28% 22%, rgba(255,255,255,.2), transparent 55%)',
                  }}
                />
                <div className="absolute bottom-3 left-3 rounded-full bg-black/30 px-[9px] py-1 font-mono text-[11px] text-white backdrop-blur-sm">
                  {p.track_count} tracks
                </div>
              </div>
              <div className="mt-[11px] truncate text-[17px] font-bold">{p.name}</div>
              <div className="mt-0.5 truncate text-[13px] text-muted-2">
                {p.description || 'Playlist'}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* uploads list */}
      <div className="mb-3.5 flex items-baseline justify-between">
        <div className="font-display text-[25px] font-bold tracking-[-.5px]">Your uploads</div>
        <span className="font-mono text-[12px] text-faint">local files</span>
      </div>
      <div className="overflow-hidden rounded-[20px] border-[1.5px] border-line-2 bg-white">
        {uploadsQ.isLoading ? (
          Array.from({ length: 4 }).map((_, i) => (
            <div
              key={i}
              className="flex items-center gap-4 border-b border-line-sep px-[18px] py-[13px] last:border-b-0"
            >
              <div className="h-[50px] w-[50px] flex-none animate-pulse rounded-[11px] bg-[#e7ead9]" />
              <div className="flex-1">
                <div className="h-4 w-1/3 animate-pulse rounded bg-[#e7ead9]" />
                <div className="mt-2 h-3 w-1/4 animate-pulse rounded bg-[#eef0e2]" />
              </div>
            </div>
          ))
        ) : uploads.length === 0 ? (
          <div className="px-[18px] py-12 text-center">
            <div className="text-[16px] font-semibold text-muted">No uploads yet</div>
            <div className="mt-1 text-[14px] text-faint">
              Drop a few tracks above and they'll show up here.
            </div>
          </div>
        ) : (
          uploads.map((t) => {
            const fresh = freshIds.has(t.id)
            return (
              <TrackRow
                key={t.key}
                track={t}
                onPlay={() => play(t, uploads)}
                whenLabel={fresh ? 'Just now' : undefined}
                fresh={fresh}
              />
            )
          })
        )}
      </div>

      <PlaylistFormModal
        open={creating}
        onClose={() => setCreating(false)}
        onSaved={(p) => navigate(`/library/playlists/${p.id}`)}
      />
    </div>
  )
}

const PlaylistsNotice = ({ text }: { text: string }) => (
  <div className="mb-[38px] rounded-[20px] border-[1.5px] border-dashed border-line bg-[#fcfdf6] px-[18px] py-12 text-center">
    <div className="text-[15px] text-muted-2">{text}</div>
  </div>
)

export default LibraryView
