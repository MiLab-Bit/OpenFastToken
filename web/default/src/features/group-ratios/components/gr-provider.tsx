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
import React, { useState } from 'react'
import useDialogState from '@/hooks/use-dialog'
import { type GroupRatio, type GroupRatioDialogType } from '../types'

type GroupRatioContextType = {
  open: GroupRatioDialogType | null
  setOpen: (str: GroupRatioDialogType | null) => void
  currentRow: GroupRatio | null
  setCurrentRow: React.Dispatch<React.SetStateAction<GroupRatio | null>>
  refreshTrigger: number
  triggerRefresh: () => void
}

const GroupRatioContext = React.createContext<GroupRatioContextType | null>(
  null
)

export function GroupRatioProvider({
  children,
}: {
  children: React.ReactNode
}) {
  const [open, setOpen] = useDialogState<GroupRatioDialogType>(null)
  const [currentRow, setCurrentRow] = useState<GroupRatio | null>(null)
  const [refreshTrigger, setRefreshTrigger] = useState(0)

  const triggerRefresh = () => setRefreshTrigger((prev) => prev + 1)

  return (
    <GroupRatioContext
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
    </GroupRatioContext>
  )
}

export const useGroupRatioContext = () => {
  const context = React.useContext(GroupRatioContext)

  if (!context) {
    throw new Error(
      'useGroupRatioContext has to be used within <GroupRatioProvider>'
    )
  }

  return context
}
