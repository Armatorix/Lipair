import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import ProtectedRoute from './components/ProtectedRoute'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import DashboardPage from './pages/DashboardPage'
import TournamentListPage from './pages/TournamentListPage'
import TournamentCreatePage from './pages/TournamentCreatePage'
import TournamentDetailPage from './pages/TournamentDetailPage'
import OAuthCallbackPage from './pages/OAuthCallbackPage'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Layout><TournamentListPage /></Layout>} />
        <Route path="/login" element={<Layout><LoginPage /></Layout>} />
        <Route path="/register" element={<Layout><RegisterPage /></Layout>} />
        <Route path="/auth/callback" element={<OAuthCallbackPage />} />
        <Route path="/dashboard" element={<Layout><ProtectedRoute><DashboardPage /></ProtectedRoute></Layout>} />
        <Route path="/tournaments/new" element={<Layout><ProtectedRoute><TournamentCreatePage /></ProtectedRoute></Layout>} />
        <Route path="/tournaments/:id" element={<Layout><TournamentDetailPage /></Layout>} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
