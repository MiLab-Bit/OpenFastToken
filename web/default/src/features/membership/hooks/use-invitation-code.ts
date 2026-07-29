import { useCallback } from 'react'
import { useInvitationCode as apiUseInvitationCode } from '../api'
import type { ApiResponse } from '../types'

export function useInvitationCode() {
  const using = false // placeholder

  const useCode = useCallback(
    async (code: string): Promise<ApiResponse<{ membership_level: string; message: string }>> => {
      try {
        const res = await apiUseInvitationCode({ code })
        return res
      } catch (error) {
        console.error('Failed to use invitation code:', error)
        return { success: false, message: 'Network error' }
      }
    },
    []
  )

  return { using, useCode }
}