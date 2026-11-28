import { useMemo, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { AlertCircle, ArrowLeft, Bug, Headset } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'

export function NotFoundView() {
  const location = useLocation()
  const navigate = useNavigate()
  const [showReport, setShowReport] = useState(false)
  const [issueText, setIssueText] = useState('')
  const [submitted, setSubmitted] = useState(false)

  const referenceId = useMemo(() => `404-${Date.now().toString(36).toUpperCase()}`, [])
  const currentTime = useMemo(() => new Date().toLocaleString('id-ID'), [])

  const handleSubmitReport = (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitted(true)
    setTimeout(() => {
      setShowReport(false)
      setSubmitted(false)
      setIssueText('')
    }, 1000)
  }

  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/30 p-4">
      <div className="w-full max-w-lg space-y-4">
        <Card>
          <CardContent className="pt-8 pb-6 px-6 text-center space-y-5">
            <AlertCircle className="h-16 w-16 text-muted-foreground mx-auto" />
            <div className="space-y-1">
              <div className="text-5xl font-bold text-muted-foreground">404</div>
              <h1 className="text-xl font-semibold">Halaman Tidak Ditemukan</h1>
              <p className="text-sm text-muted-foreground">
                Maaf, halaman yang Anda cari tidak ada atau telah dipindahkan.
              </p>
            </div>

            <div className="flex flex-wrap justify-center gap-3">
              <Button onClick={() => navigate('/dashboard')} className="min-h-[44px]">
                <ArrowLeft className="mr-2 h-4 w-4" />
                Kembali ke Dashboard
              </Button>
              <Button variant="outline" onClick={() => setShowReport(true)} className="min-h-[44px]">
                <Bug className="mr-2 h-4 w-4" />
                Laporkan Masalah
              </Button>
            </div>

            <Separator />

            <div className="text-xs text-muted-foreground space-y-1">
              <div>
                <strong>URL:</strong> {location.pathname}
              </div>
              <div>
                <strong>Waktu:</strong> {currentTime}
              </div>
              <div>
                <strong>ID Referensi:</strong> {referenceId}
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="bg-primary text-primary-foreground">
          <CardContent className="py-4 px-5 flex items-center justify-center gap-3 text-center">
            <Headset className="h-5 w-5 shrink-0" />
            <div className="text-xs">
              <div className="font-medium">Butuh bantuan? Hubungi IT Support</div>
              <div className="opacity-85">support@ajnusa.com | ahmad@ajnusa.com</div>
            </div>
          </CardContent>
        </Card>
      </div>

      <Dialog open={showReport} onOpenChange={setShowReport}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Laporkan Masalah</DialogTitle>
            <DialogDescription>
              Sertakan detail halaman yang sedang Anda coba akses.
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleSubmitReport} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="issue-desc">Deskripsi Masalah</Label>
              <textarea
                id="issue-desc"
                rows={4}
                value={issueText}
                onChange={(e) => setIssueText(e.target.value)}
                placeholder="Jelaskan masalah yang Anda alami..."
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                required
              />
            </div>
            {submitted && (
              <div className="text-xs text-emerald-600 font-medium">
                Laporan dicatat. Terima kasih.
              </div>
            )}
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setShowReport(false)}>
                Batal
              </Button>
              <Button type="submit">Kirim Laporan</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}
