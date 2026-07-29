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
import { useTranslation } from 'react-i18next'
import { MessageCircle } from 'lucide-react'
import { toast } from 'sonner'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { UserAuthForm } from '@/features/auth/sign-in/components/user-auth-form'
import { PhoneLoginForm } from '@/features/auth/components/phone-login-form'
import { EmailLoginForm } from '@/features/auth/sign-in/components/email-login-form'
import { useStatus } from '@/hooks/use-status'
import { wechatLoginByCode } from '@/features/auth/api'
import { useAuthRedirect } from '@/features/auth/hooks/use-auth-redirect'
import type { AuthFormProps } from '@/features/auth/types'

type LoginMode = 'email' | 'username' | 'phone'

export function UnifiedAuthPage({
  className,
  redirectTo,
  ...props
}: AuthFormProps) {
  const { t } = useTranslation()
  const [activeMode, setActiveMode] = useState<LoginMode>('email')
  const { status } = useStatus()
  const hasWeChatLogin = Boolean(status?.wechat_login)

  return (
    <Card className={`w-full max-w-sm mx-auto ${className || ''}`} {...props}>
      <CardHeader className="space-y-1 pb-3">
        <CardTitle className="text-xl font-bold text-center">
          {t('Welcome to FastToken')}
        </CardTitle>
        <CardDescription className="text-center text-sm">
          {t('Sign in to your account')}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {/* 登录方式切换 */}
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

        {/* 登录表单 */}
        <div className="pt-2">
          {activeMode === 'email' ? (
            <EmailLoginForm redirectTo={redirectTo} />
          ) : activeMode === 'username' ? (
            <UserAuthForm redirectTo={redirectTo} />
          ) : (
            <PhoneLoginForm redirectTo={redirectTo} />
          )}
        </div>

        {/* 微信登录 */}
        {hasWeChatLogin && (
          <div className="pt-2 border-t">
            <WeChatLoginButton redirectTo={redirectTo} />
          </div>
        )}

        {/* 注册链接 */}
        {status?.register_enabled !== false && (
          <div className="text-center text-sm text-muted-foreground">
            {t("Don't have an account?")}{' '}
            <a
              href="/sign-up"
              className="text-primary underline-offset-4 hover:underline"
            >
              {t('Sign up')}
            </a>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

// 微信登录按钮（简单图标型）
function WeChatLoginButton({ redirectTo: _redirectTo }: { redirectTo?: string }) {
  const { t } = useTranslation()
  const [code, setCode] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [showInput, setShowInput] = useState(false)
  const { handleLoginSuccess } = useAuthRedirect()
  const { status } = useStatus()

  const wechatQrCodeUrl = status?.wechat_qrcode || status?.wechat_qr_code || ''

  const handleWeChatLogin = () => {
    if (wechatQrCodeUrl) {
      window.open(wechatQrCodeUrl, '_blank')
    }
    setShowInput(true)
  }

  const handleSubmitCode = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!code.trim()) {
      toast.error(t('请输入验证码'))
      return
    }

    setIsLoading(true)
    try {
      const res = await wechatLoginByCode(code)
      if (res?.success) {
        await handleLoginSuccess(res.data as { id?: number } | null)
        toast.success(t('Signed in via WeChat'))
      } else {
        toast.error(res?.message || t('Login failed'))
      }
    } catch {
      toast.error(t('Login failed'))
    } finally {
      setIsLoading(false)
    }
  }

  if (showInput) {
    return (
      <form onSubmit={handleSubmitCode} className="space-y-2">
        <input
          type="text"
          id="wechat-verification-code"
          name="wechatCode"
          placeholder={t('Enter verification code')}
          value={code}
          onChange={(e) => setCode(e.target.value)}
          className="w-full px-3 py-2 text-sm border rounded-md"
        />
        <div className="flex gap-2">
          <Button type="submit" disabled={isLoading} size="sm" className="flex-1">
            {isLoading ? t('Verifying...') : t('Verify')}
          </Button>
          <Button type="button" variant="outline" size="sm" onClick={() => setShowInput(false)}>
            {t('Cancel')}
          </Button>
        </div>
      </form>
    )
  }

  return (
    <div className="flex flex-col items-center gap-2">
      {wechatQrCodeUrl && (
        <img
          src={wechatQrCodeUrl}
          alt={t('WeChat QR Code')}
          className="w-32 h-32 rounded-lg border cursor-pointer"
          onClick={handleWeChatLogin}
        />
      )}
      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={handleWeChatLogin}
        className="w-full"
      >
        <MessageCircle className="w-4 h-4 mr-2" />
        {t('Continue with WeChat')}
      </Button>
    </div>
  )
}
