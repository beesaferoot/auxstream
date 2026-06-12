import { useEffect, useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { createPlaylist, updatePlaylist } from '../utils/api'
import { Playlist } from '../interfaces/playlists'
import { useToast } from './ui/Toast'

interface PlaylistFormModalProps {
  open: boolean
  onClose: () => void
  /** When provided, the modal edits this playlist instead of creating one. */
  playlist?: Playlist
  /** Called with the saved playlist after a successful create/update. */
  onSaved?: (playlist: Playlist) => void
}

const fieldClass =
  'w-full rounded-xl border-[1.5px] border-line bg-[#fbfcf6] px-3.5 py-3 text-[16px] text-ink-text outline-none transition-colors focus:border-lime focus:bg-white'

const PlaylistFormModal = ({ open, onClose, playlist, onSaved }: PlaylistFormModalProps) => {
  const qc = useQueryClient()
  const { toast } = useToast()
  const editing = !!playlist

  const [name, setName] = useState(playlist?.name ?? '')
  const [description, setDescription] = useState(playlist?.description ?? '')
  const [isPublic, setIsPublic] = useState(playlist?.is_public ?? false)
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && !saving) onClose()
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [open, saving, onClose])

  if (!open) return null

  const submit = async () => {
    if (!name.trim()) {
      toast({ title: 'Name your playlist', status: 'error' })
      return
    }
    setSaving(true)
    try {
      const input = { name: name.trim(), description: description.trim(), is_public: isPublic }
      const saved =
        editing && playlist
          ? await updatePlaylist(playlist.id, input)
          : await createPlaylist(input)
      qc.invalidateQueries({ queryKey: ['library', 'playlists'] })
      qc.invalidateQueries({ queryKey: ['playlist', saved.id] })
      toast({ title: editing ? 'Playlist updated' : 'Playlist created', status: 'success' })
      onSaved?.(saved)
      onClose()
    } catch (e) {
      toast({
        title: editing ? 'Could not update playlist' : 'Could not create playlist',
        description: e instanceof Error ? e.message : 'Please try again',
        status: 'error',
      })
    } finally {
      setSaving(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-[90] flex animate-aux-pop items-center justify-center p-6"
      style={{ background: 'rgba(12,14,8,.46)', backdropFilter: 'blur(15px)', WebkitBackdropFilter: 'blur(15px)' }}
      onClick={(e) => e.stopPropagation()}
      onMouseDown={(e) => {
        if (e.target === e.currentTarget && !saving) onClose()
      }}
    >
      <div className="w-[min(460px,94vw)] animate-aux-spin-in rounded-hero border-[1.5px] border-[#ece9dc] bg-[#fffdf9] p-[34px] shadow-modal">
        <div className="font-display text-[28px] font-extrabold tracking-[-.8px]">
          {editing ? 'Edit playlist' : 'New playlist'}
        </div>
        <div className="mb-6 mt-1 text-[15px] text-muted-2">
          {editing ? 'Update the details for this playlist.' : 'Give your playlist a name to get started.'}
        </div>

        <label className="mb-1.5 block text-[13px] font-bold text-muted">Name</label>
        <input
          className={`${fieldClass} mb-3.5`}
          placeholder="e.g. Citrus Mornings"
          value={name}
          onChange={(e) => setName(e.target.value)}
          autoFocus
        />

        <label className="mb-1.5 block text-[13px] font-bold text-muted">Description</label>
        <textarea
          className={`${fieldClass} mb-3.5 resize-none`}
          placeholder="Optional"
          rows={2}
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />

        <button
          type="button"
          onClick={() => setIsPublic((v) => !v)}
          className="mb-6 flex w-full items-center justify-between rounded-xl border-[1.5px] border-line bg-[#fbfcf6] px-3.5 py-3 text-left"
        >
          <div>
            <div className="text-[15px] font-bold text-ink-text">Public playlist</div>
            <div className="text-[13px] text-muted-2">Anyone with the link can view it.</div>
          </div>
          <span
            className={`relative h-6 w-11 flex-none rounded-full transition-colors ${isPublic ? 'bg-lime' : 'bg-line'}`}
          >
            <span
              className={`absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-all ${isPublic ? 'left-[22px]' : 'left-0.5'}`}
            />
          </span>
        </button>

        <div className="flex justify-end gap-3">
          <button
            onClick={onClose}
            disabled={saving}
            className="rounded-pill border-[1.5px] border-line px-5 py-3 text-[15px] font-bold text-muted transition-colors hover:border-ink disabled:opacity-60"
          >
            Cancel
          </button>
          <button
            onClick={submit}
            disabled={saving}
            className="rounded-pill bg-lime px-6 py-3 text-[15px] font-extrabold text-ink shadow-lime transition-all hover:-translate-y-px hover:shadow-lime-hover disabled:opacity-60"
          >
            {saving ? 'Saving…' : editing ? 'Save changes' : 'Create playlist'}
          </button>
        </div>
      </div>
    </div>
  )
}

export default PlaylistFormModal
