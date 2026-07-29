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
import { type ColumnDef } from '@tanstack/react-table'
import { useTranslation } from 'react-i18next'
import { Checkbox } from '@/components/ui/checkbox'
import { DataTableColumnHeader } from '@/components/data-table'
import { StatusBadge } from '@/components/status-badge'
import { TableId } from '@/components/table-id'
import { SUBUSER_ROLE_STATUSES } from '../constants'
import { type EnterpriseSubUser } from '../types'
import { DataTableRowActions } from './data-table-row-actions'

export function useEnterpriseSubUsersColumns(): ColumnDef<EnterpriseSubUser>[] {
  const { t } = useTranslation()
  return [
    {
      id: 'select',
      meta: { label: t('Select') },
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected()}
          indeterminate={table.getIsSomePageRowsSelected()}
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label={t('Select all')}
          className='translate-y-[2px]'
        />
      ),
      cell: ({ row }) => (
        <Checkbox
          checked={row.getIsSelected()}
          onCheckedChange={(value) => row.toggleSelected(!!value)}
          aria-label={t('Select row')}
          className='translate-y-[2px]'
        />
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'id',
      meta: { label: t('ID'), mobileHidden: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('ID')} />
      ),
      cell: ({ row }) => {
        return (
          <TableId value={row.getValue('id') as number} className='w-[60px]' />
        )
      },
    },
    {
      accessorKey: 'username',
      meta: { label: t('Username'), mobileTitle: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Username')} />
      ),
      cell: ({ row }) => {
        return (
          <div className='max-w-[150px] truncate font-medium'>
            {row.getValue('username')}
          </div>
        )
      },
    },
    {
      accessorKey: 'display_name',
      meta: { label: t('Display Name'), mobileTitle: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Display Name')} />
      ),
      cell: ({ row }) => {
        return (
          <div className='max-w-[150px] truncate'>
            {row.getValue('display_name')}
          </div>
        )
      },
    },
    {
      accessorKey: 'email',
      meta: { label: t('Email'), mobileHidden: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Email')} />
      ),
      cell: ({ row }) => {
        return (
          <div className='max-w-[200px] truncate text-sm text-muted-foreground'>
            {row.getValue('email')}
          </div>
        )
      },
    },
    {
      accessorKey: 'role',
      meta: { label: t('Role'), mobileBadge: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Role')} />
      ),
      cell: ({ row }) => {
        const roleValue = row.getValue('role') as number
        const roleConfig = SUBUSER_ROLE_STATUSES[roleValue]

        if (!roleConfig) {
          return null
        }

        return (
          <StatusBadge
            label={t(roleConfig.labelKey)}
            variant={roleConfig.variant}
            copyable={false}
          />
        )
      },
    },
    {
      accessorKey: 'group',
      meta: { label: t('Group'), mobileHidden: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Group')} />
      ),
      cell: ({ row }) => {
        const group = row.getValue('group') as string
        return (
          <div className='max-w-[120px] truncate text-sm text-muted-foreground'>
            {group || '-'}
          </div>
        )
      },
    },
    {
      id: 'actions',
      cell: ({ row }) => <DataTableRowActions row={row} />,
    },
  ]
}
