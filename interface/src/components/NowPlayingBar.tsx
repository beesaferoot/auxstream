import { usePlayer } from '../context/PlayerContext'
import Cover from './Cover'
import { fmtSeconds } from '../lib/track'
import { PlayIcon, PauseIcon, PrevIcon, NextIcon, VolumeIcon, ExpandIcon } from './Icons'

const EQ_DELAYS = ['0s', '.3s', '.15s', '.45s']

/** Persistent floating now-playing bar. The audio engine itself lives in PlayerContext. */
const NowPlayingBar = () => {
  const {
    current,
    isPlaying,
    currentTime,
    duration,
    progress,
    volume,
    toggle,
    next,
    prev,
    seekFraction,
    setVolume,
    openPlayer,
  } = usePlayer()

  const fractionFromEvent = (e: React.MouseEvent<HTMLDivElement>) => {
    const r = e.currentTarget.getBoundingClientRect()
    return Math.min(1, Math.max(0, (e.clientX - r.left) / r.width))
  }

  const eqState = isPlaying ? 'running' : 'paused'
  const hasTrack = !!current
  const pct = `${(progress * 100).toFixed(1)}%`

  return (
    <div className="absolute bottom-4 left-[18px] right-[18px] flex h-[84px] items-center gap-[18px] rounded-bar bg-ink px-[18px] text-[#f4f6e9] shadow-bar">
      {/* track cluster -> open player */}
      <button
        onClick={() => hasTrack && openPlayer()}
        disabled={!hasTrack}
        className="flex w-[260px] flex-none items-center gap-3.5 text-left transition-opacity enabled:hover:opacity-85 disabled:cursor-default"
      >
        {hasTrack ? (
          <Cover
            fruit={current.fruit}
            thumbnail={current.thumbnail}
            alt={current.title}
            className="h-[54px] w-[54px] flex-none rounded-[11px] shadow-[0_6px_14px_rgba(0,0,0,.4)]"
          />
        ) : (
          <div className="h-[54px] w-[54px] flex-none rounded-[11px] bg-surface-dark" />
        )}
        <div className="min-w-0 flex-1">
          <div className="truncate text-[15px] font-bold">
            {current?.title ?? 'Nothing playing'}
          </div>
          <div className="truncate text-[13px] text-muted-dark-3">
            {current?.artist ?? 'Pick a track to start'}
          </div>
        </div>
        {hasTrack && (
          <div className="flex h-[18px] flex-none items-end gap-[2.5px]">
            {EQ_DELAYS.map((d, i) => (
              <div
                key={i}
                className="w-[3px] origin-bottom animate-aux-eq rounded-[2px] bg-lime"
                style={{ height: '100%', animationDelay: d, animationPlayState: eqState }}
              />
            ))}
          </div>
        )}
      </button>

      {/* transport */}
      <div className="flex flex-none items-center gap-3.5">
        <button
          onClick={prev}
          disabled={!hasTrack}
          className="cursor-pointer p-1 text-muted-dark transition-colors enabled:hover:text-white disabled:opacity-40"
        >
          <PrevIcon size={22} />
        </button>
        <button
          onClick={toggle}
          disabled={!hasTrack}
          className="flex h-[46px] w-[46px] items-center justify-center rounded-full bg-lime text-ink transition-transform enabled:hover:scale-105 disabled:opacity-40"
        >
          {isPlaying ? <PauseIcon size={20} /> : <PlayIcon size={20} />}
        </button>
        <button
          onClick={next}
          disabled={!hasTrack}
          className="cursor-pointer p-1 text-muted-dark transition-colors enabled:hover:text-white disabled:opacity-40"
        >
          <NextIcon size={22} />
        </button>
      </div>

      {/* scrubber */}
      <div className="flex min-w-0 flex-1 items-center gap-3">
        <span className="w-[38px] text-right font-mono text-[12px] text-muted-dark-3">
          {fmtSeconds(currentTime)}
        </span>
        <div
          onClick={(e) => hasTrack && seekFraction(fractionFromEvent(e))}
          className="relative h-1.5 flex-1 cursor-pointer rounded bg-border-dark"
        >
          <div className="absolute left-0 top-0 h-1.5 rounded bg-lime" style={{ width: pct }} />
          <div
            className="absolute top-[-4px] h-3.5 w-3.5 -translate-x-[7px] rounded-full border-2 border-lime bg-white"
            style={{ left: pct }}
          />
        </div>
        <span className="w-[38px] font-mono text-[12px] text-muted-dark-3">
          {duration > 0 ? fmtSeconds(duration) : current?.durationLabel ?? '0:00'}
        </span>
      </div>

      {/* volume + expand */}
      <div className="flex flex-none items-center gap-3">
        <span className="text-muted-dark-3">
          <VolumeIcon size={20} />
        </span>
        <div
          onClick={(e) => setVolume(fractionFromEvent(e))}
          className="relative h-[5px] w-[74px] cursor-pointer rounded-[3px] bg-border-dark"
        >
          <div
            className="absolute left-0 top-0 h-[5px] rounded-[3px] bg-muted-dark"
            style={{ width: `${volume * 100}%` }}
          />
        </div>
        <button
          onClick={() => hasTrack && openPlayer()}
          disabled={!hasTrack}
          className="cursor-pointer p-1 text-muted-dark-3 transition-colors enabled:hover:text-lime disabled:opacity-40"
        >
          <ExpandIcon size={20} />
        </button>
      </div>
    </div>
  )
}

export default NowPlayingBar
