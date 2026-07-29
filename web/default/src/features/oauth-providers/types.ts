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
// OAuth Provider Schema & Types
// ============================================================================

export const oauthProviderTypeValues = ['oidc', 'oauth2'] as const

export const oauthProviderSchema = z.object({
  id: z.number(),
  name: z.string(),
  type: z.enum(oauthProviderTypeValues),
  client_id: z.string(),
  client_secret: z.string(),
  issuer_url: z.string(),
  auth_url: z.string(),
  token_url: z.string(),
  userinfo_url: z.string(),
  scopes: z.string(),
  enabled: z.boolean(),
})

export type OAuthProvider = z.infer<typeof oauthProviderSchema>

// ============================================================================
// API Request/Response Types
// ============================================================================

export interface ApiResponse<T = unknown> {
  success: boolean
  message?: string
  data?: T
}

export interface GetOAuthProvidersParams {
  p?: number
  page_size?: number
}

export interface GetOAuthProvidersResponse {
  success: boolean
  message?: string
  data?: {
    items: OAuthProvider[]
    total: number
    page: number
    page_size: number
  }
}

export interface OAuthProviderFormData {
  name: string
  type: 'oidc' | 'oauth2'
  client_id: string
  client_secret: string
  issuer_url: string
  auth_url: string
  token_url: string
  userinfo_url: string
  scopes: string
  enabled: boolean
}

// ============================================================================
// Dialog Types
// ============================================================================

export type OAuthProviderDialogType = 'create' | 'update' | 'delete' | 'view'
