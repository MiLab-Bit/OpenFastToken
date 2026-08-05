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
import { useState } from 'react'
import { Loader2, Smartphone } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { useCountdown } from '@/hooks/use-countdown'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { sendPhoneCodeForBind, bindPhone, unbindPhone } from '../../api'

// ============================================================================
// Phone Bind Dialog Component
// ============================================================================

interface PhoneBindDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentPhone?: string
  onSuccess: () => void
}

export function PhoneBindDialog({
  open,
  onOpenChange,
  currentPhone,
  onSuccess,
}: PhoneBindDialogProps) {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [sendingCode, setSendingCode] = useState(false)
  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [password, setPassword] = useState('')
  const [mode, setMode] = useState<'bind' | 'unbind'>(
    currentPhone ? 'unbind' : 'bind'
  )
  const {
    secondsLeft,
    isActive,
    start: startCountdown,
    reset: resetCountdown,
  } = useCountdown({
    initialSeconds: 60,
  })

  // Reset mode when dialog opens with different phone state
  const handleOpenChange = (newOpen: boolean) => {
    if (!loading) {
      onOpenChange(newOpen)
      if (!newOpen) {
        resetForm()
      }
    }
  }

  const resetForm = () => {
    setPhone('')
    setCode('')
    setPassword('')
    resetCountdown()
    setMode(currentPhone ? 'unbind' : 'bind')
  }

  const handleSendCode = async () => {
    const targetPhone = phone || currentPhone || ''
    if (!targetPhone || !/^1\d{10}$/.test(targetPhone)) {
      toast.error(t('Please enter a valid 11-digit phone number'))
      return
    }

    try {
      setSendingCode(true)
      const res = await sendPhoneCodeForBind(targetPhone)

      // If send_code doesn't accept purpose, try without it
      if (!res.success && res.message?.includes('purpose')) {
        const fallbackRes = await sendPhoneCodeForBind(targetPhone)
        if (fallbackRes.success) {
          toast.success(t('Verification code sent!'))
          startCountdown()
        } else {
          toast.error(fallbackRes.message || t('Failed to send verification code'))
        }
      } else if (res.success) {
        toast.success(t('Verification code sent! Please check your phone.'))
        startCountdown()
      } else {
        toast.error(res.message || t('Failed to send verification code'))
      }
    } catch {
      toast.error(t('Failed to send verification code'))
    } finally {
      setSendingCode(false)
    }
  }

  const handleBind = async () => {
    if (mode === 'bind') {
      if (!phone || !code) {
        toast.error(t('Please enter phone number and verification code'))
        return
      }
      try {
        setLoading(true)
        const response = await bindPhone(phone, code)

        if (response.success) {
          toast.success(t('Phone bound successfully!'))
          onOpenChange(false)
          onSuccess()
          resetForm()
        } else {
          toast.error(response.message || t('Failed to bind phone'))
        }
      } catch {
        toast.error(t('Failed to bind phone'))
      } finally {
        setLoading(false)
      }
    } else {
      // Unbind mode
      if (!code) {
        toast.error(t('Please enter verification code'))
        return
      }
      try {
        setLoading(true)
        const response = await unbindPhone(code, password || undefined)

        if (response.success) {
          toast.success(t('Phone unbound successfully!'))
          onOpenChange(false)
          onSuccess()
          resetForm()
        } else {
          toast.error(response.message || t('Failed to unbind phone'))
        }
      } catch {
        toast.error(t('Failed to unbind phone'))
      } finally {
        setLoading(false)
      }
    }
  }

  const switchMode = () => {
    const newMode = mode === 'bind' ? 'unbind' : 'bind'
    setMode(newMode)
    setPhone('')
    setCode('')
    setPassword('')
    resetCountdown()
  }

  const isBound = !!currentPhone && mode === 'unbind'

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>
            {isBound ? t('Unbind Phone') : t('Bind Phone Number')}
          </DialogTitle>
          <DialogDescription>
            {isBound
              ? t('Current phone: {{phone}}. Verify to unbind.', {
                  phone: currentPhone,
                })
              : t('Bind a phone number to your account.')}
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4 py-4'>
          {/* Mode toggle */}
          {currentPhone && (
            <div className='flex justify-center'>
              <Button
                type='button'
                variant='ghost'
                size='sm'
                onClick={switchMode}
                disabled={loading}
                className='text-xs text-muted-foreground'
              >
                {mode === 'bind'
                  ? t('Switch to Unbind')
                  : t('Switch to Bind')}
              </Button>
            </div>
          )}

          {mode === 'bind' && (
            <div className='space-y-2'>
              <Label htmlFor='phone'>{t('Phone Number')}</Label>
              <div className='relative'>
                <Smartphone className='text-muted-foreground absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2' />
                <Input
                  id='phone'
                  type='tel'
                  value={phone}
                  onChange={(e) =>
                    setPhone(e.target.value.replace(/\D/g, '').slice(0, 11))
                  }
                  placeholder={t('Enter your 11-digit phone number')}
                  disabled={loading}
                  maxLength={11}
                  className='pl-10'
                />
              </div>
            </div>
          )}

          <div className='space-y-2'>
            <Label htmlFor='code'>{t('Verification Code')}</Label>
            <div className='flex gap-2'>
              <Input
                id='code'
                value={code}
                onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                placeholder={t('Enter code')}
                disabled={loading}
                maxLength={6}
              />
              <Button
                type='button'
                variant='outline'
                onClick={handleSendCode}
                disabled={sendingCode || isActive || (mode === 'bind' && !phone && !currentPhone)}
              >
                {isActive
                  ? `${secondsLeft}s`
                  : sendingCode
                    ? t('Sending...')
                    : t('Send')}
              </Button>
            </div>
          </div>

          {mode === 'unbind' && (
            <div className='space-y-2'>
              <Label htmlFor='password'>{t('Password')}</Label>
              <Input
                id='password'
                type='password'
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder={t('Enter your password to confirm')}
                disabled={loading}
              />
            </div>
          )}
        </div>

        <DialogFooter>
          <Button
            type='button'
            variant='outline'
            onClick={() => handleOpenChange(false)}
            disabled={loading}
          >
            {t('Cancel')}
          </Button>
          <Button
            type='button'
            onClick={handleBind}
            disabled={
              loading ||
              (mode === 'bind' ? !phone || !code : !code)
            }
          >
            {loading && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
            {loading
              ? t('Processing...')
              : isBound
                ? t('Unbind')
                : t('Bind Phone')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
