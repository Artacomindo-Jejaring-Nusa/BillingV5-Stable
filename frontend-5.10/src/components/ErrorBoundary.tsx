import { Component, type ErrorInfo, type ReactNode } from 'react'
import { AlertCircle, RefreshCw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
  message: string
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { hasError: false, message: '' }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, message: error.message || 'Terjadi kesalahan saat merender halaman.' }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('[ErrorBoundary]', error, info)
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="p-6 flex items-center justify-center min-h-[60vh]">
          <Card className="max-w-md w-full">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-destructive text-lg">
                <AlertCircle className="h-5 w-5" />
                Terjadi Kesalahan Tampilan
              </CardTitle>
              <CardDescription>{this.state.message}</CardDescription>
            </CardHeader>
            <CardContent className="flex gap-2">
              <Button
                onClick={() => {
                  this.setState({ hasError: false, message: '' })
                  window.location.reload()
                }}
              >
                <RefreshCw className="mr-2 h-4 w-4" />
                Muat Ulang
              </Button>
            </CardContent>
          </Card>
        </div>
      )
    }
    return this.props.children
  }
}
