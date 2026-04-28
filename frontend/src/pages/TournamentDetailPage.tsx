import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { tournamentApi, Tournament, TournamentPlayer, Round, Pairing, SPORT_OPTIONS } from '../api/client'

const sportMap = Object.fromEntries(SPORT_OPTIONS.map((s) => [s.value, s]))
import { useAuth } from '../hooks/useAuth'

const resultLabels: Record<string, string> = {
  white_wins: 'White wins (1-0)',
  black_wins: 'Black wins (0-1)',
  draw: 'Draw (½-½)',
  bye: 'Bye (+)',
  pending: 'Pending',
}

const statusColors: Record<string, string> = {
  draft: 'bg-gray-100 text-gray-800',
  registration: 'bg-blue-100 text-blue-800',
  active: 'bg-green-100 text-green-800',
  completed: 'bg-purple-100 text-purple-800',
}

export default function TournamentDetailPage() {
  const { id } = useParams<{ id: string }>()
  const { user, isAuthenticated } = useAuth()
  const navigate = useNavigate()

  const [tournament, setTournament] = useState<Tournament | null>(null)
  const [players, setPlayers] = useState<TournamentPlayer[]>([])
  const [rounds, setRounds] = useState<Round[]>([])
  const [selectedRound, setSelectedRound] = useState<Round | null>(null)
  const [pairings, setPairings] = useState<Pairing[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [actionError, setActionError] = useState('')

  // Add player form
  const [addPlayerName, setAddPlayerName] = useState('')
  const [addPlayerRating, setAddPlayerRating] = useState('')
  const [addingPlayer, setAddingPlayer] = useState(false)

  useEffect(() => {
    if (!id) return
    Promise.all([
      tournamentApi.get(id),
      tournamentApi.listPlayers(id),
      tournamentApi.listRounds(id),
    ]).then(([tRes, pRes, rRes]) => {
      setTournament(tRes.data)
      setPlayers(pRes.data)
      setRounds(rRes.data)
      if (rRes.data.length > 0) {
        const last = rRes.data[rRes.data.length - 1]
        setSelectedRound(last)
        tournamentApi.listPairings(id, last.id).then((pr) => setPairings(pr.data))
      }
    }).catch(() => setError('Failed to load tournament'))
      .finally(() => setLoading(false))
  }, [id])

  const handleRoundSelect = async (round: Round) => {
    setSelectedRound(round)
    if (!id) return
    const pr = await tournamentApi.listPairings(id, round.id)
    setPairings(pr.data)
  }

  const handleStart = async () => {
    if (!id) return
    setActionError('')
    try {
      const res = await tournamentApi.start(id)
      setTournament(res.data)
      const rRes = await tournamentApi.listRounds(id)
      setRounds(rRes.data)
      if (rRes.data.length > 0) {
        const last = rRes.data[rRes.data.length - 1]
        setSelectedRound(last)
        const pr = await tournamentApi.listPairings(id, last.id)
        setPairings(pr.data)
      }
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } }
      setActionError(error.response?.data?.message || 'Failed to start tournament')
    }
  }

  const handleDelete = async () => {
    if (!id || !window.confirm('Delete this tournament?')) return
    try {
      await tournamentApi.delete(id)
      navigate('/')
    } catch {
      setActionError('Failed to delete tournament')
    }
  }

  const handleCreateRound = async () => {
    if (!id) return
    setActionError('')
    try {
      const res = await tournamentApi.createRound(id)
      const newRound = res.data
      setRounds((prev) => [...prev, newRound])
      setSelectedRound(newRound)
      const pr = await tournamentApi.listPairings(id, newRound.id)
      setPairings(pr.data)
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } }
      setActionError(error.response?.data?.message || 'Failed to create round')
    }
  }

  const handleAddPlayer = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!id || !addPlayerName.trim()) return
    setAddingPlayer(true)
    try {
      const res = await tournamentApi.addPlayer(id, {
        name: addPlayerName,
        rating: addPlayerRating ? parseInt(addPlayerRating) : undefined,
      })
      setPlayers((prev) => [...prev, res.data])
      setAddPlayerName('')
      setAddPlayerRating('')
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } }
      setActionError(error.response?.data?.message || 'Failed to add player')
    } finally {
      setAddingPlayer(false)
    }
  }

  const handleRemovePlayer = async (playerId: string) => {
    if (!id || !window.confirm('Remove this player?')) return
    try {
      await tournamentApi.removePlayer(id, playerId)
      setPlayers((prev) => prev.filter((p) => p.id !== playerId))
    } catch {
      setActionError('Failed to remove player')
    }
  }

  const handleUpdatePairing = async (pairingId: string, result: string) => {
    if (!id || !selectedRound) return
    try {
      const res = await tournamentApi.updatePairing(id, selectedRound.id, pairingId, { result })
      setPairings((prev) => prev.map((p) => (p.id === pairingId ? res.data : p)))
      // Refresh players for updated scores
      const pRes = await tournamentApi.listPlayers(id)
      setPlayers(pRes.data)
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } }
      setActionError(error.response?.data?.message || 'Failed to update pairing')
    }
  }

  const isOrganizer = isAuthenticated() && user?.id === tournament?.organizer_id

  if (loading) {
    return <div className="text-center py-12 text-gray-500">Loading tournament...</div>
  }
  if (error || !tournament) {
    return <div className="text-center py-12 text-red-500">{error || 'Tournament not found'}</div>
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-white rounded-lg shadow p-6">
        <div className="flex justify-between items-start">
          <div>
            <div className="flex items-center gap-3 mb-2">
              <span className="text-2xl" title={sportMap[tournament.sport_type]?.label}>
                {sportMap[tournament.sport_type]?.icon ?? '🏆'}
              </span>
              <h1 className="text-2xl font-bold text-gray-900">{tournament.name}</h1>
              <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${statusColors[tournament.status]}`}>
                {tournament.status}
              </span>
            </div>
            {tournament.description && (
              <p className="text-gray-500 mb-2">{tournament.description}</p>
            )}
            <div className="flex gap-4 text-sm text-gray-500 flex-wrap">
              <span>Sport: {sportMap[tournament.sport_type]?.label ?? tournament.sport_type}</span>
              <span className="capitalize">System: {tournament.pairing_system.replace('_', ' ')}</span>
              {tournament.time_control && <span>Time: {tournament.time_control}</span>}
              {tournament.start_date && (
                <span>Start: {new Date(tournament.start_date).toLocaleDateString()}</span>
              )}
            </div>
          </div>
          {isOrganizer && (
            <div className="flex gap-2">
              {(tournament.status === 'draft' || tournament.status === 'registration') && (
                <button
                  onClick={handleStart}
                  className="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded-md text-sm font-medium"
                >
                  Start Tournament
                </button>
              )}
              {tournament.status === 'active' && (
                <button
                  onClick={handleCreateRound}
                  className="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-md text-sm font-medium"
                >
                  New Round
                </button>
              )}
              {tournament.status === 'draft' && (
                <button
                  onClick={handleDelete}
                  className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-md text-sm font-medium"
                >
                  Delete
                </button>
              )}
            </div>
          )}
        </div>
        {actionError && (
          <div className="mt-3 bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-md text-sm">
            {actionError}
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Players */}
        <div className="lg:col-span-1">
          <div className="bg-white rounded-lg shadow">
            <div className="px-6 py-4 border-b border-gray-200">
              <h2 className="text-lg font-semibold text-gray-900">
                Players ({players.length})
              </h2>
            </div>
            <ul className="divide-y divide-gray-200">
              {players.map((p) => (
                <li key={p.id} className="px-6 py-3 flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-gray-900">{p.name}</p>
                    <p className="text-xs text-gray-500">
                      Score: {p.score}
                      {p.rating && ` · Rating: ${p.rating}`}
                    </p>
                  </div>
                  {isOrganizer && tournament.status === 'draft' && (
                    <button
                      onClick={() => handleRemovePlayer(p.id)}
                      className="text-xs text-red-500 hover:text-red-700"
                    >
                      Remove
                    </button>
                  )}
                </li>
              ))}
            </ul>

            {isOrganizer && tournament.status !== 'completed' && (
              <form onSubmit={handleAddPlayer} className="px-6 py-4 border-t border-gray-200 space-y-3">
                <p className="text-sm font-medium text-gray-700">Add Player</p>
                <input
                  type="text"
                  placeholder="Player name *"
                  value={addPlayerName}
                  onChange={(e) => setAddPlayerName(e.target.value)}
                  required
                  className="block w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                />
                <input
                  type="number"
                  placeholder="Rating (optional)"
                  value={addPlayerRating}
                  onChange={(e) => setAddPlayerRating(e.target.value)}
                  className="block w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                />
                <button
                  type="submit"
                  disabled={addingPlayer}
                  className="w-full py-2 px-4 bg-indigo-600 hover:bg-indigo-700 text-white text-sm rounded-md disabled:opacity-50"
                >
                  {addingPlayer ? 'Adding...' : 'Add Player'}
                </button>
              </form>
            )}
          </div>
        </div>

        {/* Rounds & Pairings */}
        <div className="lg:col-span-2 space-y-4">
          {/* Round selector */}
          {rounds.length > 0 && (
            <div className="bg-white rounded-lg shadow">
              <div className="px-6 py-4 border-b border-gray-200">
                <h2 className="text-lg font-semibold text-gray-900">Rounds</h2>
              </div>
              <div className="px-6 py-4 flex flex-wrap gap-2">
                {rounds.map((r) => (
                  <button
                    key={r.id}
                    onClick={() => handleRoundSelect(r)}
                    className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                      selectedRound?.id === r.id
                        ? 'bg-indigo-600 text-white'
                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                    }`}
                  >
                    Round {r.round_number}
                  </button>
                ))}
              </div>
            </div>
          )}

          {/* Pairings */}
          {selectedRound && (
            <div className="bg-white rounded-lg shadow">
              <div className="px-6 py-4 border-b border-gray-200">
                <h2 className="text-lg font-semibold text-gray-900">
                  Round {selectedRound.round_number} Pairings
                </h2>
              </div>
              {pairings.length === 0 ? (
                <div className="px-6 py-4 text-gray-500 text-sm">No pairings yet.</div>
              ) : (
                <div className="divide-y divide-gray-200">
                  {pairings.map((p) => (
                    <div key={p.id} className="px-6 py-4">
                      <div className="flex items-center justify-between mb-2">
                        <div className="flex items-center space-x-3 flex-1">
                          {p.board_number && (
                            <span className="text-xs font-medium text-gray-400 w-8">#{p.board_number}</span>
                          )}
                          <div className="flex-1">
                            <div className="flex items-center gap-2">
                              <span className="text-sm font-medium text-gray-900">
                                {p.white_player_name || 'Unknown'} (White)
                              </span>
                              <span className="text-gray-400">vs</span>
                              <span className="text-sm font-medium text-gray-900">
                                {p.black_player_name || 'Unknown'} (Black)
                              </span>
                            </div>
                          </div>
                        </div>
                        <div className="ml-4">
                          <span className={`inline-flex items-center px-2 py-1 rounded text-xs font-medium ${
                            p.result === 'pending' ? 'bg-gray-100 text-gray-600' : 'bg-green-100 text-green-800'
                          }`}>
                            {resultLabels[p.result] || p.result}
                          </span>
                        </div>
                      </div>
                      {isOrganizer && (
                        <div className="mt-2 flex flex-wrap gap-2">
                          {['white_wins', 'black_wins', 'draw', 'bye', 'pending'].map((result) => (
                            <button
                              key={result}
                              onClick={() => handleUpdatePairing(p.id, result)}
                              disabled={p.result === result}
                              className={`text-xs px-2 py-1 rounded border transition-colors ${
                                p.result === result
                                  ? 'bg-indigo-600 text-white border-indigo-600'
                                  : 'bg-white text-gray-600 border-gray-300 hover:bg-gray-50'
                              } disabled:cursor-not-allowed`}
                            >
                              {result === 'white_wins' ? '1-0' :
                               result === 'black_wins' ? '0-1' :
                               result === 'draw' ? '½-½' :
                               result === 'bye' ? 'Bye' : 'Pending'}
                            </button>
                          ))}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}

          {rounds.length === 0 && (
            <div className="bg-white rounded-lg shadow px-6 py-8 text-center text-gray-500">
              {tournament.status === 'draft' || tournament.status === 'registration'
                ? 'Start the tournament to generate the first round.'
                : 'No rounds yet.'}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
