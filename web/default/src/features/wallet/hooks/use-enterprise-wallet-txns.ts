/*
Copyright (C) 2023-2026 智企惠

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact abovetigers@qq.com
*/
import { useState, useEffect, useCallback } from 'react'
import i18next from 'i18next'
import { toast } from 'sonner'
import { getTenantWalletTxns, isApiSuccess } from '../api'
import type { EnterpriseWalletTxn } from '../types'

// ============================================================================
// Enterprise Wallet Txns Hook
//
// Mirrors useBillingHistory but for the admin-only enterprise main-wallet
// fund flow. No keyword search / order-completion in Phase 1; adds an
// explicit `error` + `retry` surface and only fetches when `enabled`
// (the dialog passes `open`) so we don't hit the API while hidden.
// ============================================================================

interface UseEnterpriseWalletTxnsOptions {
  /** Initial page number */
  initialPage?: number
  /** Initial page size */
  initialPageSize?: number
  /** When false, no request is sent (e.g. dialog closed) */
  enabled?: boolean
}

export function useEnterpriseWalletTxns(
  options: UseEnterpriseWalletTxnsOptions = {}
) {
  const { initialPage = 1, initialPageSize = 20, enabled = true } = options
  const [items, setItems] = useState<EnterpriseWalletTxn[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(initialPage)
  const [pageSize, setPageSize] = useState(initialPageSize)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(false)

  const fetchTxns = useCallback(async () => {
    if (!enabled) return
    setLoading(true)
    setError(false)
    try {
      const response = await getTenantWalletTxns(page, pageSize)
      if (isApiSuccess(response) && response.data) {
        setItems(response.data.items || [])
        setTotal(response.data.total || 0)
      } else {
        setError(true)
        toast.error(response.message || i18next.t('Failed to load fund flow'))
        setItems([])
        setTotal(0)
      }
    } catch (err) {
      console.error('Failed to fetch enterprise wallet txns:', err)
      setError(true)
      toast.error(i18next.t('Failed to load fund flow'))
      setItems([])
      setTotal(0)
    } finally {
      setLoading(false)
    }
  }, [enabled, page, pageSize])

  useEffect(() => {
    if (enabled) {
      void fetchTxns()
    }
  }, [fetchTxns, enabled])

  const handlePageChange = useCallback((newPage: number) => {
    setPage(newPage)
  }, [])

  const handlePageSizeChange = useCallback((newPageSize: number) => {
    setPageSize(newPageSize)
    setPage(1) // reset to first page when changing page size
  }, [])

  return {
    items,
    total,
    page,
    pageSize,
    loading,
    error,
    handlePageChange,
    handlePageSizeChange,
    retry: fetchTxns,
  }
}
