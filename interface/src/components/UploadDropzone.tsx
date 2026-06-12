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

/** Full-width drag-and-drop dropzone wired to the real bulk-upload endpoint. */
const UploadDropzone = ({ onUploaded }: UploadDropzoneProps) => {
  const { isAuthenticated, userName } = useAuth()
  const { openAuth } = useUI()
  const { toast } = useToast()
  const inputRef = useRef<HTMLInputElement>(null)
  const [dragOver, setDragOver] = useState(false)
  const [uploading, setUploading] = useState(false)
  const [progress, setProgress] = useState(0)

  const requireAuth = (): boolean => {
    if (!isAuthenticated) {
      toast({ title: 'Sign in to upload', description: 'Create an account to add your tracks.', status: 'info' })
      openAuth('signin')
      return false
    }
    return true
  }

  const handleFiles = async (fileList: FileList | null) => {
    setDragOver(false)
    if (!requireAuth()) return
    const files = Array.from(fileList || [])
    if (!files.length) return

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
      toast({
        title: 'Upload failed',
        description: e instanceof Error ? e.message : 'Please try again',
        status: 'error',
      })
    } finally {
      setUploading(false)
      setProgress(0)
      if (inputRef.current) inputRef.current.value = ''
    }
  }

  const title = uploading
    ? `Uploading… ${progress}%`
    : dragOver
      ? 'Drop to add to your library'
      : 'Upload your tracks'

  return (
    <label
      onDragOver={(e) => {
        e.preventDefault()
        if (!dragOver) setDragOver(true)
      }}
      onDragLeave={(e) => {
        e.preventDefault()
        setDragOver(false)
      }}
      onDrop={(e) => {
        e.preventDefault()
        handleFiles(e.dataTransfer.files)
      }}
      className={`flex cursor-pointer items-center gap-5 rounded-[20px] border-2 border-dashed p-[20px_22px] transition-colors ${
        dragOver ? 'border-lime bg-[#f1fbda]' : 'border-[#cdd3b6] bg-[#fcfdf6]'
      }`}
    >
      <input
        ref={inputRef}
        type="file"
        accept="audio/*"
        multiple
        className="hidden"
        onChange={(e) => handleFiles(e.target.files)}
      />
      <div className="flex h-[58px] w-[58px] flex-none items-center justify-center rounded-2xl border-[1.5px] border-[#cdeb8a] bg-lime-tint text-[#4f7a00]">
        <UploadIcon size={26} />
      </div>
      <div className="flex-1">
        <div className="font-display text-[21px] font-bold tracking-[-.4px]">{title}</div>
        <div className="mt-[3px] text-[15px] text-muted-2">
          Drag tracks here or browse — MP3, WAV, FLAC, M4A. They land in your unified library
          instantly.
        </div>
      </div>
      <div className="flex flex-none items-center gap-2.5 rounded-pill bg-ink px-[22px] py-3 text-[16px] font-extrabold text-lime">
        <PlusIcon size={18} />
        Browse files
      </div>
    </label>
  )
}

export default UploadDropzone
