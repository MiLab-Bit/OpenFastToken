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

// ============================================================================
// OAuth Provider Type Configuration
// ============================================================================

export const OAUTH_PROVIDER_TYPES = {
  OIDC: 'oidc',
  OAUTH2: 'oauth2',
} as const

export const OAUTH_PROVIDER_TYPE_OPTIONS = [
  { label: 'OIDC', value: OAUTH_PROVIDER_TYPES.OIDC },
  { label: 'OAuth 2.0', value: OAUTH_PROVIDER_TYPES.OAUTH2 },
] as const

// ============================================================================
// Error Messages
// ============================================================================

export const ERROR_MESSAGES = {
  UNEXPECTED: 'An unexpected error occurred',
  LOAD_FAILED: 'Failed to load OAuth providers',
  CREATE_FAILED: 'Failed to create OAuth provider',
  UPDATE_FAILED: 'Failed to update OAuth provider',
  DELETE_FAILED: 'Failed to delete OAuth provider',
  NAME_REQUIRED: 'Provider name is required',
  TYPE_REQUIRED: 'Provider type is required',
  CLIENT_ID_REQUIRED: 'Client ID is required',
  CLIENT_SECRET_REQUIRED: 'Client secret is required',
} as const

// ============================================================================
// Success Messages
// ============================================================================

export const SUCCESS_MESSAGES = {
  OAUTH_PROVIDER_CREATED: 'OAuth provider created successfully',
  OAUTH_PROVIDER_UPDATED: 'OAuth provider updated successfully',
  OAUTH_PROVIDER_DELETED: 'OAuth provider deleted successfully',
  OAUTH_ENABLED: 'OAuth provider enabled successfully',
  OAUTH_DISABLED: 'OAuth provider disabled successfully',
} as const
