import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { tournamentApi, Tournament } from '../api/client'

const statusColors: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-800',
  registration: 'bg-blue-100 text-blue-800',
  active: 'bg-green-100 text-green-800',
  completed: 'bg-purple-100 text-purple-800',
}

const systemLabels: Record<string, string> = {
  round_robin: 'Round Robin',
  knockout: 'Knockout',
  swiss: 'Swiss',
  scheveningen: 'Scheveningen',
  manual: 'Manual',
}

export default function TournamentListPage() {
  const [tournaments, setTournaments] = useState<Tournament[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    tournamentApi.list()
      .then((res) => setTournaments(res.data))
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [])

  return (
    <div>
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Tournaments</h1>
        <Link
          to="/tournaments/new"
          className="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-md text-sm font-medium transition-colors"
        >
          + New Tournament
        </Link>
      </div>

      {loading ? (
        <div className="text-center py-12 text-gray-500">Loading tournaments...</div>
      ) : tournaments.length === 0 ? (
        <div className="text-center py-12">
          <p className="text-gray-500 text-lg mb-4">No tournaments yet.</p>
          <Link
            to="/tournaments/new"
            className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
          >
            Create first tournament
          </Link>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {tournaments.map((t) => (
            <Link
              key={t.id}
              to={`/tournaments/${t.id}`}
              className="bg-white rounded-lg shadow hover:shadow-md transition-shadow p-6 block"
            >
              <div className="flex justify-between items-start mb-3">
                <h2 className="text-lg font-semibold text-gray-900 truncate flex-1 mr-2">{t.name}</h2>
                <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium whitespace-nowrap ${statusColors[t.status] || 'bg-gray-100 text-gray-800'}`}>
                  {t.status}
                </span>
              </div>
              {t.description && (
                <p className="text-sm text-gray-500 mb-3 line-clamp-2">{t.description}</p>
              )}
              <div className="flex items-center justify-between">
                <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-indigo-50 text-indigo-700">
                  {systemLabels[t.pairing_system] || t.pairing_system}
                </span>
                {t.start_date && (
                  <span className="text-xs text-gray-400">
                    {new Date(t.start_date).toLocaleDateString()}
                  </span>
                )}
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
