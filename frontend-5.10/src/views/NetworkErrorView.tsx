import { useMemo, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { AlertTriangle, RefreshCw, RotateCcw, WifiOff } from 'lucide-react'
import { checkServerHealth } from '@/services/api'
import { errorStorage } from '@/services/errorStorage'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

export function NetworkErrorView() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const [isRetrying, setIsRetrying] = useState(false)
  const [retryFeedback, setRetryFeedback] = useState('')

  const errorId = searchParams.get('errorId') || ''
  const errorDetails = useMemo(() => (errorId ? errorStorage.getError(errorId) : null), [errorId])

  const errorStatus = errorDetails?.errorStatus || 503
  const errorTitle =
    errorDetails?.errorType === 'timeout'
      ? 'Koneksi Timeout'
      : errorDetails?.errorType === 'server'
        ? 'Server Tidak Tersedia'
        : 'Gagal Terhubung ke Server'

  const errorMessage =
    errorDetails?.message ||
    'Sistem tidak dapat menjangkau server backend. Pastikan koneksi jaringan dan server berjalan normal.'

  const handleRetry = async () => {
    setIsRetrying(true)
    setRetryFeedback('')
    const res = await checkServerHealth()
    setIsRetrying(false)
    if (res.success) {
      navigate('/dashboard')
    } else {
      setRetryFeedback('Server masih belum merespons. Silakan coba beberapa saat lagi.')
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/30 p-4">
      <div className="w-full max-w-lg space-y-4">
        <Card>
          <CardContent className="pt-8 pb-6 px-6 text-center space-y-5">
            <WifiOff className="h-16 w-16 text-destructive mx-auto" />
            <div className="space-y-1">
              <div className="text-4xl font-bold text-destructive">{errorStatus}</div>
              <h1 className="text-xl font-semibold">{errorTitle}</h1>
              <p className="text-sm text-muted-foreground">{errorMessage}</p>
            </div>

            {retryFeedback && (
              <div className="rounded-md border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-xs text-amber-700 dark:text-amber-300 flex items-center justify-center gap-2">
                <AlertTriangle className="h-4 w-4 shrink-0" />
                <span>{retryFeedback}</span>
              </div>
            )}

            <div className="flex flex-wrap justify-center gap-3">
              <Button onClick={handleRetry} disabled={isRetrying} className="min-h-[44px]">
                <RefreshCw className={`mr-2 h-4 w-4 ${isRetrying ? 'animate-spin' : ''}`} />
                {isRetrying ? 'Menghubungkan...' : 'Coba Lagi'}
              </Button>
              <Button variant="outline" onClick={() => window.location.reload()} className="min-h-[44px]">
                <RotateCcw className="mr-2 h-4 w-4" />
                Refresh Halaman
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="py-4 px-5 text-xs text-muted-foreground space-y-1.5">
            <div className="font-semibold text-foreground">Solusi Cepat:</div>
            <ul className="list-disc pl-4 space-y-1">
              <li>Periksa koneksi internet atau jaringan lokal Anda.</li>
              <li>Pastikan layanan backend API sedang berjalan.</li>
              <li>Hubungi tim IT jika gangguan berlanjut.</li>
            </ul>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
