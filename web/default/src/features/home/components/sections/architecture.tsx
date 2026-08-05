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
import { useTranslation } from 'react-i18next'
import { AnimateInView } from '@/components/animate-in-view'

const telcoItems = [
  'Model Serving — DeepSeek-V4-Pro, Qwen3.5 & more',
  'Standard OpenAI-compatible API endpoint',
  'GPU compute clusters, elastic inference scaling',
  'Carrier-grade backbone, mainland low-latency delivery',
]

const platformItems = [
  'One API key → all models across every channel',
  'Smart routing, weighted LB, auto failover between providers',
  'Rate limiting, concurrency control & abuse prevention',
  'Request-level observability, per-call log analysis',
]

export function Architecture() {
  const { t } = useTranslation()

  return (
    <section className='relative z-10 border-t border-[#e2ded8] px-6 py-24 md:py-32'>
      <div className='mx-auto max-w-6xl'>

        {/* Header */}
        <AnimateInView className='mb-16 text-center md:mb-20'>
          <p className='mb-3 text-xs font-medium uppercase tracking-widest text-[#6b6b6b]'>
            {t('Powered By')}
          </p>
          <h2 className='font-serif text-3xl font-bold text-stone-text'>
            {t('China Telecom Cloud delivers the models')}
            <br className='md:hidden' />
            <span className='hidden text-[#6b6b6b] md:inline'> — </span>
            {t('FastToken delivers the developer experience')}
          </h2>
        </AnimateInView>

        {/* Two-column comparison */}
        <div className='grid gap-8 md:grid-cols-2 md:gap-12'>

          {/* --- China Telecom Cloud --- */}
          <AnimateInView
            delay={100}
            animation='fade-up'
            className='rounded-2xl border border-[#e2ded8] bg-white p-8 shadow-sm transition-all hover:border-accent-brand'
          >
            <div className='mb-6 flex items-center gap-3'>
              <img
                src='/china-telecom-logo.jpg'
                alt='China Telecom Cloud'
                className='h-8 w-auto'
              />
              <span className='text-xs font-medium uppercase tracking-wider text-[#6b6b6b]'>
                {t('Model Layer')}
              </span>
            </div>
            <h3 className='mb-4 font-serif text-lg font-bold text-stone-text'>
              {t('China Telecom AI Gateway')}
            </h3>
            <ul className='space-y-3'>
              {telcoItems.map((item) => (
                <li key={item} className='flex items-start gap-2.5 text-sm leading-relaxed text-[#6b6b6b]'>
                  <span className='mt-1.5 block size-1.5 shrink-0 rounded-full bg-accent-brand' />
                  {t(item)}
                </li>
              ))}
            </ul>
          </AnimateInView>

          {/* --- FastToken --- */}
          <AnimateInView
            delay={250}
            animation='fade-up'
            className='rounded-2xl border border-[#e2ded8] bg-white p-8 shadow-sm transition-all hover:border-accent-brand'
          >
            <div className='mb-6 flex items-center gap-3'>
              <img
                src='/logo.svg'
                alt='FastToken'
                className='h-8 w-auto'
              />
              <span className='text-xs font-medium uppercase tracking-wider text-[#6b6b6b]'>
                {t('SaaS Layer')}
              </span>
            </div>
            <h3 className='mb-4 font-serif text-lg font-bold text-stone-text'>
              {t('FastToken')}
            </h3>
            <ul className='space-y-3'>
              {platformItems.map((item) => (
                <li key={item} className='flex items-start gap-2.5 text-sm leading-relaxed text-[#6b6b6b]'>
                  <span className='mt-1.5 block size-1.5 shrink-0 rounded-full bg-accent-brand' />
                  {t(item)}
                </li>
              ))}
            </ul>
          </AnimateInView>

        </div>

        {/* Bottom formula */}
        <AnimateInView delay={500} animation='fade-up' className='mt-12 text-center'>
          <div className='inline-flex items-center gap-3 rounded-full border border-[#e2ded8] bg-white px-5 py-2.5 text-sm shadow-sm'>
            <span className='font-semibold text-accent-brand'>{t('China Telecom Models')}</span>
            <span className='text-lg text-[#6b6b6b]'>+</span>
            <span className='font-semibold text-accent-brand'>{t('FastToken Platform')}</span>
            <span className='text-[#6b6b6b]'>=</span>
            <span className='font-semibold text-stone-text'>{t('Production-ready AI API')}</span>
          </div>
        </AnimateInView>

      </div>
    </section>
  )
}
