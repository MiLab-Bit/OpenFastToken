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
import { Loader2, Pencil } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
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
import { updateUserProfile } from '../../api'

// ============================================================================
// Edit Profile Dialog Component
// ============================================================================

interface EditProfileDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentDisplayName?: string
  currentUsername?: string
  onSuccess: () => void
}

export function EditProfileDialog({
  open,
  onOpenChange,
  currentDisplayName,
  currentUsername,
  onSuccess,
}: EditProfileDialogProps) {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const [displayName, setDisplayName] = useState(currentDisplayName || '')

  // Sync with props when dialog opens
  const handleOpenChange = (newOpen: boolean) => {
    if (!loading) {
      onOpenChange(newOpen)
      if (newOpen) {
        setDisplayName(currentDisplayName || '')
      }
    }
  }

  const handleSave = async () => {
    if (!displayName.trim()) {
      toast.error(t('Display name cannot be empty'))
      return
    }

    try {
      setLoading(true)
      const response = await updateUserProfile({
        display_name: displayName.trim(),
      })

      if (response.success) {
        toast.success(t('Profile updated successfully!'))
        onOpenChange(false)
        onSuccess()
      } else {
        toast.error(response.message || t('Failed to update profile'))
      }
    } catch {
      toast.error(t('Failed to update profile'))
    } finally {
      setLoading(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className='sm:max-w-md'>
        <DialogHeader>
          <DialogTitle>{t('Edit Profile')}</DialogTitle>
          <DialogDescription>
            {t('Update your display name and public information.')}
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4 py-4'>
          {/* Username (read-only) */}
          <div className='space-y-2'>
            <Label htmlFor='username'>{t('Username')}</Label>
            <Input
              id='username'
              value={currentUsername || ''}
              disabled
              className='bg-muted/50'
            />
            <p className='text-muted-foreground text-xs'>
              {t('Username cannot be changed')}
            </p>
          </div>

          {/* Display name (editable) */}
          <div className='space-y-2'>
            <Label htmlFor='display_name'>{t('Display Name')}</Label>
            <Input
              id='display_name'
              value={displayName}
              onChange={(e) => setDisplayName(e.target.value)}
              placeholder={t('Enter your display name')}
              disabled={loading}
              maxLength={50}
            />
          </div>
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
            onClick={handleSave}
            disabled={loading || !displayName.trim()}
          >
            {loading && <Loader2 className='mr-2 h-4 w-4 animate-spin' />}
            {loading ? t('Saving...') : t('Save Changes')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

// ============================================================================
// Inline edit button for profile header
// ============================================================================

interface ProfileEditButtonProps {
  onClick: () => void
}

export function ProfileEditButton({ onClick }: ProfileEditButtonProps) {
  const { t } = useTranslation()
  return (
    <Button
      variant='ghost'
      size='icon'
      className='text-muted-foreground hover:text-foreground h-7 w-7 shrink-0 rounded-lg'
      onClick={onClick}
      aria-label={t('Edit Profile')}
    >
      <Pencil className='h-3.5 w-3.5' />
    </Button>
  )
}
