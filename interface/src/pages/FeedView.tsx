import { useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { getTrendingTracks } from '../utils/api'
import { fromTrack, PlayableTrack } from '../lib/track'
import { usePlayer } from '../context/PlayerContext'
import { useToast } from '../components/ui/Toast'
import Cover from '../components/Cover'
import TrackCard from '../components/TrackCard'
import TrackRow from '../components/TrackRow'
import { PlayIcon, PlusIcon, SearchIcon } from '../components/Icons'

const TEN_MIN = 600_000

function greeting(): { kicker: string; title: string } {
  const now = new Date()
  const day = now.toLocaleDateString(undefined, { weekday: 'long' })
  const h = now.getHours()
  const part = h < 12 ? 'good morning' : h < 18 ? 'good afternoon' : 'good evening'
  return { kicker: `${day} · ${part}`, title: 'Fresh today' }
}

const SectionHeading = ({
  title,
  chip,
  onSeeAll,
}: {
  title: string
  chip?: string
  onSeeAll?: () => void
}) => (
  <div className="mb-4 flex items-baseline justify-between">
    <div className="flex items-baseline gap-3">
      <div className="font-display text-[25px] font-bold tracking-[-.5px]">{title}</div>
      {chip && (
        <div className="rounded-full border border-line bg-white px-[9px] py-[3px] font-mono text-[12px] text-faint">
          {chip}
        </div>
      )}
    </div>
    {onSeeAll && (
      <span onClick={onSeeAll} className="cursor-pointer text-[14px] font-semibold text-muted-3">
        See all →
      </span>
    )}
  </div>
)

const FeedView = () => {
  const navigate = useNavigate()
  const { play, enqueue } = usePlayer()
  const { toast } = useToast()
  const { kicker, title } = useMemo(greeting, [])

  const trendingQ = useQuery({
    queryKey: ['feed', 'trending'],
    queryFn: () => getTrendingTracks(12, 1, 'trending', 30),
    staleTime: TEN_MIN,
  })
  const recentQ = useQuery({
    queryKey: ['feed', 'recent'],
    queryFn: () => getTrendingTracks(8, 1, 'recent'),
    staleTime: TEN_MIN,
  })

  const trending: PlayableTrack[] = useMemo(
    () => (trendingQ.data?.data ?? []).map(fromTrack),
    [trendingQ.data]
  )
  const recent: PlayableTrack[] = useMemo(
    () => (recentQ.data?.data ?? []).map(fromTrack),
    [recentQ.data]
  )

  const hero = trending[0]
  const grid = trending.slice(0, 6)
  const loading = trendingQ.isLoading

  return (
    <div className="px-[44px] pt-[34px]">
      {/* topbar */}
      <div className="mb-[30px] flex animate-aux-up items-center justify-between gap-5">
        <div>
          <div className="font-mono text-[12px] uppercase tracking-[2px] text-muted-2">{kicker}</div>
          <div className="mt-1.5 font-display text-[40px] font-extrabold leading-none tracking-[-1.2px]">
            {title}
          </div>
        </div>
        <button
          onClick={() => navigate('/search')}
          className="flex items-center gap-2.5 rounded-pill border-[1.5px] border-line bg-white py-[11px] pl-[18px] pr-4 text-[15px] text-muted shadow-[0_2px_10px_rgba(20,30,0,.04)] transition-all hover:border-lime hover:shadow-[0_4px_18px_rgba(120,170,0,.12)]"
        >
          <SearchIcon size={17} strokeWidth={2.2} />
          Search every source
          <span className="rounded-[7px] border border-line bg-[#f1f2e7] px-[7px] py-0.5 font-mono text-[12px] text-muted-3">
            ⌘K
          </span>
        </button>
      </div>

      {/* hero */}
      {hero ? (
        <div className="relative mb-[38px] flex animate-aux-up gap-[30px] overflow-hidden rounded-hero bg-ink p-6 text-[#f4f6e9]">
          <div
            className="absolute right-[-60px] top-[-80px] h-[340px] w-[340px] animate-aux-glow rounded-full"
            style={{
              background: 'radial-gradient(circle,#b6f03c,transparent 62%)',
              opacity: 0.5,
              filter: 'blur(8px)',
            }}
          />
          <button
            onClick={() => play(hero, trending)}
            className="relative h-[280px] w-[280px] flex-none"
          >
            <Cover
              fruit={hero.fruit}
              thumbnail={hero.thumbnail}
              alt={hero.title}
              className="h-full w-full rounded-[20px] shadow-hero"
            >
              <div className="absolute left-4 top-4 rounded-full bg-black/30 px-[9px] py-1 font-mono text-[11px] tracking-[1px] backdrop-blur-sm">
                {hero.source}
              </div>
              <div className="absolute bottom-4 right-4 flex h-[58px] w-[58px] items-center justify-center rounded-full bg-lime text-ink shadow-[0_8px_22px_rgba(0,0,0,.35)]">
                <PlayIcon size={24} />
              </div>
            </Cover>
          </button>
          <div className="relative z-[1] flex flex-1 flex-col justify-center">
            <div className="font-mono text-[12px] uppercase tracking-[3px] text-lime">
              ◆ Squeezed pick
            </div>
            <div className="my-2 mt-3 font-display text-[54px] font-extrabold leading-[.98] tracking-[-1.6px]">
              {hero.title}
            </div>
            <div className="mb-2 text-[20px] text-muted-dark-2">{hero.artist}</div>
            <div className="mb-[26px] max-w-[460px] text-[16px] leading-[1.45] text-[#878c75]">
              Pulled together from all your sources — the one track to press play on right now.
            </div>
            <div className="flex gap-3">
              <button
                onClick={() => play(hero, trending)}
                className="flex items-center gap-2.5 rounded-pill bg-lime px-[26px] py-3.5 text-[17px] font-extrabold text-ink shadow-lime transition-all hover:-translate-y-0.5 hover:shadow-lime-hover"
              >
                <PlayIcon size={20} />
                Play now
              </button>
              <button
                onClick={() => {
                  const added = enqueue(hero)
                  toast({
                    title: added ? 'Added to queue' : 'Already in queue',
                    description: hero.title,
                    status: added ? 'success' : 'info',
                  })
                }}
                className="flex items-center gap-2.5 rounded-pill border-[1.5px] border-border-dark-3 px-[22px] py-3.5 text-[16px] font-bold text-[#f4f6e9] transition-colors hover:border-lime hover:text-lime"
              >
                <PlusIcon size={18} />
                Queue
              </button>
            </div>
          </div>
        </div>
      ) : loading ? (
        <div className="mb-[38px] h-[328px] animate-pulse rounded-hero bg-[#e7ead9]" />
      ) : (
        <EmptyHero onSearch={() => navigate('/search')} onUpload={() => navigate('/library')} />
      )}

      {/* trending */}
      <div className="animate-aux-up">
        <SectionHeading title="Trending now" chip="across sources" onSeeAll={() => navigate('/search')} />
        {loading ? (
          <GridSkeleton />
        ) : (
          <div className="mb-10 grid grid-cols-[repeat(auto-fill,minmax(168px,1fr))] gap-[18px]">
            {grid.map((t) => (
              <TrackCard key={t.key} track={t} onPlay={() => play(t, trending)} />
            ))}
          </div>
        )}
      </div>

      {/* recently added */}
      {(recent.length > 0 || recentQ.isLoading) && (
        <div className="animate-aux-up">
          <SectionHeading title="Recently added" onSeeAll={() => navigate('/library')} />
          <div className="overflow-hidden rounded-[20px] border-[1.5px] border-line-2 bg-white">
            {recentQ.isLoading
              ? Array.from({ length: 4 }).map((_, i) => <RowSkeleton key={i} />)
              : recent.map((t) => (
                  <TrackRow key={t.key} track={t} onPlay={() => play(t, recent)} />
                ))}
          </div>
        </div>
      )}
    </div>
  )
}

const EmptyHero =({ onSearch, onUpload }: { onSearch: () => void; onUpload: () => void }) => (
  <div className="mb-[38px] flex flex-col items-start gap-4 rounded-hero bg-ink p-8 text-[#f4f6e9]">
    <div className="font-mono text-[12px] uppercase tracking-[3px] text-lime">◆ Squeezed pick</div>
    <div className="font-display text-[40px] font-extrabold leading-none tracking-[-1.4px]">
      Nothing fresh yet
    </div>
    <div className="max-w-[460px] text-[16px] leading-[1.45] text-[#878c75]">
      Search across YouTube, SoundCloud and your library, or upload your own tracks to get the feed
      flowing.
    </div>
    <div className="flex gap-3">
      <button
        onClick={onSearch}
        className="rounded-pill bg-lime px-[26px] py-3.5 text-[17px] font-extrabold text-ink shadow-lime transition-all hover:-translate-y-0.5 hover:shadow-lime-hover"
      >
        Search music
      </button>
      <button
        onClick={onUpload}
        className="rounded-pill border-[1.5px] border-border-dark-3 px-[22px] py-3.5 text-[16px] font-bold text-[#f4f6e9] transition-colors hover:border-lime hover:text-lime"
      >
        Upload tracks
      </button>
    </div>
  </div>
)

const GridSkeleton = () => (
  <div className="mb-10 grid grid-cols-[repeat(auto-fill,minmax(168px,1fr))] gap-[18px]">
    {Array.from({ length: 6 }).map((_, i) => (
      <div key={i} className="p-2.5">
        <div className="aspect-square w-full animate-pulse rounded-[14px] bg-[#e7ead9]" />
        <div className="mt-3 h-4 w-3/4 animate-pulse rounded bg-[#e7ead9]" />
        <div className="mt-2 h-3 w-1/2 animate-pulse rounded bg-[#eef0e2]" />
      </div>
    ))}
  </div>
)

const RowSkeleton = () => (
  <div className="flex items-center gap-4 border-b border-line-sep px-[18px] py-[13px] last:border-b-0">
    <div className="h-[50px] w-[50px] flex-none animate-pulse rounded-[11px] bg-[#e7ead9]" />
    <div className="flex-1">
      <div className="h-4 w-1/3 animate-pulse rounded bg-[#e7ead9]" />
      <div className="mt-2 h-3 w-1/4 animate-pulse rounded bg-[#eef0e2]" />
    </div>
  </div>
)

export default FeedView
