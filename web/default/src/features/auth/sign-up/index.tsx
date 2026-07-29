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
import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useStatus } from '@/hooks/use-status'
import { AuthLayout } from '../auth-layout'
import { TermsFooter } from '../components/terms-footer'
import { SignUpForm } from './components/sign-up-form'
import { EmailSignUpForm } from './components/email-sign-up-form'
import { PhoneSignUpForm } from '../components/phone-sign-up-form'

type SignUpMode = 'email' | 'username' | 'phone'

export function SignUp() {
  const { t } = useTranslation()
  const [activeMode, setActiveMode] = useState<SignUpMode>('email')
  const { status } = useStatus()

  return (
    <AuthLayout>
      <div className="w-full space-y-8">
        <div className="space-y-2">
          <h2 className="text-center text-2xl font-semibold tracking-tight sm:text-left">
            {t('Create an account')}
          </h2>
          <p className="text-muted-foreground text-left text-sm sm:text-base">
            {t('Already have an account?')}{' '}
            <Link
              to="/sign-in"
              className="hover:text-primary font-medium underline underline-offset-4"
            >
              {t('Sign in')}
            </Link>
            .
          </p>
        </div>

        {/* 注册方式切换 */}
        <div className="flex gap-1 p-1 bg-muted rounded-lg">
          <button
            onClick={() => setActiveMode('email')}
            className={`flex-1 py-1.5 text-sm font-medium rounded-md transition-colors ${
              activeMode === 'email'
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            {t('Email')}
          </button>
          <button
            onClick={() => setActiveMode('username')}
            className={`flex-1 py-1.5 text-sm font-medium rounded-md transition-colors ${
              activeMode === 'username'
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            {t('Account')}
          </button>
          <button
            onClick={() => setActiveMode('phone')}
            className={`flex-1 py-1.5 text-sm font-medium rounded-md transition-colors ${
              activeMode === 'phone'
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            {t('Phone')}
          </button>
        </div>

        {/* 注册表单 */}
        <div className="pt-2">
          {activeMode === 'email' ? (
            <EmailSignUpForm />
          ) : activeMode === 'username' ? (
            <SignUpForm />
          ) : (
            <PhoneSignUpForm />
          )}
        </div>

        <TermsFooter
          variant="sign-up"
          status={status}
          className="text-center"
        />
      </div>
    </AuthLayout>
  )
}
