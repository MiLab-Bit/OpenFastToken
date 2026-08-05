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
import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'
import {
  ArrowRight,
  BookOpen,
  Check,
  ChevronDown,
  ChevronUp,
  Circle,
  CreditCard,
  FileText,
  KeyRound,
  ListChecks,
  RadioTower,
  ShieldCheck,
  TerminalSquare,
  Bot,
  Globe,
  type LucideIcon,
} from 'lucide-react'
import { motion, useReducedMotion } from 'motion/react'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/auth-store'
import { getUserModels } from '@/lib/api'
import { MOTION_TRANSITION } from '@/lib/motion'

type CodeLanguage = 'python' | 'javascript' | 'curl'
import { ROLE } from '@/lib/roles'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { CopyButton } from '@/components/copy-button'
import {
  CardStaggerContainer,
  CardStaggerItem,
} from '@/components/page-transition'
import { fetchTokenKey, getApiKeys } from '@/features/keys/api'
import type { ApiKey } from '@/features/keys/types'
import {
  useApiInfo,
  useDashboardContentVisibility,
} from '../../hooks/use-status-data'
import { AnnouncementsPanel } from './announcements-panel'
import { ApiInfoPanel } from './api-info-panel'
import { FAQPanel } from './faq-panel'
import { PerformanceHealthPanel } from './performance-health-panel'
import { SummaryCards } from './summary-cards'
import { UptimePanel } from './uptime-panel'

const SETUP_GUIDE_VISIBILITY_STORAGE_KEY =
  'dashboard_overview_setup_guide_expanded'

const SETUP_GUIDE_CODE_PATTERN = [
  'const request = await client.responses.create({',
  "  model: 'gpt-4.1-mini',",
  "  input: 'Start routing traffic',",
  '})',
  '',
  'if (request.output_text) {',
  '  console.log(request.output_text)',
  '}',
].join('\n')

type DashboardActionPath =
  | '/keys'
  | '/wallet'
  | '/playground'
  | '/channels'
  | '/usage-logs'
  | '/pricing'

interface StartStep {
  title: string
  description: string
  to: DashboardActionPath
  icon: LucideIcon
  completed: boolean
}

interface QuickAction {
  title: string
  description: string
  to: DashboardActionPath
  icon: LucideIcon
  adminOnly?: boolean
}

interface ApiConfigGuide {
  baseUrl: string
  models: string[]
  hasApiKey: boolean
  keyName: string
}

interface HeroSignal {
  label: string
  value: string
  icon: LucideIcon
}

function getSavedSetupGuideExpanded(): boolean | null {
  if (typeof window === 'undefined') return null
  const saved = window.localStorage.getItem(SETUP_GUIDE_VISIBILITY_STORAGE_KEY)
  if (saved === 'expanded') return true
  if (saved === 'collapsed') return false
  return null
}

function saveSetupGuideExpanded(expanded: boolean): void {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(
    SETUP_GUIDE_VISIBILITY_STORAGE_KEY,
    expanded ? 'expanded' : 'collapsed'
  )
}

function getCurrentOrigin(): string {
  if (typeof window === 'undefined') return ''
  return window.location.origin
}

function getPreferredKey(keys: ApiKey[]): ApiKey | null {
  return keys.find((item) => item.status === 1) ?? keys[0] ?? null
}

function buildSafeCodeExamples(args: {
  baseUrl: string
  model: string
}): { python: string; javascript: string; curl: string } {
  const cleanBase = args.baseUrl.replace(/\/v1\/?$/, '')
  const fullUrl = `${cleanBase}/v1/chat/completions`
  
  return {
    python: `from openai import OpenAI

client = OpenAI(
    base_url="${args.baseUrl}",
    api_key="YOUR_API_KEY"  # 在 API Keys 页面创建
)

response = client.chat.completions.create(
    model="${args.model}",
    messages=[{"role": "user", "content": "Hello!"}]
)

print(response.choices[0].message.content)
`,
    javascript: `import OpenAI from "openai";

const client = new OpenAI({
  baseURL: "${args.baseUrl}",
  apiKey: "YOUR_API_KEY",  // 在 API Keys 页面创建
});

const response = await client.chat.completions.create({
  model: "${args.model}",
  messages: [{ role: "user", content: "Hello!" }],
});

console.log(response.choices[0].message.content);
`,
    curl: `curl ${fullUrl} \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -d '{\"model\":\"${args.model}\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello!\"}]}'`,
  }
}

