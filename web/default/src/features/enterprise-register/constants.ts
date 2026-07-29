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

// ============================================================================
// Form Default Values
// ============================================================================

export const ENTERPRISE_REGISTER_FORM_DEFAULTS = {
  name: '',
  credit_code: '',
  contact_name: '',
  contact_phone: '',
  contact_email: '',
  business_license: '',
  invitation_code: '',
}

// ============================================================================
// Error Messages
// ============================================================================

export const ERROR_MESSAGES = {
  UNEXPECTED: 'An unexpected error occurred',
  REGISTRATION_FAILED: 'Failed to register enterprise',
} as const

// ============================================================================
// Success Messages
// ============================================================================

export const SUCCESS_MESSAGES = {
  REGISTRATION_SUCCESS: 'Enterprise registration submitted successfully',
} as const
