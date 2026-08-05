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
import { api } from '@/lib/api'
import type { TenantInfoResponse, TenantMembersResponse } from './types'

// ============================================================================
// Tenant Self-Service (member facing)
//
// No enterprise id is ever sent from the client: the backend resolves the
// tenant from the authenticated session, so these endpoints are inherently
// safe against cross-tenant enumeration.
// ============================================================================

/** Fetch the public profile of the tenant the current user belongs to. */
export async function getTenantInfo(): Promise<TenantInfoResponse> {
  const res = await api.get('/api/user/tenant/info')
  return res.data
}

/** Fetch a paginated, redacted list of members in the current tenant. */
export async function getTenantMembers(
  p = 1,
  pageSize = 20
): Promise<TenantMembersResponse> {
  const res = await api.get(
    `/api/user/tenant/members?p=${p}&page_size=${pageSize}`
  )
  return res.data
}
