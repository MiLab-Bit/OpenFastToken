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
import { useState, useRef, useCallback } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
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
import { CheckCircle, Upload, FileText, Loader2, X } from 'lucide-react'
import { registerEnterprise } from '../api'
import { api } from '@/lib/api'
import { ERROR_MESSAGES, SUCCESS_MESSAGES } from '../constants'
import {
  enterpriseRegisterFormSchema,
  type EnterpriseRegisterFormValues,
} from '../types'

export function EnterpriseRegisterForm() {
  const { t } = useTranslation()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [_licenseFile, setLicenseFile] = useState<File | null>(null)
  const [licenseUploading, setLicenseUploading] = useState(false)
  const [licenseUrl, setLicenseUrl] = useState('')
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [isSuccess, setIsSuccess] = useState(false)

  const form = useForm<EnterpriseRegisterFormValues>({
    resolver: zodResolver(enterpriseRegisterFormSchema),
    defaultValues: {
      name: '',
      credit_code: '',
      contact_name: '',
      contact_phone: '',
      contact_email: '',
      business_license: '',
      invitation_code: '',
    },
  })

  // Handle business license file upload
  const handleLicenseUpload = useCallback(async (file: File) => {
    if (!file) return
    setLicenseUploading(true)
    try {
      const formData = new FormData()
      formData.append('file', file)
      const res = await api.post('/api/misc/upload', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
      })
        if (res.data?.success && res.data?.data?.url) {
        setLicenseUrl(res.data.data.url)
        // Update form value
        form.setValue('business_license', res.data.data.url)
      } else {
        alert(res.data?.message || t('上传失败'))
      }
    } catch (err) {
      console.error('Upload error:', err)
      alert(t('文件上传失败，请重试'))
    } finally {
      setLicenseUploading(false)
    }
  }, [form])

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      setLicenseFile(file)
      handleLicenseUpload(file)
    }
  }

  const removeLicense = () => {
    setLicenseFile(null)
    setLicenseUrl('')
    form.setValue('business_license', '')
    if (fileInputRef.current) fileInputRef.current.value = ''
  }

  const onSubmit = async (data: EnterpriseRegisterFormValues) => {
    setIsSubmitting(true)
    try {
      const payload = {
        name: data.name,
        ...(data.credit_code ? { credit_code: data.credit_code } : {}),
        ...(data.contact_name ? { contact_name: data.contact_name } : {}),
        ...(data.contact_phone ? { contact_phone: data.contact_phone } : {}),
        ...(data.contact_email ? { contact_email: data.contact_email } : {}),
        ...(licenseUrl ? { business_license: licenseUrl } : {}),
        ...(data.invitation_code ? { invitation_code: data.invitation_code } : {}),
      }

      const result = await registerEnterprise(payload)
      if (result.success) {
        toast.success(t(SUCCESS_MESSAGES.REGISTRATION_SUCCESS))
        setIsSuccess(true)
      }
    } catch {
      toast.error(t(ERROR_MESSAGES.REGISTRATION_FAILED))
    } finally {
      setIsSubmitting(false)
    }
  }

  if (isSuccess) {
    return (
      <div className='flex flex-col items-center justify-center gap-4 py-12 text-center'>
        <CheckCircle className='h-16 w-16 text-green-500' />
        <h2 className='text-2xl font-semibold'>
          {t('Registration Submitted')}
        </h2>
        <p className='max-w-md text-muted-foreground'>
          {t(
            'Your enterprise registration has been submitted successfully. Our team will review your application and get back to you shortly.'
          )}
        </p>
      </div>
    )
  }

  return (
    <Form {...form}>
      <form
        onSubmit={form.handleSubmit(onSubmit)}
        className='w-full max-w-lg space-y-6'
      >
        <FormField
          control={form.control}
          name='name'
          render={({ field }) => (
            <FormItem>
              <FormLabel>
                {t('企业名称')}
                <span className='text-destructive'> *</span>
              </FormLabel>
              <FormControl>
                <Input
                  {...field}
                  autoComplete='organization'
                  placeholder={t('请输入企业名称')}
                  disabled={isSubmitting}
                />
              </FormControl>
              <FormDescription>
                {t('请输入您企业的完整名称')}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name='credit_code'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('统一社会信用代码')}</FormLabel>
              <FormControl>
                <Input
                  {...field}
                  placeholder={t('请输入统一社会信用代码（选填）')}
                  disabled={isSubmitting}
                />
              </FormControl>
              <FormDescription>
                {t('18位统一社会信用代码')}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name='contact_name'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('联系人姓名')}</FormLabel>
              <FormControl>
                <Input
                  {...field}
                  autoComplete='name'
                  placeholder={t('请输入联系人姓名（选填）')}
                  disabled={isSubmitting}
                />
              </FormControl>
              <FormDescription>
                {t('企业主要联系人姓名')}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name='contact_phone'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('联系电话')}</FormLabel>
              <FormControl>
                <Input
                  {...field}
                  autoComplete='tel'
                  placeholder={t('请输入联系电话（选填）')}
                  disabled={isSubmitting}
                />
              </FormControl>
              <FormDescription>
                {t('企业主要联系电话')}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name='contact_email'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('联系人邮箱')}</FormLabel>
              <FormControl>
                <Input
                  {...field}
                  type='email'
                  autoComplete='email'
                  placeholder={t('请输入联系人邮箱（选填）')}
                  disabled={isSubmitting}
                />
              </FormControl>
              <FormDescription>
                {t('用于接收审核结果通知')}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* 营业执照上传 */}
        <FormField
          control={form.control}
          name='business_license'
          render={() => (
            <FormItem>
              <FormLabel>{t('营业执照')}</FormLabel>
              <FormControl>
                <input
                  ref={fileInputRef}
                  type='file'
                  accept='.jpg,.jpeg,.png,.webp,.pdf'
                  onChange={handleFileChange}
                  disabled={isSubmitting || licenseUploading}
                  className='hidden'
                />
              </FormControl>
              <div className='space-y-2'>
                {licenseUrl ? (
                  <div className='flex items-center gap-3 rounded-lg border border-green-200 bg-green-50 p-3'>
                    <FileText className='h-5 w-5 text-green-600 flex-shrink-0' />
                    <span className='text-sm text-green-700 truncate flex-1'>{t('已上传营业执照')}</span>
                    <button type='button' onClick={removeLicense} className='text-muted-foreground hover:text-red-500'>
                      <X className='h-4 w-4' />
                    </button>
                  </div>
                ) : (
                  <button
                    type='button'
                    onClick={() => fileInputRef.current?.click()}
                    disabled={isSubmitting || licenseUploading}
                    className='flex items-center gap-2 rounded-lg border-2 border-dashed border-border p-4 hover:border-blue-400 hover:bg-blue-50 transition-colors w-full'
                  >
                    {licenseUploading ? (
                      <>
                        <Loader2 className='h-5 w-5 animate-spin text-blue-500' />
                        <span className='text-sm text-blue-600'>{t('上传中...')}</span>
                      </>
                    ) : (
                      <>
                        <Upload className='h-5 w-5 text-muted-foreground' />
                        <span className='text-sm text-muted-foreground'>{t('点击上传营业执照（jpg/png/webp/pdf，最大10MB）')}</span>
                      </>
                    )}
                  </button>
                )}
              </div>
              <FormDescription>{t('请上传企业营业执照扫描件或照片')}</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        {/* 企业认证邀请码（必填） */}
        <FormField
          control={form.control}
          name='invitation_code'
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t('企业认证邀请码')}<span className='text-destructive'> *</span>
              </FormLabel>
              <FormControl>
                <Input
                  {...field}
                  placeholder={t('请输入企业认证邀请码')}
                  disabled={isSubmitting}
                />
              </FormControl>
              <FormDescription>{t('输入企业认证邀请码，升级为黄金或铂金会员，享受更低折扣')}</FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <Button type='submit' disabled={isSubmitting} className='w-full'>
          {isSubmitting ? t('提交中...') : t('提交注册')}
        </Button>
      </form>
    </Form>
  )
}
