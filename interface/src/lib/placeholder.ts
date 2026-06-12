import { FruitName } from './covers'

// PLACEHOLDER DATA — there is no playlists backend yet. These render the Library
// "Your playlists" section so the surface matches the design. Replace with a real
// `GET /playlists` (or similar) when the API exists; the card shape below is the
// contract the UI expects.
export interface PlaylistStub {
  id: string
  name: string
  sub: string
  count: number
  fruit: FruitName
}

export const PLAYLISTS: PlaylistStub[] = [
  { id: 'pl-citrus-mornings', name: 'Citrus Mornings', sub: 'Bright & fast', count: 18, fruit: 'lime' },
  { id: 'pl-late-pulp', name: 'Late Pulp', sub: 'Wind-down mix', count: 24, fruit: 'plum' },
  { id: 'pl-aggregated-hits', name: 'Aggregated Hits', sub: 'All sources', count: 32, fruit: 'coral' },
  { id: 'pl-local-crate', name: 'Local Crate', sub: 'Your uploads', count: 9, fruit: 'teal' },
]
