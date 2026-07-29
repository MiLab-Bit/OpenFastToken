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
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'
import { ROLE } from '@/lib/roles'
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
import { PasswordInput } from '@/components/password-input'
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
import {
  createSubUser,
  updateSubUser,
  getSubUser,
} from '../api'
import { SUCCESS_MESSAGES, SUBUSER_FORM_DEFAULT_VALUES } from '../constants'
import {
  enterpriseSubUserFormSchema,
  type EnterpriseSubUserFormValues,
  type EnterpriseSubUser,
} from '../types'
import { useEnterpriseSubUsers } from './enterprise-subusers-provider'

type EnterpriseSubUsersMutateDrawerProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow?: EnterpriseSubUser
}

export function EnterpriseSubUsersMutateDrawer({
  open,
  onOpenChange,
  currentRow,
}: EnterpriseSubUsersMutateDrawerProps) {
  const { t } = useTranslation()
  const isUpdate = !!currentRow
  const { triggerRefresh, enterpriseId } = useEnterpriseSubUsers()
  const [isSubmitting, setIsSubmitting] = useState(false)

  const form = useForm<EnterpriseSubUserFormValues>({
    resolver: zodResolver(enterpriseSubUserFormSchema),
    defaultValues: SUBUSER_FORM_DEFAULT_VALUES,
  })

  // Load existing data when updating
  useEffect(() => {
    if (open && isUpdate && currentRow) {
      getSubUser(enterpriseId, currentRow.id).then((result) => {
        if (result.success && result.data) {
          form.reset({
            username: result.data.username,
            display_name: result.data.display_name,
            email: result.data.email,
            password: '',
            role: result.data.role,
            group: result.data.group || '',
          })
        }
      })
    } else if (open && !isUpdate) {
      form.reset(SUBUSER_FORM_DEFAULT_VALUES)
    }
  }, [open, isUpdate, currentRow, form, enterpriseId])

  const onSubmit = async (data: EnterpriseSubUserFormValues) => {
    setIsSubmitting(true)
    try {
      const payload = {
        username: data.username,
        display_name: data.display_name,
        email: data.email,
        role: data.role,
        ...(data.password ? { password: data.password } : {}),
        ...(data.group ? { group: data.group } : {}),
      }

      if (isUpdate && currentRow) {
        const result = await updateSubUser(enterpriseId, {
          ...payload,
          id: currentRow.id,
        })
        if (result.success) {
          toast.success(t(SUCCESS_MESSAGES.SUBUSER_UPDATED))
          onOpenChange(false)
          triggerRefresh()
        }
      } else {
        const result = await createSubUser(enterpriseId, payload)
        if (result.success) {
          toast.success(t(SUCCESS_MESSAGES.SUBUSER_CREATED))
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
              ? t('Update Sub-User')
              : t('Create Sub-User')}
          </SheetTitle>
          <SheetDescription>
            {isUpdate
              ? t('Update the sub-user by providing necessary info.')
              : t('Add a new sub-user by providing necessary info.')}{' '}
            {t("Click save when you're done.")}
          </SheetDescription>
        </SheetHeader>
        <Form {...form}>
          <form
            id='subuser-form'
            onSubmit={form.handleSubmit(onSubmit)}
            className={sideDrawerFormClassName()}
          >
            <SideDrawerSection>
              <FormField
                control={form.control}
                name='username'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Username')}</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder={t('Enter username')}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('Unique username for the sub-user account')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='display_name'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Display Name')}</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder={t('Enter display name')}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('Display name shown in the system')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='email'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Email')}</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        type='email'
                        placeholder={t('Enter email address')}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('Email address for the sub-user')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='password'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>
                      {t('Password')}
                      {!isUpdate && (
                        <span className='text-destructive'> *</span>
                      )}
                    </FormLabel>
                    <FormControl>
                      <PasswordInput
                        {...field}
                        placeholder={
                          isUpdate
                            ? t('Leave empty to keep current password')
                            : t('Enter password')
                        }
                      />
                    </FormControl>
                    <FormDescription>
                      {isUpdate
                        ? t('Leave empty to keep current password unchanged')
                        : t('Password must be at least 8 characters')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='role'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Role')}</FormLabel>
                    <Select
                      onValueChange={(value) =>
                        field.onChange(parseInt(value ?? '', 10))
                      }
                      defaultValue={String(field.value)}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder={t('Select a role')} />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value={String(ROLE.USER)}>
                          {t('User')}
                        </SelectItem>
                        <SelectItem value={String(ROLE.ADMIN)}>
                          {t('Admin')}
                        </SelectItem>
                      </SelectContent>
                    </Select>
                    <FormDescription>
                      {t('Select the permission role for this sub-user')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='group'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Group')}</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder={t('Enter group (optional)')}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('Optional group assignment for the sub-user')}
                    </FormDescription>
                    <FormMessage />
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
          <Button form='subuser-form' type='submit' disabled={isSubmitting}>
            {isSubmitting ? t('Saving...') : t('Save changes')}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
