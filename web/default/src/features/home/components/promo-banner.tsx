/*
Copyright (C) 2023-2026 FastToken

活动宣传横幅组件 - Portfolio风格UI
支持动态内容配置
*/
import { useState, useEffect } from 'react'
import { X, Gift, Sparkles, ExternalLink } from 'lucide-react'
import { motion, AnimatePresence } from 'motion/react'
import { Button } from '@/components/ui/button'
import { useSystemConfig } from '@/hooks/use-system-config'

export interface PromoConfig {
  enabled: boolean
  title: string
  subtitle: string
  ctaText: string
  ctaLink: string
  badgeText: string
  partnerText: string
}

// 运营配置：活动默认满额充值金额（元）。该值可被系统设置 promoBanner 覆盖。
const PROMO_DEFAULT_RECHARGE_AMOUNT_CNY = 1000

const defaultConfig: PromoConfig = {
  enabled: false,
  title: `单次充值${PROMO_DEFAULT_RECHARGE_AMOUNT_CNY}元`,
  subtitle: '充值满额享额外配额赠送',
  ctaText: '立即充值',
  ctaLink: '/wallet',
  badgeText: '限时活动',
  partnerText: '电信 × FastToken 联合出品',
}

// Portfolio stone palette colors
const stoneColors = {
  bg: 'from-[#1c1c1c] via-[#1c1c1c] to-[#af4028]',
  text: 'text-white',
  badge: 'bg-primary text-primary-foreground',
  accent: 'bg-accent-brand',
}

// 举牌小人SVG组件 — restyled for stone palette
function CharacterMascot() {
  return (
    <motion.svg
      width='80'
      height='100'
      viewBox='0 0 80 100'
      className='absolute -left-2 top-1/2 -translate-y-1/2'
      initial={{ y: -50, opacity: 0, rotate: -10 }}
      animate={{ y: -50, opacity: 1, rotate: 0 }}
      transition={{ type: 'spring', stiffness: 200, damping: 15 }}
    >
      {/* Body */}
      <motion.ellipse
        cx='40'
        cy='70'
        rx='20'
        ry='25'
        fill='white'
        initial={{ scale: 0 }}
        animate={{ scale: 1 }}
        transition={{ delay: 0.2 }}
      />
      {/* Head */}
      <motion.circle
        cx='40'
        cy='35'
        r='18'
        fill='white'
        initial={{ scale: 0 }}
        animate={{ scale: 1 }}
        transition={{ delay: 0.1 }}
      />
      {/* Eyes */}
      <motion.circle
        cx='34'
        cy='32'
        r='3'
        fill='#1c1c1c'
        initial={{ scale: 0 }}
        animate={{ scale: 1 }}
        transition={{ delay: 0.3 }}
      />
      <motion.circle
        cx='46'
        cy='32'
        r='3'
        fill='#1c1c1c'
        initial={{ scale: 0 }}
        animate={{ scale: 1 }}
        transition={{ delay: 0.3 }}
      />
      {/* Sign board */}
      <motion.rect
        x='55'
        y='45'
        width='20'
        height='30'
        rx='3'
        fill='#af4028'
        initial={{ x: 20, opacity: 0 }}
        animate={{ x: 0, opacity: 1 }}
        transition={{ delay: 0.4, type: 'spring' }}
      />
      {/* Sign text */}
      <motion.text
        x='65'
        y='62'
        fontSize='8'
        fill='white'
        textAnchor='middle'
        fontWeight='bold'
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.6 }}
      >
        PROMO
      </motion.text>
      {/* Star decorations */}
      <motion.circle
        cx='15'
        cy='20'
        r='2'
        fill='#af4028'
        animate={{ scale: [1, 1.5, 1], opacity: [0.5, 1, 0.5] }}
        transition={{ repeat: Infinity, duration: 2 }}
      />
      <motion.circle
        cx='70'
        cy='15'
        r='1.5'
        fill='#af4028'
        animate={{ scale: [1, 1.3, 1], opacity: [0.3, 0.8, 0.3] }}
        transition={{ repeat: Infinity, duration: 1.5, delay: 0.5 }}
      />
    </motion.svg>
  )
}

// Particle effect — unchanged, already white dots
function SparkleEffect() {
  return (
    <div className='pointer-events-none absolute inset-0 overflow-hidden'>
      {[...Array(6)].map((_, i) => (
        <motion.div
          key={i}
          className='absolute h-1 w-1 rounded-full bg-white'
          style={{
            left: `${20 + i * 15}%`,
            top: `${30 + (i % 2) * 40}%`,
          }}
          animate={{
            scale: [0, 1, 0],
            opacity: [0, 1, 0],
            y: [0, -20],
          }}
          transition={{
            duration: 2,
            repeat: Infinity,
            delay: i * 0.3,
          }}
        />
      ))}
    </div>
  )
}

