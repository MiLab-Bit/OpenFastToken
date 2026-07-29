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
import { useTranslation } from 'react-i18next'
import { SectionPageLayout } from '@/components/layout'
import { PageTransition } from '@/components/page-transition'
import { EnterpriseRegisterForm } from './components/enterprise-register-form'

export function EnterpriseRegister() {
  const { t } = useTranslation()

  return (
    <PageTransition>
      <SectionPageLayout>
        <SectionPageLayout.Title>
          {t('企业自助注册')}
        </SectionPageLayout.Title>
        <SectionPageLayout.Content>
          <div className='flex flex-col items-center py-8'>
            <div className='w-full max-w-lg'>
              <p className='mb-6 text-center text-muted-foreground'>
                {t(
                  '请填写以下信息完成企业注册。提交后我们将尽快审核您的申请。'
                )}
              </p>
              <EnterpriseRegisterForm />
            </div>
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>
    </PageTransition>
  )
}
