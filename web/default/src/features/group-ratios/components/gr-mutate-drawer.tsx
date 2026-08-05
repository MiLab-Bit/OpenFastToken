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
import { createGroupRatio, updateGroupRatio, getGroupRatio } from '../api'
import { SUCCESS_MESSAGES } from '../constants'
import { type GroupRatio } from '../types'
import { useGroupRatioContext } from './gr-provider'

const groupRatioFormSchema = z.object({
  group_name: z.string().min(1, 'Group name is required').max(50),
  model_ratios: z.string().min(1, 'Model ratios JSON is required'),
})

type GroupRatioFormValues = z.infer<typeof groupRatioFormSchema>

interface GroupRatioMutateDrawerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentRow?: GroupRatio
}

export function GroupRatioMutateDrawer({
  open,
  onOpenChange,
  currentRow,
}: GroupRatioMutateDrawerProps) {
  const { t } = useTranslation()
  const isUpdate = !!currentRow
  const { triggerRefresh } = useGroupRatioContext()
  const [isSubmitting, setIsSubmitting] = useState(false)

  const form = useForm<GroupRatioFormValues>({
    resolver: zodResolver(groupRatioFormSchema),
    defaultValues: {
      group_name: '',
      model_ratios: '{}',
    },
  })

  useEffect(() => {
    if (open && isUpdate && currentRow) {
      getGroupRatio(currentRow.id).then((result) => {
        if (result.success && result.data) {
          form.reset({
            group_name: result.data.group_name,
            model_ratios: result.data.model_ratios,
          })
        }
      })
    } else if (open && !isUpdate) {
      form.reset({
        group_name: '',
        model_ratios: '{}',
      })
    }
  }, [open, isUpdate, currentRow, form])

  const onSubmit = async (data: GroupRatioFormValues) => {
    setIsSubmitting(true)
    try {
      if (isUpdate && currentRow) {
        const result = await updateGroupRatio(currentRow.id, data)
        if (result.success) {
          toast.success(t(SUCCESS_MESSAGES.GROUP_RATIO_UPDATED))
          onOpenChange(false)
          triggerRefresh()
        }
      } else {
        const result = await createGroupRatio(data)
        if (result.success) {
          toast.success(t(SUCCESS_MESSAGES.GROUP_RATIO_CREATED))
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
              ? t('Update Group Ratio')
              : t('Create Group Ratio')}
          </SheetTitle>
          <SheetDescription>
            {isUpdate
              ? t('Update the group ratio configuration.')
              : t('Add a new group ratio configuration.')}{' '}
            {t('Click save when you&apos;re done.')}
          </SheetDescription>
        </SheetHeader>
        <Form {...form}>
          <form
            id='group-ratio-form'
            onSubmit={form.handleSubmit(onSubmit)}
            className={sideDrawerFormClassName()}
          >
            <SideDrawerSection>
              <FormField
                control={form.control}
                name='group_name'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Group Name')}</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder={t('Enter group name')}
                      />
                    </FormControl>
                    <FormDescription>
                      {t('The name of the group for ratio configuration')}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name='model_ratios'
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('Model Ratios (JSON)')}</FormLabel>
                    <FormControl>
                      <Textarea
                        {...field}
                        placeholder={t(
                          'Enter model ratios as JSON, e.g. {"gpt-4": 1.0, "gpt-3.5": 0.5}'
                        )}
                        className='min-h-[200px] font-mono text-sm'
                      />
                    </FormControl>
                    <FormDescription>
                      {t('JSON mapping of model names to ratio values')}
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
          <Button form='group-ratio-form' type='submit' disabled={isSubmitting}>
            {isSubmitting ? t('Saving...') : t('Save changes')}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}
