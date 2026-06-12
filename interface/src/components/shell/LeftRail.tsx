import { useState } from 'react'
import { NavLink, useLocation } from 'react-router-dom'
import { useAuth } from '../../context/AuthContext'
import { useUI } from '../../context/UIContext'
import { FeedIcon, SearchIcon, LibraryIcon, LogoGlyph } from '../Icons'
import AccountMenu from './AccountMenu'

const NAV = [
  { to: '/', label: 'Feed', Icon: FeedIcon, end: true },
  { to: '/search', label: 'Search', Icon: SearchIcon, end: false },
  { to: '/library', label: 'Library', Icon: LibraryIcon, end: false },
]

const LeftRail = () => {
  const location = useLocation()
  const { isAuthenticated, userInitial } = useAuth()
  const { openAuth } = useUI()
  const [menuOpen, setMenuOpen] = useState(false)

  const isActive = (to: string, end: boolean) =>
    end ? location.pathname === to : location.pathname.startsWith(to)

  return (
    <div className="relative z-[50] flex flex-col items-center gap-2 bg-ink py-[22px]">
      {/* Logo tile */}
      <div className="mb-[18px] flex h-[42px] w-[42px] items-center justify-center rounded-tile bg-lime text-ink shadow-logo">
        <LogoGlyph size={22} />
      </div>

      {NAV.map(({ to, label, Icon, end }) => {
        const active = isActive(to, end)
        return (
          <NavLink
            key={to}
            to={to}
            className={`flex w-[62px] flex-col items-center gap-1.5 rounded-nav py-[11px] transition-colors ${
              active ? 'bg-lime text-ink' : 'bg-transparent text-[#7d8268] hover:text-[#cdd3b6]'
            }`}
          >
            <Icon size={22} />
            <span className="text-[10px] font-bold tracking-[.3px]">{label}</span>
          </NavLink>
        )
      })}

      <div className="flex-1" />

      {/* Avatar / account */}
      <div className="relative">
        <button
          onClick={() => (isAuthenticated ? setMenuOpen((o) => !o) : openAuth('signin'))}
          title={isAuthenticated ? 'Account' : 'Sign in'}
          className="flex h-10 w-10 items-center justify-center rounded-full border-2 text-[15px] font-extrabold text-white transition-transform hover:scale-105"
          style={{
            background: 'linear-gradient(135deg,#ff8a3d,#ff3d7f)',
            borderColor: menuOpen ? '#b6f03c' : '#2a2e20',
          }}
        >
          {isAuthenticated ? userInitial : '+'}
        </button>
        {menuOpen && isAuthenticated && <AccountMenu onClose={() => setMenuOpen(false)} />}
      </div>
    </div>
  )
}

export default LeftRail
