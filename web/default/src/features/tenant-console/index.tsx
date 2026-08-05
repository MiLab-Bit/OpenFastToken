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
import { useQuery } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { SectionPageLayout } from '@/components/layout'
import { PageTransition } from '@/components/page-transition'
import { getTenantInfo } from './api'
import { TenantInfoCard } from './components/tenant-info-card'
import { TenantMembersTable } from './components/tenant-members-table'
import { TenantResourceLinks } from './components/tenant-resource-links'

// ============================================================================
// Tenant Console — member-facing self-service view
//
// Counterpart of the platform-admin `/enterprises` pages: this route is guarded
// by `user.enterprise_id > 0` instead of a role check, and talks exclusively to
// the `/api/user/tenant/**` endpoints, which resolve the tenant server-side.
// ============================================================================

export function TenantConsole() {
  const { t } = useTranslation()

  const { data, isLoading } = useQuery({
    queryKey: ['tenant-info'],
    queryFn: getTenantInfo,
  })

  const info = data?.data
  const joined = Boolean(info?.joined)

  return (
    <PageTransition>
      <SectionPageLayout>
        <SectionPageLayout.Title>
          {t('Tenant Console')}
        </SectionPageLayout.Title>
        <SectionPageLayout.Content>
          <div className='space-y-4'>
            <TenantInfoCard info={info} isLoading={isLoading} />

            {joined && (
              <>
                <TenantResourceLinks />

                <div className='space-y-2'>
                  <h3 className='text-sm font-medium'>{t('Members')}</h3>
                  <TenantMembersTable />
                </div>
              </>
            )}
          </div>
        </SectionPageLayout.Content>
      </SectionPageLayout>
    </PageTransition>
  )
}
