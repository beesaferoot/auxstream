import { useEffect, useState } from 'react'
import { createPortal } from 'react-dom'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { getPlaylists, addTrackToPlaylist } from '../utils/api'
import { Playlist } from '../interfaces/playlists'
import { useToast } from './ui/Toast'
import { PlusIcon } from './Icons'
import PlaylistFormModal from './PlaylistFormModal'

const PANEL_W = 240

interface AddToPlaylistMenuProps {
  trackId: string
  onClose: () => void
  /** Bounding rect of the trigger; the panel is positioned against it. */
  anchor: DOMRect | null
}

/** Add-to-playlist popover. Rendered in a portal with fixed positioning so it's never
 *  clipped by an `overflow-hidden` list or trapped in a stacking context. */
const AddToPlaylistMenu = ({ trackId, onClose, anchor }: AddToPlaylistMenuProps) => {
  const qc = useQueryClient()
  const { toast } = useToast()
  const [showForm, setShowForm] = useState(false)

  const { data: playlists = [], isLoading } = useQuery({
    queryKey: ['library', 'playlists'],
    queryFn: getPlaylists,
  })

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !showForm) onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [onClose, showForm])

  const add = async (playlistId: string, name: string) => {
    try {
      await addTrackToPlaylist(playlistId, trackId)
      qc.invalidateQueries({ queryKey: ['playlist', playlistId] })
      qc.invalidateQueries({ queryKey: ['library', 'playlists'] })
      toast({ title: `Added to ${name}`, status: 'success' })
    } catch (e) {
      toast({
        title: 'Could not add to playlist',
        description: e instanceof Error ? e.message : 'Please try again',
        status: 'error',
      })
    }
    onClose()
  }

  if (!anchor) return null

  const margin = 8
  const left = Math.min(
    Math.max(margin, anchor.right - PANEL_W),
    window.innerWidth - PANEL_W - margin
  )
  // Flip above the trigger when there isn't room below (e.g. the now-playing bar).
  const openAbove = anchor.bottom > window.innerHeight - 340
  const pos = openAbove
    ? { left, bottom: window.innerHeight - anchor.top + margin }
    : { left, top: anchor.bottom + margin }

  return createPortal(
    <>
      <div
        className="fixed inset-0 z-[70]"
        onClick={(e) => {
          e.stopPropagation()
          onClose()
        }}
      />
      <div
        style={pos}
        className="fixed z-[71] w-[240px] animate-aux-pop overflow-hidden rounded-[16px] border-[1.5px] border-[#e7e9da] bg-white p-1.5 shadow-menu"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="px-2.5 pb-1.5 pt-1 font-mono text-[10px] uppercase tracking-[2px] text-faint">
          Add to playlist
        </div>

        <button
          onClick={() => setShowForm(true)}
          className="mb-1 flex w-full items-center gap-2.5 rounded-[11px] px-2.5 py-2.5 text-left text-[14px] font-bold text-ink-text transition-colors hover:bg-[#f5f7ec]"
        >
          <span className="flex h-6 w-6 items-center justify-center rounded-full bg-lime text-ink">
            <PlusIcon size={14} />
          </span>
          New playlist
        </button>

        <div className="max-h-[240px] overflow-y-auto">
          {isLoading ? (
            <div className="px-2.5 py-3 text-[13px] text-faint">Loading…</div>
          ) : playlists.length === 0 ? (
            <div className="px-2.5 py-3 text-[13px] text-faint">No playlists yet.</div>
          ) : (
            playlists.map((p: Playlist) => (
              <button
                key={p.id}
                onClick={() => add(p.id, p.name)}
                className="flex w-full items-center justify-between gap-2 rounded-[11px] px-2.5 py-2.5 text-left text-[14px] text-[#2d3022] transition-colors hover:bg-[#f5f7ec]"
              >
                <span className="truncate">{p.name}</span>
                <span className="flex-none font-mono text-[10px] text-faint">{p.track_count}</span>
              </button>
            ))
          )}
        </div>
      </div>

      <PlaylistFormModal
        open={showForm}
        onClose={() => setShowForm(false)}
        onSaved={(p) => add(p.id, p.name)}
      />
    </>,
    document.body
  )
}

export default AddToPlaylistMenu