const CODE_EXAMPLE_LABELS = { python: 'Python', javascript: 'JavaScript', curl: 'cURL' } as const
function SetupGuideBackdrop(props: { compact?: boolean }) {
    return (
    <>
      <div
        className={cn(
          'pointer-events-none absolute inset-0 bg-[linear-gradient(112deg,oklch(0.97_0.04_250/.92)_0%,oklch(0.95_0.08_315/.82)_38%,oklch(0.96_0.12_92/.78)_74%,oklch(0.94_0.1_132/.62)_100%)] dark:opacity-25',
          props.compact
            ? '[mask-image:linear-gradient(90deg,black_0%,black_48%,transparent_74%)] opacity-55'
            : 'opacity-85'
        )}
        aria-hidden='true'
      />
      <div
        className={cn(
          'pointer-events-none absolute inset-y-0 right-0 hidden overflow-hidden font-mono text-lime-100/75 sm:block dark:text-lime-200/25',
          props.compact ? 'w-1/2 opacity-45' : 'w-[58%] opacity-75'
        )}
        aria-hidden='true'
      >
        <pre
          className={cn(
            'absolute right-3 [mask-image:linear-gradient(90deg,transparent_0%,black_30%,black_82%,transparent_100%)] text-right tracking-[0.38em] whitespace-pre',
            props.compact
              ? '-top-6 text-[9px] leading-4'
              : 'top-1 text-[11px] leading-5'
          )}
        >
          {SETUP_GUIDE_CODE_PATTERN}
        </pre>
      </div>
      <div
        className='from-background/35 to-background/70 dark:from-background/20 dark:to-background/80 pointer-events-none absolute inset-0 bg-linear-to-b via-transparent'
        aria-hidden='true'
      />
    </>
  )
}

function StartStepItem(props: {
  step: StartStep
  index: number
  isLast: boolean
}) {
  const Icon = props.step.icon
  const StatusIcon = props.step.completed ? Check : Circle

  return (
    <li className='relative flex gap-3 pb-2.5 last:pb-0'>
      {!props.isLast && (
        <span
          className='bg-border absolute top-9 bottom-0 left-4 w-px'
          aria-hidden='true'
        />
      )}
      <span
        className={cn(
          'bg-background relative z-10 flex size-8 shrink-0 items-center justify-center rounded-lg border shadow-xs',
          props.step.completed && 'border-success/30 bg-success/10'
        )}
      >
        <StatusIcon
          className={props.step.completed ? 'text-success size-4' : 'size-4'}
          aria-hidden='true'
        />
      </span>

      <Link
        to={props.step.to}
        className='bg-background/70 hover:bg-muted/50 focus-visible:ring-ring flex min-w-0 flex-1 items-center justify-between gap-3 rounded-xl border px-3 py-2.5 text-left shadow-xs transition-colors outline-none focus-visible:ring-2'
      >
        <span className='flex min-w-0 items-start gap-2.5'>
          <span className='bg-muted mt-0.5 flex size-7 shrink-0 items-center justify-center rounded-lg'>
            <Icon className='size-3.5' aria-hidden='true' />
          </span>
          <span className='flex min-w-0 flex-col gap-0.5'>
            <span className='flex items-center gap-2 text-sm font-medium'>
              <span className='text-muted-foreground font-mono text-xs tabular-nums'>
                {props.index + 1}.
              </span>
              <span className='truncate'>{props.step.title}</span>
            </span>
            <span className='text-muted-foreground line-clamp-1 text-xs'>
              {props.step.description}
            </span>
          </span>
        </span>
        <ArrowRight
          className='text-muted-foreground size-4 shrink-0'
          aria-hidden='true'
        />
      </Link>
    </li>
  )
}

