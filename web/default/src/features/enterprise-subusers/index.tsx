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
import { EnterpriseSubUsersDialogs } from './components/enterprise-subusers-dialogs'
import { EnterpriseSubUsersPrimaryButtons } from './components/enterprise-subusers-primary-buttons'
import { EnterpriseSubUsersProvider } from './components/enterprise-subusers-provider'
import { EnterpriseSubUsersTable } from './components/enterprise-subusers-table'

export function EnterpriseSubUsers({ enterpriseId }: { enterpriseId: number }) {
  const { t } = useTranslation()
  return (
    <EnterpriseSubUsersProvider enterpriseId={enterpriseId}>
      <PageTransition>
        <SectionPageLayout>
          <SectionPageLayout.Title>
            {t('Sub-User Management')}
          </SectionPageLayout.Title>
          <SectionPageLayout.Actions>
            <EnterpriseSubUsersPrimaryButtons />
          </SectionPageLayout.Actions>
          <SectionPageLayout.Content>
            <EnterpriseSubUsersTable />
          </SectionPageLayout.Content>
        </SectionPageLayout>
      </PageTransition>

      <EnterpriseSubUsersDialogs />
    </EnterpriseSubUsersProvider>
  )
}
