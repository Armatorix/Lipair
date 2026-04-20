import { useEffect } from 'react'
import { useSearchParams, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import { userApi } from '../api/client'

export default function OAuthCallbackPage() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const { setAuth } = useAuthStore()

  useEffect(() => {
    const token = searchParams.get('token')
    if (!token) {
      navigate('/login?error=no_token')
      return
    }
    // Temporarily set token so userApi can use it
    useAuthStore.setState({ token })
    userApi.getMe()
      .then((res) => {
        setAuth(res.data, token)
        navigate('/dashboard')
      })
      .catch(() => {
        useAuthStore.setState({ token: null })
        navigate('/login?error=auth_failed')
      })
  }, [searchParams, setAuth, navigate])

  return (
    <div className="min-h-screen flex items-center justify-center">
      <div className="text-center">
        <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600 mb-4"></div>
        <p className="text-gray-600">Signing you in...</p>
      </div>
    </div>
  )
}
