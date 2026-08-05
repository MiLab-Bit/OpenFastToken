/*
Copyright (C) 2023-2026 FastToken

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact hello@fasttoken.example.com
*/
import { useCallback, useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { api } from '@/lib/api'
import { toast } from 'sonner'

export type GiftStatus = 'active' | 'used' | 'expired'

export interface AdminUserGift {
  id: number
  user_id: number
  user_email: string
  username: string
  gift_type: string
  gift_name: string
  trade_no: string
  status: GiftStatus
  description: string
  create_time: number
  update_time: number
}

export type GiftStatusFilter = 'all' | GiftStatus

interface AdminGiftsResponse {
  success: boolean
  gifts: AdminUserGift[]
  total: number
  page: number
}

const PAGE_SIZE = 20

export function useGiftsAdmin() {
  const { t } = useTranslation()
  const [gifts, setGifts] = useState<AdminUserGift[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [statusFilter, setStatusFilter] = useState<GiftStatusFilter>('all')
  const [loading, setLoading] = useState(true)
  const [updatingId, setUpdatingId] = useState<number | null>(null)

  const fetchGifts = useCallback(async () => {
    try {
      setLoading(true)
      const params: Record<string, unknown> = {
        page,
        page_size: PAGE_SIZE,
      }
      if (statusFilter !== 'all') {
        params.status = statusFilter
      }
      const res = await api.get<AdminGiftsResponse>('/api/admin/gifts', {
        params,
      })
      if (res.data?.success) {
        setGifts(res.data.gifts ?? [])
        setTotal(res.data.total ?? 0)
      }
    } catch {
      // error handled by interceptor
    } finally {
      setLoading(false)
    }
  }, [page, statusFilter])

  useEffect(() => {
    fetchGifts()
  }, [fetchGifts])

  const updateStatus = useCallback(
    async (id: number, status: GiftStatus) => {
      try {
        setUpdatingId(id)
        const res = await api.post(`/api/admin/gifts/${id}/status`, { status })
        if (res.data?.success) {
          toast.success(t('Status updated'))
          await fetchGifts()
          return true
        }
        toast.error(t('Failed to update status'))
        return false
      } catch {
        toast.error(t('Failed to update status'))
        return false
      } finally {
        setUpdatingId(null)
      }
    },
    [fetchGifts]
  )

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE))

  return {
    gifts,
    total,
    page,
    totalPages,
    statusFilter,
    loading,
    updatingId,
    setPage,
    setStatusFilter,
    updateStatus,
  }
}
