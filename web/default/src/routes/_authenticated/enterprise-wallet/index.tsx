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

import { createFileRoute, redirect } from '@tanstack/react-router'
import { useAuthStore } from '@/stores/auth-store'
import { EnterpriseWalletPage } from '@/features/enterprise-wallet'

export const Route = createFileRoute('/_authenticated/enterprise-wallet/')({
  beforeLoad: () => {
    const { auth } = useAuthStore.getState()
    // Enterprise wallet is member-scoped: personal accounts are bounced to 403,
    // mirroring the guard used by /tenant and /enterprises.
    if (!auth.user || !auth.user.enterprise_id) {
      throw redirect({ to: '/403' })
    }
  },
  component: EnterpriseWalletPage,
})
