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
import { useState, useCallback, useRef, useEffect } from 'react'
import { z } from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Loader2, LogIn, Smartphone } from 'lucide-react'
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
import { sendPhoneCode, phoneLogin } from '@/features/auth/api'
import { useAuthRedirect } from '@/features/auth/hooks/use-auth-redirect'
import type { AuthFormProps } from '@/features/auth/types'

// ============================================================================
// Form Schema
// ============================================================================

const phoneLoginFormSchema = z.object({
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
    .optional()
    .refine(
      (val) => !val || (val.length >= 8 && val.length <= 20),
      { message: '密码长度必须在8-20个字符之间' }
    ),
})

type PhoneLoginFormValues = z.infer<typeof phoneLoginFormSchema>

// ============================================================================
// Constants
// ============================================================================

const SMS_CODE_COUNTDOWN = 60 // seconds
const CHINA_PHONE_REGEX = /^1\d{10}$/

// ============================================================================
// Component
// ============================================================================

export function PhoneLoginForm({
  className,
  redirectTo,
  ...props
}: AuthFormProps) {
  const { t } = useTranslation()
  const [isLoading, setIsLoading] = useState(false)
  const [isSendingCode, setIsSendingCode] = useState(false)
  const [countdown, setCountdown] = useState(0)
  const [showPassword, setShowPassword] = useState(false)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const { handleLoginSuccess } = useAuthRedirect()

  const form = useForm<PhoneLoginFormValues>({
    resolver: zodResolver(phoneLoginFormSchema),
    defaultValues: {
      phone: '',
      code: '',
      password: '',
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
      const res = await sendPhoneCode(phone, 'login')
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
  async function onSubmit(data: PhoneLoginFormValues) {
    setIsLoading(true)
    try {
      const payload: { phone: string; code: string; password?: string } = {
        phone: data.phone,
        code: data.code,
      }
      if (data.password) {
        payload.password = data.password
      }

      const res = await phoneLogin(payload)

      if (res.success) {
        await handleLoginSuccess(res.data as { id?: number } | null, redirectTo)
        toast.success(t('Welcome back!'))
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
        className={cn('grid gap-4', className)}
        {...props}
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

        {/* Optional Password Toggle */}
        {!showPassword && (
          <Button
            type="button"
            variant="link"
            size="sm"
            className="justify-start p-0 text-muted-foreground h-auto"
            onClick={() => setShowPassword(true)}
          >
            {t('Use password login instead')}
          </Button>
        )}

        {/* Optional Password Field */}
        {showPassword && (
          <FormField
            control={form.control}
            name="password"
            render={({ field }) => (
              <FormItem className="relative">
                <FormLabel>{t('Password (Optional)')}</FormLabel>
                <FormControl>
                  <PasswordInput
                    placeholder={t('Enter password if you have one')}
                    {...field}
                    value={field.value || ''}
                  />
                </FormControl>
                <FormMessage />
                <Button
                  type="button"
                  variant="link"
                  size="sm"
                  className="absolute end-0 -top-0.5 z-10 text-muted-foreground h-auto p-0 text-xs"
                  onClick={() => {
                    form.setValue('password', '')
                    setShowPassword(false)
                  }}
                >
                  {t('Hide')}
                </Button>
              </FormItem>
            )}
          />
        )}

        {/* Submit Button */}
        <Button
          type="submit"
          className="mt-2 w-full justify-center gap-2"
          disabled={isLoading}
        >
          {isLoading ? <Loader2 className="animate-spin" /> : <LogIn />}
          {t('Sign in')}
        </Button>
      </form>
    </Form>
  )
}
