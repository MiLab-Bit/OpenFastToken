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
import { useEffect, useState } from 'react'
import type { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import { useStatus } from '@/hooks/use-status'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { PasswordInput } from '@/components/password-input'
import { Turnstile } from '@/components/turnstile'
import { register } from '@/features/auth/api'
import { LegalConsent } from '@/features/auth/components/legal-consent'
import { useAuthRedirect } from '@/features/auth/hooks/use-auth-redirect'
import { useEmailVerification } from '@/features/auth/hooks/use-email-verification'
import { useTurnstile } from '@/features/auth/hooks/use-turnstile'
import {
  getAffiliateCode,
  saveAffiliateCode,
} from '@/features/auth/lib/storage'
import { emailRegisterSchema } from '@/features/auth/constants'

export function EmailSignUpForm() {
  const { t } = useTranslation()
  const [isLoading, setIsLoading] = useState(false)
  const [verificationCode, setVerificationCode] = useState('')
  const [agreedToLegal, setAgreedToLegal] = useState(false)
  const legalConsentErrorMessage = t('请先同意法律条款')

  const { status } = useStatus()
  const {
    isTurnstileEnabled,
    turnstileSiteKey,
    turnstileToken,
    setTurnstileToken,
    validateTurnstile,
  } = useTurnstile()
  const { redirectToLogin } = useAuthRedirect()
  const {
    isSending: isSendingCode,
    secondsLeft,
    isActive,
    sendCode,
  } = useEmailVerification({
    turnstileToken,
    validateTurnstile,
  })

  const form = useForm<z.infer<typeof emailRegisterSchema>>({
    resolver: zodResolver(emailRegisterSchema),
    defaultValues: {
      email: '',
      password: '',
      confirmPassword: '',
    },
  })

  const emailValue = form.watch('email')
  const hasUserAgreement = Boolean(status?.user_agreement_enabled)
  const hasPrivacyPolicy = Boolean(status?.privacy_policy_enabled)
  const requiresLegalConsent = hasUserAgreement || hasPrivacyPolicy
  const turnstileReady = !isTurnstileEnabled || Boolean(turnstileToken)

  useEffect(() => {
    if (requiresLegalConsent) {
      setAgreedToLegal(false)
    } else {
      setAgreedToLegal(true)
    }
  }, [requiresLegalConsent])

  useEffect(() => {
    const aff = new URLSearchParams(window.location.search).get('aff')?.trim()
    if (aff) {
      saveAffiliateCode(aff)
    }
  }, [])

  async function onSubmit(data: z.infer<typeof emailRegisterSchema>) {
    if (requiresLegalConsent && !agreedToLegal) {
      toast.error(legalConsentErrorMessage)
      return
    }

    if (!verificationCode) {
      toast.error(t('请输入验证码'))
      return
    }

    if (!validateTurnstile()) return

    setIsLoading(true)
    try {
      const res = await register({
        username: data.email, // 用邮箱作为用户名
        password: data.password,
        email: data.email,
        verification_code: verificationCode,
        aff_code: getAffiliateCode(),
        turnstile: turnstileToken,
      })

      if (res?.success) {
        toast.success(t('Account created! Please sign in'))
        redirectToLogin()
      } else {
        toast.error(res?.message || t('Failed to create account'))
      }
    } catch {
      // Errors are handled by global interceptor
    } finally {
      setIsLoading(false)
    }
  }

  async function handleSendVerificationCode() {
    const email = form.getValues('email')
    if (!email) {
      toast.error(t('请先输入邮箱地址'))
      return
    }
    const result = await sendCode(email)
    if (result === false) return // sendCode 内部已处理错误提示
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn('grid gap-4')}
      >
        {/* Email Field */}
        <FormField
          control={form.control}
          name="email"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Email')}</FormLabel>
              <FormControl>
                <Input
                  type="text"
                  placeholder={t('name@example.com')}
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Verification Code Field */}
        <div className="flex items-end gap-2">
          <div className="flex-1">
            <Input
              placeholder={t('Verification code')}
              value={verificationCode}
              onChange={(e) => setVerificationCode(e.target.value)}
            />
          </div>
          <Button
            variant="outline"
            type="button"
            disabled={
              isLoading ||
              isSendingCode ||
              isActive ||
              !emailValue ||
              !turnstileReady
            }
            onClick={handleSendVerificationCode}
          >
            {isActive ? (
              t('Resend ({{seconds}}s)', { seconds: secondsLeft })
            ) : isSendingCode ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              t('Send code')
            )}
          </Button>
        </div>

        {/* Password Field */}
        <FormField
          control={form.control}
          name="password"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Password')}</FormLabel>
              <FormControl>
                <PasswordInput
                  placeholder={t('Enter password (8-20 characters)')}
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Confirm Password Field */}
        <FormField
          control={form.control}
          name="confirmPassword"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Confirm password')}</FormLabel>
              <FormControl>
                <PasswordInput
                  placeholder={t('Confirm password')}
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Turnstile */}
        {isTurnstileEnabled && (
          <div className="mt-2">
            <Turnstile
              siteKey={turnstileSiteKey}
              onVerify={setTurnstileToken}
            />
          </div>
        )}

        <LegalConsent
          status={status}
          checked={agreedToLegal}
          onCheckedChange={setAgreedToLegal}
          className="mt-1"
        />

        {/* Submit Button */}
        <Button
          type="submit"
          className="mt-2 w-full justify-center gap-2"
          disabled={
            isLoading ||
            (requiresLegalConsent && !agreedToLegal) ||
            !turnstileReady
          }
        >
          {isLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
          {t('Create account')}
        </Button>
      </form>
    </Form>
  )
}
