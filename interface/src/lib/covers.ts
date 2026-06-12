// Procedural "fruit" cover palettes — the fallback artwork used whenever a track
// has no real thumbnail. Each palette is [light, mid, deep]. See the design handoff
// README ("Color — fruit cover palettes").

export type FruitName = 'lime' | 'coral' | 'lemon' | 'berry' | 'teal' | 'plum'

export const PAL: Record<FruitName, [string, string, string]> = {
  lime: ['#d8ff6e', '#5ba20a', '#1c3a00'],
  coral: ['#ffb489', '#ff6a3d', '#7a2400'],
  lemon: ['#ffe873', '#f5b800', '#7a5a00'],
  berry: ['#ff8fb8', '#ff3d7f', '#5e0a2e'],
  teal: ['#7df5df', '#1bcfae', '#064a3e'],
  plum: ['#c6a8ff', '#9b6bff', '#2e0a5e'],
}

const FRUITS = Object.keys(PAL) as FruitName[]

/** Cover gradient string for a fruit palette. */
export function cover(fruit: FruitName): string {
  const [light, mid, deep] = PAL[fruit]
  return (
    'radial-gradient(60% 55% at 75% 82%, rgba(255,255,255,.16), transparent 58%),' +
    `radial-gradient(125% 125% at 18% 12%, ${light} 0%, ${mid} 44%, ${deep} 100%)`
  )
}

/** The "mid" color of a palette — used for the player-mode background glow. */
export function glow(fruit: FruitName): string {
  return PAL[fruit][1]
}

/**
 * Deterministically pick a fruit palette from a seed (track id or title) so a given
 * track always falls back to the same cover across renders and sessions.
 */
export function pickFruit(seed: string): FruitName {
  let h = 0
  for (let i = 0; i < seed.length; i++) {
    h = (h * 31 + seed.charCodeAt(i)) | 0
  }
  return FRUITS[Math.abs(h) % FRUITS.length]
}
