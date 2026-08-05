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
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact hello@fasttoken.example.com
*/

// ============================================================================
// Tenant Console API Types
//
// Mirrors the backend payload of `controller/tenant_self.go`:
//   GET /api/user/tenant/info
//   GET /api/user/tenant/members
// The server never accepts a client-supplied enterprise id — the tenant is
// always resolved from the authenticated session context.
// ============================================================================

export interface ApiResponse<T = unknown> {
  success: boolean
  message?: string
  data?: T
}

/** Public profile of the enterprise (tenant) the current user belongs to. */
export interface TenantInfo {
  /** `false` when the current user is a personal (non-enterprise) account. */
  joined: boolean
  enterprise_id: number
  name?: string
  /** `silver` | `gold` | `platinum` */
  membership_level?: string
  /** `pending` | `approved` | `rejected` */
  status?: string
  /** Billing discount rate derived from the membership level, e.g. `0.7`. */
  discount_rate?: number
  contact_name?: string
  created_at?: number
  approved_at?: number
  total_members?: number
  active_members?: number
  admin_count?: number
}

/** Redacted member record — no email / phone / quota is exposed. */
export interface TenantMember {
  id: number
  username: string
  display_name: string
  /** `member` | `admin` */
  role: string
  /** `active` | `inactive` */
  status: string
}

export interface TenantMembersData {
  joined: boolean
  items: TenantMember[]
  total: number
  page: number
  page_size: number
}

export type TenantInfoResponse = ApiResponse<TenantInfo>

export type TenantMembersResponse = ApiResponse<TenantMembersData>
