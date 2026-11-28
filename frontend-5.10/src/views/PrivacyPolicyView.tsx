import { Link } from 'react-router-dom'
import { LogIn, ShieldCheck } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

const SECTIONS = [
  {
    id: 's1',
    num: '01',
    title: 'Identitas Pengembang & Ruang Lingkup',
    body: 'Kebijakan Privasi ini mengatur bagaimana PT Artacomindo Jejaring Nusa ("Kami", "Perusahaan", atau "Jelantik") mengumpulkan, menggunakan, memproses, menyimpan, dan melindungi informasi data pribadi pengguna aplikasi Portal Ajnusa serta portal web resmi di https://jpo.jelantik.com.',
  },
  {
    id: 's2',
    num: '02',
    title: 'Data yang Dikumpulkan',
    body: 'Kami mengumpulkan data identitas pelanggan (nama, alamat instalasi, nomor telepon/WhatsApp, email), data teknis layanan FTTH (ID pelanggan, status langganan, perangkat OLT/ODP), serta riwayat tagihan dan laporan gangguan teknis.',
  },
  {
    id: 's3',
    num: '03',
    title: 'Tujuan Penggunaan',
    body: 'Data digunakan untuk aktivasi dan pemeliharaan layanan internet FTTH, penerbitan dan penagihan invoice bulanan, pengiriman notifikasi layanan melalui WhatsApp, serta penanganan trouble ticket oleh tim teknis.',
  },
  {
    id: 's4',
    num: '04',
    title: 'Pihak Ketiga',
    body: 'Kami hanya membagikan data sebatas kebutuhan operasional kepada mitra payment gateway (seperti Xendit untuk pemrosesan pembayaran invoice) dan penyedia gateway notifikasi resmi, sesuai peraturan yang berlaku.',
  },
  {
    id: 's5',
    num: '05',
    title: 'Keamanan & Penyimpanan Data',
    body: 'Akses ke sistem dilindungi dengan autentikasi JWT terenkripsi, pembatasan hak akses berbasis peran (RBAC), serta pencatatan log aktivitas untuk menjaga integritas dan kerahasiaan data.',
  },
  {
    id: 's6',
    num: '06',
    title: 'Hak Pengguna & Hubungi Kami',
    body: 'Pengguna berhak meminta pembaruan atau koreksi data pribadi melalui layanan dukungan PT Artacomindo Jejaring Nusa di support@ajnusa.com.',
  },
]

export function PrivacyPolicyView() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <nav className="border-b bg-card sticky top-0 z-20">
        <div className="max-w-5xl mx-auto px-4 h-16 flex items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <img src="/portal_logo.svg" alt="Portal Ajnusa" className="h-8 w-8 object-contain" />
            <div>
              <div className="text-sm font-bold">Portal Ajnusa</div>
              <div className="text-xs text-muted-foreground">PT Artacomindo Jejaring Nusa</div>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <span className="hidden sm:inline-flex items-center gap-1.5 text-xs text-muted-foreground">
              <ShieldCheck className="h-4 w-4 text-emerald-600" />
              Dokumen Publik
            </span>
            <Button asChild size="sm">
              <Link to="/login">
                <LogIn className="mr-1.5 h-4 w-4" />
                Login Portal
              </Link>
            </Button>
          </div>
        </div>
      </nav>

      <header className="border-b bg-muted/30">
        <div className="max-w-5xl mx-auto px-4 py-10 space-y-2">
          <div className="text-xs font-semibold tracking-wider text-primary uppercase">Kebijakan Privasi</div>
          <h1 className="text-3xl font-bold tracking-tight">Kebijakan Privasi Portal Ajnusa</h1>
          <p className="text-sm text-muted-foreground">
            PT Artacomindo Jejaring Nusa: Sistem Manajemen FTTH & Billing Jelantik
          </p>
        </div>
      </header>

      <div className="max-w-5xl mx-auto px-4 py-8 grid md:grid-cols-[220px_1fr] gap-8">
        <aside className="hidden md:block">
          <div className="sticky top-24 space-y-2">
            <div className="text-xs font-semibold text-muted-foreground uppercase">Daftar Isi</div>
            <ol className="space-y-1.5 text-sm">
              {SECTIONS.map((s) => (
                <li key={s.id}>
                  <a href={`#${s.id}`} className="text-muted-foreground hover:text-foreground transition-colors">
                    {s.num}. {s.title}
                  </a>
                </li>
              ))}
            </ol>
          </div>
        </aside>

        <main className="space-y-4">
          {SECTIONS.map((s) => (
            <Card key={s.id} id={s.id} className="scroll-mt-24">
              <CardContent className="pt-6 space-y-2">
                <div className="text-xs font-semibold text-primary">{s.num}</div>
                <h2 className="text-lg font-semibold">{s.title}</h2>
                <p className="text-sm leading-relaxed text-muted-foreground">{s.body}</p>
              </CardContent>
            </Card>
          ))}
        </main>
      </div>
    </div>
  )
}
