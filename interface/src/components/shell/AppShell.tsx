import { ReactNode } from 'react'
import LeftRail from './LeftRail'
import NowPlayingBar from '../NowPlayingBar'
import PlayerOverlay from '../PlayerOverlay'
import AuthModal from '../AuthModal'

interface AppShellProps {
  children: ReactNode
}

/** Two-column app shell: fixed left rail + internally-scrolling main, with the
 *  persistent now-playing bar, the immersive Player overlay, and the auth modal
 *  layered over everything. */
const AppShell = ({ children }: AppShellProps) => {
  return (
    <div className="relative h-screen overflow-hidden">
      <div className="grid h-screen grid-cols-[86px_1fr]">
        <LeftRail />
        <main
          className="relative overflow-hidden"
          style={{
            background:
              'radial-gradient(120% 80% at 100% 0%, #eef3df 0%, #f7f8ef 46%)',
          }}
        >
          <div className="h-screen overflow-y-auto pb-[124px]">{children}</div>
          <NowPlayingBar />
        </main>
      </div>

      <PlayerOverlay />
      <AuthModal />
    </div>
  )
}

export default AppShell
