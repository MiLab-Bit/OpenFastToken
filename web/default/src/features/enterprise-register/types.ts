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
// Enterprise Registration Form Schema
// ============================================================================

export const enterpriseRegisterFormSchema = z.object({
  name: z.string().min(1, '企业名称不能为空').max(200, '企业名称不能超过200个字符'),
  credit_code: z.string().max(50, '统一社会信用代码不能超过50个字符').optional().or(z.literal('')),
  contact_name: z.string().max(100, '联系人姓名不能超过100个字符').optional().or(z.literal('')),
  contact_phone: z.string().max(30, '联系电话不能超过30个字符').optional().or(z.literal('')),
  contact_email: z.string().email('请输入有效的邮箱地址').max(100, '联系人邮箱不能超过100个字符').optional().or(z.literal('')),
  business_license: z.string().optional(), // 营业执照文件 URL
  invitation_code: z.string().min(1, '请输入企业认证邀请码'),
})

export type EnterpriseRegisterFormValues = z.infer<typeof enterpriseRegisterFormSchema>

// ============================================================================
// API Response Types
// ============================================================================

export interface ApiResponse<T = unknown> {
  success: boolean
  message?: string
  data?: T
}

export interface EnterpriseRegisterPayload {
  name: string
  credit_code?: string
  contact_name?: string
  contact_phone?: string
  contact_email?: string
  business_license?: string
  invitation_code?: string
}
