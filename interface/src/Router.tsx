import { BrowserRouter, Route, Routes } from 'react-router-dom'
import AppShell from './components/shell/AppShell'
import FeedView from './pages/FeedView'
import SearchView from './pages/SearchView'
import LibraryView from './pages/LibraryView'
import PlaylistDetailView from './pages/PlaylistDetailView'
import NotFound from './components/NotFound'

const Router = () => {
  return (
    <BrowserRouter future={{ v7_startTransition: true }}>
      <AppShell>
        <Routes>
          <Route path="/" element={<FeedView />} />
          <Route path="/search" element={<SearchView />} />
          <Route path="/library" element={<LibraryView />} />
          <Route path="/library/playlists/:id" element={<PlaylistDetailView />} />
          <Route path="*" element={<NotFound />} />
        </Routes>
      </AppShell>
    </BrowserRouter>
  )
}

export default Router
