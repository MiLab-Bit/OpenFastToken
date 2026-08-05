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
import { Shield, Zap, Lock, Globe } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { cn } from '@/lib/utils'

interface HeroBrandShowcaseProps {
  className?: string
}

export function HeroBrandShowcase({ className }: HeroBrandShowcaseProps) {
  const { t } = useTranslation()

  return (
    <div
      className={cn(
        'mx-auto mt-16 w-full max-w-3xl',
        className
      )}
    >
      <div
        className={cn(
          'overflow-hidden rounded-2xl border bg-white shadow-sm',
          'border-[#e2ded8]'
        )}
      >
        {/* Top decorative bar — accent color */}
        <div className='h-1 w-full bg-accent-brand' />

        {/* Brand display area */}
        <div className='flex flex-col items-center px-8 py-12 md:py-16'>
          {/* Dual brand side-by-side */}
          <div className='flex flex-col items-center gap-6 md:flex-row md:items-center md:gap-10 lg:gap-16'>
            {/* China Telecom Logo */}
            <div className='flex items-center justify-center'>
              <img
                src='/china-telecom-logo.jpg'
                alt='中国电信'
                className='h-24 w-auto object-contain md:h-32 lg:h-36'
              />
            </div>

            {/* Joint branding badge */}
            <div className='flex flex-col items-center gap-2 text-[#6b6b6b] md:flex-row'>
              <div className='hidden h-px w-12 bg-gradient-to-r from-transparent to-[#e2ded8] md:block' />
              <span className='rounded-full border border-[#e2ded8] bg-stone-bg px-4 py-1.5 text-xs font-medium tracking-wider text-[#6b6b6b]'>
                {t('联合打造')}
              </span>
              <div className='hidden h-px w-12 bg-gradient-to-l from-transparent to-[#e2ded8] md:block' />
            </div>

            {/* Zhiqihui Logo */}
            <div className='flex items-center justify-center'>
              <img
                src='/zhiqihui-logo.jpg'
                alt='FastToken'
                className='h-20 w-auto object-contain md:h-24 lg:h-28'
              />
            </div>
          </div>

          {/* Tagline */}
          <div className='mt-8 text-center md:mt-10'>
            <p className='mx-auto max-w-2xl text-sm leading-relaxed text-[#6b6b6b] md:text-base lg:text-lg'>
              {t('中国电信与FastToken强强联合，为企业提供安全、稳定、高效的 AI 服务基础设施，助力数字化转型')}
            </p>
          </div>

          {/* Core advantage tags */}
          <div className='mt-6 flex flex-wrap items-center justify-center gap-2 px-4 md:mt-8 md:gap-3'>
            <AdvantageTag Icon={Shield} text={t('企业级安全')} />
            <AdvantageTag Icon={Zap} text={t('低延迟响应')} />
            <AdvantageTag Icon={Lock} text={t('数据合规')} />
            <AdvantageTag Icon={Globe} text={t('全国覆盖')} />
          </div>
        </div>

        {/* Bottom service commitment bar */}
        <div className='border-t border-[#e2ded8] bg-stone-bg px-6 py-4'>
          <div className='flex flex-wrap items-center justify-center gap-4 text-xs text-[#6b6b6b] md:gap-6'>
            <span className='flex items-center gap-1.5'>
              <span className='inline-block h-1.5 w-1.5 rounded-full bg-accent-brand' />
              {t('7×24 小时服务')}
            </span>
            <span className='flex items-center gap-1.5'>
              <span className='inline-block h-1.5 w-1.5 rounded-full bg-accent-brand' />
              {t('电信级稳定性')}
            </span>
            <span className='flex items-center gap-1.5'>
              <span className='inline-block h-1.5 w-1.5 rounded-full bg-accent-brand' />
              {t('专业运维支持')}
            </span>
          </div>
        </div>
      </div>
    </div>
  )
}

function AdvantageTag({ Icon, text }: { Icon: React.ComponentType<{ className?: string }>; text: string }) {
  return (
    <div className='flex items-center gap-1.5 rounded-full border border-[#e2ded8] bg-white px-3 py-1.5 text-sm text-stone-text'>
      <Icon className='size-3.5 text-accent-brand' />
      <span>{text}</span>
    </div>
  )
}
