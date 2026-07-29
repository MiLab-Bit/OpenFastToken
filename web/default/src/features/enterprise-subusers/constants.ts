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
import { type StatusBadgeProps } from '@/components/status-badge'
import { ROLE } from '@/lib/roles'

// ============================================================================
// Sub-User Role Configuration
// ============================================================================

export const SUBUSER_ROLE_STATUSES: Record<
  number,
  Pick<StatusBadgeProps, 'variant'> & {
    labelKey: string
    value: number
  }
> = {
  [ROLE.ADMIN]: {
    labelKey: 'Admin',
    variant: 'warning',
    value: ROLE.ADMIN,
  },
  [ROLE.USER]: {
    labelKey: 'User',
    variant: 'success',
    value: ROLE.USER,
  },
  [ROLE.GUEST]: {
    labelKey: 'Guest',
    variant: 'neutral',
    value: ROLE.GUEST,
  },
} as const

// ============================================================================
// Form Default Values
// ============================================================================

export const SUBUSER_FORM_DEFAULT_VALUES = {
  username: '',
  display_name: '',
  email: '',
  password: '',
  role: ROLE.USER,
  group: '',
}

// ============================================================================
// Error Messages
// ============================================================================

export const ERROR_MESSAGES = {
  UNEXPECTED: 'An unexpected error occurred',
  LOAD_FAILED: 'Failed to load sub-users',
  SEARCH_FAILED: 'Failed to search sub-users',
  CREATE_FAILED: 'Failed to create sub-user',
  UPDATE_FAILED: 'Failed to update sub-user',
  DELETE_FAILED: 'Failed to delete sub-user',
} as const

// ============================================================================
// Success Messages
// ============================================================================

export const SUCCESS_MESSAGES = {
  SUBUSER_CREATED: 'Sub-user created successfully',
  SUBUSER_UPDATED: 'Sub-user updated successfully',
  SUBUSER_DELETED: 'Sub-user deleted successfully',
} as const
