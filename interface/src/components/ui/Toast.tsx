/* eslint-disable react-refresh/only-export-components -- provider + hook co-located by design */
import {
  createContext,
  useCallback,
  useContext,
  useRef,
  useState,
  ReactNode,
} from 'react'
import { createPortal } from 'react-dom'

type ToastStatus = 'success' | 'error' | 'info'

interface ToastItem {
  id: number
  title: string
  description?: string
  status: ToastStatus
}

interface ToastInput {
  title: string
  description?: string
  status?: ToastStatus
  duration?: number
}

interface ToastContextValue {
  toast: (input: ToastInput) => void
}

const ToastContext = createContext<ToastContextValue | undefined>(undefined)

export const useToast = () => {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast must be used within a ToastProvider')
  return ctx
}

interface ToastStyle {
  kicker: string
  accent: string
  chipBg: string
  chipBorder: string
  icon: JSX.Element
}

const checkIcon = (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.6} strokeLinecap="round" strokeLinejoin="round">
    <path d="M20 6L9 17l-5-5" />
  </svg>
)
const alertIcon = (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.4} strokeLinecap="round" strokeLinejoin="round">
    <line x1="12" y1="8" x2="12" y2="13" />
    <line x1="12" y1="17" x2="12" y2="17" />
    <circle cx="12" cy="12" r="9" />
  </svg>
)
const infoIcon = (
  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2.4} strokeLinecap="round" strokeLinejoin="round">
    <line x1="12" y1="11" x2="12" y2="16" />
    <line x1="12" y1="8" x2="12" y2="8" />
    <circle cx="12" cy="12" r="9" />
  </svg>
)

const STYLES: Record<ToastStatus, ToastStyle> = {
  success: { kicker: 'Fresh', accent: '#5ba20a', chipBg: '#eafabf', chipBorder: '#cdeb8a', icon: checkIcon },
  error: { kicker: 'Hold up', accent: '#c43d3d', chipBg: '#fbeeee', chipBorder: '#f1d6d6', icon: alertIcon },
  info: { kicker: 'Heads up', accent: '#5d7a14', chipBg: '#f1f2e7', chipBorder: '#e4e6d6', icon: infoIcon },
}

export const ToastProvider = ({ children }: { children: ReactNode }) => {
  const [items, setItems] = useState<ToastItem[]>([])
  const idRef = useRef(0)

  const toast = useCallback((input: ToastInput) => {
    const id = ++idRef.current
    const item: ToastItem = {
      id,
      title: input.title,
      description: input.description,
      status: input.status ?? 'info',
    }
    setItems((prev) => [...prev, item])
    const duration = input.duration ?? 4000
    window.setTimeout(() => {
      setItems((prev) => prev.filter((t) => t.id !== id))
    }, duration)
  }, [])

  return (
    <ToastContext.Provider value={{ toast }}>
      {children}
      {createPortal(
        <div className="pointer-events-none fixed bottom-6 right-6 z-[200] flex w-[360px] max-w-[calc(100vw-2rem)] flex-col gap-3">
          {items.map((t) => {
            const s = STYLES[t.status]
            return (
              <div
                key={t.id}
                className="pointer-events-auto flex animate-aux-pop items-start gap-3.5 rounded-[18px] border-[1.5px] border-line bg-white p-3.5 shadow-menu"
              >
                <div
                  className="flex h-10 w-10 flex-none items-center justify-center rounded-[13px] border-[1.5px]"
                  style={{ background: s.chipBg, borderColor: s.chipBorder, color: s.accent }}
                >
                  {s.icon}
                </div>
                <div className="min-w-0 flex-1 pt-0.5">
                  <div
                    className="font-mono text-[10px] uppercase tracking-[2px]"
                    style={{ color: s.accent }}
                  >
                    {s.kicker}
                  </div>
                  <div className="mt-0.5 text-[15px] font-bold leading-tight text-ink-text">
                    {t.title}
                  </div>
                  {t.description && (
                    <div className="mt-1 text-[13px] leading-snug text-muted-2">{t.description}</div>
                  )}
                </div>
              </div>
            )
          })}
        </div>,
        document.body
      )}
    </ToastContext.Provider>
  )
}
