/*
Copyright (C) 2023-2026 FastToken
*/
// ============================================================================
// Membership Type Definitions
// ============================================================================

/** Generic API response */
export interface ApiResponse<T = unknown> {
  success: boolean
  message?: string
  data?: T
}

/** Membership levels */
export type MembershipLevel = 'silver' | 'gold' | 'platinum'

/** Membership info for current user */
export interface MembershipInfo {
  membership_level: MembershipLevel
  discount_rate: number
  is_active: boolean
  expire_time: number
}

/** Membership level config (for UI display) */
export interface MembershipLevelConfig {
  key: MembershipLevel
  label: string
  labelZh: string
  discountRate: number
  discountLabel: string
  color: string
  bgColor: string
  icon: string
  description: string
}

/** Invitation code record */
export interface InvitationCode {
  id: number
  code: string
  type: MembershipLevel
  used_by: number
  enterprise_id: number
  expires_at: number
  remark: string
  created_by: number
  created_at: number
}

/** Invitation code create request */
export interface InvitationCodeCreateRequest {
  type: MembershipLevel
  count: number
  expires_in?: number
  remark?: string
}

/** Invitation code list response */
export interface InvitationCodeListResponse {
  codes: InvitationCode[]
  total: number
  page: number
}

/** Invitation code stats */
export interface InvitationCodeStats {
  total: number
  used: number
  unused: number
  expired: number
}

/** Use invitation code request */
export interface UseInvitationCodeRequest {
  code: string
}

/** Enterprise record */
export interface Enterprise {
  id: number
  name: string
  credit_code: string
  contact_name: string
  contact_phone: string
  contact_email: string
  status: 'pending' | 'approved' | 'rejected'
  membership_level: MembershipLevel
  approved_at: number
  approved_by: number
  reject_reason: string
  created_at: number
  updated_at: number
}

/** Enterprise create request */
export interface EnterpriseCreateRequest {
  name: string
  credit_code: string
  contact_name?: string
  contact_phone?: string
  contact_email?: string
  remark?: string
}

/** Enterprise list response */
export interface EnterpriseListResponse {
  enterprises: Enterprise[]
  total: number
  page: number
}