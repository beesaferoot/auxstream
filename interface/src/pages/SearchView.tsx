import { useEffect, useMemo, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { searchTracks } from '../utils/api'
import { fromSearchResult, PlayableTrack } from '../lib/track'
import { usePlayer } from '../context/PlayerContext'
import TrackRow from '../components/TrackRow'
import { SearchIcon } from '../components/Icons'

type Filter = 'All' | 'YouTube' | 'SoundCloud' | 'Local'
const FILTERS: Filter[] = ['All', 'YouTube', 'SoundCloud', 'Local']
const SOURCE_PARAM: Record<Filter, 'local' | 'youtube' | 'soundcloud' | undefined> = {
  All: undefined,
  YouTube: 'youtube',
  SoundCloud: 'soundcloud',
  Local: 'local',
}

const SearchView = () => {
  const { play } = usePlayer()
  const [query, setQuery] = useState('')
  const [debounced, setDebounced] = useState('')
  const [filter, setFilter] = useState<Filter>('All')
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    inputRef.current?.focus()
  }, [])

  // Debounce typing before hitting the API.
  useEffect(() => {
    const id = window.setTimeout(() => setDebounced(query.trim()), 350)
    return () => window.clearTimeout(id)
  }, [query])

  const { data, isFetching, isError } = useQuery({
    queryKey: ['search', debounced, filter],
    queryFn: () => searchTracks(debounced, SOURCE_PARAM[filter], 30),
    enabled: debounced.length > 0,
    retry: false,
    staleTime: 60_000,
  })

  const results: PlayableTrack[] = useMemo(
    () => (data?.results ?? []).map(fromSearchResult),
    [data]
  )

  const hasQuery = debounced.length > 0

  return (
    <div className="animate-aux-pop px-[44px] pt-[34px]">
      {/* command bar */}
      <div className="mb-4 flex items-center gap-3.5 rounded-[18px] border-2 border-ink bg-white px-5 py-4 shadow-[0_14px_34px_rgba(20,30,0,.1)]">
        <SearchIcon size={24} strokeWidth={2.4} className="text-ink" />
        <input
          ref={inputRef}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search songs, artists, or paste a link…"
          className="flex-1 bg-transparent text-[22px] font-semibold text-ink-text outline-none placeholder:text-faint"
        />
        <span className="font-mono text-[12px] text-faint">
          {hasQuery ? `${results.length} results` : ''}
        </span>
      </div>

      {/* filter chips */}
      <div className="mb-[26px] flex gap-[9px]">
        {FILTERS.map((f) => {
          const active = filter === f
          return (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={`rounded-[24px] border-[1.5px] px-4 py-2 text-[14px] font-bold transition-colors ${
                active
                  ? 'border-ink bg-ink text-lime'
                  : 'border-line bg-white text-muted hover:border-lime'
              }`}
            >
              {f}
            </button>
          )
        })}
      </div>

      {/* results */}
      {!hasQuery ? (
        <IdleState />
      ) : isFetching && results.length === 0 ? (
        <div className="space-y-1">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="flex items-center gap-4 p-3">
              <div className="h-[54px] w-[54px] flex-none animate-pulse rounded-[11px] bg-[#e7ead9]" />
              <div className="flex-1">
                <div className="h-4 w-1/3 animate-pulse rounded bg-[#e7ead9]" />
                <div className="mt-2 h-3 w-1/4 animate-pulse rounded bg-[#eef0e2]" />
              </div>
            </div>
          ))}
        </div>
      ) : isError ? (
        <div className="py-16 text-center">
          <div className="text-[18px] font-semibold text-muted">
            {filter === 'All' ? 'Search is unavailable right now' : `${filter} search is unavailable right now`}
          </div>
          <div className="mt-1 text-[14px] text-faint">
            {filter === 'All' || filter === 'Local'
              ? 'Please try again in a moment.'
              : 'This source isn’t connected yet — try “Local” or “All”.'}
          </div>
        </div>
      ) : results.length === 0 ? (
        <div className="py-16 text-center">
          <div className="text-[18px] font-semibold text-muted">No results found</div>
          <div className="mt-1 text-[14px] text-faint">Try a different search term or source.</div>
        </div>
      ) : (
        <div>
          {results.map((t) => (
            <TrackRow
              key={t.key}
              track={t}
              onPlay={() => play(t, results)}
              variant="search"
              sourcePill
            />
          ))}
        </div>
      )}
    </div>
  )
}

const IdleState = () => (
  <div className="flex flex-col items-center gap-3 py-20 text-center">
    <div className="flex h-16 w-16 items-center justify-center rounded-full bg-white text-faint shadow-cover">
      <SearchIcon size={28} strokeWidth={2.2} />
    </div>
    <div className="text-[18px] font-semibold text-muted">Search every source at once</div>
    <div className="max-w-[360px] text-[14px] text-faint">
      One bar across YouTube, SoundCloud and your local library. Start typing to see results.
    </div>
  </div>
)

export default SearchView
