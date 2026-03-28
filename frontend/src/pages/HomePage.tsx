import { useNavigate } from 'react-router-dom'
import { useGetAuthMeQuery, usePostAuthLogoutMutation } from '@/api/client'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

const HomePage = () => {
  const navigate = useNavigate()
  const { data: me } = useGetAuthMeQuery()
  const [logout, { isLoading }] = usePostAuthLogoutMutation()

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/40 px-4">
      <Card className="w-full max-w-sm text-center">
        <CardHeader>
          <CardTitle>Welcome</CardTitle>
          <CardDescription>{me?.email}</CardDescription>
        </CardHeader>
        <CardContent>
          <Button variant="outline" onClick={handleLogout} disabled={isLoading} className="w-full">
            {isLoading ? 'Signing out…' : 'Sign out'}
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}

export default HomePage
