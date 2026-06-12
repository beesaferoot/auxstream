import { useState, ReactNode } from 'react'
import { cover, FruitName } from '../lib/covers'

interface CoverProps {
  fruit: FruitName
  thumbnail?: string
  alt?: string
  /** Sizing / radius classes (e.g. "w-full aspect-square rounded-[14px]"). */
  className?: string
  /** Show the soft top-left highlight overlay used on gradient covers. */
  highlight?: boolean
  /** Overlay content (play FAB, source tag, badges). */
  children?: ReactNode
}

/**
 * Track artwork: renders the real thumbnail when present (falling back to the
 * deterministic fruit gradient on missing/broken images), else the gradient.
 */
const Cover = ({
  fruit,
  thumbnail,
  alt = '',
  className = '',
  highlight = true,
  children,
}: CoverProps) => {
  const [broken, setBroken] = useState(false)
  const showImage = thumbnail && !broken

  return (
    <div
      className={`relative overflow-hidden ${className}`}
      style={{ background: cover(fruit) }}
    >
      {showImage && (
        <img
          src={thumbnail}
          alt={alt}
          className="absolute inset-0 h-full w-full object-cover"
          onError={() => setBroken(true)}
        />
      )}
      {!showImage && highlight && (
        <div
          className="absolute inset-0"
          style={{
            background:
              'radial-gradient(60% 55% at 28% 22%, rgba(255,255,255,.22), transparent 55%)',
          }}
        />
      )}
      {children}
    </div>
  )
}

export default Cover
