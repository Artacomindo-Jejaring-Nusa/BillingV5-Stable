import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import type { AxiosError } from 'axios'
import {
  Activity,
  AlertCircle,
  CheckCircle2,
  ChevronLeft,
  Eye,
  EyeOff,
  KeyRound,
  Lock,
  Mail,
  Receipt,
  ShieldCheck,
  Users,
} from 'lucide-react'
import apiClient from '@/services/api'
import { useAuthStore } from '@/stores/auth'
import { getDefaultHomeRoute } from '@/router/guards'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

interface ErrorResponse {
  detail?: string
  message?: string
}

export function LoginView() {
  const navigate = useNavigate()
  const login = useAuthStore((s) => s.login)

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [rememberMe, setRememberMe] = useState(false)
  const [showPassword, setShowPassword] = useState(false)

  const [showForgotPassword, setShowForgotPassword] = useState(false)
  const [resetEmail, setResetEmail] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [resetToken, setResetToken] = useState('')

  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [successMessage, setSuccessMessage] = useState('')

  useEffect(() => {
    const rememberedEmail = localStorage.getItem('remember_email')
    if (rememberedEmail) {
      setEmail(rememberedEmail)
      setRememberMe(true)
    }
  }, [])

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email.trim() || !password) {
      setError('Email dan kata sandi harus diisi')
      return
    }
    setError('')
    setLoading(true)
    try {
      const success = await login(email.trim(), password)
      if (success) {
        if (rememberMe) {
          localStorage.setItem('remember_email', email.trim())
        } else {
          localStorage.removeItem('remember_email')
        }
        const target = getDefaultHomeRoute(useAuthStore.getState().hasPermission) || '/dashboard'
        navigate(target)
      } else {
        setError('Email atau kata sandi salah')
      }
    } catch (err) {
      const axErr = err as AxiosError<ErrorResponse>
      setError(
        axErr.response?.data?.detail ||
          axErr.response?.data?.message ||
          'Terjadi kesalahan saat login'
      )
    } finally {
      setLoading(false)
    }
  }

  const handleResetPassword = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!resetEmail.trim() || !newPassword || !resetToken.trim()) {
      setError('Email, kata sandi baru, dan token harus diisi')
      return
    }
    setError('')
    setSuccessMessage('')
    setLoading(true)
    try {
      const response = await apiClient.post<{ message: string }>('/users/reset-password', {
        email: resetEmail.trim(),
        new_password: newPassword,
        token: resetToken.trim(),
      })
      setSuccessMessage(response.data.message || 'Kata sandi berhasil diatur ulang.')
      setShowForgotPassword(false)
    } catch (err) {
      const axErr = err as AxiosError<ErrorResponse>
      setError(
        axErr.response?.data?.detail ||
          axErr.response?.data?.message ||
          'Terjadi kesalahan saat mengatur ulang kata sandi'
      )
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex flex-col lg:flex-row bg-white text-slate-900">
      {/* Mobile Brand Header (< 640px) */}
      <div className="sm:hidden bg-[#0f2a8f] text-white px-6 py-6 space-y-3">
        <div className="flex items-center justify-between">
          <img src="/icon_light.ico" alt="Jelantik Logo" className="h-10 w-10 object-contain" />
          <span className="rounded border border-white/25 bg-white/10 px-2.5 py-0.5 text-xs font-semibold">
            AJNUSA V5.10
          </span>
        </div>
        <div>
          <h2 className="text-xl font-bold tracking-tight">WELCOME!</h2>
          <p className="text-xs text-blue-100">Portal FTTH & Billing Ajnusa V5.10</p>
        </div>
      </div>

      {/* Left Panel: Sign-in Form */}
      <div className="flex-1 flex flex-col justify-between px-6 py-8 sm:px-12 lg:px-16">
        <div className="flex-1 flex flex-col justify-center max-w-sm w-full mx-auto">
          {/* Centered Jelantik Logo */}
          <div className="flex justify-center mb-6">
            <img
              src="/icon_dark.ico"
              alt="Jelantik Logo"
              className="h-16 w-16 object-contain"
              onError={(e) => {
                ;(e.currentTarget as HTMLImageElement).src = '/Jelantik 1.svg'
              }}
            />
          </div>

          {!showForgotPassword ? (
            <form onSubmit={handleLogin} className="space-y-5" noValidate>
              <div className="text-center space-y-1">
                <h1 className="text-2xl font-bold tracking-tight text-slate-900">
                  Selamat Datang Kembali
                </h1>
                <p className="text-sm text-slate-500">
                  Silakan login untuk mengakses portal Anda
                </p>
              </div>

              {successMessage && (
                <div
                  role="status"
                  className="flex items-center gap-2 rounded-md border border-emerald-500/30 bg-emerald-50 p-3 text-sm text-emerald-700"
                >
                  <CheckCircle2 className="h-4 w-4 shrink-0" />
                  <span>{successMessage}</span>
                </div>
              )}

              <div className="space-y-4">
                <div className="space-y-1.5">
                  <Label htmlFor="login-email" className="text-slate-700 font-medium">
                    Alamat Email
                  </Label>
                  <div className="relative">
                    <Mail className="h-4 w-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
                    <Input
                      id="login-email"
                      type="email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      placeholder="Masukkan alamat email Anda"
                      disabled={loading}
                      autoComplete="email"
                      className="pl-9 h-11 bg-white border-slate-200 text-slate-900 placeholder:text-slate-400"
                    />
                  </div>
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="login-password" className="text-slate-700 font-medium">
                    Kata Sandi
                  </Label>
                  <div className="relative">
                    <Lock className="h-4 w-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
                    <Input
                      id="login-password"
                      type={showPassword ? 'text' : 'password'}
                      value={password}
                      onChange={(e) => setPassword(e.target.value)}
                      placeholder="Masukkan kata sandi Anda"
                      disabled={loading}
                      autoComplete="current-password"
                      className="pl-9 pr-10 h-11 bg-white border-slate-200 text-slate-900 placeholder:text-slate-400"
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword((v) => !v)}
                      aria-label={showPassword ? 'Sembunyikan kata sandi' : 'Lihat kata sandi'}
                      className="absolute right-2.5 top-1/2 -translate-y-1/2 p-1 text-slate-400 hover:text-slate-700"
                    >
                      {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </button>
                  </div>
                </div>
              </div>

              {error && (
                <div
                  role="alert"
                  className="flex items-start gap-2.5 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700"
                >
                  <AlertCircle className="h-4 w-4 mt-0.5 shrink-0 text-red-600" />
                  <div>
                    <div className="font-medium">{error}</div>
                    <div className="text-xs text-red-600/90">Periksa kembali data Anda dan coba lagi.</div>
                  </div>
                </div>
              )}

              <div className="flex items-center justify-between text-sm">
                <label className="inline-flex items-center gap-2 cursor-pointer select-none text-slate-600">
                  <input
                    type="checkbox"
                    checked={rememberMe}
                    onChange={(e) => setRememberMe(e.target.checked)}
                    disabled={loading}
                    className="h-4 w-4 rounded border-slate-300 text-[#1d4ed8]"
                  />
                  <span>Ingat Saya</span>
                </label>
                <button
                  type="button"
                  onClick={() => {
                    setError('')
                    setSuccessMessage('')
                    setShowForgotPassword(true)
                  }}
                  className="text-[#1d4ed8] hover:underline font-medium"
                >
                  Lupa kata sandi?
                </button>
              </div>

              <Button
                type="submit"
                disabled={loading}
                className="w-full h-11 bg-[#1d4ed8] hover:bg-[#1e40af] text-white font-semibold tracking-wide"
              >
                {loading ? 'MEMPROSES...' : 'MASUK'}
              </Button>

              <div className="flex items-center justify-center gap-2 text-xs text-slate-500 pt-1">
                <ShieldCheck className="h-4 w-4 text-slate-400" />
                <span>Akses aman ke portal Anda</span>
              </div>
            </form>
          ) : (
            <form onSubmit={handleResetPassword} className="space-y-5" noValidate>
              <button
                type="button"
                onClick={() => {
                  setShowForgotPassword(false)
                  setError('')
                }}
                className="inline-flex items-center gap-1 text-sm text-slate-500 hover:text-slate-800"
              >
                <ChevronLeft className="h-4 w-4" />
                <span>Kembali ke login</span>
              </button>

              <div className="space-y-1">
                <h1 className="text-2xl font-bold tracking-tight text-slate-900">
                  Atur Ulang Kata Sandi
                </h1>
                <p className="text-sm text-slate-500">
                  Masukkan email, token reset, dan kata sandi baru Anda
                </p>
              </div>

              <div className="space-y-4">
                <div className="space-y-1.5">
                  <Label htmlFor="reset-email" className="text-slate-700 font-medium">
                    Alamat Email
                  </Label>
                  <div className="relative">
                    <Mail className="h-4 w-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
                    <Input
                      id="reset-email"
                      type="email"
                      value={resetEmail}
                      onChange={(e) => setResetEmail(e.target.value)}
                      placeholder="Masukkan alamat email Anda"
                      disabled={loading}
                      className="pl-9 h-11 bg-white border-slate-200 text-slate-900"
                    />
                  </div>
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="reset-password" className="text-slate-700 font-medium">
                    Kata Sandi Baru
                  </Label>
                  <div className="relative">
                    <Lock className="h-4 w-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
                    <Input
                      id="reset-password"
                      type="password"
                      value={newPassword}
                      onChange={(e) => setNewPassword(e.target.value)}
                      placeholder="Masukkan kata sandi baru"
                      disabled={loading}
                      className="pl-9 h-11 bg-white border-slate-200 text-slate-900"
                    />
                  </div>
                </div>

                <div className="space-y-1.5">
                  <Label htmlFor="reset-token" className="text-slate-700 font-medium">
                    Token Reset
                  </Label>
                  <div className="relative">
                    <KeyRound className="h-4 w-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
                    <Input
                      id="reset-token"
                      type="text"
                      value={resetToken}
                      onChange={(e) => setResetToken(e.target.value)}
                      placeholder="Masukkan token reset"
                      disabled={loading}
                      className="pl-9 h-11 bg-white border-slate-200 text-slate-900"
                    />
                  </div>
                </div>
              </div>

              {error && (
                <div
                  role="alert"
                  className="flex items-start gap-2 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700"
                >
                  <AlertCircle className="h-4 w-4 mt-0.5 shrink-0 text-red-600" />
                  <span>{error}</span>
                </div>
              )}

              <Button
                type="submit"
                disabled={loading}
                className="w-full h-11 bg-[#1d4ed8] hover:bg-[#1e40af] text-white font-semibold"
              >
                {loading ? 'MEMPROSES...' : 'ATUR ULANG KATA SANDI'}
              </Button>
            </form>
          )}
        </div>

        {/* Legal Footer */}
        <footer className="pt-6 text-center text-xs text-slate-500">
          <div className="flex items-center justify-center gap-2.5 flex-wrap">
            <a
              href="https://www.ajnusa.com"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-slate-800 hover:underline"
            >
              www.ajnusa.com
            </a>
            <span>|</span>
            <a
              href="https://www.jelantik.com"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-slate-800 hover:underline"
            >
              www.jelantik.com
            </a>
            <span>|</span>
            <Link to="/privacy-policy" className="hover:text-slate-800 hover:underline">
              Kebijakan Privasi
            </Link>
          </div>
        </footer>
      </div>

      {/* Right Panel: Brand & FTTH Network Illustration */}
      <div className="hidden lg:flex lg:w-[50%] flex-col justify-between bg-gradient-to-br from-[#0f2a8f] via-[#0b237a] to-[#081754] text-white p-12">
        {/* Masthead */}
        <div className="flex items-center justify-between">
          <span className="text-xs font-semibold tracking-widest text-blue-200/90 uppercase">
            AJNUSA / CONNECTED SERVICES
          </span>
          <span className="rounded border border-white/25 bg-white/10 px-2.5 py-0.5 text-xs font-semibold">
            V5.10
          </span>
        </div>

        {/* Welcome + Capabilities + FTTH Illustration */}
        <div className="space-y-8 my-auto py-8">
          <div className="space-y-2">
            <h2 className="text-4xl font-bold tracking-tight">WELCOME!</h2>
            <p className="text-base text-blue-100/90">Portal FTTH & Billing Ajnusa V5.10</p>
          </div>

          <div className="flex flex-wrap gap-2.5">
            <div className="inline-flex items-center gap-2 rounded-md border border-white/15 bg-white/10 px-3 py-1.5 text-xs font-medium">
              <Activity className="h-3.5 w-3.5 text-[#2bb5f0]" />
              <span>Monitoring Jaringan</span>
            </div>
            <div className="inline-flex items-center gap-2 rounded-md border border-white/15 bg-white/10 px-3 py-1.5 text-xs font-medium">
              <Receipt className="h-3.5 w-3.5 text-[#2bb5f0]" />
              <span>Billing & Tagihan</span>
            </div>
            <div className="inline-flex items-center gap-2 rounded-md border border-white/15 bg-white/10 px-3 py-1.5 text-xs font-medium">
              <Users className="h-3.5 w-3.5 text-[#2bb5f0]" />
              <span>Manajemen Pelanggan</span>
            </div>
          </div>

          {/* FTTH Network Illustration */}
          <div className="pt-2" aria-hidden="true">
            <svg viewBox="0 0 520 260" fill="none" className="w-full max-w-md mx-auto">
              <path d="M260 112 L120 185" stroke="rgba(232, 245, 255, 0.45)" strokeWidth="2" strokeDasharray="5 5" />
              <path d="M260 112 L400 185" stroke="rgba(232, 245, 255, 0.45)" strokeWidth="2" strokeDasharray="5 5" />
              <path d="M260 112 L260 212" stroke="rgba(232, 245, 255, 0.45)" strokeWidth="2" strokeDasharray="5 5" />
              <circle cx="190" cy="148" r="4" fill="#ff6b3d" />
              <circle cx="330" cy="148" r="4" fill="#ff6b3d" />
              <circle cx="260" cy="166" r="4" fill="#ff6b3d" />

              <g transform="translate(204, 24)">
                <path d="M36 16 Q56 0 76 16" stroke="#2bb5f0" strokeWidth="2.5" strokeLinecap="round" />
                <path d="M44 24 Q56 14 68 24" stroke="#ffffff" strokeWidth="2.5" strokeLinecap="round" />
                <rect x="4" y="32" width="104" height="64" rx="10" fill="#ffffff" />
                <rect x="14" y="44" width="84" height="12" rx="4" fill="#e8f5ff" />
                <rect x="14" y="64" width="52" height="10" rx="4" fill="#e8f5ff" />
                <circle cx="88" cy="69" r="4" fill="#ff6b3d" />
                <rect x="18" y="82" width="12" height="6" rx="2" fill="#0b1f6b" />
                <rect x="36" y="82" width="12" height="6" rx="2" fill="#0b1f6b" />
              </g>

              <g transform="translate(83, 168)">
                <path d="M37 4 L4 26 L10 26 L10 66 L64 66 L64 26 L70 26 Z" fill="#e8f5ff" />
                <path d="M37 0 L0 25 L74 25 Z" fill="#ffffff" />
                <rect x="28" y="40" width="18" height="26" rx="2" fill="#0f2a8f" />
                <circle cx="37" cy="31" r="5" fill="#2bb5f0" />
              </g>

              <g transform="translate(231, 196)">
                <path d="M29 4 L4 20 L8 20 L8 52 L50 52 L50 20 L54 20 Z" fill="#e8f5ff" />
                <path d="M29 0 L0 19 L58 19 Z" fill="#ffffff" />
                <rect x="22" y="32" width="14" height="20" rx="2" fill="#0f2a8f" />
                <circle cx="29" cy="24" r="4" fill="#2bb5f0" />
              </g>

              <g transform="translate(363, 168)">
                <path d="M37 4 L4 26 L10 26 L10 66 L64 66 L64 26 L70 26 Z" fill="#e8f5ff" />
                <path d="M37 0 L0 25 L74 25 Z" fill="#ffffff" />
                <rect x="28" y="40" width="18" height="26" rx="2" fill="#0f2a8f" />
                <circle cx="37" cy="31" r="5" fill="#2bb5f0" />
              </g>
            </svg>
          </div>
        </div>

        {/* Brand Signature */}
        <div className="space-y-0.5">
          <p className="text-base font-semibold">Koneksi yang menyatukan.</p>
          <p className="text-xs text-blue-200/80">PT. Artacomindo Jejaring Nusa</p>
        </div>
      </div>
    </div>
  )
}
