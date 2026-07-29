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
import { OAuthProviderDialogs } from './components/op-dialogs'
import { OAuthProviderPrimaryButtons } from './components/op-primary-buttons'
import { OAuthProviderProvider } from './components/op-provider'
import { OAuthProviderTable } from './components/op-table'

export function OAuthProviders() {
  const { t } = useTranslation()
  return (
    <OAuthProviderProvider>
      <SectionPageLayout>
        <SectionPageLayout.Title>
          {t('OAuth Provider Management')}
        </SectionPageLayout.Title>
        <SectionPageLayout.Actions>
          <OAuthProviderPrimaryButtons />
        </SectionPageLayout.Actions>
        <SectionPageLayout.Content>
          <OAuthProviderTable />
        </SectionPageLayout.Content>
      </SectionPageLayout>

      <OAuthProviderDialogs />
    </OAuthProviderProvider>
  )
}
