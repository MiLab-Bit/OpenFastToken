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
import type { ApiResponse, EnterpriseRegisterPayload } from './types'

// ============================================================================
// Enterprise Registration
// ============================================================================

export async function registerEnterprise(
  data: EnterpriseRegisterPayload
): Promise<ApiResponse> {
  const res = await api.post('/api/enterprise', data)
  return res.data
}
