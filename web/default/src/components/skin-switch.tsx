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
import { Check, Palette } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'
import { useSkin } from '@/context/skin-provider'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'

/**
 * 顶栏「主题」切换器：与设置面板内的皮肤选择器共用 SkinProvider，
 * 直接驱动 <html data-skin>，不再提供浅色/深色/系统三色切换。
 */
export function SkinSwitch() {
  const { t } = useTranslation()
  const { skin, skins, setSkin } = useSkin()

  return (
    <DropdownMenu modal={false}>
      <DropdownMenuTrigger
        render={<Button variant='ghost' size='icon' className='h-9 w-9' />}
      >
        <Palette className='size-[1.2rem]' />
        <span className='sr-only'>{t('Theme')}</span>
      </DropdownMenuTrigger>
      <DropdownMenuContent align='end'>
        {skins.map((s) => (
          <DropdownMenuItem key={s.id} onClick={() => setSkin(s.id)}>
            {t(`skin.${s.id}.name`, s.name)}
            <Check
              size={14}
              className={cn('ms-auto', skin !== s.id && 'hidden')}
            />
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
