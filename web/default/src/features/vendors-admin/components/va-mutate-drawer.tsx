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
import { Textarea } from '@/components/ui/textarea'
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
import { createVendor, updateVendor, getVendor } from '../api'
import { SUCCESS_MESSAGES } from '../constants'
import { type Vendor } from '../types'
import { useVendorContext } from './va-provider'

const vendorFormSchema = z.object({
  name: z.string().min(1, 'Vendor name is required'),
  icon: z.string(),
  description: z.string(),
})

type VendorFormValues = z.infer<typeof vendorFormSchema>

interface VendorMutateDrawerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow?: Vendor
}

export function VendorMutateDrawer({
  open,
  onOpenChange,
  currentRow,
}: VendorMutateDrawerProps) {
  const { t } = useTranslation()
  const isUpdate = !!currentRow
  const { triggerRefresh } = useVendorContext()
  const [isSubmitting, setIsSubmitting] = useState(false)

  const form = useForm<VendorFormValues>({
    resolver: zodResolver(vendorFormSchema),
    defaultValues: {
      name: '',
      icon: '',
      description: '',
    },
  })

  useEffect(() => {
    if (open && isUpdate && currentRow) {
      getVendor(currentRow.id).then((result) => {
        if (result.success && result.data) {
          form.reset({
            name: result.data.name,
            icon: result.data.icon,
            description: result.data.description,
          })
        }
      })
    } else if (open && !isUpdate) {
      form.reset({
        name: '',
        icon: '',
        description: '',
      })
    }
  }, [open, isUpdate, currentRow, form])

  const onSubmit = async (data: VendorFormValues) => {
    setIsSubmitting(true)
    try {
      if (isUpdate && currentRow) {
        const result = await updateVendor({
          id: currentRow.id,
          ...data,
        })
        if (result.success) {
          toast.success(t(SUCCESS_MESSAGES.VENDOR_UPDATED))
          onOpenChange(false)
          triggerRefresh()
        }
      } else {
        const result = await createVendor(data)
        if (result.success) {
          toast.success(t(SUCCESS_MESSAGES.VENDOR_CREATED))
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
            {isUpdate ? t('Update Vendor') : t('Create Vendor')}
          </SheetTitle>
          <SheetDescription>
            {isUpdate
              ? t('Update the vendor information.')
              : t('Add a new vendor.')}{' '}
            {t('Click save when you&apos;re done.')}
          </SheetDescription>
        </SheetHeader>
        <Form {...form}>
          <form
            id='vendor-form'
            onSubmit={form.handleSubmit(onSubmit)}
            className={sideDrawerFormClassName()}
          >
            <SideDrawerSection>
              <FormField
                control={form.control}
                name='name'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Vendor Name')}</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder={t('Enter vendor name')}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='icon'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Icon URL')}</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder='https://example.com/icon.png'
                      />
                    </FormControl>
                    <FormDescription>
                      {t('URL to the vendor icon image')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='description'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Description')}</FormLabel>
                    <FormControl>
                      <Textarea
                        {...field}
                        placeholder={t('Enter vendor description')}
                        className='min-h-[100px]'
                      />
                    </FormControl>
                    <FormDescription>
                      {t('A brief description of the vendor')}
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
          <Button form='vendor-form' type='submit' disabled={isSubmitting}>
            {isSubmitting ? t('Saving...') : t('Save changes')}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
