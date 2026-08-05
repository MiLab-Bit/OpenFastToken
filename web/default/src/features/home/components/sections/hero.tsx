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
import { Link } from '@tanstack/react-router'
import { ArrowRight } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { HeroBrandShowcase } from '../hero-brand-showcase'

interface HeroProps {
  className?: string
  isAuthenticated?: boolean
}

export function Hero(props: HeroProps) {
  const { t } = useTranslation()

  return (
    <section className='relative z-10 flex min-h-screen flex-col items-center justify-center bg-stone-bg px-6 text-center'>
      <div className='flex max-w-3xl flex-col items-center'>
        {/* Title — Portfolio serif hero style */}
        <h1
          className='landing-animate-fade-up font-serif text-6xl font-bold tracking-tighter text-stone-text md:text-8xl'
          style={{ animationDelay: '0ms' }}
        >
          FastToken
        </h1>

        {/* Subtitle — italic serif muted */}
        <p
          className='landing-animate-fade-up mt-6 font-serif text-xl italic text-[#6b6b6b] opacity-0 md:text-2xl'
          style={{ animationDelay: '80ms' }}
        >
          {t('AI API Gateway & Unified Management')}
        </p>

        {/* Accent divider */}
        <div
          className='landing-animate-fade-up my-12 h-px w-24 bg-accent-brand opacity-0'
          style={{ animationDelay: '120ms' }}
        />

        {/* Buttons — restyled to stone palette */}
        <div
          className='landing-animate-fade-up flex items-center gap-3 opacity-0'
          style={{ animationDelay: '160ms' }}
        >
          {props.isAuthenticated ? (
            <Button
              className='group rounded-lg bg-stone-text text-white transition-all hover:bg-accent-brand'
              render={<Link to='/dashboard' />}
            >
              {t('Go to Dashboard')}
              <ArrowRight className='ml-1 size-3.5 transition-transform duration-200 group-hover:translate-x-0.5' />
            </Button>
          ) : (
            <>
              <Button
                className='group rounded-lg bg-stone-text text-white transition-all hover:bg-accent-brand'
                render={<Link to='/sign-up' />}
              >
                {t('Get Started')}
                <ArrowRight className='ml-1 size-3.5 transition-transform duration-200 group-hover:translate-x-0.5' />
              </Button>
              <Button
                variant='outline'
                className='rounded-lg border-stone-text/20 bg-white text-stone-text transition-all hover:border-accent-brand hover:bg-white hover:text-accent-brand'
                render={<Link to='/pricing' />}
              >
                {t('View Pricing')}
              </Button>
            </>
          )}
        </div>
      </div>

      {/* Brand showcase */}
      <div
        className='landing-animate-fade-up w-full opacity-0'
        style={{ animationDelay: '300ms' }}
      >
        <HeroBrandShowcase />
      </div>

      {/* Scroll indicator */}
      <div className='absolute bottom-8 animate-bounce text-sm text-[#6b6b6b]'>
        {t('向下滚动探索')}
      </div>
    </section>
  )
}
