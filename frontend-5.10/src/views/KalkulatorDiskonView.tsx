import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Calculator,
  Check,
  CheckCircle2,
  ChevronDown,
  ChevronUp,
  Circle,
  Copy,
  Info,
  Layers,
  Package,
  Percent,
  PiggyBank,
  RefreshCw,
  Tag,
  Wallet,
} from 'lucide-react'
import apiClient from '@/services/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

interface Brand {
  id_brand: string
  brand: string
  pajak: number
}

interface Paket {
  id: number
  id_brand: string
  nama_paket: string
  harga: number
}

interface DiskonResult {
  nama_paket: string
  nama_brand: string
  harga_paket: number
  pajak_persen: number
  pajak_amount: number
  subtotal_sebelum_diskon: number
  persentase_diskon: number
  diskon_amount: number
  harga_final: number
  detail_perhitungan: string
}

function formatCurrency(value?: number | null): string {
  if (typeof value !== 'number' || Number.isNaN(value)) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(value)
}

export function KalkulatorDiskonView() {
  const [brandList, setBrandList] = useState<Brand[]>([])
  const [paketList, setPaketList] = useState<Paket[]>([])
  const [selectedBrandId, setSelectedBrandId] = useState('')
  const [selectedPaketId, setSelectedPaketId] = useState('')
  const [persentaseDiskon, setPersentaseDiskon] = useState(0)
  const [calculationResult, setCalculationResult] = useState<DiskonResult | null>(null)
  const [calculating, setCalculating] = useState(false)
  const [showDetail, setShowDetail] = useState(false)
  const [copied, setCopied] = useState(false)

  const fetchInitialData = useCallback(async () => {
    const [bRes, pRes] = await Promise.allSettled([
      apiClient.get('/harga_layanan'),
      apiClient.get('/paket_layanan'),
    ])
    if (bRes.status === 'fulfilled') {
      const list = Array.isArray(bRes.value.data) ? bRes.value.data : bRes.value.data?.data || []
      setBrandList(list)
    }
    if (pRes.status === 'fulfilled') {
      const list = Array.isArray(pRes.value.data) ? pRes.value.data : pRes.value.data?.data || []
      setPaketList(list)
    }
  }, [])

  useEffect(() => {
    void fetchInitialData()
  }, [fetchInitialData])

  const selectedBrand = useMemo(
    () => brandList.find((b) => String(b.id_brand) === selectedBrandId) || null,
    [brandList, selectedBrandId],
  )

  const filteredPaketList = useMemo(() => {
    if (!selectedBrandId) return []
    return paketList.filter((p) => String(p.id_brand) === selectedBrandId)
  }, [paketList, selectedBrandId])

  const selectedPaket = useMemo(
    () => filteredPaketList.find((p) => String(p.id) === selectedPaketId) || null,
    [filteredPaketList, selectedPaketId],
  )

  const completionPercentage = useMemo(() => {
    let completed = 0
    if (selectedBrand) completed += 33.33
    if (selectedPaket) completed += 33.33
    if (persentaseDiskon > 0) completed += 33.34
    return Math.round(completed)
  }, [selectedBrand, selectedPaket, persentaseDiskon])

  useEffect(() => {
    if (!selectedBrand || !selectedPaket || persentaseDiskon <= 0) {
      setCalculationResult(null)
      return
    }
    const timer = window.setTimeout(async () => {
      setCalculating(true)
      try {
        const res = await apiClient.post<DiskonResult>('/calculator/diskon', {
          id_brand: selectedBrand.id_brand,
          paket_layanan_id: selectedPaket.id,
          persentase_diskon: persentaseDiskon,
        })
        setCalculationResult(res.data)
      } catch (err) {
        console.error('Gagal melakukan kalkulasi diskon:', err)
        setCalculationResult(null)
      } finally {
        setCalculating(false)
      }
    }, 500)
    return () => window.clearTimeout(timer)
  }, [selectedBrand, selectedPaket, persentaseDiskon])

  const handleBrandChange = (val: string) => {
    setSelectedBrandId(val)
    setSelectedPaketId('')
    setCalculationResult(null)
  }

  const handleDiscountChange = (raw: string) => {
    const num = Number(raw)
    if (Number.isNaN(num)) {
      setPersentaseDiskon(0)
      return
    }
    setPersentaseDiskon(Math.max(0, Math.min(100, num)))
  }

  const resetCalculation = () => {
    setSelectedBrandId('')
    setSelectedPaketId('')
    setPersentaseDiskon(0)
    setCalculationResult(null)
    setShowDetail(false)
  }

  const copyResult = async () => {
    if (!calculationResult) return
    try {
      await navigator.clipboard.writeText((calculationResult.detail_perhitungan || '').trim())
      setCopied(true)
      window.setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Gagal menyalin hasil:', err)
    }
  }

  return (
    <div className="p-4 md:p-6 space-y-6">
      <div className="flex items-center gap-3.5">
        <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-primary text-primary-foreground">
          <Percent className="h-5 w-5" />
        </div>
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Kalkulator Diskon</h2>
          <p className="text-sm text-muted-foreground">Simulasi perhitungan harga dengan diskon untuk layanan FTTH</p>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Form Kalkulasi */}
        <Card className="lg:col-span-5">
          <CardHeader className="pb-4 border-b">
            <CardTitle className="text-base flex items-center gap-2">
              <Layers className="h-4 w-4 text-primary" />
              Form Kalkulasi
            </CardTitle>
          </CardHeader>
          <CardContent className="pt-5 space-y-5">
            <div className="space-y-1.5">
              <Label className="flex items-center gap-1.5">
                {selectedBrand ? (
                  <CheckCircle2 className="h-4 w-4 text-emerald-600" />
                ) : (
                  <Circle className="h-4 w-4 text-muted-foreground" />
                )}
                Pilih Brand
              </Label>
              <Select value={selectedBrandId} onValueChange={handleBrandChange}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Pilih brand layanan" />
                </SelectTrigger>
                <SelectContent>
                  {brandList.map((b) => (
                    <SelectItem key={b.id_brand} value={String(b.id_brand)}>
                      {b.brand}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-1.5">
              <Label className="flex items-center gap-1.5">
                {selectedPaket ? (
                  <CheckCircle2 className="h-4 w-4 text-emerald-600" />
                ) : (
                  <Circle className="h-4 w-4 text-muted-foreground" />
                )}
                Pilih Paket Layanan
              </Label>
              <Select value={selectedPaketId} onValueChange={setSelectedPaketId} disabled={!selectedBrandId}>
                <SelectTrigger className="w-full">
                  <SelectValue placeholder={selectedBrandId ? 'Pilih paket layanan' : 'Pilih brand terlebih dahulu'} />
                </SelectTrigger>
                <SelectContent>
                  {filteredPaketList.map((p) => (
                    <SelectItem key={p.id} value={String(p.id)}>
                      {p.nama_paket}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="diskon-percent" className="flex items-center gap-1.5">
                {persentaseDiskon > 0 ? (
                  <CheckCircle2 className="h-4 w-4 text-emerald-600" />
                ) : (
                  <Circle className="h-4 w-4 text-muted-foreground" />
                )}
                Persentase Diskon (%)
              </Label>
              <div className="flex items-center gap-3">
                <Input
                  id="diskon-percent"
                  type="number"
                  min={0}
                  max={100}
                  step={1}
                  disabled={!selectedPaket}
                  value={persentaseDiskon}
                  onChange={(e) => handleDiscountChange(e.target.value)}
                  className="w-24"
                  placeholder="20"
                />
                <input
                  type="range"
                  min={0}
                  max={100}
                  step={1}
                  disabled={!selectedPaket}
                  value={persentaseDiskon}
                  onChange={(e) => handleDiscountChange(e.target.value)}
                  className="flex-1 accent-primary cursor-pointer disabled:opacity-50"
                />
                <span className="text-sm font-bold w-10 text-right">{persentaseDiskon}%</span>
              </div>
            </div>

            {/* Progress Pengisian */}
            <div className="rounded-lg bg-muted/50 p-4 space-y-2.5">
              <div className="flex items-center justify-between text-xs font-medium">
                <span>Progress Pengisian</span>
                <span className="font-bold text-primary">{completionPercentage}%</span>
              </div>
              <div className="h-1.5 w-full rounded-full bg-muted overflow-hidden">
                <div
                  className="h-full bg-primary transition-all duration-300"
                  style={{ width: `${completionPercentage}%` }}
                />
              </div>
              <div className="flex items-center justify-between pt-1 text-xs text-muted-foreground">
                <span className="flex items-center gap-1">
                  {selectedBrand ? (
                    <CheckCircle2 className="h-3.5 w-3.5 text-emerald-600" />
                  ) : (
                    <Circle className="h-3.5 w-3.5" />
                  )}
                  Brand
                </span>
                <span className="flex items-center gap-1">
                  {selectedPaket ? (
                    <CheckCircle2 className="h-3.5 w-3.5 text-emerald-600" />
                  ) : (
                    <Circle className="h-3.5 w-3.5" />
                  )}
                  Paket
                </span>
                <span className="flex items-center gap-1">
                  {persentaseDiskon > 0 ? (
                    <CheckCircle2 className="h-3.5 w-3.5 text-emerald-600" />
                  ) : (
                    <Circle className="h-3.5 w-3.5" />
                  )}
                  Diskon
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Hasil Simulasi */}
        <Card className="lg:col-span-7">
          <CardHeader className="pb-4 border-b flex flex-row items-center justify-between">
            <CardTitle className="text-base flex items-center gap-2">
              <Calculator className="h-4 w-4 text-emerald-600" />
              Hasil Simulasi
            </CardTitle>
            {calculating && (
              <span className="text-xs text-muted-foreground flex items-center gap-1.5">
                <RefreshCw className="h-3.5 w-3.5 animate-spin" />
                Menghitung...
              </span>
            )}
          </CardHeader>
          <CardContent className="pt-6">
            {!calculationResult ? (
              <div className="py-10 text-center space-y-4">
                <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-muted">
                  <Percent className="h-8 w-8 text-muted-foreground" />
                </div>
                <div>
                  <h3 className="text-base font-semibold">Siap Menghitung Diskon!</h3>
                  <p className="text-sm text-muted-foreground mt-1">
                    Lengkapi form di sebelah kiri untuk melihat simulasi harga dengan diskon
                  </p>
                </div>
                <div className="mx-auto max-w-xs space-y-2 text-left text-xs pt-2">
                  {[
                    { done: Boolean(selectedBrand), label: 'Pilih Brand' },
                    { done: Boolean(selectedPaket), label: 'Pilih Paket Layanan' },
                    { done: persentaseDiskon > 0, label: 'Masukkan Persentase Diskon' },
                  ].map((item) => (
                    <div key={item.label} className="flex items-center gap-2">
                      {item.done ? (
                        <CheckCircle2 className="h-4 w-4 text-emerald-600" />
                      ) : (
                        <Circle className="h-4 w-4 text-muted-foreground" />
                      )}
                      <span className={item.done ? 'font-medium text-foreground' : 'text-muted-foreground'}>
                        {item.label}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            ) : (
              <div className="space-y-5">
                <div>
                  <Badge className="bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 mb-2">
                    <CheckCircle2 className="mr-1 h-3.5 w-3.5" />
                    Kalkulasi Berhasil
                  </Badge>
                  <p className="text-xs text-muted-foreground">{calculationResult.nama_brand}</p>
                  <p className="text-lg font-bold">{calculationResult.nama_paket}</p>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div className="rounded-lg border p-3.5 flex items-center gap-3">
                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-sky-50 text-sky-600 dark:bg-sky-950/40">
                      <Package className="h-4 w-4" />
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground">Harga Paket</p>
                      <p className="text-sm font-bold">{formatCurrency(calculationResult.harga_paket)}</p>
                    </div>
                  </div>
                  <div className="rounded-lg border p-3.5 flex items-center gap-3">
                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-amber-50 text-amber-600 dark:bg-amber-950/40">
                      <Percent className="h-4 w-4" />
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground">Pajak ({calculationResult.pajak_persen}%)</p>
                      <p className="text-sm font-bold">{formatCurrency(calculationResult.pajak_amount)}</p>
                    </div>
                  </div>
                  <div className="rounded-lg border p-3.5 flex items-center gap-3">
                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-blue-50 text-blue-600 dark:bg-blue-950/40">
                      <Wallet className="h-4 w-4" />
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground">Subtotal</p>
                      <p className="text-sm font-bold">{formatCurrency(calculationResult.subtotal_sebelum_diskon)}</p>
                    </div>
                  </div>
                  <div className="rounded-lg border border-rose-200 bg-rose-50/40 dark:border-rose-900 dark:bg-rose-950/20 p-3.5 flex items-center gap-3">
                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-rose-100 text-rose-600 dark:bg-rose-950">
                      <Tag className="h-4 w-4" />
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground">Diskon ({calculationResult.persentase_diskon}%)</p>
                      <p className="text-sm font-bold text-rose-600 dark:text-rose-400">
                        -{formatCurrency(calculationResult.diskon_amount)}
                      </p>
                    </div>
                  </div>
                </div>

                <div className="rounded-xl border border-emerald-200 bg-emerald-50/50 dark:border-emerald-900 dark:bg-emerald-950/20 p-4 space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2.5">
                      <CheckCircle2 className="h-5 w-5 text-emerald-600" />
                      <span className="text-sm font-medium">Harga Final</span>
                    </div>
                    <span className="text-2xl font-extrabold text-emerald-600 dark:text-emerald-400">
                      {formatCurrency(calculationResult.harga_final)}
                    </span>
                  </div>
                  <div className="flex items-center gap-3 rounded-lg bg-emerald-100/70 dark:bg-emerald-950/60 px-3.5 py-2.5">
                    <PiggyBank className="h-5 w-5 text-emerald-700 dark:text-emerald-400 shrink-0" />
                    <div>
                      <p className="text-xs text-emerald-800 dark:text-emerald-300">Anda hemat</p>
                      <p className="text-base font-bold text-emerald-700 dark:text-emerald-300">
                        {formatCurrency(calculationResult.diskon_amount)}
                      </p>
                    </div>
                  </div>
                </div>

                {/* Collapsible Detail Perhitungan */}
                <div className="rounded-lg border">
                  <button
                    type="button"
                    onClick={() => setShowDetail((v) => !v)}
                    className="w-full flex items-center justify-between px-4 py-3 text-sm font-medium hover:bg-muted/40 transition-colors"
                  >
                    <span className="flex items-center gap-2">
                      <Info className="h-4 w-4 text-sky-600" />
                      Detail Perhitungan
                    </span>
                    {showDetail ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
                  </button>
                  {showDetail && (
                    <div className="border-t px-4 py-3 bg-muted/30">
                      <pre className="whitespace-pre-wrap text-xs font-mono text-muted-foreground">
                        {calculationResult.detail_perhitungan}
                      </pre>
                    </div>
                  )}
                </div>

                <div className="flex flex-wrap gap-2">
                  <Button variant="outline" onClick={resetCalculation}>
                    <RefreshCw className="mr-2 h-4 w-4" />
                    Hitung Ulang
                  </Button>
                  <Button variant="secondary" onClick={() => void copyResult()}>
                    {copied ? <Check className="mr-2 h-4 w-4 text-emerald-600" /> : <Copy className="mr-2 h-4 w-4" />}
                    {copied ? 'Tersalin' : 'Salin Hasil'}
                  </Button>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
