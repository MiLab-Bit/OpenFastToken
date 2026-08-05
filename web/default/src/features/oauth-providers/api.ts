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
import type {
  OAuthProvider,
  ApiResponse,
  GetOAuthProvidersParams,
  GetOAuthProvidersResponse,
  OAuthProviderFormData,
} from './types'

// ============================================================================
// OAuth Provider Management
// ============================================================================

export async function getOAuthProviders(
  params: GetOAuthProvidersParams = {}
): Promise<GetOAuthProvidersResponse> {
  const { p = 1, page_size = 10 } = params
  const res = await api.get(
    `/api/custom-oauth-provider/?p=${p}&page_size=${page_size}`
  )
  return res.data
}

export async function getOAuthProvider(
  id: number
): Promise<ApiResponse<OAuthProvider>> {
  const res = await api.get(`/api/custom-oauth-provider/${id}`)
  return res.data
}

export async function createOAuthProvider(
  data: OAuthProviderFormData
): Promise<ApiResponse<OAuthProvider>> {
  const res = await api.post('/api/custom-oauth-provider/', data)
  return res.data
}

export async function updateOAuthProvider(
  id: number,
  data: OAuthProviderFormData
): Promise<ApiResponse<OAuthProvider>> {
  const res = await api.put(`/api/custom-oauth-provider/${id}`, data)
  return res.data
}

export async function deleteOAuthProvider(
  id: number
): Promise<ApiResponse> {
  const res = await api.delete(`/api/custom-oauth-provider/${id}`)
  return res.data
}
