import { useState } from 'react'
import { login as apiLogin, register as apiRegister } from '../utils/api'
import { useAuth } from '../context/AuthContext'
import { useUI } from '../context/UIContext'
import { useToast } from './ui/Toast'
import { LogoGlyph } from './Icons'

const fieldClass =
  'w-full rounded-xl border-[1.5px] border-line bg-[#fbfcf6] px-3.5 py-3 text-[16px] text-ink-text outline-none transition-colors focus:border-lime focus:bg-white'
const labelClass = 'mb-1.5 block text-[13px] font-bold text-muted'

/**
 * Sign in / sign up in a centered card over the blurred, dimmed app.
 * Opened on demand via UIContext (anonymous browsing is allowed).
 */
const AuthModal = () => {
  const { authOpen, authMode, closeAuth, setAuthMode } = useUI()
  const { login: authLogin, checkAuth } = useAuth()
  const { toast } = useToast()

  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)

  if (!authOpen) return null

  const isSignup = authMode === 'signup'

  const reset = () => {
    setName('')
    setEmail('')
    setPassword('')
  }

  const validate = (): string | null => {
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) return 'Enter a valid email address'
    if (password.length < 6) return 'Password must be at least 6 characters'
    return null
  }

  const doLogin = async () => {
    const res = await apiLogin({ email, password })
    authLogin(res.data.access_token, email)
    checkAuth()
    toast({ title: 'Welcome back', description: email, status: 'success' })
    reset()
    closeAuth()
  }

  const submit = async () => {
    const err = validate()
    if (err) {
      toast({ title: 'Check your details', description: err, status: 'error' })
      return
    }
    setLoading(true)
    try {
      if (isSignup) {
        await apiRegister({ email, password })
        await doLogin()
      } else {
        await doLogin()
      }
    } catch (e) {
      toast({
        title: isSignup ? 'Could not create account' : 'Login failed',
        description: e instanceof Error ? e.message : 'Please try again',
        status: 'error',
      })
    } finally {
      setLoading(false)
    }
  }

  const onKey = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') submit()
  }

  return (
    <div
      className="absolute inset-0 z-[80] flex animate-aux-pop items-center justify-center p-6"
      style={{
        background: 'rgba(12,14,8,.46)',
        backdropFilter: 'blur(15px)',
        WebkitBackdropFilter: 'blur(15px)',
      }}
      onMouseDown={(e) => {
        if (e.target === e.currentTarget) closeAuth()
      }}
    >
      <div className="w-[min(424px,94vw)] animate-aux-spin-in rounded-hero border-[1.5px] border-[#ece9dc] bg-[#fffdf9] p-[36px_34px] shadow-modal">
        <div className="mb-[26px] flex items-center gap-2.5">
          <div className="flex h-[38px] w-[38px] items-center justify-center rounded-xl bg-lime text-ink shadow-[0_6px_16px_rgba(182,240,60,.35)]">
            <LogoGlyph size={20} />
          </div>
          <span className="font-display text-[21px] font-extrabold tracking-[-.4px]">auxstream</span>
        </div>

        <div className="font-display text-[30px] font-extrabold leading-[1.05] tracking-[-1px]">
          {isSignup ? 'Create your account' : 'Welcome back'}
        </div>
        <div className="mb-6 mt-1.5 text-[15px] text-muted-2">
          {isSignup
            ? 'One account for every source — start aggregating in seconds.'
            : 'Sign in to pick up where the music left off.'}
        </div>

        {isSignup && (
          <div className="mb-3.5">
            <label className={labelClass}>Name</label>
            <input
              className={fieldClass}
              placeholder="Your name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              onKeyDown={onKey}
            />
          </div>
        )}

        <div className="mb-3.5">
          <label className={labelClass}>Email</label>
          <input
            type="email"
            className={fieldClass}
            placeholder="you@email.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            onKeyDown={onKey}
          />
        </div>

        <div className="mb-3.5">
          <div className="mb-1.5 flex items-baseline justify-between">
            <label className="text-[13px] font-bold text-muted">Password</label>
            {!isSignup && (
              <span className="cursor-pointer text-[13px] font-semibold text-[#5d7a14]">Forgot?</span>
            )}
          </div>
          <input
            type="password"
            className={fieldClass}
            placeholder="••••••••"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            onKeyDown={onKey}
          />
        </div>

        <button
          onClick={submit}
          disabled={loading}
          className="mt-2 w-full rounded-pill bg-lime py-[15px] text-[17px] font-extrabold text-ink shadow-lime transition-all hover:-translate-y-px hover:shadow-lime-hover disabled:opacity-60"
        >
          {loading ? 'Please wait…' : isSignup ? 'Create account' : 'Log in'}
        </button>

        <div className="my-[22px] flex items-center gap-3">
          <div className="h-px flex-1 bg-[#ece9dc]" />
          <span className="font-mono text-[11px] text-faint-2">OR</span>
          <div className="h-px flex-1 bg-[#ece9dc]" />
        </div>

        <div className="text-center text-[15px] text-muted-2">
          {isSignup ? 'Already have an account?' : 'New to AuxStream?'}{' '}
          <span
            onClick={() => setAuthMode(isSignup ? 'signin' : 'signup')}
            className="cursor-pointer border-b-2 border-lime font-extrabold text-ink-text"
          >
            {isSignup ? 'Log in' : 'Create one'}
          </span>
        </div>
      </div>
    </div>
  )
}

export default AuthModal
