// Deterministic 56-bar waveform heights for the Player overlay scrubber.
// Ported from the design prototype's procedural formula so the silhouette is stable.

export const WAVE_BARS = 56

export interface WaveBar {
  /** Height percentage 0–100. */
  h: number
  /** Animation delay string, e.g. "0.18s". */
  delay: string
}

/** The static bar silhouette (heights + per-bar animation delays). */
export const WAVEFORM: WaveBar[] = Array.from({ length: WAVE_BARS }, (_, i) => {
  const raw =
    22 +
    Math.round(
      Math.abs(Math.sin(i * 0.55) * Math.cos(i * 0.21) + Math.sin(i * 0.13)) * 60
    )
  return {
    h: Math.min(100, raw),
    delay: `${(i % 7) * 0.09}s`,
  }
})

/** Whether bar `i` is in the "played" region given a 0–1 progress fraction. */
export function isPlayed(i: number, progress: number): boolean {
  return i / WAVE_BARS <= progress
}
