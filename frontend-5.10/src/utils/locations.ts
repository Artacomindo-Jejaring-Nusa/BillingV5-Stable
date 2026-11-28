import apiClient from '@/services/api'

export interface LocationItem {
  name: string
  is_active: boolean
}

export const DEFAULT_LOCATIONS: string[] = [
  'Tambun',
  'Rusun Pinus Elok',
  'Luar Pinus Elok',
  'Rusun Pulogebang',
  'Rusun Cakung KM2',
  'Rusun Tipar Cakung',
  'Rusun Albo',
  'Rusun Nagrak',
  'Waringin',
  'Parama',
  'Casanova',
]

export const LOCATIONS_SETTING_KEY = 'pelanggan_locations'

export function parseLocationsSetting(raw?: string | null): LocationItem[] {
  if (!raw || !raw.trim()) {
    return DEFAULT_LOCATIONS.map((name) => ({ name, is_active: true }))
  }
  try {
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed) || parsed.length === 0) {
      return DEFAULT_LOCATIONS.map((name) => ({ name, is_active: true }))
    }
    const items: LocationItem[] = parsed
      .map((entry) => {
        if (typeof entry === 'string') {
          return { name: entry.trim(), is_active: true }
        }
        if (entry && typeof entry.name === 'string') {
          return { name: entry.name.trim(), is_active: entry.is_active !== false }
        }
        return null
      })
      .filter((item): item is LocationItem => Boolean(item && item.name))

    return items.length > 0
      ? items
      : DEFAULT_LOCATIONS.map((name) => ({ name, is_active: true }))
  } catch {
    return DEFAULT_LOCATIONS.map((name) => ({ name, is_active: true }))
  }
}

export async function fetchAllLocations(): Promise<LocationItem[]> {
  try {
    const res = await apiClient.get(`/system/settings/${LOCATIONS_SETTING_KEY}`, {
      headers: { 'X-Silent': 'true' },
    })
    return parseLocationsSetting(res.data?.data?.value)
  } catch {
    return DEFAULT_LOCATIONS.map((name) => ({ name, is_active: true }))
  }
}

export async function fetchActiveLocations(): Promise<string[]> {
  const all = await fetchAllLocations()
  const active = all.filter((l) => l.is_active).map((l) => l.name)
  return active.length > 0 ? active : DEFAULT_LOCATIONS
}
