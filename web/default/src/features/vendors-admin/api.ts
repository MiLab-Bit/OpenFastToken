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
  Vendor,
  ApiResponse,
  GetVendorsParams,
  GetVendorsResponse,
  VendorFormData,
} from './types'

// ============================================================================
// Vendor Management
// ============================================================================

export async function getVendors(
  params: GetVendorsParams = {}
): Promise<GetVendorsResponse> {
  const { p = 1, page_size = 10, keyword = '' } = params

  if (keyword) {
    const res = await api.get(
      `/api/vendors/search?keyword=${encodeURIComponent(keyword)}&p=${p}&page_size=${page_size}`
    )
    return res.data
  }

  const res = await api.get(`/api/vendors/?p=${p}&page_size=${page_size}`)
  return res.data
}

export async function getVendor(
  id: number
): Promise<ApiResponse<Vendor>> {
  const res = await api.get(`/api/vendors/${id}`)
  return res.data
}

export async function createVendor(
  data: VendorFormData
): Promise<ApiResponse<Vendor>> {
  const res = await api.post('/api/vendors/', data)
  return res.data
}

export async function updateVendor(
  data: VendorFormData & { id: number }
): Promise<ApiResponse<Vendor>> {
  const res = await api.put('/api/vendors/', data)
  return res.data
}

export async function deleteVendor(id: number): Promise<ApiResponse> {
  const res = await api.delete(`/api/vendors/${id}`)
  return res.data
}
