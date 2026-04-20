import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { tournamentApi, Tournament } from '../api/client'
import { useAuth } from '../hooks/useAuth'

export default function DashboardPage() {
  const { user } = useAuth()
  const [tournaments, setTournaments] = useState<Tournament[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    tournamentApi.list().then((res) => {
      const myTournaments = res.data.filter((t) => t.organizer_id === user?.id)
      setTournaments(myTournaments)
    }).catch(console.error).finally(() => setLoading(false))
  }, [user?.id])

  const statusColors: Record<string, string> = {
    draft: 'bg-gray-100 text-gray-800',
    registration: 'bg-blue-100 text-blue-800',
    active: 'bg-green-100 text-green-800',
    completed: 'bg-purple-100 text-purple-800',
  }

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">My Dashboard</h1>
          <p className="text-gray-500 mt-1">Welcome back, {user?.name}!</p>
        </div>
        <Link
          to="/tournaments/new"
          className="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-md text-sm font-medium transition-colors"
        >
          + New Tournament
        </Link>
      </div>

      <div className="bg-white rounded-lg shadow">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">My Tournaments</h2>
        </div>
        {loading ? (
          <div className="p-6 text-center text-gray-500">Loading...</div>
        ) : tournaments.length === 0 ? (
          <div className="p-6 text-center">
            <p className="text-gray-500 mb-4">You haven't created any tournaments yet.</p>
            <Link
              to="/tournaments/new"
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
            >
              Create your first tournament
            </Link>
          </div>
        ) : (
          <ul className="divide-y divide-gray-200">
            {tournaments.map((t) => (
              <li key={t.id} className="px-6 py-4 hover:bg-gray-50 transition-colors">
                <Link to={`/tournaments/${t.id}`} className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-900">{t.name}</p>
                    <p className="text-sm text-gray-500 capitalize">{t.pairing_system.replace('_', ' ')}</p>
                  </div>
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${statusColors[t.status] || 'bg-gray-100 text-gray-800'}`}>
                    {t.status}
                  </span>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  )
}
