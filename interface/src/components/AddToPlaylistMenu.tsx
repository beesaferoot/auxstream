import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { getPlaylists, addTrackToPlaylist } from '../utils/api'
import { Playlist } from '../interfaces/playlists'
import { useToast } from './ui/Toast'
import { PlusIcon } from './Icons'
import PlaylistFormModal from './PlaylistFormModal'

interface AddToPlaylistMenuProps {
  trackId: string
  onClose: () => void
  /** Positioning classes for the popover panel (e.g. "right-0 bottom-full mb-2"). */
  className?: string
}

/** Popover that adds a track to one of the user's playlists, or a brand-new one.
 *  Assumes the user is authenticated (callers gate on auth before opening it). */
const AddToPlaylistMenu = ({ trackId, onClose, className = '' }: AddToPlaylistMenuProps) => {
  const qc = useQueryClient()
  const { toast } = useToast()
  const [showForm, setShowForm] = useState(false)

  const { data: playlists = [], isLoading } = useQuery({
    queryKey: ['library', 'playlists'],
    queryFn: getPlaylists,
  })

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

  return (
    <>
      {/* click-away */}
      <div className="fixed inset-0 z-[70]" onClick={onClose} />
      <div
        className={`absolute z-[71] w-[240px] animate-aux-pop overflow-hidden rounded-[16px] border-[1.5px] border-[#e7e9da] bg-white p-1.5 shadow-menu ${className}`}
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
    </>
  )
}

export default AddToPlaylistMenu
