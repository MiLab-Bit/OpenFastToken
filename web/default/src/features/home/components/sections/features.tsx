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
import {
  Zap,
  Shield,
  Globe,
  Code,
  Gauge,
  DollarSign,
  Users,
  Headset,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { AnimateInView } from '@/components/animate-in-view'

interface FeaturesProps {
  className?: string
}

export function Features(_props: FeaturesProps) {
  const { t } = useTranslation()

  const features = [
    {
      id: 'fast',
      num: '01',
      title: t('Lightning Fast'),
      desc: t(
        'Optimized network architecture ensures millisecond response times'
      ),
      span: 'md:col-span-2',
      icon: <Zap className='size-4 text-accent-brand' />,
      visual: (
        <div className='mt-4 grid grid-cols-3 place-items-center gap-2'>
          {['DeepSeek-V4-Pro', 'Qwen3.5', 'Qwen3-Max', 'Qwen3-VL', 'Qwen-Flash', 'DeepSeek-V3'].map(
            (name) => (
              <div
                key={name}
                className='flex items-center justify-center rounded-lg border border-[#e2ded8] bg-stone-bg px-3 py-2 text-xs text-[#6b6b6b] transition-all duration-300 hover:border-accent-brand/30 hover:bg-accent-brand/5'
              >
                {name}
              </div>
            )
          )}
        </div>
      ),
    },
    {
      id: 'secure',
      num: '02',
      title: t('Secure & Reliable'),
      desc: t(
        'Enterprise-grade security with comprehensive permission management'
      ),
      span: 'md:col-span-1',
      icon: <Shield className='size-4 text-accent-brand' />,
      visual: (
        <div className='mt-4 flex items-center justify-center'>
          <div className='relative'>
            <div className='flex size-16 items-center justify-center rounded-2xl border border-accent-brand/20 bg-accent-brand/5'>
              <Shield
                className='size-7 text-accent-brand/70'
                strokeWidth={1.5}
              />
            </div>
            <div className='absolute -top-1 -right-1 flex size-4 items-center justify-center rounded-full bg-accent-brand'>
              <svg
                className='size-2.5 text-white'
                fill='none'
                viewBox='0 0 24 24'
                stroke='currentColor'
                strokeWidth={3}
              >
                <path
                  strokeLinecap='round'
                  strokeLinejoin='round'
                  d='m4.5 12.75 6 6 9-13.5'
                />
              </svg>
            </div>
          </div>
        </div>
      ),
    },
    {
      id: 'global',
      num: '03',
      title: t('Global Coverage'),
      desc: t('Multi-region deployment for stable global access'),
      span: 'md:col-span-2',
      icon: <Globe className='size-4 text-accent-brand' />,
      visual: (
        <div className='mt-4 space-y-2'>
          {[t('Load Balancing'), t('Rate Limiting'), t('Cost Tracking')].map(
            (step, i) => (
              <div key={step} className='flex items-center gap-2'>
                <div
                  className={`flex size-6 items-center justify-center rounded-full text-[10px] font-bold ${
                    i === 1
                      ? 'border border-accent-brand/30 bg-accent-brand/20 text-accent-brand'
                      : 'border border-[#e2ded8] bg-stone-bg text-[#6b6b6b]'
                  }`}
                >
                  {i + 1}
                </div>
                <div className='h-px flex-1 bg-[#e2ded8]' />
                <span className='text-xs text-[#6b6b6b]'>{step}</span>
              </div>
            )
          )}
        </div>
      ),
    },
    {
      id: 'developer',
      num: '04',
      title: t('Developer Friendly'),
      desc: t('Compatible API routes for common AI application workflows'),
      span: 'md:col-span-1',
      icon: <Code className='size-4 text-accent-brand' />,
      visual: (
        <div className='mt-4 flex items-center justify-center gap-3'>
          <div className='flex -space-x-2'>
            {['API', 'SDK', 'CLI', 'Docs'].map((n) => (
              <div
                key={n}
                className='flex size-8 items-center justify-center rounded-full border-2 border-white bg-gradient-to-br from-[#f4f2ed] to-[#e2ded8] text-[9px] font-bold text-[#6b6b6b]'
              >
                {n}
              </div>
            ))}
          </div>
          <div className='flex items-center gap-1.5 text-xs text-[#6b6b6b]'>
            <Code className='size-3.5 text-accent-brand' />
            {t('Multi-protocol Compatible')}
          </div>
        </div>
      ),
    },
  ]

  const additionalFeatures = [
    {
      icon: <Gauge className='size-5' strokeWidth={1.5} />,
      title: t('High Performance'),
      desc: t('Support for high concurrency with automatic load balancing'),
    },
    {
      icon: <DollarSign className='size-5' strokeWidth={1.5} />,
      title: t('Transparent Billing'),
      desc: t('Pay-as-you-go with real-time usage monitoring'),
    },
    {
      icon: <Users className='size-5' strokeWidth={1.5} />,
      title: t('Team Collaboration'),
      desc: t('Multi-user management with flexible permission allocation'),
    },
    {
      icon: <Headset className='size-5' strokeWidth={1.5} />,
      title: t('Enterprise Support'),
      desc: t('Professional team guarantee, SLA commitment, and dedicated technical support'),
    },
  ]

  return (
    <section className='relative z-10 border-t border-[#e2ded8] px-6 py-24 md:py-32'>
      <div className='mx-auto max-w-6xl'>
        {/* Section heading */}
        <AnimateInView className='mb-16 max-w-lg'>
          <p className='mb-3 text-xs font-medium uppercase tracking-widest text-[#6b6b6b]'>
            {t('Core Features')}
          </p>
          <h2 className='font-serif text-3xl font-bold leading-tight text-stone-text'>
            {t('Built for developers,')}{' '}{t('designed for scale')}
          </h2>
        </AnimateInView>

        {/* Bento grid — restyled with white cards */}
        <div className='grid gap-px overflow-hidden rounded-xl border border-[#e2ded8] bg-[#e2ded8] md:grid-cols-3'>
          {features.map((f, i) => (
            <AnimateInView
              key={f.id}
              delay={i * 100}
              animation='scale-in'
              className={`group border bg-white p-7 shadow-sm transition-all duration-300 hover:border-accent-brand md:p-8 ${f.span}`}
            >
              <div className='mb-3 flex items-center justify-center gap-3'>
                <span className='flex size-7 items-center justify-center rounded-md border border-[#e2ded8] bg-stone-bg text-[10px] font-semibold tabular-nums text-[#6b6b6b]'>
                  {f.num}
                </span>
                <h3 className='text-sm font-semibold text-stone-text'>{f.title}</h3>
              </div>
              <p className='text-center text-sm leading-relaxed text-[#6b6b6b]'>
                {f.desc}
              </p>
              {f.visual}
            </AnimateInView>
          ))}
        </div>

        {/* Additional features row */}
        <div className='mt-12 grid grid-cols-2 gap-8 md:grid-cols-4 md:gap-12'>
          {additionalFeatures.map((f, i) => (
            <AnimateInView
              key={f.title}
              delay={i * 100}
              animation='fade-up'
              className='flex flex-col items-center text-center'
            >
              <div className='mb-3 flex size-12 items-center justify-center rounded-xl border border-[#e2ded8] bg-white text-accent-brand shadow-sm transition-colors group-hover:text-stone-text'>
                {f.icon}
              </div>
              <h3 className='mb-1.5 text-sm font-semibold text-stone-text'>{f.title}</h3>
              <p className='max-w-[200px] text-xs leading-relaxed text-[#6b6b6b]'>
                {f.desc}
              </p>
            </AnimateInView>
          ))}
        </div>
      </div>
    </section>
  )
}
