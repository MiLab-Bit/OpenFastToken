import { useTranslation } from 'react-i18next'
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

You should have received received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@fasttoken.com
*/
import { useState } from 'react'
import {
  Shield,
  BarChart3,
  Key,
  Network,
  Globe,
  Rocket,
  Lock,
  Users,
  TrendingUp,
  Sparkles,
  Building2,
} from 'lucide-react'
import { PublicLayout } from '@/components/layout'

interface FeatureCard {
  icon: React.ComponentType<{ className?: string }>
  title: string
  desc: string
  color: string
  bgColor: string
}

// featureCards built inside component with t()

// advantages built inside component with t()

// stats built inside component with t()

export function About() {
  const { t } = useTranslation()

  const [activeTab, setActiveTab] = useState<'features' | 'advantages' | 'brand'>('features')

  // i18n-aware data (must be inside component for t())
  const featureCards: FeatureCard[] = [
    {
      icon: Globe,
      title: t('about.featureUnifiedGateway.title'),
      desc: t('about.featureUnifiedGateway.desc'),
      color: 'text-blue-600',
      bgColor: 'bg-blue-50',
    },
    {
      icon: Shield,
      title: t('about.featureEnterpriseSecurity.title'),
      desc: t('about.featureEnterpriseSecurity.desc'),
      color: 'text-purple-600',
      bgColor: 'bg-purple-50',
    },
    {
      icon: BarChart3,
      title: t('about.featureGranularBilling.title'),
      desc: t('about.featureGranularBilling.desc'),
      color: 'text-cyan-600',
      bgColor: 'bg-cyan-50',
    },
    {
      icon: Key,
      title: t('about.featureKeyManagement.title'),
      desc: t('about.featureKeyManagement.desc'),
      color: 'text-emerald-600',
      bgColor: 'bg-emerald-50',
    },
    {
      icon: Network,
      title: t('about.featureSmartRouting.title'),
      desc: t('about.featureSmartRouting.desc'),
      color: 'text-orange-600',
      bgColor: 'bg-orange-50',
    },
    {
      icon: Sparkles,
      title: t('about.featureMultiModel.title'),
      desc: t('about.featureMultiModel.desc'),
      color: 'text-pink-600',
      bgColor: 'bg-pink-50',
    },
  ]

  const advantages = [
    {
      icon: Lock,
      title: t('about.advantageSecureReliable.title'),
      desc: t('about.advantageSecureReliable.desc'),
      color: 'text-blue-600',
      bgColor: 'bg-blue-50',
    },
    {
      icon: Users,
      title: t('about.advantageMultiTenant.title'),
      desc: t('about.advantageMultiTenant.desc'),
      color: 'text-purple-600',
      bgColor: 'bg-purple-50',
    },
    {
      icon: TrendingUp,
      title: t('about.advantagePerformance.title'),
      desc: t('about.advantagePerformance.desc'),
      color: 'text-cyan-600',
      bgColor: 'bg-cyan-50',
    },
    {
      icon: Rocket,
      title: t('about.advantageInnovation.title'),
      desc: t('about.advantageInnovation.desc'),
      color: 'text-emerald-600',
      bgColor: 'bg-emerald-50',
    },
  ]

  const stats = [
    { label: t('about.statsSupportedModels'), value: '50+', suffix: t('about.statsSuffixUnit') },
    { label: t('about.statsChannelTypes'), value: '30+', suffix: t('about.statsSuffixKind') },
    { label: t('about.statsResponseTime'), value: '<50', suffix: t('about.statsSuffixMs') },
    { label: t('about.statsUptime'), value: '99.9', suffix: t('about.statsSuffixPercent') },
  ]

  return (
    <PublicLayout>
      {/* 渐变背景 + 动态装饰 */}
      <div className="min-h-screen relative overflow-hidden">
        {/* 渐变背景 */}
        <div className="absolute inset-0 bg-gradient-to-br from-blue-50 via-white to-purple-50 
                        dark:from-slate-950 dark:via-slate-900 dark:to-blue-950" />
        
        {/* 动态装饰球 */}
        <div className="absolute -top-40 -right-40 w-96 h-96 bg-blue-400 rounded-full 
                        mix-blend-multiply filter blur-3xl opacity-20 animate-blob" />
        <div className="absolute -bottom-40 -left-40 w-96 h-96 bg-purple-400 rounded-full 
                        mix-blend-multiply filter blur-3xl opacity-20 animate-blob animation-delay-2000" />
        <div className="absolute top-40 left-1/2 w-96 h-96 bg-cyan-400 rounded-full 
                        mix-blend-multiply filter blur-3xl opacity-20 animate-blob animation-delay-4000" />
        
        {/* Hero Section */}
        <div className="relative bg-background/80 backdrop-blur-sm border-b border-border/60">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24 sm:py-32">
            <div className="text-center">
              <h1 className="text-4xl font-bold tracking-tight text-foreground sm:text-5xl lg:text-6xl">{t('about.hero.prefix')}<span className="text-gradient-blue">FastToken</span>
              </h1>
              <p className="mt-6 text-lg leading-8 text-muted-foreground max-w-2xl mx-auto">{t('about.hero.subtitle')}</p>
              
              {/* 统计数字 */}
              <div className="mt-12 grid grid-cols-2 gap-8 sm:grid-cols-4">
                {stats.map((stat, idx) => (
                  <div
                    key={idx}
                    className="backdrop-blur-lg bg-card/60 
                               rounded-2xl p-6 shadow-xl border border-white/20
                               hover:scale-105 transition-transform duration-300"
                  >
                    <div className="text-4xl font-bold text-gradient-blue">{stat.value}</div>
                    <div className="text-sm text-muted-foreground mt-1">{stat.label}</div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Tabs */}
        <div className="sticky top-16 z-10 bg-background/90 backdrop-blur-md border-b border-border/60">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex space-x-8">
              {[
                { key: 'features', label: t('about.tabCoreFunctions') },
                { key: 'advantages', label: t('about.tabProductAdvantages') },
                { key: 'brand', label: t('about.tabBrandInfo') },
              ].map((tab) => (
                <button
                  key={tab.key}
                  onClick={() => setActiveTab(tab.key as typeof activeTab)}
                  className={`py-4 px-1 border-b-2 font-medium text-sm transition-all duration-300 ${
                    activeTab === tab.key
                      ? 'border-blue-600 text-blue-600 dark:border-blue-400 dark:text-blue-400 scale-105'
                      : 'border-transparent text-muted-foreground hover:text-foreground hover:scale-105'
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Content */}
        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
          {activeTab === 'features' && (
            <div>
              <h2 className="text-3xl font-bold text-foreground mb-8 text-gradient-blue inline-block">{t('about.section.features')}</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {featureCards.map((feature, index) => (
                  <div
                    key={index}
                    className="relative group bg-card rounded-2xl p-6 
                               shadow-sm border border-border 
                               hover:shadow-2xl hover:-translate-y-2 transition-all duration-300
                               overflow-hidden"
                  >
                    {/* 光泽效果 */}
                    <div className="absolute inset-0 opacity-0 group-hover:opacity-100 
                                    transition-opacity duration-300
                                    bg-gradient-to-r from-transparent via-white/10 to-transparent 
                                    -skew-x-12 group-hover:animate-shine" />
                    
                    <div className="flex items-start space-x-4">
                      <div className="flex-shrink-0">
                        <div className={`w-12 h-12 ${feature.bgColor} dark:bg-opacity-20 rounded-xl 
                                        flex items-center justify-center 
                                        group-hover:scale-110 transition-transform duration-300`}>
                          <feature.icon className={`w-6 h-6 ${feature.color}`} />
                        </div>
                      </div>
                      <div className="flex-1">
                        <h3 className="text-lg font-semibold text-foreground mb-2">
                          {feature.title}
                        </h3>
                        <p className="text-sm text-muted-foreground leading-relaxed">
                          {feature.desc}
                        </p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {activeTab === 'advantages' && (
            <div>
              <h2 className="text-3xl font-bold text-foreground mb-8 text-gradient-blue inline-block">{t('about.section.advantages')}</h2>
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                {advantages.map((adv, index) => (
                  <div
                    key={index}
                    className="bg-card rounded-2xl p-8 
                               shadow-sm border border-border
                               hover:shadow-2xl hover:-translate-y-2 transition-all duration-300"
                  >
                    <div className="flex items-start space-x-4">
                      <div className="flex-shrink-0">
                        <div className={`w-14 h-14 ${adv.bgColor} dark:bg-opacity-20 rounded-xl 
                                        flex items-center justify-center`}>
                          <adv.icon className={`w-7 h-7 ${adv.color}`} />
                        </div>
                      </div>
                      <div className="flex-1">
                        <h3 className="text-xl font-semibold text-foreground mb-3">
                          {adv.title}
                        </h3>
                        <p className="text-muted-foreground leading-relaxed">
                          {adv.desc}
                        </p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {activeTab === 'brand' && (
            <div>
              <h2 className="text-3xl font-bold text-foreground mb-8 text-gradient-blue inline-block">{t('about.section.brand')}</h2>
              
              {/* Company Info */}
              <div className="bg-card rounded-2xl p-8 shadow-sm border border-border mb-8
                             hover:shadow-2xl transition-all duration-300">
                <div className="flex items-start space-x-4 mb-6">
                  <div className="w-16 h-16 bg-gradient-to-br from-blue-600 to-purple-600 rounded-2xl flex items-center justify-center
                                  shadow-lg shadow-blue-500/30">
                    <Building2 className="w-8 h-8 text-white" />
                  </div>
                  <div>
                    <h3 className="text-2xl font-bold text-foreground">{t('about.brand.company')}</h3>
                    <p className="text-muted-foreground mt-1">{t('about.brand.companyTagline')}</p>
                  </div>
                </div>
                <p className="text-muted-foreground leading-relaxed">{t('about.brand.companyDesc')}</p>
              </div>

              {/* Product Info */}
              <div className="bg-card rounded-2xl p-8 shadow-sm border border-border mb-8
                             hover:shadow-2xl transition-all duration-300">
                <h3 className="text-xl font-semibold text-foreground mb-6">{t('about.brand.productTitle')}</h3>
                <div className="space-y-4">
                  <div className="flex justify-between py-3 border-b border-border">
                    <span className="text-muted-foreground">{t('about.brand.productNameLabel')}</span>
                    <span className="font-medium text-gradient-blue">FastToken</span>
                  </div>
                  <div className="flex justify-between py-3 border-b border-border">
                    <span className="text-muted-foreground">{t('about.brand.domainLabel')}</span>
                    <a href="https://openfasttoken.example" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:text-blue-700 dark:text-blue-400">
                      https://openfasttoken.example
                    </a>
                  </div>
                  <div className="flex justify-between py-3 border-b border-border">
                    <span className="text-muted-foreground">{t('about.brand.docsLabel')}</span>
                    <a href="https://openfasttoken.example/docs" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:text-blue-700 dark:text-blue-400">
                      https://openfasttoken.example/docs
                    </a>
                  </div>
                  <div className="flex justify-between py-3 border-b border-border">
                    <span className="text-muted-foreground">{t('about.brand.licenseLabel')}</span>
                    <span className="font-medium text-foreground">{t('about.brand.licenseValue')}</span>
                  </div>
                  <div className="flex justify-between py-3 border-b border-slate-700">
                    <span className="text-muted-foreground">{t('about.brand.sourceLabel')}</span>
                    <a href="https://github.com/MiLab-Bit/OpenFastToken" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:text-blue-700 dark:text-blue-400">
                      https://github.com/MiLab-Bit/OpenFastToken
                    </a>
                  </div>
                </div>
              </div>

              {/* Copyright */}
              <div className="bg-slate-50 dark:bg-slate-800/50 rounded-2xl p-6 text-center">
                <p className="text-sm text-muted-foreground">{t('about.copyright.notice')}</p>
                <p className="text-sm text-muted-foreground mt-1">{t('about.copyright.dev')}</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </PublicLayout>
  )
}
