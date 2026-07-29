import { useState, useCallback, useEffect } from 'react'
import { getMembershipInfo } from '../api'
import type { MembershipInfo } from '../types'

export function useMembership() {
  const [membershipInfo, setMembershipInfo] = useState<MembershipInfo | null>(
    null
  )
  const [loading, setLoading] = useState(true)

  const refresh = useCallback(async () => {
    try {
      setLoading(true)
      const res = await getMembershipInfo()
      if (res.success && res.data) {
        setMembershipInfo(res.data)
      }
    } catch (error) {
      console.error('Failed to fetch membership info:', error)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    refresh()
  }, [refresh])

  return { membershipInfo, loading, refresh }
}