function ApiConfigCard(props: {
  config: ApiConfigGuide
  signals: HeroSignal[]
}) {
  const { t } = useTranslation()
  const [codeLang, setCodeLang] = useState<CodeLanguage>('python')
  const shouldReduceMotion = useReducedMotion()
  
  const firstModel = props.config.models[0] || 'deepseek-v3'
  const examples = buildSafeCodeExamples({ baseUrl: props.config.baseUrl, model: firstModel })
  const currentCode = examples[codeLang] || ''

  return (
    <motion.div
      initial={shouldReduceMotion ? false : { opacity: 0, y: 10, scale: 0.98 }}
      animate={shouldReduceMotion ? undefined : { opacity: 1, y: 0, scale: 1 }}
      transition={MOTION_TRANSITION.slow}
      className="bg-background/75 relative overflow-hidden rounded-2xl border p-3 shadow-sm backdrop-blur"
    >
      {!shouldReduceMotion && (
        <motion.div
          className="via-foreground/30 pointer-events-none absolute inset-x-0 top-0 h-px bg-linear-to-r from-transparent to-transparent"
          animate={{ x: ['-100%', '100%'] }}
          transition={{ duration: 3.2, repeat: Infinity, ease: 'easeInOut' }}
          aria-hidden="true"
        />
      )}

      <div className="flex items-center justify-between gap-3 border-b pb-3">
        <div className="flex min-w-0 items-center gap-2">
          <span className="bg-muted flex size-8 shrink-0 items-center justify-center rounded-lg">
            <TerminalSquare className="size-4" aria-hidden="true" />
          </span>
          <div className="min-w-0">
            <div className="truncate text-sm font-medium">
              {t('API Configuration')}
            </div>
            <div className="text-muted-foreground truncate text-xs">
              {props.config.hasApiKey ? props.config.keyName : t('Create an API key to get started')}
            </div>
          </div>
        </div>
        {!props.config.hasApiKey ? (
          <Button size="sm" variant="outline" render={<Link to="/keys" />}>
            {t('Create API Key')}
          </Button>
        ) : (
          <CopyButton
            value={currentCode}
            variant="outline"
            size="sm"
            className="h-7 gap-1.5 px-2 text-xs"
            tooltip={t('Copy code example')}
            successTooltip={t('Copied!')}
            aria-label={t('Copy code example')}
          >
            {t('Copy')}
          </CopyButton>
        )}
      </div>

      {/* One-line intro */}
      <p className="text-muted-foreground text-xs mt-2">
        {t('OpenAI-compatible API endpoint examples')}
      </p>

      {/* Base URL display */}
      <div className="my-3 flex items-center gap-2 rounded-xl bg-foreground/[0.03] px-3 py-2">
        <Globe className="text-muted-foreground size-4 shrink-0" />
        <span className="text-xs text-muted-foreground shrink-0">{t('Base URL')}</span>
        <code className="flex-1 truncate text-xs font-mono">{props.config.baseUrl}</code>
        <CopyButton value={props.config.baseUrl} variant="ghost" size="icon" className="size-6" />
      </div>

      {/* Code example with language switcher */}
      <div className="bg-foreground/[0.035] overflow-hidden rounded-xl">
        <div className="flex items-center border-b px-2 py-1.5 gap-1">
          {(Object.entries(CODE_EXAMPLE_LABELS) as [CodeLanguage, string][]).map(([key, label]) => (
            <button
              key={key}
              type="button"
              onClick={() => setCodeLang(key)}
              className={`rounded-md px-2 py-1 text-[11px] font-medium transition-colors ${
                codeLang === key
                  ? 'bg-foreground/10 text-foreground'
                  : 'text-muted-foreground hover:text-foreground'
              }`}
            >
              {label}
            </button>
          ))}
        </div>
        <pre className="max-h-[200px] p-3 overflow-auto"><code className="text-muted-foreground text-[11px] leading-relaxed whitespace-pre">{currentCode}</code></pre>
      </div>

      {/* Available models */}
      {props.config.models.length > 0 && (
        <div className="mt-3 flex flex-wrap gap-1.5">
          <Bot className="text-muted-foreground size-3.5 mt-0.5 shrink-0" />
          {props.config.models.slice(0, 4).map((m) => (
            <kbd key={m} className="bg-muted rounded-md px-1.5 py-0.5 text-[11px] font-mono">{m}</kbd>
          ))}
          {props.config.models.length > 4 && (
            <span className="text-muted-foreground text-[11px]">+{props.config.models.length - 4}</span>
          )}
        </div>
      )}

      {/* Status signals */}
      <div className="grid gap-2 mt-3">
        {props.signals.map((signal) => {
          const Icon = signal.icon
          return (
            <div key={signal.label} className="bg-muted/40 flex items-center justify-between gap-3 rounded-xl px-3 py-2">
              <span className="flex min-w-0 items-center gap-2">
                <Icon className="text-muted-foreground size-3.5 shrink-0" aria-hidden="true" />
                <span className="truncate text-xs font-medium">{signal.label}</span>
              </span>
              <span className="text-muted-foreground shrink-0 text-xs">{signal.value}</span>
            </div>
          )
        })}
      </div>
    </motion.div>
  )
}
function QuickActionItem(props: { action: QuickAction }) {
  const Icon = props.action.icon

  return (
    <Button
      variant='outline'
      className='h-auto justify-start rounded-xl px-3 py-3 text-left'
      render={<Link to={props.action.to} />}
    >
      <span className='bg-muted flex size-9 shrink-0 items-center justify-center rounded-lg'>
        <Icon className='size-4' aria-hidden='true' />
      </span>
      <span className='flex min-w-0 flex-1 flex-col gap-0.5'>
        <span className='truncate text-sm font-medium'>
          {props.action.title}
        </span>
        <span className='text-muted-foreground line-clamp-2 text-xs leading-relaxed'>
          {props.action.description}
        </span>
      </span>
    </Button>
  )
}

