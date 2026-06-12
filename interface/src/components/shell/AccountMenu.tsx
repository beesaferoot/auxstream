import { useAuth } from '../../context/AuthContext'
import { useToast } from '../ui/Toast'
import { LogoutIcon } from '../Icons'

interface AccountMenuProps {
  onClose: () => void
}

/** Popover anchored to the rail avatar: identity + log out. */
const AccountMenu = ({ onClose }: AccountMenuProps) => {
  const { userName, userEmail, userInitial, logout } = useAuth()
  const { toast } = useToast()

  const handleLogout = async () => {
    await logout()
    onClose()
    toast({ title: 'Logged out', status: 'info' })
  }

  return (
    <>
      {/* click-away */}
      <div className="fixed inset-0 z-[60]" onClick={onClose} />
      <div className="absolute bottom-0 left-[54px] z-[61] w-[248px] animate-aux-pop rounded-[18px] border-[1.5px] border-[#e7e9da] bg-white p-2 shadow-menu">
        <div className="flex items-center gap-3 px-2.5 pb-3 pt-2.5">
          <div
            className="flex h-[42px] w-[42px] flex-none items-center justify-center rounded-full text-[16px] font-extrabold text-white"
            style={{ background: 'linear-gradient(135deg,#ff8a3d,#ff3d7f)' }}
          >
            {userInitial}
          </div>
          <div className="min-w-0">
            <div className="truncate text-[15px] font-bold text-ink-text">{userName}</div>
            <div className="truncate text-[13px] text-muted-3">{userEmail || ''}</div>
          </div>
        </div>

        <div className="mx-1 my-1.5 h-px bg-line-sep" />

        <button
          onClick={handleLogout}
          className="flex w-full items-center gap-[11px] rounded-[11px] px-2.5 py-2.5 text-left text-[15px] font-bold text-danger transition-colors hover:bg-[#fbeeee]"
        >
          <LogoutIcon size={18} />
          Log out
        </button>
      </div>
    </>
  )
}

export default AccountMenu
