/*
Copyright (C) 2023-2026 OpenFastToken
*/
import { api } from '@/lib/api'
import type {
  ApiResponse,
  MembershipInfo,
  InvitationCode,
  InvitationCodeCreateRequest,
  InvitationCodeListResponse,
  InvitationCodeStats,
  EnterpriseCreateRequest,
  EnterpriseListResponse,
  Enterprise,
  UseInvitationCodeRequest,
} from './types'

// ============================================================================
// Membership APIs (用户)
// ============================================================================

/** 获取当前用户会员信息 */
export async function getMembershipInfo(): Promise<ApiResponse<MembershipInfo>> {
  const res = await api.get('/api/user/membership')
  return res.data
}

/** 使用邀请码升级会员 */
export async function useInvitationCode(
  request: UseInvitationCodeRequest
): Promise<ApiResponse<{ membership_level: string; message: string }>> {
  const res = await api.post('/api/user/use_invitation_code', request)
  return res.data
}

// ============================================================================
// Invitation Code APIs (管理员)
// ============================================================================

/** 创建邀请码 */
export async function createInvitationCodes(
  request: InvitationCodeCreateRequest
): Promise<ApiResponse<InvitationCode[]>> {
  const res = await api.post('/api/user/invitation_code', request)
  return res.data
}

/** 获取邀请码列表 */
export async function listInvitationCodes(
  params?: { type?: string; used?: string; page?: number; page_size?: number }
): Promise<ApiResponse<InvitationCodeListResponse>> {
  const query = new URLSearchParams()
  if (params?.type) query.set('type', params.type)
  if (params?.used) query.set('used', params.used)
  if (params?.page) query.set('page', params.page.toString())
  if (params?.page_size) query.set('page_size', params.page_size.toString())
  const res = await api.get(`/api/user/invitation_codes?${query}`)
  return res.data
}

/** 获取邀请码统计 */
export async function getInvitationCodeStats(): Promise<
  ApiResponse<InvitationCodeStats>
> {
  const res = await api.get('/api/user/invitation_code/stats')
  return res.data
}

/** 删除邀请码 */
export async function deleteInvitationCode(id: number): Promise<ApiResponse> {
  const res = await api.delete(`/api/user/invitation_code/${id}`)
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