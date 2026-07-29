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
// Group Ratio Schema & Types
// ============================================================================

export const groupRatioSchema = z.object({
  id: z.number(),
  group_name: z.string(),
  model_ratios: z.string(),
  created_time: z.number(),
  updated_time: z.number(),
})

export type GroupRatio = z.infer<typeof groupRatioSchema>

// ============================================================================
// API Request/Response Types
// ============================================================================

export interface ApiResponse<T = unknown> {
  success: boolean
  message?: string
  data?: T
}

export interface GetGroupRatiosParams {
  p?: number
  page_size?: number
}

export interface GetGroupRatiosResponse {
  success: boolean
  message?: string
  data?: {
    items: GroupRatio[]
    total: number
    page: number
    page_size: number
  }
}

export interface GroupRatioFormData {
  id?: number
  group_name: string
  model_ratios: string
}

// ============================================================================
// Dialog Types
// ============================================================================

export type GroupRatioDialogType = 'create' | 'update' | 'delete' | 'view'
