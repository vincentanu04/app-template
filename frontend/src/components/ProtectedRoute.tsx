import { type ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useGetAuthMeQuery } from '@/api/client'

const ProtectedRoute = ({ children }: { children: ReactNode }) => {
  const { data, isLoading, isError } = useGetAuthMeQuery()

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center text-muted-foreground text-sm">
        Loading…
      </div>
    )
  }

  if (isError || !data) {
    return <Navigate to="/login" replace />
  }

  return <>{children}</>
}

export default ProtectedRoute
