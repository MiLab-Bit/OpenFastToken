/*
Copyright (C) 2023-2026 智企惠

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/

import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/auth-store'
import { SectionPageLayout } from '@/components/layout'
import { PageTransition } from '@/components/page-transition'
import { EnterpriseWalletCard } from '@/features/wallet/components/enterprise-wallet-card'
import { TenantMembersTable } from '@/features/tenant-console/components/tenant-members-table'

// ============================================================================
// Enterprise Wallet — a dedicated interface, separate from the personal Wallet.
//
// Aggregates the enterprise wallet card (main balance, member quota, grant /
// recharge / fund-flow actions) with the tenant member roster. Talks to the
// same `/api/user/tenant/**` endpoints as the embedded card; the tenant is
// resolved server-side from the session.
// ============================================================================

export function EnterpriseWalletPage() {
  const { t } = useTranslation()
  const user = useAuthStore((state) => state.auth.user)

  return (
    <PageTransition>
      <SectionPageLayout>
        <SectionPageLayout.Title>{t('Enterprise Wallet')}</SectionPageLayout.Title>
        <SectionPageLayout.Content>
          <div className='space-y-4'>
            <EnterpriseWalletCard enterpriseId={user?.enterprise_id} />

            <div className='space-y-2'>
              <h3 className='text-sm font-medium'>{t('Members')}</h3>
              <TenantMembersTable />
            </div>
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>
    </PageTransition>
  )
}
