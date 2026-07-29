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
import { useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet'
import {
  SideDrawerSection,
  sideDrawerContentClassName,
  sideDrawerFooterClassName,
  sideDrawerFormClassName,
  sideDrawerHeaderClassName,
} from '@/components/drawer-layout'
import { createOAuthProvider, updateOAuthProvider, getOAuthProvider } from '../api'
import { OAUTH_PROVIDER_TYPE_OPTIONS, SUCCESS_MESSAGES } from '../constants'
import { type OAuthProvider } from '../types'
import { useOAuthProviderContext } from './op-provider'

const oauthProviderFormSchema = z.object({
  name: z.string().min(1, 'Provider name is required'),
  type: z.enum(['oidc', 'oauth2']),
  client_id: z.string().min(1, 'Client ID is required'),
  client_secret: z.string().min(1, 'Client secret is required'),
  issuer_url: z.string(),
  auth_url: z.string(),
  token_url: z.string(),
  userinfo_url: z.string(),
  scopes: z.string(),
  enabled: z.boolean(),
})

type OAuthProviderFormValues = z.infer<typeof oauthProviderFormSchema>

interface OAuthProviderMutateDrawerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow?: OAuthProvider
}

export function OAuthProviderMutateDrawer({
  open,
  onOpenChange,
  currentRow,
}: OAuthProviderMutateDrawerProps) {
  const { t } = useTranslation()
  const isUpdate = !!currentRow
  const { triggerRefresh } = useOAuthProviderContext()
  const [isSubmitting, setIsSubmitting] = useState(false)

  const form = useForm<OAuthProviderFormValues>({
    resolver: zodResolver(oauthProviderFormSchema),
    defaultValues: {
      name: '',
      type: 'oidc',
      client_id: '',
      client_secret: '',
      issuer_url: '',
      auth_url: '',
      token_url: '',
      userinfo_url: '',
      scopes: 'openid profile email',
      enabled: true,
    },
  })

  useEffect(() => {
    if (open && isUpdate && currentRow) {
      getOAuthProvider(currentRow.id).then((result) => {
        if (result.success && result.data) {
          form.reset({
            name: result.data.name,
            type: result.data.type,
            client_id: result.data.client_id,
            client_secret: result.data.client_secret,
            issuer_url: result.data.issuer_url,
            auth_url: result.data.auth_url,
            token_url: result.data.token_url,
            userinfo_url: result.data.userinfo_url,
            scopes: result.data.scopes,
            enabled: result.data.enabled,
          })
        }
      })
    } else if (open && !isUpdate) {
      form.reset({
        name: '',
        type: 'oidc',
        client_id: '',
        client_secret: '',
        issuer_url: '',
        auth_url: '',
        token_url: '',
        userinfo_url: '',
        scopes: 'openid profile email',
        enabled: true,
      })
    }
  }, [open, isUpdate, currentRow, form])

  const onSubmit = async (data: OAuthProviderFormValues) => {
    setIsSubmitting(true)
    try {
      if (isUpdate && currentRow) {
        const result = await updateOAuthProvider(currentRow.id, data)
        if (result.success) {
          toast.success(t(SUCCESS_MESSAGES.OAUTH_PROVIDER_UPDATED))
          onOpenChange(false)
          triggerRefresh()
        }
      } else {
        const result = await createOAuthProvider(data)
        if (result.success) {
          toast.success(t(SUCCESS_MESSAGES.OAUTH_PROVIDER_CREATED))
          onOpenChange(false)
          triggerRefresh()
        }
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <Sheet
      open={open}
      onOpenChange={(v) => {
        onOpenChange(v)
        if (!v) {
          form.reset()
        }
      }}
    >
      <SheetContent className={sideDrawerContentClassName('sm:max-w-[600px]')}>
        <SheetHeader className={sideDrawerHeaderClassName()}>
          <SheetTitle>
            {isUpdate
              ? t('Update OAuth Provider')
              : t('Create OAuth Provider')}
          </SheetTitle>
          <SheetDescription>
            {isUpdate
              ? t('Update the OAuth provider configuration.')
              : t('Add a new OAuth provider.')}{' '}
            {t('Click save when you&apos;re done.')}
          </SheetDescription>
        </SheetHeader>
        <Form {...form}>
          <form
            id='oauth-provider-form'
            onSubmit={form.handleSubmit(onSubmit)}
            className={sideDrawerFormClassName()}
          >
            <SideDrawerSection>
              <FormField
                control={form.control}
                name='name'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Provider Name')}</FormLabel>
                    <FormControl>
                      <Input {...field} placeholder={t('Enter provider name')} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='type'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Provider Type')}</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder={t('Select type')} />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {OAUTH_PROVIDER_TYPE_OPTIONS.map((opt) => (
                          <SelectItem key={opt.value} value={opt.value}>
                            {opt.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='issuer_url'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Issuer URL')}</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder='https://example.com'
                      />
                    </FormControl>
                    <FormDescription>
                      {t('OIDC discovery issuer URL')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='client_id'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Client ID')}</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder={t('Enter client ID')}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='client_secret'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Client Secret')}</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        type='password'
                        placeholder={t('Enter client secret')}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='auth_url'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Authorization URL')}</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder='https://example.com/authorize'
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='token_url'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Token URL')}</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder='https://example.com/token'
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='userinfo_url'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('UserInfo URL')}</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder='https://example.com/userinfo'
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='scopes'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Scopes')}</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder='openid profile email'
                      />
                    </FormControl>
                    <FormDescription>
                      {t('Space-separated list of OAuth scopes')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='enabled'
                render={({ field }) => (
                  <FormItem className='flex items-center justify-between'>
                    <div>
                      <FormLabel>{t('Enabled')}</FormLabel>
                      <FormDescription>
                        {t('Enable or disable this OAuth provider')}
                      </FormDescription>
                    </div>
                    <FormControl>
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />
            </SideDrawerSection>
          </form>
        </Form>
        <SheetFooter className={sideDrawerFooterClassName()}>
          <SheetClose render={<Button variant='outline' />}>
            {t('Close')}
          </SheetClose>
          <Button form='oauth-provider-form' type='submit' disabled={isSubmitting}>
            {isSubmitting ? t('Saving...') : t('Save changes')}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
