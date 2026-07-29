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
import * as React from 'react'
import * as z from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Code2, Eye, ShieldAlert } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import {
  Alert,
  AlertAction,
  AlertDescription,
  AlertTitle,
} from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Separator } from '@/components/ui/separator'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { RiskAcknowledgementDialog } from '@/components/risk-acknowledgement-dialog'
import { confirmPaymentCompliance } from '../api'
import {
  SettingsForm,
  SettingsSwitchContent,
  SettingsSwitchItem,
} from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useUpdateOption } from '../hooks/use-update-option'
import { AmountDiscountVisualEditor } from './amount-discount-visual-editor'
import { AmountOptionsVisualEditor } from './amount-options-visual-editor'
import { PaymentMethodsVisualEditor } from './payment-methods-visual-editor'
import {
  formatJsonForEditor,
  getJsonError,
  normalizeJsonForComparison,
  removeTrailingSlash,
} from './utils'

const paymentSchema = z.object({
  PayAddress: z.string().refine((value) => {
    const trimmed = value.trim()
    if (!trimmed) return true
    return /^https?:\/\//.test(trimmed)
  }, 'Provide a valid callback URL starting with http:// or https://'),
  EpayId: z.string(),
  EpayKey: z.string(),
  Price: z.coerce.number().min(0),
  MinTopUp: z.coerce.number().min(0),
  CustomCallbackAddress: z.string().refine((value) => {
    const trimmed = value.trim()
    if (!trimmed) return true
    return /^https?:\/\//.test(trimmed)
  }, 'Provide a valid URL starting with http:// or https://'),
  PayMethods: z.string().superRefine((value, ctx) => {
    const error = getJsonError(value)
    if (error) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: error,
      })
    }
  }),
  AmountOptions: z.string().superRefine((value, ctx) => {
    const error = getJsonError(value, (parsed) => Array.isArray(parsed))
    if (error) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: error,
      })
    }
  }),
  AmountDiscount: z.string().superRefine((value, ctx) => {
    const error = getJsonError(
      value,
      (parsed) =>
        !!parsed && typeof parsed === 'object' && !Array.isArray(parsed)
    )
    if (error) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: error,
      })
    }
  }),
  AlipayEnabled: z.boolean(),
  AlipayAppId: z.string(),
  AlipayPrivateKey: z.string(),
  AlipayPublicKey: z.string(),
  AlipaySandbox: z.boolean(),
  AlipayNotifyUrl: z.string(),
  AlipayUnitPrice: z.coerce.number().min(0),
  AlipayMinTopUp: z.coerce.number().min(0),
  AlipaySellerId: z.string(),
  WechatEnabled: z.boolean(),
  WechatAppId: z.string(),
  WechatMchId: z.string(),
  WechatApiV3Key: z.string(),
  WechatPrivateKey: z.string(),
  WechatSerialNo: z.string(),
  WechatNotifyUrl: z.string(),
  WechatUnitPrice: z.coerce.number().min(0),
  WechatMinTopUp: z.coerce.number().min(0),
})

type PaymentFormValues = z.infer<typeof paymentSchema>

const CURRENT_COMPLIANCE_TERMS_VERSION = 'v1'

type PaymentComplianceDefaults = {
  confirmed: boolean
  termsVersion: string
  confirmedAt: number
  confirmedBy: number
}

type PaymentSettingsSectionProps = {
  defaultValues: PaymentFormValues
  complianceDefaults: PaymentComplianceDefaults
}

