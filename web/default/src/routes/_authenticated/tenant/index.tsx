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
import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { TenantConsole } from '@/features/tenant-console'

export const Route = createFileRoute('/_authenticated/tenant/')({
  beforeLoad: () => {
    const { auth } = useAuthStore.getState()

    // Tenant console is member-scoped: personal accounts (enterprise_id 0 or
    // absent) are bounced to 403, mirroring the role guard used by /enterprises.
    if (!auth.user || !auth.user.enterprise_id) {
      throw redirect({
        to: '/403',
      })
    }
  },
  component: TenantConsole,
})
