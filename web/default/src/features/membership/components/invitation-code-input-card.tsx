/*
Copyright (C) 2023-2026 FastToken
*/
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { toast } from 'sonner'
import { useInvitationCode } from '../hooks/use-invitation-code'

interface InvitationCodeInputCardProps {
  onCodeUsed: () => void
  disabled: boolean
}

export function InvitationCodeInputCard({
  onCodeUsed,
  disabled,
}: InvitationCodeInputCardProps) {
  const { i18n } = useTranslation()
  const [code, setCode] = useState('')
  const { using, useCode } = useInvitationCode()
  const isZh = i18n.language === 'zh' || i18n.language?.startsWith('zh')

  const handleUseCode = async () => {
    if (!code.trim()) {
      toast.error(isZh ? '请输入邀请码' : 'Please enter an invitation code')
      return
    }

    const result = await useCode(code.trim())
    if (result.success) {
      toast.success(
        isZh ? '邀请码使用成功！会员已升级' : 'Invitation code applied! Membership upgraded'
      )
      setCode('')
      onCodeUsed()
    } else {
      toast.error(result.message || (isZh ? '邀请码无效' : 'Invalid invitation code'))
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>
          {isZh ? '🎫 使用邀请码升级会员' : '🎫 Upgrade with Invitation Code'}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className='mb-4 text-sm text-muted-foreground'>
          {isZh
            ? '输入企业认证邀请码，升级为黄金或铂金会员，享受更低折扣'
            : 'Enter an enterprise invitation code to upgrade to Gold or Platinum membership for lower discount rates'}
        </p>
        <div className='flex gap-2'>
          <Input
            placeholder={
              isZh ? '请输入邀请码（如 FAGGUUPA）' : 'Enter invitation code (e.g. FAGGUUPA)'
            }
            value={code}
            onChange={(e) => setCode(e.target.value)}
            disabled={disabled || using}
            className='flex-1'
            maxLength={20}
          />
          <Button
            onClick={handleUseCode}
            disabled={disabled || using || !code.trim()}
          >
            {using
              ? isZh
                ? '使用中...'
                : 'Applying...'
              : isZh
                ? '使用'
                : 'Apply'}
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}