export function PaymentSettingsSection({
  defaultValues,
  complianceDefaults,
}: PaymentSettingsSectionProps) {
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const updateOption = useUpdateOption()
  const initialRef = React.useRef(defaultValues)
  const defaultsSignature = React.useMemo(
    () => JSON.stringify(defaultValues),
    [defaultValues]
  )

  const [payMethodsVisualMode, setPayMethodsVisualMode] = React.useState(true)
  const [amountOptionsVisualMode, setAmountOptionsVisualMode] =
    React.useState(true)
  const [amountDiscountVisualMode, setAmountDiscountVisualMode] =
    React.useState(true)
  const [showComplianceDialog, setShowComplianceDialog] = React.useState(false)

  const complianceStatements = React.useMemo(
    () => [
      t(
        'You have legally obtained authorization for the connected model APIs, accounts, keys, and quotas.'
      ),
      t(
        'You commit to using upstream APIs, accounts, keys, quotas, and service capabilities only within the scope of lawful authorization obtained from upstream service providers, model service providers, or relevant rights holders, and will not conduct unauthorized resale, trafficking, distribution, or other non-compliant commercialization.'
      ),
      t(
        'If you provide generative AI services to the public in mainland China, you will fulfill legal obligations including filing, security assessment, content safety, complaint handling, generated content labeling, log retention, and personal information protection.'
      ),
      t(
        'You commit not to use this system to implement, assist with, or indirectly implement acts that violate applicable laws and regulations, regulatory requirements, platform rules, public interests, or the lawful rights and interests of third parties.'
      ),
      t(
        'You understand and independently bear legal responsibility arising from deployment, operation, and charging behavior.'
      ),
      t(
        'You understand this compliance reminder is only for risk notice and does not constitute legal advice, a compliance review conclusion, or a guarantee of the legality of your use of this system; you should consult professional legal or compliance advisors based on your actual business scenario.'
      ),
    ],
    [t]
  )

  const complianceRequiredText = t(
    'I have read and understood the above compliance reminder, acknowledge the related legal risks, and confirm that I bear legal responsibility arising from deployment, operation, and charging behavior.'
  )
  const complianceRequiredTextParts = React.useMemo(
    () => [
      {
        type: 'input' as const,
        text: t('I have read and understood the above compliance reminder'),
      },
      { type: 'static' as const, text: t('，') },
      {
        type: 'input' as const,
        text: t('acknowledge the related legal risks'),
      },
      { type: 'static' as const, text: t('，and ') },
      {
        type: 'input' as const,
        text: t(
          'confirm that I bear legal responsibility arising from deployment'
        ),
      },
      { type: 'static' as const, text: t('、') },
      {
        type: 'input' as const,
        text: t('operation and charging behavior'),
      },
    ],
    [t]
  )

  const complianceConfirmed =
    complianceDefaults.confirmed &&
    complianceDefaults.termsVersion === CURRENT_COMPLIANCE_TERMS_VERSION

  const confirmComplianceMutation = useMutation({
    mutationFn: confirmPaymentCompliance,
    onSuccess: (data) => {
      if (data.success) {
        toast.success(t('Compliance confirmed successfully'))
        setShowComplianceDialog(false)
        queryClient.invalidateQueries({ queryKey: ['system-options'] })
      } else {
        toast.error(data.message || t('Failed to confirm compliance'))
      }
    },
    onError: (error: Error) => {
      toast.error(error.message || t('Failed to confirm compliance'))
    },
  })

  const form = useForm({
    resolver: zodResolver(paymentSchema),
    mode: 'onChange', // Enable real-time validation
    defaultValues: {
      ...defaultValues,
      PayMethods: formatJsonForEditor(defaultValues.PayMethods),
      AmountOptions: formatJsonForEditor(defaultValues.AmountOptions),
      AmountDiscount: formatJsonForEditor(defaultValues.AmountDiscount),
    },
  })

  React.useEffect(() => {
    const parsedDefaults = JSON.parse(defaultsSignature) as PaymentFormValues
    initialRef.current = parsedDefaults
    form.reset({
      ...parsedDefaults,
      PayMethods: formatJsonForEditor(parsedDefaults.PayMethods),
      AmountOptions: formatJsonForEditor(parsedDefaults.AmountOptions),
      AmountDiscount: formatJsonForEditor(parsedDefaults.AmountDiscount),
    })
  }, [defaultsSignature, form])

  const onSubmit = async (values: PaymentFormValues) => {
    const sanitized = {
      PayAddress: removeTrailingSlash(values.PayAddress),
      EpayId: values.EpayId.trim(),
      EpayKey: values.EpayKey.trim(),
      Price: values.Price,
      MinTopUp: values.MinTopUp,
      CustomCallbackAddress: removeTrailingSlash(values.CustomCallbackAddress),
      PayMethods: values.PayMethods.trim(),
      AmountOptions: values.AmountOptions.trim(),
      AmountDiscount: values.AmountDiscount.trim(),
      AlipayEnabled: values.AlipayEnabled,
      AlipayAppId: values.AlipayAppId.trim(),
      AlipayPrivateKey: values.AlipayPrivateKey.trim(),
      AlipayPublicKey: values.AlipayPublicKey.trim(),
      AlipaySandbox: values.AlipaySandbox,
      AlipayNotifyUrl: values.AlipayNotifyUrl.trim(),
      AlipayUnitPrice: values.AlipayUnitPrice,
      AlipayMinTopUp: values.AlipayMinTopUp,
      AlipaySellerId: values.AlipaySellerId.trim(),
      WechatEnabled: values.WechatEnabled,
      WechatAppId: values.WechatAppId.trim(),
      WechatMchId: values.WechatMchId.trim(),
      WechatApiV3Key: values.WechatApiV3Key.trim(),
      WechatPrivateKey: values.WechatPrivateKey.trim(),
      WechatSerialNo: values.WechatSerialNo.trim(),
      WechatNotifyUrl: values.WechatNotifyUrl.trim(),
      WechatUnitPrice: values.WechatUnitPrice,
      WechatMinTopUp: values.WechatMinTopUp,
    }

    const initial = {
      PayAddress: removeTrailingSlash(initialRef.current.PayAddress),
      EpayId: initialRef.current.EpayId.trim(),
      EpayKey: initialRef.current.EpayKey.trim(),
      Price: initialRef.current.Price,
      MinTopUp: initialRef.current.MinTopUp,
      CustomCallbackAddress: removeTrailingSlash(
        initialRef.current.CustomCallbackAddress
      ),
      PayMethods: initialRef.current.PayMethods.trim(),
      AmountOptions: initialRef.current.AmountOptions.trim(),
      AmountDiscount: initialRef.current.AmountDiscount.trim(),
      AlipayEnabled: initialRef.current.AlipayEnabled,
      AlipayAppId: initialRef.current.AlipayAppId.trim(),
      AlipayPrivateKey: initialRef.current.AlipayPrivateKey.trim(),
      AlipayPublicKey: initialRef.current.AlipayPublicKey.trim(),
      AlipaySandbox: initialRef.current.AlipaySandbox,
      AlipayNotifyUrl: initialRef.current.AlipayNotifyUrl.trim(),
      AlipayUnitPrice: initialRef.current.AlipayUnitPrice,
      AlipayMinTopUp: initialRef.current.AlipayMinTopUp,
      AlipaySellerId: initialRef.current.AlipaySellerId.trim(),
      WechatEnabled: initialRef.current.WechatEnabled,
      WechatAppId: initialRef.current.WechatAppId.trim(),
      WechatMchId: initialRef.current.WechatMchId.trim(),
      WechatApiV3Key: initialRef.current.WechatApiV3Key.trim(),
      WechatPrivateKey: initialRef.current.WechatPrivateKey.trim(),
      WechatSerialNo: initialRef.current.WechatSerialNo.trim(),
      WechatNotifyUrl: initialRef.current.WechatNotifyUrl.trim(),
      WechatUnitPrice: initialRef.current.WechatUnitPrice,
      WechatMinTopUp: initialRef.current.WechatMinTopUp,
    }

    const updates: Array<{ key: string; value: string | number | boolean }> = []

    if (sanitized.PayAddress !== initial.PayAddress) {
      updates.push({ key: 'PayAddress', value: sanitized.PayAddress })
    }

    if (sanitized.EpayId !== initial.EpayId) {
      updates.push({ key: 'EpayId', value: sanitized.EpayId })
    }

    if (sanitized.EpayKey && sanitized.EpayKey !== initial.EpayKey) {
      updates.push({ key: 'EpayKey', value: sanitized.EpayKey })
    }

    if (sanitized.Price !== initial.Price) {
      updates.push({ key: 'Price', value: sanitized.Price })
    }

    if (sanitized.MinTopUp !== initial.MinTopUp) {
      updates.push({ key: 'MinTopUp', value: sanitized.MinTopUp })
    }

    if (sanitized.CustomCallbackAddress !== initial.CustomCallbackAddress) {
      updates.push({
        key: 'CustomCallbackAddress',
        value: sanitized.CustomCallbackAddress,
      })
    }

    if (
      normalizeJsonForComparison(sanitized.PayMethods) !==
      normalizeJsonForComparison(initial.PayMethods)
    ) {
      updates.push({ key: 'PayMethods', value: sanitized.PayMethods })
    }

    if (
      normalizeJsonForComparison(sanitized.AmountOptions) !==
      normalizeJsonForComparison(initial.AmountOptions)
    ) {
      updates.push({
        key: 'payment_setting.amount_options',
        value: sanitized.AmountOptions,
      })
    }

    if (
      normalizeJsonForComparison(sanitized.AmountDiscount) !==
      normalizeJsonForComparison(initial.AmountDiscount)
    ) {
      updates.push({
        key: 'payment_setting.amount_discount',
        value: sanitized.AmountDiscount,
      })
    }

    if (sanitized.AlipayEnabled !== initial.AlipayEnabled) {
      updates.push({ key: 'AlipayEnabled', value: sanitized.AlipayEnabled })
    }

    if (sanitized.AlipayAppId !== initial.AlipayAppId) {
      updates.push({ key: 'AlipayAppId', value: sanitized.AlipayAppId })
    }

    if (
      sanitized.AlipayPrivateKey &&
      sanitized.AlipayPrivateKey !== initial.AlipayPrivateKey
    ) {
      updates.push({ key: 'AlipayPrivateKey', value: sanitized.AlipayPrivateKey })
    }

    if (
      sanitized.AlipayPublicKey &&
      sanitized.AlipayPublicKey !== initial.AlipayPublicKey
    ) {
      updates.push({ key: 'AlipayPublicKey', value: sanitized.AlipayPublicKey })
    }

    if (sanitized.AlipaySandbox !== initial.AlipaySandbox) {
      updates.push({ key: 'AlipaySandbox', value: sanitized.AlipaySandbox })
    }

    if (sanitized.AlipayNotifyUrl !== initial.AlipayNotifyUrl) {
      updates.push({ key: 'AlipayNotifyUrl', value: sanitized.AlipayNotifyUrl })
    }

    if (sanitized.AlipayUnitPrice !== initial.AlipayUnitPrice) {
      updates.push({ key: 'AlipayUnitPrice', value: sanitized.AlipayUnitPrice })
    }

    if (sanitized.AlipayMinTopUp !== initial.AlipayMinTopUp) {
      updates.push({ key: 'AlipayMinTopUp', value: sanitized.AlipayMinTopUp })
    }

    if (sanitized.AlipaySellerId !== initial.AlipaySellerId) {
      updates.push({ key: 'AlipaySellerId', value: sanitized.AlipaySellerId })
    }

    if (sanitized.WechatEnabled !== initial.WechatEnabled) {
      updates.push({ key: 'WechatEnabled', value: sanitized.WechatEnabled })
    }

    if (sanitized.WechatAppId !== initial.WechatAppId) {
      updates.push({ key: 'WechatAppId', value: sanitized.WechatAppId })
    }

    if (sanitized.WechatMchId !== initial.WechatMchId) {
      updates.push({ key: 'WechatMchId', value: sanitized.WechatMchId })
    }

    if (
      sanitized.WechatApiV3Key &&
      sanitized.WechatApiV3Key !== initial.WechatApiV3Key
    ) {
      updates.push({ key: 'WechatApiV3Key', value: sanitized.WechatApiV3Key })
    }

    if (
      sanitized.WechatPrivateKey &&
      sanitized.WechatPrivateKey !== initial.WechatPrivateKey
    ) {
      updates.push({ key: 'WechatPrivateKey', value: sanitized.WechatPrivateKey })
    }

    if (sanitized.WechatSerialNo !== initial.WechatSerialNo) {
      updates.push({ key: 'WechatSerialNo', value: sanitized.WechatSerialNo })
    }

    if (sanitized.WechatNotifyUrl !== initial.WechatNotifyUrl) {
      updates.push({ key: 'WechatNotifyUrl', value: sanitized.WechatNotifyUrl })
    }

    if (sanitized.WechatUnitPrice !== initial.WechatUnitPrice) {
      updates.push({ key: 'WechatUnitPrice', value: sanitized.WechatUnitPrice })
    }

    if (sanitized.WechatMinTopUp !== initial.WechatMinTopUp) {
      updates.push({ key: 'WechatMinTopUp', value: sanitized.WechatMinTopUp })
    }

    for (const update of updates) {
      await updateOption.mutateAsync(update)
    }
  }

  return (
    <SettingsSection title={t('Payment Gateway')}>
      {!complianceConfirmed ? (
        <Alert variant='destructive' className='mb-6'>
          <ShieldAlert className='h-4 w-4' />
          <AlertTitle>{t('Compliance confirmation required')}</AlertTitle>
          <AlertDescription>
            <div className='space-y-3'>
              <p>
                {t(
                  'Payment, redemption codes, subscription plans, and invitation rewards are locked until the root administrator confirms the compliance terms.'
                )}
              </p>
              <ol className='list-decimal space-y-1 pl-5'>
                {complianceStatements.map((statement) => (
                  <li key={statement}>{statement}</li>
                ))}
              </ol>
            </div>
          </AlertDescription>
          <AlertAction>
            <Button
              type='button'
              size='sm'
              variant='destructive'
              onClick={() => setShowComplianceDialog(true)}
            >
              {t('Confirm compliance')}
            </Button>
          </AlertAction>
        </Alert>
      ) : (
        <Alert className='mb-6'>
          <AlertTitle>{t('Compliance confirmed')}</AlertTitle>
          <AlertDescription>
            {t('Confirmed at {{time}} by user #{{userId}}', {
              time: complianceDefaults.confirmedAt
                ? new Date(
                    complianceDefaults.confirmedAt * 1000
                  ).toLocaleString()
                : '-',
              userId: complianceDefaults.confirmedBy || '-',
            })}
          </AlertDescription>
        </Alert>
      )}

      <RiskAcknowledgementDialog
        open={showComplianceDialog}
        onOpenChange={setShowComplianceDialog}
        title={t('Confirm compliance terms')}
        description={t(
          'This confirmation unlocks payment, redemption code, subscription plan, and invitation reward features. Please read the statements carefully.'
        )}
        items={complianceStatements}
        requiredText={complianceRequiredText}
        requiredTextParts={complianceRequiredTextParts}
        inputPrompt={t('Please type the following text to confirm:')}
        inputPlaceholder={t('Type the confirmation text here')}
        mismatchHint={t('The entered text does not match the required text.')}
        confirmText={t('Confirm and enable')}
        isLoading={confirmComplianceMutation.isPending}
        onConfirm={() => confirmComplianceMutation.mutate()}
      />

      { }
      <Form {...form}>
        <SettingsForm
          onSubmit={form.handleSubmit(onSubmit)}
          className={cn(
            'gap-y-8',
            !complianceConfirmed && 'pointer-events-none opacity-40'
          )}
          data-no-autosubmit='true'
        >
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending}
            saveLabel='Save all settings'
          />
          <div className='space-y-4'>
            <div>
              <h3 className='text-lg font-medium'>{t('General Settings')}</h3>
              <p className='text-muted-foreground text-sm'>
                {t('Shared configuration for all payment gateways')}
              </p>
            </div>

            <div className='grid gap-6 md:grid-cols-2'>
              <FormField
                control={form.control}
                name='Price'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Price (local currency / USD)')}</FormLabel>
                    <FormControl>
                      <Input
                        type='number'
                        step='0.01'
                        min={0}
                        value={(field.value ?? 0) as number}
                        onChange={(event) =>
                          field.onChange(event.target.valueAsNumber)
                        }
                      />
                    </FormControl>
                    <FormDescription>
                      {t(
                        'How much to charge for each US dollar of balance'
                      )}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='MinTopUp'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Minimum top-up (USD)')}</FormLabel>
                    <FormControl>
                      <Input
                        type='number'
                        step='0.01'
                        min={0}
                        value={(field.value ?? 0) as number}
                        onChange={(event) =>
                          field.onChange(event.target.valueAsNumber)
                        }
                      />
                    </FormControl>
                    <FormDescription>
                      {t('Smallest USD amount users can recharge')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name='PayMethods'
              render={({ field }) => (
                <FormItem>
                  <div className='mb-2 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between'>
                    <FormLabel>{t('Payment methods')}</FormLabel>
                    <Button
                      type='button'
                      variant='outline'
                      size='sm'
                      onClick={() =>
                        setPayMethodsVisualMode(!payMethodsVisualMode)
                      }
                      className='w-full sm:w-auto'
                    >
                      {payMethodsVisualMode ? (
                        <>
                          <Code2 className='mr-2 h-3 w-3' />
                          {t('JSON Editor')}
                        </>
                      ) : (
                        <>
                          <Eye className='mr-2 h-3 w-3' />
                          {t('Visual Editor')}
                        </>
                      )}
                    </Button>
                  </div>
                  <FormControl>
                    {payMethodsVisualMode ? (
                      <PaymentMethodsVisualEditor
                        value={field.value}
                        onChange={field.onChange}
                      />
                    ) : (
                      <Textarea
                        rows={4}
                        placeholder={t(
                          '[{"name":"支付宝","type":"alipay","color":"#1677FF"}]'
                        )}
                        {...field}
                        onChange={(event) => field.onChange(event.target.value)}
                      />
                    )}
                  </FormControl>
                  <FormDescription>
                    {t(
                      'Configure available payment methods. Provide a JSON array.'
                    )}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className='grid gap-6 md:grid-cols-2 md:items-start'>
              <FormField
                control={form.control}
                name='AmountOptions'
                render={({ field }) => (
                  <FormItem>
                    <div className='mb-2 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between'>
                      <FormLabel>{t('Top-up amount options')}</FormLabel>
                      <Button
                        type='button'
                        variant='outline'
                        size='sm'
                        onClick={() =>
                          setAmountOptionsVisualMode(!amountOptionsVisualMode)
                        }
                        className='w-full sm:w-auto'
                      >
                        {amountOptionsVisualMode ? (
                          <>
                            <Code2 className='mr-2 h-3 w-3' />
                            {t('JSON Editor')}
                          </>
                        ) : (
                          <>
                            <Eye className='mr-2 h-3 w-3' />
                            {t('Visual Editor')}
                          </>
                        )}
                      </Button>
                    </div>
                    <FormControl>
                      {amountOptionsVisualMode ? (
                        <AmountOptionsVisualEditor
                          value={field.value}
                          onChange={field.onChange}
                        />
                      ) : (
                        <Textarea
                          rows={4}
                          placeholder='[10, 20, 50, 100]'
                          {...field}
                          onChange={(event) =>
                            field.onChange(event.target.value)
                          }
                        />
                      )}
                    </FormControl>
                    <FormDescription>
                      {t('Preset recharge amounts (JSON array)')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='AmountDiscount'
                render={({ field }) => (
                  <FormItem>
                    <Alert className='mb-3'>
                      <AlertTitle>{t('Deprecated')}</AlertTitle>
                      <AlertDescription>
                        {t(
                          'Amount discount is deprecated. Discounts and gifts are now controlled by the "Recharge Gift" settings; this field is ignored by the server.'
                        )}
                      </AlertDescription>
                    </Alert>
                    <div className='mb-2 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between'>
                      <FormLabel>{t('Amount discount')}</FormLabel>
                      <Button
                        type='button'
                        variant='outline'
                        size='sm'
                        onClick={() =>
                          setAmountDiscountVisualMode(!amountDiscountVisualMode)
                        }
                        className='w-full sm:w-auto'
                      >
                        {amountDiscountVisualMode ? (
                          <>
                            <Code2 className='mr-2 h-3 w-3' />
                            {t('JSON Editor')}
                          </>
                        ) : (
                          <>
                            <Eye className='mr-2 h-3 w-3' />
                            {t('Visual Editor')}
                          </>
                        )}
                      </Button>
                    </div>
                    <FormControl>
                      {amountDiscountVisualMode ? (
                        <AmountDiscountVisualEditor
                          value={field.value}
                          onChange={field.onChange}
                        />
                      ) : (
                        <Textarea
                          rows={4}
                          placeholder='{"100":0.95,"200":0.9}'
                          {...field}
                          onChange={(event) =>
                            field.onChange(event.target.value)
                          }
                        />
                      )}
                    </FormControl>
                    <FormDescription>
                      {t(
                        'Deprecated — changes here no longer take effect. Configure gifts in the "Recharge Gift" section.'
                      )}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
          </div>

          <Separator />

          <div className='space-y-4'>
            <div>
              <h3 className='text-lg font-medium'>{t('Alipay Gateway')}</h3>
              <p className='text-muted-foreground text-sm'>
                {t('Configuration for Alipay payment integration')}
              </p>
            </div>

            <div className='grid gap-6 md:grid-cols-2'>
              <FormField
                control={form.control}
                name='AlipayEnabled'
                render={({ field }) => (
                  <SettingsSwitchItem>
                    <SettingsSwitchContent>
                      <FormLabel>{t('Enable Alipay')}</FormLabel>
                      <FormDescription>
                        {t('Enable Alipay as a payment method')}
                      </FormDescription>
                    </SettingsSwitchContent>
                    <FormControl>
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                  </SettingsSwitchItem>
                )}
              />

              <FormField
                control={form.control}
                name='AlipaySandbox'
                render={({ field }) => (
                  <SettingsSwitchItem>
                    <SettingsSwitchContent>
                      <FormLabel>{t('Sandbox Mode')}</FormLabel>
                      <FormDescription>
                        {t('Use Alipay sandbox environment for testing')}
                      </FormDescription>
                    </SettingsSwitchContent>
                    <FormControl>
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                  </SettingsSwitchItem>
                )}
              />
            </div>

            <div className='grid gap-6 md:grid-cols-2'>
              <FormField
                control={form.control}
                name='AlipayAppId'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('App ID')}</FormLabel>
                    <FormControl>
                      <Input
                        placeholder='20210xxxxxxxxxxxxx'
                        autoComplete='off'
                        {...field}
                        onChange={(event) => field.onChange(event.target.value)}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('Alipay application ID')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='AlipaySellerId'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Seller ID')}</FormLabel>
                    <FormControl>
                      <Input
                        placeholder='2088xxxxxxxxxxxx'
                        autoComplete='off'
                        {...field}
                        onChange={(event) => field.onChange(event.target.value)}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('Alipay seller/partner ID')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name='AlipayPrivateKey'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Private Key')}</FormLabel>
                  <FormControl>
                    <Textarea
                      rows={4}
                      placeholder={t('Paste your Alipay private key (PEM format)')}
                      autoComplete='off'
                      {...field}
                      onChange={(event) => field.onChange(event.target.value)}
                    />
                  </FormControl>
                  <FormDescription>
                    {t('RSA2/SHA256 private key for signing Alipay requests')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name='AlipayPublicKey'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Alipay Public Key')}</FormLabel>
                  <FormControl>
                    <Textarea
                      rows={4}
                      placeholder={t('Paste the Alipay public key (PEM format)')}
                      autoComplete='off'
                      {...field}
                      onChange={(event) => field.onChange(event.target.value)}
                    />
                  </FormControl>
                  <FormDescription>
                    {t('Alipay public key for verifying callbacks')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className='grid gap-6 md:grid-cols-2'>
              <FormField
                control={form.control}
                name='AlipayNotifyUrl'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Notify URL')}</FormLabel>
                    <FormControl>
                      <Input
                        placeholder={t('https://example.com/api/alipay/notify')}
                        {...field}
                        onChange={(event) => field.onChange(event.target.value)}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('Alipay async notification URL')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='AlipayUnitPrice'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>
                      {t('Unit price (local currency / USD)')}
                    </FormLabel>
                    <FormControl>
                      <Input
                        type='number'
                        step='0.01'
                        min={0}
                        value={(field.value ?? 0) as number}
                        onChange={(event) =>
                          field.onChange(event.target.valueAsNumber)
                        }
                      />
                    </FormControl>
                    <FormDescription>
                      {t('e.g., 7 means 7 CNY per USD')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name='AlipayMinTopUp'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Minimum top-up (USD)')}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      step='0.01'
                      min={0}
                      value={(field.value ?? 0) as number}
                      onChange={(event) =>
                        field.onChange(event.target.valueAsNumber)
                      }
                    />
                  </FormControl>
                  <FormDescription>
                    {t('Minimum recharge amount in USD for Alipay')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <Separator />

          <div className='space-y-4'>
            <div>
              <h3 className='text-lg font-medium'>{t('WeChat Pay Gateway')}</h3>
              <p className='text-muted-foreground text-sm'>
                {t('Configuration for WeChat Pay integration')}
              </p>
            </div>

            <FormField
              control={form.control}
              name='WechatEnabled'
              render={({ field }) => (
                <SettingsSwitchItem>
                  <SettingsSwitchContent>
                    <FormLabel>{t('Enable WeChat Pay')}</FormLabel>
                    <FormDescription>
                      {t('Enable WeChat Pay as a payment method')}
                    </FormDescription>
                  </SettingsSwitchContent>
                  <FormControl>
                    <Switch
                      checked={field.value}
                      onCheckedChange={field.onChange}
                    />
                  </FormControl>
                </SettingsSwitchItem>
              )}
            />

            <div className='grid gap-6 md:grid-cols-2'>
              <FormField
                control={form.control}
                name='WechatAppId'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('App ID')}</FormLabel>
                    <FormControl>
                      <Input
                        placeholder='wxXXXXXXXXXXXXXXXX'
                        autoComplete='off'
                        {...field}
                        onChange={(event) => field.onChange(event.target.value)}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('WeChat Pay application ID')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='WechatMchId'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Merchant ID')}</FormLabel>
                    <FormControl>
                      <Input
                        placeholder='1234567890'
                        autoComplete='off'
                        {...field}
                        onChange={(event) => field.onChange(event.target.value)}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('WeChat Pay merchant ID')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className='grid gap-6 md:grid-cols-2'>
              <FormField
                control={form.control}
                name='WechatApiV3Key'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('API v3 Key')}</FormLabel>
                    <FormControl>
                      <Input
                        type='password'
                        placeholder={t('Enter API v3 key')}
                        autoComplete='new-password'
                        {...field}
                        onChange={(event) => field.onChange(event.target.value)}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('WeChat Pay API v3 key (leave blank unless updating)')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='WechatSerialNo'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Certificate Serial No.')}</FormLabel>
                    <FormControl>
                      <Input
                        placeholder={t('Enter certificate serial number')}
                        autoComplete='off'
                        {...field}
                        onChange={(event) => field.onChange(event.target.value)}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('WeChat Pay platform certificate serial number')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name='WechatPrivateKey'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Private Key')}</FormLabel>
                  <FormControl>
                    <Textarea
                      rows={4}
                      placeholder={t('Paste your WeChat Pay private key (PEM format)')}
                      autoComplete='off'
                      {...field}
                      onChange={(event) => field.onChange(event.target.value)}
                    />
                  </FormControl>
                  <FormDescription>
                    {t('Merchant private key for signing WeChat Pay API v3 requests')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className='grid gap-6 md:grid-cols-2'>
              <FormField
                control={form.control}
                name='WechatNotifyUrl'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Notify URL')}</FormLabel>
                    <FormControl>
                      <Input
                        placeholder={t('https://example.com/api/wechat/notify')}
                        {...field}
                        onChange={(event) => field.onChange(event.target.value)}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('WeChat Pay callback notification URL')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='WechatUnitPrice'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>
                      {t('Unit price (local currency / USD)')}
                    </FormLabel>
                    <FormControl>
                      <Input
                        type='number'
                        step='0.01'
                        min={0}
                        value={(field.value ?? 0) as number}
                        onChange={(event) =>
                          field.onChange(event.target.valueAsNumber)
                        }
                      />
                    </FormControl>
                    <FormDescription>
                      {t('e.g., 7 means 7 CNY per USD')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name='WechatMinTopUp'
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('Minimum top-up (USD)')}</FormLabel>
                  <FormControl>
                    <Input
                      type='number'
                      step='0.01'
                      min={0}
                      value={(field.value ?? 0) as number}
                      onChange={(event) =>
                        field.onChange(event.target.valueAsNumber)
                      }
                    />
                  </FormControl>
                  <FormDescription>
                    {t('Minimum recharge amount in USD for WeChat Pay')}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
        </SettingsForm>
      </Form>
      { }
    </SettingsSection>
  )
}
