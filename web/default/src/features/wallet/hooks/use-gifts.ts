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
import { api } from '@/lib/api'

export type UserGiftStatus = 'active' | 'used' | 'expired'

export interface UserGift {
  id: number
  user_id: number
  gift_type: string
  gift_name: string
  trade_no: string
  status: UserGiftStatus
  description: string
  create_time: number
  update_time: number
}

interface UserGiftsResponse {
  success: boolean
  data: UserGift[]
}

export function useGifts() {
  const [gifts, setGifts] = useState<UserGift[]>([])
  const [loading, setLoading] = useState(true)

  const fetchGifts = useCallback(async () => {
    try {
      setLoading(true)
      const res = await api.get<UserGiftsResponse>('/api/user/gifts')
      if (res.data?.success) {
        setGifts(res.data.data ?? [])
      }
    } catch {
      // error handled by interceptor
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchGifts()
  }, [fetchGifts])

  return { gifts, loading, refetch: fetchGifts }
}
