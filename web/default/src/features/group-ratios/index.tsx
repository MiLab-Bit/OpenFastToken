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
import { useTranslation } from 'react-i18next'
import { SectionPageLayout } from '@/components/layout'
import { GroupRatioDialogs } from './components/gr-dialogs'
import { GroupRatioPrimaryButtons } from './components/gr-primary-buttons'
import { GroupRatioProvider } from './components/gr-provider'
import { GroupRatioTable } from './components/gr-table'

export function GroupRatios() {
  const { t } = useTranslation()
  return (
    <GroupRatioProvider>
      <SectionPageLayout>
        <SectionPageLayout.Title>
          {t('Group Ratio Management')}
        </SectionPageLayout.Title>
        <SectionPageLayout.Actions>
          <GroupRatioPrimaryButtons />
        </SectionPageLayout.Actions>
        <SectionPageLayout.Content>
          <GroupRatioTable />
        </SectionPageLayout.Content>
      </SectionPageLayout>

      <GroupRatioDialogs />
    </GroupRatioProvider>
  )
}
