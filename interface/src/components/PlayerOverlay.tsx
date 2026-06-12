import { useNavigate } from 'react-router-dom'
import { usePlayer } from '../context/PlayerContext'
import Cover from './Cover'
import { fmtSeconds } from '../lib/track'
import { glow } from '../lib/covers'
import { WAVEFORM, isPlayed } from '../lib/waveform'
import {
  PlayIcon,
  PauseIcon,
  PrevIcon,
  NextIcon,
  ShuffleIcon,
  RepeatIcon,
  ChevronDownIcon,
  SearchIcon,
} from './Icons'

/** Immersive full-screen now-playing experience with waveform scrubber + queue. */
const PlayerOverlay = () => {
  const navigate = useNavigate()
  const {
    current,
    queue,
    isPlaying,
    currentTime,
    duration,
    progress,
    isOpen,
    shuffle,
    repeat,
    toggle,
    next,
    prev,
    seekFraction,
    closePlayer,
    play,
    toggleShuffle,
    toggleRepeat,
  } = usePlayer()

  if (!isOpen || !current) return null

  const eqState = isPlaying ? 'running' : 'paused'
  const upNext = queue.filter((t) => t.key !== current.key)

  const seekFromEvent = (e: React.MouseEvent<HTMLDivElement>) => {
    const r = e.currentTarget.getBoundingClientRect()
    seekFraction((e.clientX - r.left) / r.width)
  }

  const goSearch = () => {
    closePlayer()
    navigate('/search')
  }

  return (
    <div className="absolute inset-0 z-40 flex animate-aux-spin-in flex-col overflow-hidden bg-player-bg text-[#f4f6e9]">
      {/* glow */}
      <div
        className="pointer-events-none absolute left-[-10%] top-[-20%] h-[80%] w-[60%] rounded-full"
        style={{
          background: `radial-gradient(circle, ${glow(current.fruit)}, transparent 62%)`,
          opacity: 0.4,
          filter: 'blur(40px)',
        }}
      />

      {/* top bar */}
      <div className="relative z-[1] flex items-center justify-between px-[34px] py-[22px]">
        <button
          onClick={closePlayer}
          className="flex items-center gap-2.5 rounded-pill border-[1.5px] border-border-dark-2 py-2.5 pl-[13px] pr-4 text-[14px] font-semibold text-muted-dark transition-colors hover:border-lime hover:text-lime"
        >
          <ChevronDownIcon size={18} />
          Back to feed
        </button>
        <div className="font-mono text-[11px] uppercase tracking-[2px] text-faint-dark">
          Player mode
        </div>
        <div className="w-[120px]" />
      </div>

      <div className="relative z-[1] flex min-h-0 flex-1 gap-[44px] px-[44px] pb-[40px] pt-1.5">
        {/* now playing */}
        <div className="flex max-w-[560px] flex-[1.15] flex-col justify-center">
          <Cover
            fruit={current.fruit}
            thumbnail={current.thumbnail}
            alt={current.title}
            className="mb-[30px] h-[min(420px,38vh)] w-[min(420px,38vh)] flex-none rounded-hero shadow-player-cover"
          >
            <div className="absolute left-[18px] top-[18px] rounded-full bg-black/30 px-[11px] py-[5px] font-mono text-[12px] tracking-[1px] backdrop-blur-sm">
              {current.source}
            </div>
          </Cover>

          <div className="mb-2 font-mono text-[12px] uppercase tracking-[3px] text-lime">
            ▸ Now playing
          </div>
          <div className="font-display text-[46px] font-extrabold leading-none tracking-[-1.4px]">
            {current.title}
          </div>
          <div className="mt-2 text-[20px] text-muted-dark-2">{current.artist}</div>

          {/* waveform */}
          <div
            onClick={seekFromEvent}
            className="my-[26px] mb-2 flex h-[54px] cursor-pointer items-center gap-[2px]"
          >
            {WAVEFORM.map((bar, i) => (
              <div
                key={i}
                className="flex-1 origin-center animate-aux-wave rounded-[2px]"
                style={{
                  height: `${bar.h}%`,
                  background: isPlayed(i, progress) ? '#b6f03c' : '#363b29',
                  animationDelay: bar.delay,
                  animationPlayState: eqState,
                }}
              />
            ))}
          </div>
          <div className="mb-6 flex justify-between font-mono text-[13px] text-faint-dark">
            <span>{fmtSeconds(currentTime)}</span>
            <span>{duration > 0 ? fmtSeconds(duration) : current.durationLabel}</span>
          </div>

          {/* transport */}
          <div className="flex items-center justify-center gap-[30px]">
            <button
              onClick={toggleShuffle}
              className={`transition-colors hover:text-lime ${shuffle ? 'text-lime' : 'text-muted-dark-3'}`}
            >
              <ShuffleIcon size={22} />
            </button>
            <button onClick={prev} className="text-[#f4f6e9] transition-transform hover:scale-110">
              <PrevIcon size={30} />
            </button>
            <button
              onClick={toggle}
              className="flex h-[74px] w-[74px] items-center justify-center rounded-full bg-lime text-ink shadow-lime-big transition-transform hover:scale-105"
            >
              {isPlaying ? <PauseIcon size={30} /> : <PlayIcon size={30} />}
            </button>
            <button onClick={next} className="text-[#f4f6e9] transition-transform hover:scale-110">
              <NextIcon size={30} />
            </button>
            <button
              onClick={toggleRepeat}
              className={`transition-colors hover:text-lime ${repeat ? 'text-lime' : 'text-muted-dark-3'}`}
            >
              <RepeatIcon size={22} />
            </button>
          </div>
        </div>

        {/* queue */}
        <div className="flex min-h-0 max-w-[520px] flex-1 flex-col">
          <button
            onClick={goSearch}
            className="mb-[22px] flex items-center gap-3 rounded-[14px] border-[1.5px] border-border-dark bg-surface-dark px-4 py-3.5 text-left text-[15px] text-faint-dark transition-colors hover:border-lime"
          >
            <SearchIcon size={20} strokeWidth={2.2} />
            Add to queue — search any source…
          </button>
          <div className="mb-3.5 flex items-baseline justify-between">
            <div className="font-display text-[22px] font-bold tracking-[-.4px]">Up next</div>
            <span className="font-mono text-[12px] text-faint-dark">{upNext.length} in queue</span>
          </div>
          <div className="flex flex-1 flex-col gap-1 overflow-y-auto">
            {upNext.map((t) => (
              <button
                key={t.key}
                onClick={() => play(t, queue)}
                className="flex items-center gap-3.5 rounded-xl p-2.5 px-2.5 text-left transition-colors hover:bg-surface-dark"
              >
                <Cover
                  fruit={t.fruit}
                  thumbnail={t.thumbnail}
                  alt={t.title}
                  className="h-12 w-12 flex-none rounded-[10px] shadow-[0_5px_12px_rgba(0,0,0,.4)]"
                />
                <div className="min-w-0 flex-1">
                  <div className="truncate text-[16px] font-bold">{t.title}</div>
                  <div className="truncate text-[13px] text-faint-dark-2">
                    {t.artist} · {t.source}
                  </div>
                </div>
                <span className="font-mono text-[12px] text-faint-dark">{t.durationLabel}</span>
              </button>
            ))}
            {upNext.length === 0 && (
              <div className="mt-4 text-[14px] text-faint-dark">The queue is empty.</div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default PlayerOverlay
