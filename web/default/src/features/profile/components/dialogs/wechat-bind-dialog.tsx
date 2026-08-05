/*
Copyright (C) 2023-2026 FastToken

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
as published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact hello@fasttoken.example.com
*/
import { useEffect, useRef, useState } from 'react'
import { QrCode, ExternalLink } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { getBindings } from '../../api'

// ============================================================================
// WeChat Bind Dialog Component
// ============================================================================

// Polling interval (ms) for checking whether the WeChat account has been bound.
const POLL_INTERVAL_MS = 2000
// Overall timeout (ms) before we give up waiting for the scan/approval.
const POLL_TIMEOUT_MS = 40000

interface WeChatBindDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSuccess: () => void
  /** WeChat QR code URL sourced from system status (status.wechat_qrcode). */
  qrUrl?: string
}

export function WeChatBindDialog({
  open,
  onOpenChange,
  onSuccess,
  qrUrl,
}: WeChatBindDialogProps) {
  const { t } = useTranslation()
  const [polling, setPolling] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Capture the latest callbacks so the polling effect (keyed only on `open`)
  // does not restart on every parent re-render due to new callback identities.
  const onSuccessRef = useRef(onSuccess)
  const onOpenChangeRef = useRef(onOpenChange)
  const popupRef = useRef<Window | null>(null)
  onSuccessRef.current = onSuccess
  onOpenChangeRef.current = onOpenChange

  useEffect(() => {
    if (!open) {
      setError(null)
      setPolling(false)
      return
    }

    let cancelled = false
    let timer: ReturnType<typeof setInterval> | undefined
    let timeout: ReturnType<typeof setTimeout> | undefined

    const stop = () => {
      if (timer) clearInterval(timer)
      if (timeout) clearTimeout(timeout)
      setPolling(false)
    }

    const finish = (success: boolean) => {
      if (cancelled) return
      cancelled = true
      stop()
      if (success) {
        popupRef.current?.close()
        toast.success(t('WeChat bound successfully'))
        onSuccessRef.current()
      }
      onOpenChangeRef.current(false)
    }

    setPolling(true)
    setError(null)

    // Binding success is judged by polling the server's ground truth:
    // once wechat_bound becomes true, the current user has completed the
    // WeChat OAuth bind flow (controller.handleOAuthBind), so we close.
    timer = setInterval(async () => {
      try {
        const res = await getBindings()
        if (cancelled) return
        if (res.success && res.data?.wechat_bound) {
          finish(true)
        }
      } catch {
        // Tolerate transient network errors; keep polling until timeout.
      }
    }, POLL_INTERVAL_MS)

    timeout = setTimeout(() => {
      if (cancelled) return
      const msg = t('WeChat binding timed out, please try again')
      setError(msg)
      toast.error(msg)
      finish(false)
    }, POLL_TIMEOUT_MS)

    return () => {
      cancelled = true
      stop()
    }
  }, [open])

  const handleOpenQrWindow = () => {
    if (!qrUrl) {
      toast.error(t('WeChat QR code is not available'))
      return
    }
    popupRef.current = window.open(qrUrl, '_blank')
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>{t('Bind WeChat Account')}</DialogTitle>
          <DialogDescription>
            {t('Scan the QR code with WeChat to bind your account')}
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4 py-4'>
          <Alert>
            <QrCode className='h-4 w-4' />
            <AlertDescription>
              {t(
                'Please use WeChat\'s "Scan QR Code" feature to complete the binding process.'
              )}
            </AlertDescription>
          </Alert>

          {qrUrl ? (
            <div className='flex flex-col items-center justify-center gap-3'>
              <img
                src={qrUrl}
                alt={t('WeChat QR Code')}
                className='h-48 w-48 rounded-lg border object-contain'
              />
              <Button
                type='button'
                variant='outline'
                size='sm'
                onClick={handleOpenQrWindow}
              >
                <ExternalLink className='mr-1 h-4 w-4' />
                {t('Open in WeChat')}
              </Button>
            </div>
          ) : (
            <p className='text-muted-foreground text-center text-sm'>
              {t(
                'WeChat QR code is not available. Please contact the administrator.'
              )}
            </p>
          )}

          {polling && (
            <p className='text-muted-foreground text-center text-xs'>
              {t('Waiting for WeChat authorization...')}
            </p>
          )}

          {error && (
            <p className='text-destructive text-center text-xs'>{error}</p>
          )}

          <p className='text-muted-foreground text-center text-xs'>
            {t('After scanning, the binding will complete automatically')}
          </p>
        </div>
      </DialogContent>
    </Dialog>
  )
}
