import { useCallback, useEffect, useState } from 'react'
import { getAdminTopUps } from '../api'
import type { AdminTopUp } from '../types'

const PAGE_SIZE = 20

export function useRechargeAdmin() {
  const [topups, setTopups] = useState<AdminTopUp[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [userId, setUserId] = useState('')
  const [keyword, setKeyword] = useState('')
  const [loading, setLoading] = useState(true)

  const fetchTopUps = useCallback(async () => {
    try {
      setLoading(true)
      const res = await getAdminTopUps({
        page,
        page_size: PAGE_SIZE,
        user_id: userId || undefined,
        keyword: keyword || undefined,
      })
      if (res?.success) {
        setTopups(res.data.items ?? [])
        setTotal(res.data.total ?? 0)
      }
    } catch {
      // error toast handled by axios interceptor
    } finally {
      setLoading(false)
    }
  }, [page, userId, keyword])

  useEffect(() => {
    fetchTopUps()
  }, [fetchTopUps])

  // 应用筛选条件：回到第 1 页并触发查询
  const applyFilter = useCallback((uid: string, kw: string) => {
    setUserId(uid)
    setKeyword(kw)
    setPage(1)
  }, [])

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  return {
    topups,
    total,
    page,
    totalPages,
    loading,
    setPage,
    applyFilter,
  }
}
