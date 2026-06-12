import Router from './Router.tsx'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from './context/AuthContext'
import { PlayerProvider } from './context/PlayerContext'
import { UIProvider } from './context/UIContext'
import { ToastProvider } from './components/ui/Toast'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
    },
  },
})

function App() {
  return (
    <AuthProvider>
      <QueryClientProvider client={queryClient}>
        <ToastProvider>
          <PlayerProvider>
            <UIProvider>
              <Router />
            </UIProvider>
          </PlayerProvider>
        </ToastProvider>
      </QueryClientProvider>
    </AuthProvider>
  )
}

export default App
