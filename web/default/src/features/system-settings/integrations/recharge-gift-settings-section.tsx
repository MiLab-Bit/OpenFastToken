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
*/
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { Plus, Trash2 } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import {
  SettingsForm,
  SettingsSwitchField,
} from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useUpdateOption } from '../hooks/use-update-option'

interface RechargeGiftTier {
  amount: number
  bonus_rate: number
}

interface RechargeGiftGift {
  enabled: boolean
  threshold: number
  name: string
  type: string
}

interface RechargeGiftSettingsSectionProps {
  defaultValues: {
    enabled: boolean
    tiers: string
    gift: string
  }
}

function asBool(v: unknown, fallback = false): boolean {
  if (typeof v === 'boolean') return v
  if (typeof v === 'string') return v === 'true' || v === '1'
  return fallback
}

function parseTiers(raw: string): RechargeGiftTier[] {
  try {
    const arr = JSON.parse(raw)
    if (!Array.isArray(arr)) return []
    return arr
      .filter((t) => t && typeof t === 'object')
      .map((t) => ({
        amount: Number((t as Record<string, unknown>).amount) || 0,
        bonus_rate: Number((t as Record<string, unknown>).bonus_rate) || 0,
      }))
  } catch {
    return []
  }
}

function parseGift(raw: string): RechargeGiftGift {
  try {
    const g = JSON.parse(raw) as Record<string, unknown>
    return {
      enabled: asBool(g?.enabled),
      threshold: Number(g?.threshold) || 0,
      name: typeof g?.name === 'string' ? g.name : '',
      type: typeof g?.type === 'string' ? g.type : '',
    }
  } catch {
    return { enabled: false, threshold: 0, name: '', type: '' }
  }
}

export function RechargeGiftSettingsSection({
  defaultValues,
}: RechargeGiftSettingsSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()

  const [enabled, setEnabled] = useState<boolean>(defaultValues.enabled)
  const [tiers, setTiers] = useState<RechargeGiftTier[]>(
    parseTiers(defaultValues.tiers)
  )
  const [gift, setGift] = useState<RechargeGiftGift>(
    parseGift(defaultValues.gift)
  )
  const [saving, setSaving] = useState(false)

  const updateTier = (idx: number, patch: Partial<RechargeGiftTier>) => {
    setTiers((prev) =>
      prev.map((tier, i) => (i === idx ? { ...tier, ...patch } : tier))
    )
  }
  const addTier = () =>
    setTiers((prev) => [...prev, { amount: 0, bonus_rate: 0 }])
  const removeTier = (idx: number) =>
    setTiers((prev) => prev.filter((_, i) => i !== idx))

  async function handleSave() {
    try {
      setSaving(true)
      const updates: Array<{ key: string; value: string }> = [
        { key: 'recharge_gift_setting.enabled', value: String(enabled) },
        {
          key: 'recharge_gift_setting.tiers',
          value: JSON.stringify(tiers),
        },
        { key: 'recharge_gift_setting.gift', value: JSON.stringify(gift) },
      ]
      for (const u of updates) {
        await updateOption.mutateAsync(u)
      }
      toast.success(t('Recharge gift settings saved'))
    } catch {
      toast.error(t('Failed to save settings'))
    } finally {
      setSaving(false)
    }
  }

  return (
    <SettingsSection title={t('Recharge Gift')}>
      <SettingsForm
        onSubmit={(e) => {
          e.preventDefault()
          void handleSave()
        }}
      >
        <SettingsPageFormActions
          onSave={() => void handleSave()}
          isSaving={saving}
          saveLabel='Save recharge gift settings'
        />

        <SettingsSwitchField
          checked={enabled}
          onCheckedChange={setEnabled}
          label={t('Enable recharge gift')}
          description={t(
            'Grant bonus quota and/or a threshold gift on recharge'
          )}
        />

        {/* Bonus tiers */}
        <div
          data-settings-form-span='full'
          className='flex flex-col gap-3 border-t pt-4'
        >
          <div className='flex items-center justify-between'>
            <span className='text-sm font-medium'>{t('Bonus tiers')}</span>
            <Button type='button' size='sm' variant='outline' onClick={addTier}>
              <Plus className='mr-1 h-4 w-4' />
              {t('Add tier')}
            </Button>
          </div>
          {tiers.length === 0 ? (
            <p className='text-muted-foreground text-xs'>
              {t('No tiers configured')}
            </p>
          ) : (
            <div className='flex flex-col gap-2'>
              {tiers.map((tier, idx) => (
                <div key={idx} className='flex items-end gap-2'>
                  <div className='flex flex-1 flex-col gap-1'>
                    <Label className='text-xs'>{t('Tier amount')}</Label>
                    <Input
                      type='number'
                      min={0}
                      value={tier.amount}
                      onChange={(e) =>
                        updateTier(idx, {
                          amount: Number(e.target.value) || 0,
                        })
                      }
                    />
                  </div>
                  <div className='flex flex-1 flex-col gap-1'>
                    <Label className='text-xs'>{t('Bonus rate (%)')}</Label>
                    <Input
                      type='number'
                      min={0}
                      step={0.1}
                      value={tier.bonus_rate}
                      onChange={(e) =>
                        updateTier(idx, {
                          bonus_rate: Number(e.target.value) || 0,
                        })
                      }
                    />
                  </div>
                  <Button
                    type='button'
                    size='icon'
                    variant='ghost'
                    onClick={() => removeTier(idx)}
                    aria-label={t('Remove tier')}
                  >
                    <Trash2 className='h-4 w-4' />
                  </Button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Threshold gift */}
        <div
          data-settings-form-span='full'
          className='flex flex-col gap-3 border-t pt-4'
        >
          <SettingsSwitchField
            checked={gift.enabled}
            onCheckedChange={(v) =>
              setGift((prev) => ({ ...prev, enabled: v }))
            }
            label={t('Threshold gift')}
            description={t(
              'Grant a named gift when a single recharge meets the threshold'
            )}
          />
          {gift.enabled && (
            <div className='grid gap-3 sm:grid-cols-3'>
              <div className='flex flex-col gap-1'>
                <Label className='text-xs'>{t('Gift threshold')}</Label>
                <Input
                  type='number'
                  min={0}
                  value={gift.threshold}
                  onChange={(e) =>
                    setGift((prev) => ({
                      ...prev,
                      threshold: Number(e.target.value) || 0,
                    }))
                  }
                />
              </div>
              <div className='flex flex-col gap-1'>
                <Label className='text-xs'>{t('Gift name')}</Label>
                <Input
                  value={gift.name}
                  placeholder={t('bonus gift name')}
                  onChange={(e) =>
                    setGift((prev) => ({ ...prev, name: e.target.value }))
                  }
                />
              </div>
              <div className='flex flex-col gap-1'>
                <Label className='text-xs'>{t('Gift type')}</Label>
                <Input
                  value={gift.type}
                  placeholder='gift_type'
                  onChange={(e) =>
                    setGift((prev) => ({ ...prev, type: e.target.value }))
                  }
                />
              </div>
            </div>
          )}
        </div>
      </SettingsForm>
    </SettingsSection>
  )
}
