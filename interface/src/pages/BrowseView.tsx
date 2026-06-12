import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getTrendingTracks } from '../utils/api'
import { fromTrack, PlayableTrack } from '../lib/track'
import { usePlayer } from '../context/PlayerContext'
import TrackRow from '../components/TrackRow'
import { ChevronDownIcon } from '../components/Icons'

const PAGE = 24

type Sort = 'trending' | 'recent'
const META: Record<Sort, { kicker: string; title: string }> = {
  trending: { kicker: 'Across sources', title: 'Trending now' },
  recent: { kicker: 'Fresh in your library', title: 'Recently added' },
}

/** Full, paginated list behind the Feed's "See all" links. */
const BrowseView = () => {
  const { sort: sortParam } = useParams()
  const sort: Sort = sortParam === 'recent' ? 'recent' : 'trending'
  const navigate = useNavigate()
  const { play } = usePlayer()

  const [page, setPage] = useState(1)
  const [all, setAll] = useState<PlayableTrack[]>([])
  const [hasMore, setHasMore] = useState(true)

  // Reset accumulated list when the sort (route) changes.
  useEffect(() => {
    setPage(1)
    setAll([])
    setHasMore(true)
  }, [sort])

  const { data, isFetching } = useQuery({
    queryKey: ['browse', sort, page],
    queryFn: () => getTrendingTracks(PAGE, page, sort, sort === 'trending' ? 30 : undefined),
    staleTime: 60_000,
  })

  useEffect(() => {
    if (!data?.data) return
    const mapped = data.data.map(fromTrack)
    setAll((prev) => (page === 1 ? mapped : [...prev, ...mapped]))
    setHasMore(data.data.length === PAGE)
  }, [data, page])

  const meta = META[sort]
  const loadingFirst = isFetching && page === 1 && all.length === 0

  return (
    <div className="animate-aux-pop px-[44px] pt-[34px]">
      <button
        onClick={() => navigate('/')}
        className="mb-5 flex items-center gap-1.5 text-[14px] font-semibold text-muted-2 transition-colors hover:text-ink"
      >
        <span className="rotate-90">
          <ChevronDownIcon size={16} />
        </span>
        Feed
      </button>

      <div className="mb-6">
        <div className="font-mono text-[12px] uppercase tracking-[2px] text-muted-2">{meta.kicker}</div>
        <div className="mt-1.5 font-display text-[40px] font-extrabold leading-none tracking-[-1.2px]">
          {meta.title}
        </div>
      </div>

      <div className="overflow-hidden rounded-[20px] border-[1.5px] border-line-2 bg-white">
        {loadingFirst ? (
          Array.from({ length: 8 }).map((_, i) => (
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
        ) : all.length === 0 ? (
          <div className="px-[18px] py-12 text-center">
            <div className="text-[16px] font-semibold text-muted">Nothing here yet</div>
            <div className="mt-1 text-[14px] text-faint">Tracks will appear as the library grows.</div>
          </div>
        ) : (
          all.map((t) => <TrackRow key={t.key} track={t} onPlay={() => play(t, all)} />)
        )}
      </div>

      {!loadingFirst && hasMore && all.length > 0 && (
        <div className="flex justify-center py-8">
          <button
            onClick={() => setPage((p) => p + 1)}
            disabled={isFetching}
            className="rounded-pill border-[1.5px] border-line bg-white px-6 py-3 text-[15px] font-bold text-muted transition-colors hover:border-lime disabled:opacity-60"
          >
            {isFetching ? 'Loading…' : 'Load more'}
          </button>
        </div>
      )}
    </div>
  )
}

export default BrowseView
