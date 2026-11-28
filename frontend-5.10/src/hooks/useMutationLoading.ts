import { useEffect, useState } from 'react'
import { subscribeMutationLoading } from '@/services/api'

export function useMutationLoading(): boolean {
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    return subscribeMutationLoading(setIsLoading)
  }, [])

  return isLoading
}
