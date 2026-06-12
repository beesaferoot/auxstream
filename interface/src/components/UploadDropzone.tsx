import { useRef, useState } from 'react'
import { createArtist, uploadTracksBulk } from '../utils/api'
import { useAuth } from '../context/AuthContext'
import { useUI } from '../context/UIContext'
import { useToast } from './ui/Toast'
import { UploadIcon, PlusIcon } from './Icons'

const ACCEPTED = ['.mp3', '.wav', '.ogg', '.m4a', '.flac']
const MAX_SIZE = 100 * 1024 * 1024 // 100MB

interface UploadDropzoneProps {
  /** Called with the saved track ids after a successful upload. */
  onUploaded: (savedIds: string[]) => void
}

/** Full-width drag-and-drop dropzone wired to the real bulk-upload endpoint, with
 *  live progress and an inline failure state. */
const UploadDropzone = ({ onUploaded }: UploadDropzoneProps) => {
  const { isAuthenticated, userName } = useAuth()
  const { openAuth } = useUI()
  const { toast } = useToast()
  const inputRef = useRef<HTMLInputElement>(null)
  const [dragOver, setDragOver] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [progress, setProgress] = useState(0)
  const [count, setCount] = useState(0)
  const [error, setError] = useState<string | null>(null)

  const handleFiles = async (fileList: FileList | null) => {
    setDragOver(false)
    if (!isAuthenticated) {
      openAuth('signin')
      return
    }
    const files = Array.from(fileList || [])
    if (!files.length) return

    setError(null)
    const valid = files.filter((f) => {
      const ext = '.' + (f.name.split('.').pop()?.toLowerCase() ?? '')
      if (!ACCEPTED.includes(ext)) {
        toast({ title: `Skipped ${f.name}`, description: `Unsupported format ${ext}`, status: 'error' })
        return false
      }
      if (f.size > MAX_SIZE) {
        toast({ title: `Skipped ${f.name}`, description: 'File too large (max 100MB)', status: 'error' })
        return false
      }
      return true
    })
    if (!valid.length) return

    setUploading(true)
    setCount(valid.length)
    setProgress(0)
    try {
      const artist = await createArtist(userName || 'You')
      const titles = valid.map((f) => f.name.replace(/\.[^/.]+$/, ''))
      const res = await uploadTracksBulk(valid, titles, artist.id, (p) => setProgress(Math.round(p)))
      const saved = res.data?.saved ?? []
      toast({
        title: 'Upload complete',
        description: `Added ${res.data?.rows ?? saved.length} track(s) to your library`,
        status: 'success',
      })
      onUploaded(saved)
    } catch (e) {
      const msg = e instanceof Error ? e.message : 'Something went wrong. Please try again.'
      setError(msg)
      toast({ title: 'Upload failed', description: msg, status: 'error' })
    } finally {
      setUploading(false)
      setProgress(0)
      if (inputRef.current) inputRef.current.value = ''
    }
  }

  const title = uploading
    ? `Uploading ${count} track${count === 1 ? '' : 's'}…`
    : dragOver
      ? 'Drop to add to your library'
      : 'Upload your tracks'

  return (
    <div>
      <label
        onDragOver={(e) => {
          e.preventDefault()
          if (!uploading && !dragOver) setDragOver(true)
        }}
        onDragLeave={(e) => {
          e.preventDefault()
          setDragOver(false)
        }}
        onDrop={(e) => {
          e.preventDefault()
          if (!uploading) handleFiles(e.dataTransfer.files)
        }}
        onClick={(e) => {
          // Don't open the OS picker mid-upload, or when a sign-in is required first.
          if (uploading) {
            e.preventDefault()
          } else if (!isAuthenticated) {
            e.preventDefault()
            openAuth('signin')
          }
        }}
        className={`flex items-center gap-5 rounded-[20px] border-2 border-dashed p-[20px_22px] transition-colors ${
          uploading ? 'cursor-default border-[#cdeb8a] bg-[#f6fbe8]' : 'cursor-pointer'
        } ${dragOver ? 'border-lime bg-[#f1fbda]' : !uploading ? 'border-[#cdd3b6] bg-[#fcfdf6]' : ''}`}
      >
        <input
          ref={inputRef}
          type="file"
          accept="audio/*"
          multiple
          disabled={uploading}
          className="hidden"
          onChange={(e) => handleFiles(e.target.files)}
        />
        <div className="flex h-[58px] w-[58px] flex-none items-center justify-center rounded-2xl border-[1.5px] border-[#cdeb8a] bg-lime-tint text-[#4f7a00]">
          {uploading ? (
            <span className="h-6 w-6 animate-spin rounded-full border-[3px] border-[#cdeb8a] border-t-[#4f7a00]" />
          ) : (
            <UploadIcon size={26} />
          )}
        </div>

        <div className="min-w-0 flex-1">
          <div className="font-display text-[21px] font-bold tracking-[-.4px]">{title}</div>
          {uploading ? (
            <div className="mt-2.5">
              <div className="h-2 w-full overflow-hidden rounded-full bg-[#e3ebc8]">
                <div
                  className="h-full rounded-full bg-lime transition-[width] duration-200 ease-out"
                  style={{ width: `${progress}%` }}
                />
              </div>
              <div className="mt-1.5 font-mono text-[12px] text-muted-2">
                {progress < 100 ? `${progress}%` : 'Finishing up…'}
              </div>
            </div>
          ) : (
            <div className="mt-[3px] text-[15px] text-muted-2">
              Drag tracks here or browse — MP3, WAV, FLAC, M4A. They land in your unified library
              instantly.
            </div>
          )}
        </div>

        <div
          className={`flex flex-none items-center gap-2.5 rounded-pill bg-ink px-[22px] py-3 text-[16px] font-extrabold text-lime ${
            uploading ? 'opacity-60' : ''
          }`}
        >
          {uploading ? (
            <>
              <span className="h-4 w-4 animate-spin rounded-full border-2 border-lime/40 border-t-lime" />
              Uploading…
            </>
          ) : (
            <>
              <PlusIcon size={18} />
              Browse files
            </>
          )}
        </div>
      </label>

      {error && (
        <div className="mt-3 flex items-start gap-3 rounded-2xl border-[1.5px] border-[#f1d6d6] bg-[#fbeeee] p-3.5">
          <div className="flex h-9 w-9 flex-none items-center justify-center rounded-xl border-[1.5px] border-[#f1d6d6] bg-white text-danger">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.4} strokeLinecap="round" strokeLinejoin="round">
              <circle cx="12" cy="12" r="9" />
              <line x1="12" y1="8" x2="12" y2="13" />
              <line x1="12" y1="17" x2="12" y2="17" />
            </svg>
          </div>
          <div className="min-w-0 flex-1 pt-0.5">
            <div className="text-[14px] font-bold text-danger">Upload failed</div>
            <div className="mt-0.5 break-words text-[13px] text-muted-2">{error}</div>
          </div>
          <button
            onClick={() => setError(null)}
            className="flex-none rounded-full px-3 py-1 text-[13px] font-bold text-muted transition-colors hover:bg-white"
          >
            Dismiss
          </button>
        </div>
      )}
    </div>
  )
}

export default UploadDropzone
