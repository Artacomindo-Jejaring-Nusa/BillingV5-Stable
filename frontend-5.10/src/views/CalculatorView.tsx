import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Calculator,
  Calendar,
  CalendarPlus,
  Check,
  CheckCircle2,
  Circle,
  Copy,
  Layers,
  Package,
  Percent,
  RefreshCw,
  Wallet,
} from 'lucide-react'
import apiClient from '@/services/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'

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

interface ProrateResult {
  harga_dasar_prorate: number
  pajak: number
  total_harga_prorate: number
  periode_hari: number
  harga_bulan_depan?: number | null
  ppn_bulan_depan?: number | null
  total_bulan_depan_dengan_ppn?: number | null
  total_keseluruhan?: number | null
}

function formatCurrency(value?: number | null): string {
  if (typeof value !== 'number' || Number.isNaN(value)) return 'Rp 0'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
  }).format(value)
}

function todayIso(): string {
  return new Date().toISOString().split('T')[0]
}

export function CalculatorView() {
  const [brandList, setBrandList] = useState<Brand[]>([])
  const [paketList, setPaketList] = useState<Paket[]>([])
  const [selectedBrandId, setSelectedBrandId] = useState('')
  const [selectedPaketId, setSelectedPaketId] = useState('')
  const [startDate, setStartDate] = useState(todayIso)
  const [includeNextMonth, setIncludeNextMonth] = useState(false)
  const [calculationResult, setCalculationResult] = useState<ProrateResult | null>(null)
  const [calculating, setCalculating] = useState(false)
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
    if (startDate) completed += 33.34
    return Math.round(completed)
  }, [selectedBrand, selectedPaket, startDate])

  useEffect(() => {
    if (!selectedBrand || !selectedPaket || !startDate) {
      setCalculationResult(null)
      return
    }
    const timer = window.setTimeout(async () => {
      setCalculating(true)
      try {
        const res = await apiClient.post<ProrateResult>('/calculator/prorate', {
          id_brand: selectedBrand.id_brand,
          paket_layanan_id: selectedPaket.id,
          tgl_mulai: startDate,
          include_ppn_next_month: includeNextMonth,
        })
        setCalculationResult(res.data)
      } catch (err) {
        console.error('Gagal melakukan kalkulasi:', err)
        setCalculationResult(null)
      } finally {
        setCalculating(false)
      }
    }, 500)
    return () => window.clearTimeout(timer)
  }, [selectedBrand, selectedPaket, startDate, includeNextMonth])

  const handleBrandChange = (val: string) => {
    setSelectedBrandId(val)
    setSelectedPaketId('')
    setCalculationResult(null)
  }

  const resetCalculation = () => {
    setSelectedBrandId('')
    setSelectedPaketId('')
    setStartDate(todayIso())
    setIncludeNextMonth(false)
    setCalculationResult(null)
  }

  const copyResult = async () => {
    if (!calculationResult) return
    const baseText = [
      'Hasil Kalkulasi Harga Prorate:',
      `Brand: ${selectedBrand?.brand || '-'}`,
      `Paket: ${selectedPaket?.nama_paket || '-'}`,
      `Periode: ${calculationResult.periode_hari} hari`,
      `Harga Dasar: ${formatCurrency(calculationResult.harga_dasar_prorate)}`,
      `Pajak: ${formatCurrency(calculationResult.pajak)}`,
      `Total: ${formatCurrency(calculationResult.total_harga_prorate)}`,
    ].join('\n')

    const nextMonthText =
      includeNextMonth && calculationResult.total_bulan_depan_dengan_ppn != null
        ? [
            '',
            'Harga Bulan Depan (Dengan PPN):',
            `- Harga Dasar: ${formatCurrency(calculationResult.harga_bulan_depan)}`,
            `- PPN (${selectedBrand?.pajak || 0}%): ${formatCurrency(calculationResult.ppn_bulan_depan)}`,
            `- Total Bulan Depan: ${formatCurrency(calculationResult.total_bulan_depan_dengan_ppn)}`,
            `Total Keseluruhan: ${formatCurrency(calculationResult.total_keseluruhan)}`,
          ].join('\n')
        : ''

    try {
      await navigator.clipboard.writeText(`${baseText}${nextMonthText}`.trim())
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
          <Calculator className="h-5 w-5" />
        </div>
        <div>
          <h2 className="text-2xl font-bold tracking-tight">Simulasi Harga</h2>
          <p className="text-sm text-muted-foreground">Simulasi perhitungan harga prorate untuk layanan FTTH</p>
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

            <div className="space-y-1.5">
              <Label htmlFor="simulasi-start-date" className="flex items-center gap-1.5">
                {startDate ? (
                  <CheckCircle2 className="h-4 w-4 text-emerald-600" />
                ) : (
                  <Circle className="h-4 w-4 text-muted-foreground" />
                )}
                Tanggal Mulai Layanan
              </Label>
              <Input
                id="simulasi-start-date"
                type="date"
                value={startDate}
                disabled={!selectedPaket}
                onChange={(e) => setStartDate(e.target.value)}
              />
            </div>

            <div className="space-y-2">
              <Label className="flex items-center gap-1.5 text-xs text-muted-foreground">
                <CalendarPlus className="h-3.5 w-3.5" />
                Opsi Harga
              </Label>
              <div className="flex items-center justify-between rounded-lg border px-3.5 py-2.5">
                <span className="text-sm">Termasuk harga bulan depan (dengan PPN)</span>
                <Switch checked={includeNextMonth} onCheckedChange={setIncludeNextMonth} />
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
                  {startDate ? (
                    <CheckCircle2 className="h-3.5 w-3.5 text-emerald-600" />
                  ) : (
                    <Circle className="h-3.5 w-3.5" />
                  )}
                  Tanggal
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
                  <Calculator className="h-8 w-8 text-muted-foreground" />
                </div>
                <div>
                  <h3 className="text-base font-semibold">Siap Menghitung!</h3>
                  <p className="text-sm text-muted-foreground mt-1">
                    Lengkapi form di sebelah kiri untuk melihat simulasi harga prorate
                  </p>
                </div>
                <div className="mx-auto max-w-xs space-y-2 text-left text-xs pt-2">
                  {[
                    { done: Boolean(selectedBrand), label: 'Pilih Brand' },
                    { done: Boolean(selectedPaket), label: 'Pilih Paket Layanan' },
                    { done: Boolean(startDate), label: 'Pilih Tanggal Mulai' },
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
                  <p className="text-xs text-muted-foreground">{selectedBrand?.brand}</p>
                  <p className="text-lg font-bold">{selectedPaket?.nama_paket}</p>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                  <div className="rounded-lg border p-3.5 flex items-center gap-3">
                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-sky-50 text-sky-600 dark:bg-sky-950/40">
                      <Calendar className="h-4 w-4" />
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground">Periode Tagihan</p>
                      <p className="text-sm font-bold">{calculationResult.periode_hari} hari</p>
                    </div>
                  </div>
                  <div className="rounded-lg border p-3.5 flex items-center gap-3">
                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-blue-50 text-blue-600 dark:bg-blue-950/40">
                      <Wallet className="h-4 w-4" />
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground">Harga Dasar (Prorate)</p>
                      <p className="text-sm font-bold">{formatCurrency(calculationResult.harga_dasar_prorate)}</p>
                    </div>
                  </div>
                  <div className="rounded-lg border p-3.5 flex items-center gap-3">
                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-amber-50 text-amber-600 dark:bg-amber-950/40">
                      <Percent className="h-4 w-4" />
                    </div>
                    <div>
                      <p className="text-xs text-muted-foreground">Pajak ({selectedBrand?.pajak || 0}%)</p>
                      <p className="text-sm font-bold">{formatCurrency(calculationResult.pajak)}</p>
                    </div>
                  </div>
                </div>

                <div className="rounded-xl border border-emerald-200 bg-emerald-50/50 dark:border-emerald-900 dark:bg-emerald-950/20 p-4 space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2.5">
                      <Package className="h-5 w-5 text-emerald-600" />
                      <span className="text-sm font-medium">Total Tagihan</span>
                    </div>
                    <span className="text-2xl font-extrabold text-emerald-600 dark:text-emerald-400">
                      {formatCurrency(calculationResult.total_harga_prorate)}
                    </span>
                  </div>

                  {includeNextMonth && calculationResult.total_bulan_depan_dengan_ppn != null && (
                    <div className="border-t pt-3 space-y-2 text-xs">
                      <div className="flex items-center justify-between">
                        <div>
                          <p className="text-muted-foreground">Harga Bulan Depan + PPN</p>
                          <p className="text-base font-bold text-foreground">
                            {formatCurrency(calculationResult.total_bulan_depan_dengan_ppn)}
                          </p>
                        </div>
                        <Badge variant="secondary">Dengan PPN</Badge>
                      </div>
                      {calculationResult.ppn_bulan_depan != null && (
                        <div className="space-y-1 text-muted-foreground">
                          <div className="flex justify-between">
                            <span>Harga Dasar Bulan Depan:</span>
                            <span className="text-foreground">{formatCurrency(calculationResult.harga_bulan_depan)}</span>
                          </div>
                          <div className="flex justify-between">
                            <span>PPN ({selectedBrand?.pajak || 0}%):</span>
                            <span className="text-emerald-600 font-medium">
                              {formatCurrency(calculationResult.ppn_bulan_depan)}
                            </span>
                          </div>
                        </div>
                      )}
                      <div className="flex items-center justify-between border-t pt-2 text-sm font-bold">
                        <span>Total Keseluruhan</span>
                        <span className="text-base">{formatCurrency(calculationResult.total_keseluruhan)}</span>
                      </div>
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
