// Inline SVG icon set, lifted from the design handoff (AuxStream.dc.html) for pixel
// fidelity. Stroke icons use currentColor; fill icons use currentColor. Pass `size`
// (px) and `className` (for color) as needed.

interface IconProps {
  size?: number
  className?: string
  strokeWidth?: number
}

const stroke = (sw = 2) => ({
  fill: 'none',
  stroke: 'currentColor',
  strokeWidth: sw,
  strokeLinecap: 'round' as const,
  strokeLinejoin: 'round' as const,
})

const svgBase = (size: number, className?: string) => ({
  width: size,
  height: size,
  viewBox: '0 0 24 24',
  className,
})

export const LogoGlyph = ({ size = 22, className }: IconProps) => (
  <svg {...svgBase(size, className)} fill="none" stroke="currentColor" strokeWidth={2.4} strokeLinecap="round" strokeLinejoin="round">
    <path d="M9 18V6l10-2v12" />
    <circle cx="6" cy="18" r="2.6" fill="currentColor" stroke="none" />
    <circle cx="16" cy="16" r="2.6" fill="currentColor" stroke="none" />
  </svg>
)

export const FeedIcon = ({ size = 22, className, strokeWidth = 2 }: IconProps) => (
  <svg {...svgBase(size, className)} {...stroke(strokeWidth)}>
    <rect x="3.5" y="3.5" width="7" height="7" rx="2" />
    <rect x="13.5" y="3.5" width="7" height="7" rx="2" />
    <rect x="3.5" y="13.5" width="7" height="7" rx="2" />
    <rect x="13.5" y="13.5" width="7" height="7" rx="2" />
  </svg>
)

export const SearchIcon = ({ size = 22, className, strokeWidth = 2 }: IconProps) => (
  <svg {...svgBase(size, className)} {...stroke(strokeWidth)}>
    <circle cx="11" cy="11" r="7" />
    <line x1="16.5" y1="16.5" x2="21" y2="21" />
  </svg>
)

export const LibraryIcon = ({ size = 22, className, strokeWidth = 2 }: IconProps) => (
  <svg {...svgBase(size, className)} {...stroke(strokeWidth)}>
    <path d="M4 19V7a2 2 0 0 1 2-2h2" />
    <path d="M11 19V5l9-1v13" />
    <circle cx="8.5" cy="18.5" r="2" />
    <circle cx="18.5" cy="16.5" r="2" />
  </svg>
)

export const PlayIcon = ({ size = 24, className }: IconProps) => (
  <svg {...svgBase(size, className)} fill="currentColor">
    <path d="M8 5v14l11-7z" />
  </svg>
)

export const PauseIcon = ({ size = 24, className }: IconProps) => (
  <svg {...svgBase(size, className)} fill="currentColor">
    <rect x="6" y="5" width="4" height="14" rx="1.3" />
    <rect x="14" y="5" width="4" height="14" rx="1.3" />
  </svg>
)

export const PrevIcon = ({ size = 22, className }: IconProps) => (
  <svg {...svgBase(size, className)} fill="currentColor">
    <path d="M7 5v14h2.2V5zM20 5l-9.3 7L20 19z" />
  </svg>
)

export const NextIcon = ({ size = 22, className }: IconProps) => (
  <svg {...svgBase(size, className)} fill="currentColor">
    <path d="M17 5v14h-2.2V5zM4 5l9.3 7L4 19z" />
  </svg>
)

export const ShuffleIcon = ({ size = 22, className, strokeWidth = 2 }: IconProps) => (
  <svg {...svgBase(size, className)} {...stroke(strokeWidth)}>
    <path d="M16 3h5v5" />
    <path d="M21 3l-7 7" />
    <path d="M8 21H3v-5" />
    <path d="M3 21l7-7" />
    <path d="M3 8V3h5" />
    <path d="M16 21h5v-5" />
  </svg>
)

export const RepeatIcon = ({ size = 22, className, strokeWidth = 2 }: IconProps) => (
  <svg {...svgBase(size, className)} {...stroke(strokeWidth)}>
    <path d="M17 2l4 4-4 4" />
    <path d="M3 11V9a4 4 0 0 1 4-4h14" />
    <path d="M7 22l-4-4 4-4" />
    <path d="M21 13v2a4 4 0 0 1-4 4H3" />
  </svg>
)

export const VolumeIcon = ({ size = 20, className, strokeWidth = 2 }: IconProps) => (
  <svg {...svgBase(size, className)} {...stroke(strokeWidth)}>
    <path d="M4 9v6h4l5 4V5L8 9z" />
    <path d="M17 8.5a5 5 0 0 1 0 7" />
  </svg>
)

export const ExpandIcon = ({ size = 20, className, strokeWidth = 2.2 }: IconProps) => (
  <svg {...svgBase(size, className)} {...stroke(strokeWidth)}>
    <path d="M8 4H4v4M16 4h4v4M16 20h4v-4M8 20H4v-4" />
  </svg>
)

export const PlusIcon = ({ size = 18, className, strokeWidth = 2.4 }: IconProps) => (
  <svg {...svgBase(size, className)} {...stroke(strokeWidth)}>
    <line x1="12" y1="5" x2="12" y2="19" />
    <line x1="5" y1="12" x2="19" y2="12" />
  </svg>
)

export const ChevronDownIcon = ({ size = 18, className, strokeWidth = 2.4 }: IconProps) => (
  <svg {...svgBase(size, className)} {...stroke(strokeWidth)}>
    <path d="M6 9l6 6 6-6" />
  </svg>
)

export const UploadIcon = ({ size = 26, className, strokeWidth = 2.2 }: IconProps) => (
  <svg {...svgBase(size, className)} {...stroke(strokeWidth)}>
    <path d="M12 16V4" />
    <path d="M7 9l5-5 5 5" />
    <path d="M4 17v2a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1v-2" />
  </svg>
)

export const GearIcon = ({ size = 18, className, strokeWidth = 2 }: IconProps) => (
  <svg {...svgBase(size, className)} {...stroke(strokeWidth)}>
    <circle cx="12" cy="12" r="3" />
    <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z" />
  </svg>
)

export const LinkIcon = ({ size = 18, className, strokeWidth = 2 }: IconProps) => (
  <svg {...svgBase(size, className)} {...stroke(strokeWidth)}>
    <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
    <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
  </svg>
)

export const LogoutIcon = ({ size = 18, className, strokeWidth = 2 }: IconProps) => (
  <svg {...svgBase(size, className)} {...stroke(strokeWidth)}>
    <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
    <path d="M16 17l5-5-5-5" />
    <path d="M21 12H9" />
  </svg>
)
