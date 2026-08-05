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
  GroupRatio,
  ApiResponse,
  GetGroupRatiosParams,
  GetGroupRatiosResponse,
  GroupRatioFormData,
} from './types'

// ============================================================================
// Group Ratio Management
// ============================================================================

export async function getGroupRatios(
  params: GetGroupRatiosParams = {}
): Promise<GetGroupRatiosResponse> {
  const { p = 1, page_size = 10 } = params
  const res = await api.get(`/api/group-ratio/?p=${p}&page_size=${page_size}`)
  return res.data
}

export async function getGroupRatio(
  id: number
): Promise<ApiResponse<GroupRatio>> {
  const res = await api.get(`/api/group-ratio/${id}`)
  return res.data
}

export async function createGroupRatio(
  data: GroupRatioFormData
): Promise<ApiResponse<GroupRatio>> {
  const res = await api.post('/api/group-ratio/', data)
  return res.data
}

export async function updateGroupRatio(
  id: number,
  data: GroupRatioFormData
): Promise<ApiResponse<GroupRatio>> {
  const res = await api.put(`/api/group-ratio/${id}`, data)
  return res.data
}

export async function deleteGroupRatio(
  id: number
): Promise<ApiResponse> {
  const res = await api.delete(`/api/group-ratio/${id}`)
  return res.data
}
