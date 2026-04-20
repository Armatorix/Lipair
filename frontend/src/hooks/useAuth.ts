import { useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import { userApi } from '../api/client'

export function useAuth() {
  const { user, token, setAuth, logout, isAuthenticated } = useAuthStore()
  return { user, token, setAuth, logout, isAuthenticated }
}

export function useOAuthCallback() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { setAuth } = useAuthStore()

  useEffect(() => {
    const token = searchParams.get('token')
    if (token) {
      userApi.getMe().then((res) => {
        setAuth(res.data, token)
        navigate('/dashboard')
      }).catch(() => {
        navigate('/login')
      })
    } else {
      navigate('/login')
    }
  }, [searchParams, setAuth, navigate])
}
