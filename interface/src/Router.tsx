import { BrowserRouter, Route, Routes } from 'react-router-dom'
import HomePage from './pages/HomePage.tsx'
import SearchPage from './pages/SearchPage.tsx'
import MusicPlayerPage from './pages/MusicPlayerPage.tsx'
import ProfilePage from './pages/ProfilePage.tsx'
import SettingsPage from './pages/SettingsPage.tsx'
import PageNotFound from './components/PageNotFound.tsx'
import Layout from './components/Layout.tsx'

const Router = () => {
  return (
    <BrowserRouter future={{ v7_startTransition: true }}>
      <Layout>
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/trending" element={<SearchPage />} />
          <Route path="/player" element={<MusicPlayerPage />} />
          <Route path="/profile" element={<ProfilePage />} />
          <Route path="/settings" element={<SettingsPage />} />
          <Route path="*" element={<PageNotFound />} />
        </Routes>
      </Layout>
    </BrowserRouter>
  )
}

export default Router
