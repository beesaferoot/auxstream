import Cover from './Cover'
import { PlayIcon } from './Icons'
import { PlayableTrack } from '../lib/track'

interface TrackRowProps {
  track: PlayableTrack
  onPlay: () => void
  /** "list" = inside a white container with dividers; "search" = standalone rounded row. */
  variant?: 'list' | 'search'
  /** Render the source as a pill (search) instead of plain mono text (lists). */
  sourcePill?: boolean
  /** Optional right-aligned "when" label (e.g. "2 days ago", "Just now"). */
  whenLabel?: string
  /** Show the lime "new" dot on the cover. */
  fresh?: boolean
}

/** A horizontal track row: cover, title/artist, source, optional when, duration, play. */
const TrackRow = ({
  track,
  onPlay,
  variant = 'list',
  sourcePill = false,
  whenLabel,
  fresh = false,
}: TrackRowProps) => {
  const isSearch = variant === 'search'
  const coverSize = isSearch ? 'h-[54px] w-[54px]' : 'h-[50px] w-[50px]'
  const playSize = isSearch ? 'h-9 w-9' : 'h-[34px] w-[34px]'

  return (
    <div
      onClick={onPlay}
      className={
        isSearch
          ? 'flex cursor-pointer items-center gap-4 rounded-[14px] p-3 transition-all hover:bg-white hover:shadow-row'
          : 'flex cursor-pointer items-center gap-4 border-b border-line-sep px-[18px] py-[13px] transition-colors last:border-b-0 hover:bg-[#f7f9ef]'
      }
    >
      <Cover
        fruit={track.fruit}
        thumbnail={track.thumbnail}
        alt={track.title}
        className={`${coverSize} flex-none rounded-[11px] shadow-cover`}
      >
        {fresh && (
          <div className="absolute right-[-4px] top-[-4px] h-4 w-4 rounded-full border-2 border-white bg-lime" />
        )}
      </Cover>

      <div className="min-w-0 flex-1">
        <div className={`truncate font-bold ${isSearch ? 'text-[17px]' : 'text-[16px]'}`}>
          {track.title}
        </div>
        <div className="truncate text-[14px] text-muted-2">{track.artist}</div>
      </div>

      {sourcePill ? (
        <span className="rounded-full border border-[#e7e9da] bg-[#f1f2e7] px-[9px] py-[3px] font-mono text-[11px] text-faint">
          {track.source}
        </span>
      ) : whenLabel ? (
        <span className="w-[120px] text-right font-mono text-[11px] tracking-[.5px] text-faint-2">
          {whenLabel}
        </span>
      ) : (
        <span className="w-[90px] text-right font-mono text-[11px] tracking-[.5px] text-faint-2">
          {track.source}
        </span>
      )}

      <span className="w-[46px] text-right font-mono text-[13px] text-faint">
        {track.durationLabel}
      </span>

      <button
        onClick={(e) => {
          e.stopPropagation()
          onPlay()
        }}
        className={`${playSize} flex flex-none items-center justify-center rounded-full border-[1.5px] border-line bg-white text-muted transition-colors hover:border-lime hover:bg-lime-tint hover:text-ink`}
      >
        <PlayIcon size={isSearch ? 16 : 15} />
      </button>
    </div>
  )
}

export default TrackRow
