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
import { z } from 'zod'

// ============================================================================
// Form Schemas
// ============================================================================

export const loginFormSchema = z.object({
  username: z.string().min(1, '请输入用户名或邮箱'),
  password: z
    .string()
    .min(1, '请输入密码')
    .min(8, '密码至少需要8个字符'),
})

// 邮箱登录专用 schema
export const emailLoginSchema = z.object({
  email: z
    .string()
    .min(1, '请输入邮箱地址')
    .email('请输入有效的邮箱地址'),
  password: z
    .string()
    .min(1, '请输入密码')
    .min(8, '密码至少需要8个字符'),
})

export const registerFormSchema = z
  .object({
    username: z.string().min(1, '请输入用户名'),
    email: z.string().optional(),
    password: z
      .string()
      .min(1, '请输入密码')
      .min(8, '密码至少需要8个字符')
      .max(20, '密码最多不能超过20个字符'),
    confirmPassword: z.string().min(1, '请确认密码'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "两次输入的密码不一致",
    path: ['confirmPassword'],
  })

// 邮箱注册专用 schema（邮箱为必填，作为登录名）
export const emailRegisterSchema = z
  .object({
    email: z
      .string()
      .min(1, '请输入邮箱地址')
      .email('请输入有效的邮箱地址'),
    password: z
      .string()
      .min(1, '请输入密码')
      .min(8, '密码至少需要8个字符')
      .max(20, '密码最多不能超过20个字符'),
    confirmPassword: z.string().min(1, '请确认密码'),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "两次输入的密码不一致",
    path: ['confirmPassword'],
  })

export const forgotPasswordFormSchema = z.object({
  email: z.string().email({
    message: '请输入有效的邮箱地址',
  }),
})

export const otpFormSchema = z.object({
  otp: z.string().min(1, '请输入验证码'),
})

// ============================================================================
// Validation Constants
// ============================================================================

export const PASSWORD_MIN_LENGTH = 8
export const PASSWORD_MAX_LENGTH = 20
export const BACKUP_CODE_LENGTH = 9 // XXXX-XXXX format
export const BACKUP_CODE_REGEX = /^[A-Z0-9]{4}-[A-Z0-9]{4}$/i

// ============================================================================
// Countdown Constants
// ============================================================================

export const EMAIL_VERIFICATION_COUNTDOWN = 30 // seconds
export const PASSWORD_RESET_COUNTDOWN = 30 // seconds

// ============================================================================
// OAuth Constants
// ============================================================================

export const OAUTH_BIND_STORAGE_KEY = 'oauth:binding:result'

// OTP Constants
export const OTP_LENGTH = 6
