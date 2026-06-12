import { useMemo, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { getTrendingTracks } from '../utils/api'
import { fromTrack, PlayableTrack } from '../lib/track'
import { usePlayer } from '../context/PlayerContext'
import { PLAYLISTS } from '../lib/placeholder'
import { cover } from '../lib/covers'
import UploadDropzone from '../components/UploadDropzone'
import TrackRow from '../components/TrackRow'

const LibraryView = () => {
  const { play } = usePlayer()
  const queryClient = useQueryClient()
  // Track ids uploaded this session so we can flag them "fresh".
  const [freshIds, setFreshIds] = useState<Set<string>>(new Set())

  const uploadsQ = useQuery({
    queryKey: ['library', 'uploads'],
    queryFn: () => getTrendingTracks(40, 1, 'recent'),
    staleTime: 60_000,
  })

  const uploads: PlayableTrack[] = useMemo(
    () => (uploadsQ.data?.data ?? []).map(fromTrack),
    [uploadsQ.data]
  )

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
            <div className="font-display text-[24px] font-bold leading-none">{PLAYLISTS.length}</div>
            <div className="font-mono text-[11px] tracking-[.5px] text-faint">PLAYLISTS</div>
          </div>
        </div>
      </div>

      <UploadDropzone onUploaded={handleUploaded} />

      {/* playlists */}
      <div className="my-4 mt-[34px] mb-4 font-display text-[25px] font-bold tracking-[-.5px]">
        Your playlists
      </div>
      <div className="mb-[38px] grid grid-cols-[repeat(auto-fill,minmax(186px,1fr))] gap-[18px]">
        {PLAYLISTS.map((p) => (
          <div
            key={p.id}
            className="cursor-pointer rounded-[18px] p-3 transition-all hover:bg-white hover:shadow-card"
          >
            <div
              className="relative aspect-square w-full overflow-hidden rounded-[14px] shadow-cover"
              style={{ background: cover(p.fruit) }}
            >
              <div
                className="absolute inset-0"
                style={{
                  background:
                    'radial-gradient(60% 55% at 28% 22%, rgba(255,255,255,.2), transparent 55%)',
                }}
              />
              <div className="absolute bottom-3 left-3 rounded-full bg-black/30 px-[9px] py-1 font-mono text-[11px] text-white backdrop-blur-sm">
                {p.count} tracks
              </div>
            </div>
            <div className="mt-[11px] truncate text-[17px] font-bold">{p.name}</div>
            <div className="mt-0.5 text-[13px] text-muted-2">{p.sub}</div>
          </div>
        ))}
      </div>

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
    </div>
  )
}

export default LibraryView
