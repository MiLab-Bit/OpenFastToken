/*
Copyright (C) 2023-2026 OpenFastToken

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

For commercial licensing, please contact support@example.com
*/
import { z } from 'zod'

// ============================================================================
// Enterprise Sub-User Schema & Types
// ============================================================================

export const enterpriseSubUserSchema = z.object({
  id: z.number(),
  username: z.string(),
  display_name: z.string(),
  email: z.string(),
  role: z.number(),
  group: z.string(),
})

export type EnterpriseSubUser = z.infer<typeof enterpriseSubUserSchema>

// ============================================================================
// Form Schemas
// ============================================================================

export const enterpriseSubUserFormSchema = z.object({
  username: z.string().min(1, '用户名不能为空').max(50, '用户名不能超过50个字符'),
  display_name: z.string().min(1, '显示名称不能为空').max(100, '显示名称不能超过100个字符'),
  email: z.string().email('请输入有效的邮箱地址').min(1, '邮箱不能为空'),
  password: z.string().min(8, '密码至少8个字符').max(100, '密码不能超过100个字符').optional().or(z.literal('')),
  role: z.number().min(0),
  group: z.string().max(50).optional().or(z.literal('')),
})

export type EnterpriseSubUserFormValues = z.infer<typeof enterpriseSubUserFormSchema>

// ============================================================================
// API Request/Response Types
// ============================================================================

export interface ApiResponse<T = unknown> {
  success: boolean
  message?: string
  data?: T
}

export interface GetSubUsersParams {
  enterpriseId: number
  p?: number
  page_size?: number
}

export interface GetSubUsersResponse {
  success: boolean
  message?: string
  data?: {
    items: EnterpriseSubUser[]
    total: number
    page: number
    page_size: number
  }
}

export interface SearchSubUsersParams {
  enterpriseId: number
  keyword?: string
  p?: number
  page_size?: number
}

export interface CreateSubUserPayload {
  username: string
  display_name: string
  email: string
  password?: string
  role: number
  group?: string
}

export interface UpdateSubUserPayload {
  id: number
  username: string
  display_name: string
  email: string
  password?: string
  role: number
  group?: string
}

// ============================================================================
// Dialog Types
// ============================================================================

export type EnterpriseSubUsersDialogType = 'create' | 'update' | 'delete' | 'view'
