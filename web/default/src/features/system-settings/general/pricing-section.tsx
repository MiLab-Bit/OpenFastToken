/*
Copyright (C) 2023-2026 OpenFastToken
*/

import * as z from "zod"
import type { Resolver } from "react-hook-form"
import { zodResolver } from "@hookform/resolvers/zod"
import { useTranslation } from "react-i18next"
import { DEFAULT_CURRENCY_CONFIG } from "@/stores/system-config-store"
import {
  Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage,
} from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import {
  Select, SelectContent, SelectGroup, SelectItem, SelectTrigger, SelectValue,
} from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { FormDirtyIndicator } from "../components/form-dirty-indicator"
import { FormNavigationGuard } from "../components/form-navigation-guard"
import { SettingsForm, SettingsSwitchItem } from "../components/settings-form-layout"
import { SettingsPageFormActions } from "../components/settings-page-context"
import { SettingsSection } from "../components/settings-section"
import { useSettingsForm } from "../hooks/use-settings-form"
import { useUpdateOption } from "../hooks/use-update-option"

const createPricingSchema = (t: (key: string) => string) =>
  z.object({
    QuotaPerUnit: z.coerce.number().min(0, t("Value must be at least 0")),
    DisplayInCurrencyEnabled: z.boolean(),
    DisplayTokenStatEnabled: z.boolean(),
    USDExchangeRate: z.coerce.number().optional(),
    general_setting: z.object({
      quota_display_type: z.enum(["CNY", "TOKENS"]),
      custom_currency_symbol: z.string().optional(),
      custom_currency_exchange_rate: z.coerce.number().optional(),
    }),
  })

type PricingFormValues = z.infer<ReturnType<typeof createPricingSchema>>

type PricingSectionProps = {
  defaultValues: PricingFormValues
}

export function PricingSection({ defaultValues }: PricingSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const pricingSchema = createPricingSchema(t)

  const { form, handleSubmit, handleReset, isDirty, isSubmitting } =
    useSettingsForm<PricingFormValues>({
      resolver: zodResolver(pricingSchema) as Resolver<any, unknown, PricingFormValues>,
      defaultValues,
      onSubmit: async (_data, changedFields) => {
        for (const [key, value] of Object.entries(changedFields)) {
          if (value === undefined || value === null) continue
          if (typeof value === "object") continue
          let serialized: string | boolean = value as string | boolean
          if (typeof value === "boolean") serialized = String(value)
          else if (typeof value === "number") serialized = Number.isFinite(value) ? String(value) : "0"
          await updateOption.mutateAsync({ key, value: serialized })
        }
      },
    })

  const displayType = form.watch("general_setting.quota_display_type") ?? "CNY"
  const showQuotaPerUnit = displayType === "TOKENS" || defaultValues.QuotaPerUnit !== DEFAULT_CURRENCY_CONFIG.quotaPerUnit

  return (
    <>
      <FormNavigationGuard when={isDirty} />
      <SettingsSection title={t("Pricing & Display")}>
        <Form {...form}>
          <SettingsForm onSubmit={handleSubmit}>
            <SettingsPageFormActions
              onSave={handleSubmit}
              onReset={handleReset}
              isSaving={updateOption.isPending || isSubmitting}
              isResetDisabled={!isDirty}
            />
            <FormDirtyIndicator isDirty={isDirty} />

            {showQuotaPerUnit && (
              <FormField
                control={form.control}
                name="QuotaPerUnit"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("Quota Per Unit")}</FormLabel>
                    <FormControl>
                      <Input type="number" step="0.01" value={field.value as number} disabled name={field.name} onBlur={field.onBlur} ref={field.ref} />
                    </FormControl>
                    <FormDescription>{t("Number of tokens per unit quota")}</FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}

            <FormField
              control={form.control}
              name="general_setting.quota_display_type"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("Display Mode")}</FormLabel>
                  <Select
                    items={[
                      { value: "CNY", label: t("CNY (元)") },
                      { value: "TOKENS", label: t("Tokens Only") },
                    ]}
                    value={field.value}
                    onValueChange={field.onChange}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder={t("Select display mode")} />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent alignItemWithTrigger={false}>
                      <SelectGroup>
                        <SelectItem value="CNY">{t("CNY (元)")}</SelectItem>
                        <SelectItem value="TOKENS">{t("Tokens Only")}</SelectItem>
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                  <FormDescription>{t("Choose how quota values are shown to users")}</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="DisplayInCurrencyEnabled"
              render={({ field }) => (
                <FormItem>
                  <SettingsSwitchItem>
                    <FormLabel>{t("Display Currency")}</FormLabel>
                    <FormControl>
                      <Switch checked={field.value} onCheckedChange={field.onChange} />
                    </FormControl>
                  </SettingsSwitchItem>
                  <FormDescription>{t("Show quota in currency format")}</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="DisplayTokenStatEnabled"
              render={({ field }) => (
                <FormItem>
                  <SettingsSwitchItem>
                    <FormLabel>{t("Display Token Stats")}</FormLabel>
                    <FormControl>
                      <Switch checked={field.value} onCheckedChange={field.onChange} />
                    </FormControl>
                  </SettingsSwitchItem>
                  <FormDescription>{t("Show token consumption statistics")}</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </SettingsForm>
        </Form>
      </SettingsSection>
    </>
  )
}