function CompactQuickAction(props: { action: QuickAction }) {
  const Icon = props.action.icon

  return (
    <Button
      variant='outline'
      size='sm'
      className='bg-background/70 h-8 min-w-24 gap-1.5 px-2.5'
      render={<Link to={props.action.to} />}
    >
      <Icon data-icon='inline-start' />
      <span>{props.action.title}</span>
    </Button>
  )
}

export function OverviewDashboard() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.auth.user)
  const { items: apiInfoItems } = useApiInfo()
  const {
    apiInfo: showApiInfoPanel,
    announcements: showAnnouncementsPanel,
    faq: showFAQPanel,
    uptimeKuma: showUptimePanel,
  } = useDashboardContentVisibility()
  const [manualSetupGuideExpanded, setManualSetupGuideExpanded] = useState<
    boolean | null
  >(() => getSavedSetupGuideExpanded())

  const requestCount = Number(user?.request_count ?? 0)
  const remainQuota = Number(user?.quota ?? 0)
  const usedQuota = Number(user?.used_quota ?? 0)
  const isAdmin = Boolean(user?.role && user.role >= ROLE.ADMIN)

  const apiKeysQuery = useQuery({
    queryKey: ['dashboard', 'overview', 'api-keys'],
    queryFn: async () => {
      const result = await getApiKeys({ p: 1, size: 10 })
      return result.success ? (result.data?.items ?? []) : []
    },
    staleTime: 60 * 1000,
  })

  const modelsQuery = useQuery({
    queryKey: ['dashboard', 'overview', 'user-models'],
    queryFn: async () => {
      const result = await getUserModels()
      return result.success ? (result.data ?? []) : []
    },
    staleTime: 5 * 60 * 1000,
  })

  const preferredKey = useMemo(
    () => getPreferredKey(apiKeysQuery.data ?? []),
    [apiKeysQuery.data]
  )

  const realKeyQuery = useQuery({
    queryKey: ['dashboard', 'overview', 'token-key', preferredKey?.id],
    queryFn: async () => {
      if (!preferredKey?.id) return ''
      const result = await fetchTokenKey(preferredKey.id)
      return result.success && result.data?.key ? `sk-${result.data.key}` : ''
    },
    enabled: Boolean(preferredKey?.id),
    staleTime: 5 * 60 * 1000,
  })

  const startSteps = useMemo<StartStep[]>(
    () => [
      {
        title: t('Create API Key'),
        description: t('Create a key for your app or service'),
        to: '/keys',
        icon: KeyRound,
        completed: Boolean(preferredKey),
      },
      {
        title: t('Add credits'),
        description: t('Keep enough balance before production traffic'),
        to: '/wallet',
        icon: CreditCard,
        completed: remainQuota > 0 || usedQuota > 0,
      },
      {
        title: t('Send a request'),
        description: t('Verify routing with Playground or your client'),
        to: '/playground',
        icon: TerminalSquare,
  Bot,
        completed: requestCount > 0,
      },
    ],
    [preferredKey, remainQuota, requestCount, t, usedQuota]
  )

    const quickActions = useMemo<QuickAction[]>(
    () => [
      {
        title: t('API Keys'),
        description: t('Create a key for your app or service'),
        to: '/keys',
        icon: KeyRound,
      },
      {
        title: t('Channels'),
        description: t('Configure upstream providers and routing.'),
        to: '/channels',
        icon: RadioTower,
        adminOnly: true,
      },
      {
        title: t('Usage Logs'),
        description: t('Inspect requests, errors, and billing details'),
        to: '/usage-logs',
        icon: FileText,
      },
      {
        title: t('Pricing'),
        description: t('Review model rates before scaling traffic'),
        to: '/pricing',
        icon: BookOpen,
      },
    ],
    [t]
  )

  const visibleQuickActions = useMemo(
    () => quickActions.filter((action) => !action.adminOnly || isAdmin),
    [isAdmin, quickActions]
  )

  const heroSignals = useMemo<HeroSignal[]>(
    () => [
      {
        label: t('Route active'),
        value: apiInfoItems.length > 0 ? t('Online') : t('Current domain'),
        icon: RadioTower,
      },
      {
        label: t('Auth configured'),
        value: preferredKey ? t('Secured') : t('Needs API key'),
        icon: ShieldCheck,
      },
    ],
    [apiInfoItems.length, preferredKey, t]
  )

  const apiConfigGuide = useMemo<ApiConfigGuide>(() => {
    const baseUrl = (apiInfoItems[0]?.url || '').replace(/\/chat\/completions$/, '') || `${getCurrentOrigin()}/v1`
    const models = modelsQuery.data?.slice(0, 5) ?? []
    const hasApiKey = Boolean(realKeyQuery.data)
    const keyName = preferredKey?.name ?? t('No API key yet')

    return { baseUrl, models, hasApiKey, keyName }
  }, [apiInfoItems, modelsQuery.data, preferredKey, realKeyQuery.data, t])

  const completedStepCount = startSteps.filter((step) => step.completed).length
  const setupComplete = completedStepCount === startSteps.length
  const setupGuideExpanded = manualSetupGuideExpanded ?? !setupComplete
  const showLeftContentPanels =
    isAdmin || showApiInfoPanel || showAnnouncementsPanel || showFAQPanel
  const showContentPanels = showLeftContentPanels || showUptimePanel

  const handleSetupGuideToggle = () => {
    const nextExpanded = !setupGuideExpanded
    setManualSetupGuideExpanded(nextExpanded)
    saveSetupGuideExpanded(nextExpanded)
  }

  return (
    <div className='flex flex-col gap-4'>
      {setupGuideExpanded ? (
        <CardStaggerContainer className='grid items-stretch gap-4 xl:grid-cols-[minmax(0,1fr)_22rem]'>
          <CardStaggerItem className='bg-card h-full overflow-hidden rounded-2xl border shadow-xs'>
            <div className='relative h-full overflow-hidden p-4 sm:p-5'>
              <SetupGuideBackdrop />
              <div className='relative grid gap-5 lg:grid-cols-[minmax(0,1fr)_21rem]'>
                <div className='flex min-w-0 flex-col gap-5'>
                  <div className='flex flex-wrap items-start justify-between gap-3'>
                    <div className='flex max-w-2xl flex-col gap-1'>
                      <div className='text-muted-foreground flex items-center gap-2 text-xs font-medium tracking-wider uppercase'>
                        <ListChecks className='size-3.5' aria-hidden='true' />
                        {t('Get started')}
                      </div>
                      <h3 className='text-xl font-semibold tracking-tight sm:text-2xl'>
                        {t('Build on your API gateway in minutes')}
                      </h3>
                      <p className='text-muted-foreground max-w-xl text-sm leading-relaxed'>
                        {t(
                          'A focused home for keys, balance, routing, and service health.'
                        )}
                      </p>
                    </div>
                    <div className='flex flex-wrap items-center gap-2'>
                      <Button
                        variant='outline'
                        size='sm'
                        onClick={handleSetupGuideToggle}
                      >
                        <ChevronUp data-icon='inline-start' />
                        {t('Hide setup guide')}
                      </Button>
                      <Button size='sm' render={<Link to='/keys' />}>
                        <KeyRound data-icon='inline-start' />
                        {t('Create API Key')}
                      </Button>
                    </div>
                  </div>

                  <ol className='bg-background/45 rounded-2xl border p-2 backdrop-blur'>
                    {startSteps.map((step, index) => (
                      <StartStepItem
                        key={step.title}
                        step={step}
                        index={index}
                        isLast={index === startSteps.length - 1}
                      />
                    ))}
                  </ol>
                </div>

                <ApiConfigCard
                  config={apiConfigGuide}
                  signals={heroSignals}
                />
              </div>
            </div>
          </CardStaggerItem>

          <CardStaggerItem className='bg-card h-full rounded-2xl border p-4 shadow-xs sm:p-5'>
            <div className='flex h-full flex-col gap-4'>
              <div className='flex flex-col gap-1'>
                <div className='text-muted-foreground text-xs font-medium tracking-wider uppercase'>
                  {t('Recommended actions')}
                </div>
                <h3 className='text-lg font-semibold tracking-tight'>
                  {t('Keep the platform ready')}
                </h3>
              </div>
              <div className='grid gap-2'>
                {visibleQuickActions.map((action) => (
                  <QuickActionItem key={action.title} action={action} />
                ))}
              </div>
            </div>
          </CardStaggerItem>
        </CardStaggerContainer>
      ) : (
        <CardStaggerContainer>
          <CardStaggerItem className='bg-card overflow-hidden rounded-2xl border shadow-xs'>
            <div className='relative overflow-hidden px-4 py-3 sm:px-5'>
              <SetupGuideBackdrop compact />
              <div className='relative flex flex-wrap items-center justify-between gap-3'>
                <div className='flex min-w-0 items-center gap-3'>
                  <span className='bg-background/70 flex size-9 shrink-0 items-center justify-center rounded-xl border shadow-xs'>
                    <Check className='text-success size-4' aria-hidden='true' />
                  </span>
                  <div className='min-w-0'>
                    <div className='flex items-center gap-2'>
                      <h3 className='truncate text-sm font-semibold'>
                        {setupComplete
                          ? t('Setup guide complete')
                          : t('Setup guide')}
                      </h3>
                      <span className='text-muted-foreground bg-background/60 rounded-md border px-2 py-0.5 text-xs'>
                        {t('Setup progress: {{completed}}/{{total}}', {
                          completed: completedStepCount,
                          total: startSteps.length,
                        })}
                      </span>
                    </div>
                    <p className='text-muted-foreground line-clamp-1 text-xs'>
                      {setupComplete
                        ? t(
                            'Your setup guide is collapsed so usage stays in focus.'
                          )
                        : t('Setup guide is collapsed. Expand it anytime.')}
                    </p>
                  </div>
                </div>

                <div className='flex flex-wrap items-center gap-2'>
                  {visibleQuickActions.map((action) => (
                    <CompactQuickAction key={action.title} action={action} />
                  ))}
                  <Button
                    variant='outline'
                    size='sm'
                    className='bg-background/70 h-8 min-w-28'
                    onClick={handleSetupGuideToggle}
                  >
                    <ChevronDown data-icon='inline-start' />
                    {t('Show setup guide')}
                  </Button>
                </div>
              </div>
            </div>
          </CardStaggerItem>
        </CardStaggerContainer>
      )}

      <SummaryCards />

      {showContentPanels && (
        <CardStaggerContainer
          className={cn(
            'grid grid-cols-1 gap-4',
            showLeftContentPanels &&
              showUptimePanel &&
              'xl:grid-cols-[minmax(0,1fr)_22rem]'
          )}
        >
          {showLeftContentPanels && (
            <div
              className={cn(
                'grid min-w-0 grid-cols-1 gap-4',
                (showApiInfoPanel || showAnnouncementsPanel || showFAQPanel) &&
                  'lg:grid-cols-2'
              )}
            >
              {isAdmin && (
                <CardStaggerItem className='lg:col-span-2'>
                  <PerformanceHealthPanel />
                </CardStaggerItem>
              )}
              {showApiInfoPanel && (
                <CardStaggerItem>
                  <ApiInfoPanel />
                </CardStaggerItem>
              )}
              {showAnnouncementsPanel && (
                <CardStaggerItem>
                  <AnnouncementsPanel />
                </CardStaggerItem>
              )}
              {showFAQPanel && (
                <CardStaggerItem>
                  <FAQPanel />
                </CardStaggerItem>
              )}
            </div>
          )}
          {showUptimePanel && (
            <CardStaggerItem>
              <UptimePanel />
            </CardStaggerItem>
          )}
        </CardStaggerContainer>
      )}
    </div>
  )
}
