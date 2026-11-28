import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'

interface PlaceholderPageProps {
  title: string
  taskNumber: number
}

export function PlaceholderPage({ title, taskNumber }: PlaceholderPageProps) {
  return (
    <div className="p-6">
      <Card className="max-w-xl">
        <CardHeader>
          <CardTitle>{title}</CardTitle>
          <CardDescription>
            Halaman ini dijadwalkan untuk migrasi pada Task {taskNumber}.
          </CardDescription>
        </CardHeader>
        <CardContent className="text-sm text-muted-foreground">
          Struktur routing dan proteksi akses sudah aktif. Konten penuh akan dimigrasi sesuai urutan task.
        </CardContent>
      </Card>
    </div>
  )
}
