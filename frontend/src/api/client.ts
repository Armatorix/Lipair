import axios from 'axios'
import { useAuthStore } from '../store/authStore'

const apiClient = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
})

apiClient.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      useAuthStore.getState().logout()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export interface User {
  id: string
  email?: string
  name: string
  google_id?: string
  lichess_id?: string
  lichess_username?: string
  avatar_url?: string
  created_at: string
  updated_at: string
}

export type SportType =
  | 'chess'
  | 'checkers'
  | 'go'
  | 'shogi'
  | 'tennis'
  | 'table_tennis'
  | 'badminton'
  | 'squash'
  | 'pool'
  | 'darts'
  | 'foosball'
  | 'other'

export const SPORT_OPTIONS: { value: SportType; label: string; icon: string }[] = [
  { value: 'chess', label: 'Chess', icon: '♛' },
  { value: 'checkers', label: 'Checkers', icon: '⛀' },
  { value: 'go', label: 'Go', icon: '⚫' },
  { value: 'shogi', label: 'Shogi', icon: '将' },
  { value: 'tennis', label: 'Tennis', icon: '🎾' },
  { value: 'table_tennis', label: 'Table Tennis', icon: '🏓' },
  { value: 'badminton', label: 'Badminton', icon: '🏸' },
  { value: 'squash', label: 'Squash', icon: '🎯' },
  { value: 'pool', label: 'Pool / Billiards', icon: '🎱' },
  { value: 'darts', label: 'Darts', icon: '🎯' },
  { value: 'foosball', label: 'Foosball', icon: '⚽' },
  { value: 'other', label: 'Other', icon: '🏆' },
]

export interface Tournament {
  id: string
  name: string
  description?: string
  organizer_id: string
  sport_type: SportType
  pairing_system: 'round_robin' | 'knockout' | 'swiss' | 'scheveningen' | 'manual'
  status: 'draft' | 'registration' | 'active' | 'completed'
  rounds_count?: number
  time_control?: string
  start_date?: string
  end_date?: string
  created_at: string
  updated_at: string
}

export interface TournamentPlayer {
  id: string
  tournament_id: string
  player_id: string
  name: string
  email?: string
  avatar_url?: string
  seed?: number
  rating?: number
  score: number
  joined_at: string
}

export interface Round {
  id: string
  tournament_id: string
  round_number: number
  status: 'pending' | 'active' | 'completed'
  created_at: string
}

export interface Pairing {
  id: string
  round_id: string
  white_player_id?: string
  black_player_id?: string
  white_player_name?: string
  black_player_name?: string
  result: 'white_wins' | 'black_wins' | 'draw' | 'bye' | 'pending'
  board_number?: number
  created_at: string
  updated_at: string
}

// Auth API
export const authApi = {
  register: (data: { name: string; email: string; password: string }) =>
    apiClient.post<{ token: string; user: User }>('/auth/register', data),
  login: (data: { email: string; password: string }) =>
    apiClient.post<{ token: string; user: User }>('/auth/login', data),
}

// User API
export const userApi = {
  getMe: () => apiClient.get<User>('/users/me'),
  updateMe: (data: { name?: string; avatar_url?: string }) =>
    apiClient.put<User>('/users/me', data),
}

// Tournament API
export const tournamentApi = {
  list: () => apiClient.get<Tournament[]>('/tournaments'),
  create: (data: Partial<Tournament>) => apiClient.post<Tournament>('/tournaments', data),
  get: (id: string) => apiClient.get<Tournament>(`/tournaments/${id}`),
  update: (id: string, data: Partial<Tournament>) => apiClient.put<Tournament>(`/tournaments/${id}`, data),
  delete: (id: string) => apiClient.delete(`/tournaments/${id}`),
  start: (id: string) => apiClient.post<Tournament>(`/tournaments/${id}/start`),
  listPlayers: (id: string) => apiClient.get<TournamentPlayer[]>(`/tournaments/${id}/players`),
  addPlayer: (id: string, data: { player_id?: string; name: string; rating?: number; seed?: number }) =>
    apiClient.post<TournamentPlayer>(`/tournaments/${id}/players`, data),
  removePlayer: (id: string, playerId: string) => apiClient.delete(`/tournaments/${id}/players/${playerId}`),
  listRounds: (id: string) => apiClient.get<Round[]>(`/tournaments/${id}/rounds`),
  createRound: (id: string) => apiClient.post<Round>(`/tournaments/${id}/rounds`),
  getRound: (id: string, roundId: string) => apiClient.get<Round>(`/tournaments/${id}/rounds/${roundId}`),
  listPairings: (id: string, roundId: string) => apiClient.get<Pairing[]>(`/tournaments/${id}/rounds/${roundId}/pairings`),
  updatePairing: (id: string, roundId: string, pairingId: string, data: { result: string }) =>
    apiClient.put<Pairing>(`/tournaments/${id}/rounds/${roundId}/pairings/${pairingId}`, data),
}

export default apiClient
