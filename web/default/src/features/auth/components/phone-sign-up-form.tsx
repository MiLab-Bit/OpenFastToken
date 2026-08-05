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
import { useState, useCallback, useRef, useEffect } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2, Smartphone } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
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
import { sendPhoneCode, phoneRegister } from '@/features/auth/api'
import { LegalConsent } from '@/features/auth/components/legal-consent'
import { useAuthRedirect } from '@/features/auth/hooks/use-auth-redirect'
import { useStatus } from '@/hooks/use-status'

// ============================================================================
// Form Schema
// ============================================================================

const phoneRegisterFormSchema = z.object({
  phone: z
    .string()
    .min(1, '请输入手机号')
    .regex(/^1\d{10}$/, '请输入有效的11位中国大陆手机号'),
  code: z
    .string()
    .min(1, '请输入验证码')
    .regex(/^\d{6}$/, '验证码必须为6位数字'),
  password: z
    .string()
    .min(8, '密码至少需要8个字符')
    .max(20, '密码最多不能超过20个字符'),
  confirmPassword: z
    .string()
    .min(1, '请确认密码'),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match.",
  path: ['confirmPassword'],
})

type PhoneRegisterFormValues = z.infer<typeof phoneRegisterFormSchema>

// ============================================================================
// Constants
// ============================================================================

const SMS_CODE_COUNTDOWN = 60 // seconds
const CHINA_PHONE_REGEX = /^1\d{10}$/

// ============================================================================
// Component
// ============================================================================

export function PhoneSignUpForm({ redirectTo }: { redirectTo?: string }) {
  const { t } = useTranslation()
  const [isLoading, setIsLoading] = useState(false)
  const [isSendingCode, setIsSendingCode] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const { handleLoginSuccess } = useAuthRedirect()

  const { status } = useStatus()
  const [agreedToLegal, setAgreedToLegal] = useState(false)
  const hasUserAgreement = Boolean(status?.user_agreement_enabled)
  const hasPrivacyPolicy = Boolean(status?.privacy_policy_enabled)
  const requiresLegalConsent = hasUserAgreement || hasPrivacyPolicy
  const legalConsentErrorMessage = t('please agree to legal terms first')

  useEffect(() => {
    if (requiresLegalConsent) {
      setAgreedToLegal(false)
    } else {
      setAgreedToLegal(true)
    }
  }, [requiresLegalConsent])

  const form = useForm<PhoneRegisterFormValues>({
    resolver: zodResolver(phoneRegisterFormSchema),
    defaultValues: {
      phone: '',
      code: '',
      password: '',
      confirmPassword: '',
    },
  })

  // Start countdown timer
  const startCountdown = useCallback(() => {
    setCountdown(SMS_CODE_COUNTDOWN)
    // Clear any existing timer
    if (timerRef.current) {
      clearInterval(timerRef.current)
    }
    
    timerRef.current = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          if (timerRef.current) {
            clearInterval(timerRef.current)
            timerRef.current = null
          }
          return 0
        }
        return prev - 1
      })
    }, 1000)
  }, [])

  // Cleanup timer on unmount
  useEffect(() => {
    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current)
        timerRef.current = null
      }
    }
  }, [])

  // Send verification code
  async function handleSendCode() {
    const phone = form.getValues('phone')
    if (!phone || !CHINA_PHONE_REGEX.test(phone)) {
      form.trigger('phone')
      toast.error(t('请先输入有效的手机号'))
      return
    }

    setIsSendingCode(true)
    try {
      const res = await sendPhoneCode(phone, 'register')
      if (res.success) {
        toast.success(t('Verification code sent'))
        startCountdown()
      } else {
        toast.error(res.message || t('Failed to send verification code'))
      }
    } catch {
      // Handled by global error interceptor
    } finally {
      setIsSendingCode(false)
    }
  }

  // Handle form submission
  async function onSubmit(data: PhoneRegisterFormValues) {
    if (requiresLegalConsent && !agreedToLegal) {
      toast.error(legalConsentErrorMessage)
      return
    }
    setIsLoading(true)
    try {
      const res = await phoneRegister({
        phone: data.phone,
        code: data.code,
        password: data.password,
      })

      if (res.success) {
        // Registration successful, backend auto-logs in the user
        await handleLoginSuccess(res.data as { id?: number } | null, redirectTo)
        toast.success(t('Registration successful!'))
      } else {
        toast.error(res.message || t('Registration failed'))
      }
    } catch {
      // Handled by global error interceptor
    } finally {
      setIsLoading(false)
    }
  }

  const codeButtonDisabled = isSendingCode || countdown > 0

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className={cn('grid gap-4')}
      >
        {/* Phone Number Field */}
        <FormField
          control={form.control}
          name="phone"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Phone Number')}</FormLabel>
              <FormControl>
                <div className="relative">
                  <Smartphone className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    placeholder={t('Enter your 11-digit phone number')}
                    className="pl-9"
                    maxLength={11}
                    inputMode="numeric"
                    autoComplete="tel"
                    {...field}
                  />
                </div>
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Verification Code Field */}
        <FormField
          control={form.control}
          name="code"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Verification Code')}</FormLabel>
              <div className="flex gap-2">
                <FormControl>
                  <Input
                    placeholder={t('Enter 6-digit code')}
                    className="flex-1"
                    maxLength={6}
                    inputMode="numeric"
                    autoComplete="one-time-code"
                    {...field}
                  />
                </FormControl>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  disabled={codeButtonDisabled}
                  onClick={handleSendCode}
                  className="shrink-0"
                >
                  {isSendingCode ? (
                    <Loader2 className="mr-1 h-3 w-3 animate-spin" />
                  ) : null}
                  {countdown > 0
                    ? t('Resend in {{seconds}}s', { seconds: countdown })
                    : t('Send Code')}
                </Button>
              </div>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* Password Field */}
        <FormField
          control={form.control}
          name="password"
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('Password')}</FormLabel>
              <FormControl>
                <PasswordInput
                  placeholder={t('Set a password (8-20 characters)')}
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
              <FormLabel>{t('Confirm Password')}</FormLabel>
              <FormControl>
                <PasswordInput
                  placeholder={t('Re-enter your password')}
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

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
          disabled={isLoading || (requiresLegalConsent && !agreedToLegal)}
        >
          {isLoading ? <Loader2 className="animate-spin" /> : null}
          {t('Sign up')}
        </Button>
      </form>
    </Form>
  )
}
