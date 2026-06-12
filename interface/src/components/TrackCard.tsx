import Cover from './Cover'
import { PlayIcon } from './Icons'
import { PlayableTrack } from '../lib/track'

interface TrackCardProps {
  track: PlayableTrack
  onPlay: () => void
}

/** Square gradient/artwork cover card with a hover play FAB (feed "Trending now"). */
const TrackCard = ({ track, onPlay }: TrackCardProps) => (
  <button
    onClick={onPlay}
    className="group block w-full rounded-[18px] p-2.5 text-left transition-all hover:bg-white hover:shadow-card"
  >
    <Cover
      fruit={track.fruit}
      thumbnail={track.thumbnail}
      alt={track.title}
      className="aspect-square w-full rounded-[14px] shadow-cover"
    >
      <div className="absolute bottom-2.5 right-2.5 flex h-[42px] w-[42px] translate-y-1.5 items-center justify-center rounded-full bg-lime text-ink opacity-0 shadow-[0_6px_16px_rgba(0,0,0,.3)] transition-all group-hover:translate-y-0 group-hover:opacity-100">
        <PlayIcon size={18} />
      </div>
    </Cover>
    <div className="mt-[11px] truncate text-[16px] font-bold">{track.title}</div>
    <div className="mt-[3px] flex items-center gap-[7px]">
      <span className="truncate text-[14px] text-muted-2">{track.artist}</span>
      <span className="flex-none font-mono text-[10px] text-faint-2">· {track.source}</span>
    </div>
  </button>
)

export default TrackCard
