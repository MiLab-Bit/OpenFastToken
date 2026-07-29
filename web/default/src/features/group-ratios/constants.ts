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
import { type TFunction } from 'i18next'

// ============================================================================
// Group Ratio Validation Constants
// ============================================================================

export const GROUP_RATIO_VALIDATION = {
  GROUP_NAME_MIN_LENGTH: 1,
  GROUP_NAME_MAX_LENGTH: 50,
  MODEL_RATIOS_MIN_LENGTH: 1,
  MODEL_RATIOS_MAX_LENGTH: 10000,
} as const

// ============================================================================
// Error Messages
// ============================================================================

export const ERROR_MESSAGES = {
  UNEXPECTED: 'An unexpected error occurred',
  LOAD_FAILED: 'Failed to load group ratios',
  CREATE_FAILED: 'Failed to create group ratio',
  UPDATE_FAILED: 'Failed to update group ratio',
  DELETE_FAILED: 'Failed to delete group ratio',
  GROUP_NAME_REQUIRED: 'Group name is required',
  GROUP_NAME_LENGTH_INVALID:
    'Group name must be between {{min}} and {{max}} characters',
  MODEL_RATIOS_REQUIRED: 'Model ratios configuration is required',
  MODEL_RATIOS_INVALID: 'Model ratios must be valid JSON',
} as const

export function getGroupRatioFormErrorMessages(t: TFunction) {
  return {
    GROUP_NAME_LENGTH_INVALID: t(ERROR_MESSAGES.GROUP_NAME_LENGTH_INVALID, {
      min: GROUP_RATIO_VALIDATION.GROUP_NAME_MIN_LENGTH,
      max: GROUP_RATIO_VALIDATION.GROUP_NAME_MAX_LENGTH,
    }),
  } as const
}

// ============================================================================
// Success Messages
// ============================================================================

export const SUCCESS_MESSAGES = {
  GROUP_RATIO_CREATED: 'Group ratio created successfully',
  GROUP_RATIO_UPDATED: 'Group ratio updated successfully',
  GROUP_RATIO_DELETED: 'Group ratio deleted successfully',
} as const
