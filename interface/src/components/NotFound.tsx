import { useNavigate } from 'react-router-dom'

const NotFound = () => {
  const navigate = useNavigate()
  return (
    <div className="flex h-full flex-col items-center justify-center px-[44px] pt-[34px] text-center">
      <div className="font-mono text-[12px] uppercase tracking-[3px] text-muted-2">Error 404</div>
      <div className="mt-3 font-display text-[54px] font-extrabold leading-none tracking-[-1.6px]">
        Lost the beat
      </div>
      <div className="mt-3 max-w-[420px] text-[16px] text-muted-2">
        That page isn't on any of our sources. Head back to the feed and keep the music going.
      </div>
      <button
        onClick={() => navigate('/')}
        className="mt-7 rounded-pill bg-lime px-[26px] py-3.5 text-[17px] font-extrabold text-ink shadow-lime transition-all hover:-translate-y-0.5 hover:shadow-lime-hover"
      >
        Back to feed
      </button>
    </div>
  )
}

export default NotFound