export function PromoBanner() {
  const [isVisible, setIsVisible] = useState(true)
  const [isDismissed, setIsDismissed] = useState(false)
  const { promoBanner } = useSystemConfig()

  // Read promo config from system config
  const promoConfig: PromoConfig = {
    ...defaultConfig,
    ...(promoBanner || {}),
  }

  // Check local storage for dismissal
  useEffect(() => {
    const dismissed = localStorage.getItem('promo-banner-dismissed')
    if (dismissed) {
      const dismissedTime = parseInt(dismissed)
      // Re-display after 24 hours
      if (Date.now() - dismissedTime < 24 * 60 * 60 * 1000) {
        setIsDismissed(true)
      }
    }
  }, [])

  const handleDismiss = () => {
    setIsVisible(false)
    localStorage.setItem('promo-banner-dismissed', Date.now().toString())
    setTimeout(() => setIsDismissed(true), 300)
  }

  if (!promoConfig.enabled || isDismissed) {
    return null
  }

  return (
    <AnimatePresence>
      {isVisible && (
        <motion.div
          initial={{ height: 0, opacity: 0 }}
          animate={{ height: 'auto', opacity: 1 }}
          exit={{ height: 0, opacity: 0 }}
          transition={{ duration: 0.3 }}
          className='relative w-full overflow-hidden'
        >
          <div
            className={`relative bg-gradient-to-r ${stoneColors.bg} py-3 px-4 shadow-lg sm:px-6 lg:px-8`}
          >
            {/* Background decoration */}
            <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAiIGhlaWdodD0iMjAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PGNpcmNsZSBjeD0iMiIgY3k9IjIiIHI9IjEiIGZpbGw9InJnYmEoMjU1LDI1NSwyNTUsMC4xKSIvPjwvc3ZnPg==')] opacity-30" />

            {/* Particle effect */}
            <SparkleEffect />

            {/* Character mascot */}
            <CharacterMascot />

            {/* Content area */}
            <div className='relative z-10 flex items-center justify-center gap-3 pl-16 pr-8'>
              {/* Badge */}
              <motion.span
                className={`hidden sm:inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-bold shadow-sm ${stoneColors.badge}`}
                animate={{ scale: [1, 1.05, 1] }}
                transition={{ repeat: Infinity, duration: 2 }}
              >
                <Sparkles className='h-3 w-3' />
                {promoConfig.badgeText}
              </motion.span>

              {/* Main title */}
              <div className='flex flex-col gap-1 sm:flex-row sm:items-center sm:gap-3'>
                <span className={`text-sm font-bold sm:text-base ${stoneColors.text}`}>
                  {promoConfig.title}
                </span>
                <span className={`text-xs sm:text-sm ${stoneColors.text} opacity-90 cursor-pointer hover:underline`}>
                  <span onClick={() => {
                    if (promoConfig.ctaLink.startsWith('/')) {
                      window.location.href = promoConfig.ctaLink
                    } else {
                      window.open(promoConfig.ctaLink, '_blank')
                    }
                  }}>
                    {promoConfig.subtitle}
                  </span>
                </span>
              </div>

              {/* CTA button */}
              <motion.div
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
              >
                <Button
                  size='sm'
                  className='h-7 bg-white px-3 py-1 text-xs font-bold text-stone-text shadow-md transition-all hover:bg-stone-bg hover:shadow-lg'
                  onClick={() => {
                    if (promoConfig.ctaLink.startsWith('/')) {
                      window.location.href = promoConfig.ctaLink
                    } else {
                      window.open(promoConfig.ctaLink, '_blank')
                    }
                  }}
                >
                  <Gift className='mr-1 h-3 w-3' />
                  {promoConfig.ctaText}
                  {!promoConfig.ctaLink.startsWith('/') && <ExternalLink className='ml-1 h-3 w-3' />}
                </Button>
              </motion.div>
            </div>

            {/* Partner label */}
            <div className={`absolute bottom-1 right-16 hidden text-[10px] ${stoneColors.text} opacity-70 sm:block`}>
              {promoConfig.partnerText}
            </div>

            {/* Close button */}
            <button
              onClick={handleDismiss}
              className={`absolute right-2 top-1/2 -translate-y-1/2 rounded-full p-1 ${stoneColors.text} opacity-70 transition-opacity hover:opacity-100`}
              aria-label='关闭'
            >
              <X className='h-4 w-4' />
            </button>
          </div>
        </motion.div>
      )}
    </AnimatePresence>
  )
}
