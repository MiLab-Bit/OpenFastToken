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
// API Functions
// ============================================================================

export {
  login,
  logout,
  register,
  sendPasswordResetEmail,
  sendEmailVerification,
  getOAuthState,
  wechatLoginByCode,
  sendPhoneCode,
  phoneLogin,
  phoneRegister,
} from './api'

// ============================================================================
// Types
// ============================================================================

export type {
  LoginPayload,
  LoginResponse,
  RegisterPayload,
  PasswordResetPayload,
  EmailVerificationPayload,
  BindEmailPayload,
  ApiResponse,
  SystemStatus,
  OAuthProvider,
  AuthFormProps,
} from './types'

// ============================================================================
// Constants & Schemas
// ============================================================================

export {
  loginFormSchema,
  registerFormSchema,
  forgotPasswordFormSchema,
  PASSWORD_MIN_LENGTH,
  PASSWORD_MAX_LENGTH,
  BACKUP_CODE_LENGTH,
  BACKUP_CODE_REGEX,
  EMAIL_VERIFICATION_COUNTDOWN,
  PASSWORD_RESET_COUNTDOWN,
} from './constants'

// ============================================================================
// Utilities
// ============================================================================

export {
  buildOIDCOAuthUrl,
  buildLinuxDOOAuthUrl,
  getAvailableOAuthProviders,
  hasOAuthProviders,
} from './lib/oauth'

export {
  saveUserId,
  getUserId,
  removeUserId,
  getAffiliateCode,
  saveAffiliateCode,
} from './lib/storage'

export {
  isValidBackupCode,
  formatBackupCode,
  cleanBackupCode,
  isValidEmail,
} from './lib/validation'

// ============================================================================
// Hooks
// ============================================================================

export { useTurnstile } from './hooks/use-turnstile'
export { useOAuthLogin } from './hooks/use-oauth-login'
export { useAuthRedirect } from './hooks/use-auth-redirect'
export { useEmailVerification } from './hooks/use-email-verification'

// ============================================================================
// Components
// ============================================================================

export { AuthLayout } from './auth-layout'
export { OAuthProviders } from './components/oauth-providers'
export { TermsFooter } from './components/terms-footer'
export { LegalConsent } from './components/legal-consent'
export { SignIn } from './sign-in'
export { SignUp } from './sign-up'
export { ForgotPassword } from './forgot-password'
