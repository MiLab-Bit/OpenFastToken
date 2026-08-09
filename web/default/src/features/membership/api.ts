/*
Copyright (C) 2023-2026 智企惠
*/
import { api } from '@/lib/api'
import type {
  ApiResponse,
  MembershipInfo,
  EnterpriseCreateRequest,
  EnterpriseListResponse,
  Enterprise,
} from './types'

// ============================================================================
// Membership APIs (用户)
// ============================================================================

/** 获取当前用户会员信息 */
export async function getMembershipInfo(): Promise<ApiResponse<MembershipInfo>> {
  const res = await api.get('/api/user/membership')
  return res.data
}


// ============================================================================
// Enterprise APIs (管理员)
// ============================================================================

/** 创建企业 */
export async function createEnterprise(
  request: EnterpriseCreateRequest
): Promise<ApiResponse<Enterprise>> {
  const res = await api.post('/api/enterprise', request)
  return res.data
}

/** 获取企业列表 */
export async function listEnterprises(
  params?: { status?: string; page?: number; page_size?: number }
): Promise<ApiResponse<EnterpriseListResponse>> {
  const query = new URLSearchParams()
  if (params?.status) query.set('status', params.status)
  if (params?.page) query.set('page', params.page.toString())
  if (params?.page_size) query.set('page_size', params.page_size.toString())
  const res = await api.get(`/api/enterprise?${query}`)
  return res.data
}

/** 审核通过企业 */
export async function approveEnterprise(id: number): Promise<ApiResponse> {
  const res = await api.post(`/api/enterprise/${id}/approve`)
  return res.data
}

/** 审核拒绝企业 */
export async function rejectEnterprise(
  id: number,
  reject_reason: string
): Promise<ApiResponse> {
  const res = await api.post(`/api/enterprise/${id}/reject`, {
    reject_reason,
  })
  return res.data
}