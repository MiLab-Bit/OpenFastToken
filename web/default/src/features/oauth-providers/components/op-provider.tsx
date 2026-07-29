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
import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog'
import { type OAuthProvider, type OAuthProviderDialogType } from '../types'

type OAuthProviderContextType = {
  open: OAuthProviderDialogType | null
  setOpen: (str: OAuthProviderDialogType | null) => void
  currentRow: OAuthProvider | null
  setCurrentRow: React.Dispatch<React.SetStateAction<OAuthProvider | null>>
  refreshTrigger: number
  triggerRefresh: () => void
}

const OAuthProviderContext =
  React.createContext<OAuthProviderContextType | null>(null)

export function OAuthProviderProvider({
  children,
}: {
  children: React.ReactNode
}) {
  const [open, setOpen] = useDialogState<OAuthProviderDialogType>(null)
  const [currentRow, setCurrentRow] = useState<OAuthProvider | null>(null)
  const [refreshTrigger, setRefreshTrigger] = useState(0)

  const triggerRefresh = () => setRefreshTrigger((prev) => prev + 1)

  return (
    <OAuthProviderContext
      value={{
        open,
        setOpen,
        currentRow,
        setCurrentRow,
        refreshTrigger,
        triggerRefresh,
      }}
    >
      {children}
    </OAuthProviderContext>
  )
}

export const useOAuthProviderContext = () => {
  const context = React.useContext(OAuthProviderContext)

  if (!context) {
    throw new Error(
      'useOAuthProviderContext has to be used within <OAuthProviderProvider>'
    )
  }

  return context
}
