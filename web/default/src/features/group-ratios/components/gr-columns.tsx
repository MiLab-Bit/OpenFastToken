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
import { type ColumnDef } from '@tanstack/react-table'
import { useTranslation } from 'react-i18next'
import { formatTimestampToDate } from '@/lib/format'
import { Checkbox } from '@/components/ui/checkbox'
import { DataTableColumnHeader } from '@/components/data-table'
import { TableId } from '@/components/table-id'
import { type GroupRatio } from '../types'
import { DataTableRowActions } from './data-table-row-actions'

export function useGroupRatioColumns(): ColumnDef<GroupRatio>[] {
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
      accessorKey: 'group_name',
      meta: { label: t('Group Name'), mobileTitle: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Group Name')} />
      ),
      cell: ({ row }) => {
        return (
          <div className='max-w-[200px] truncate font-medium'>
            {row.getValue('group_name')}
          </div>
        )
      },
    },
    {
      accessorKey: 'model_ratios',
      meta: { label: t('Model Ratios'), mobileHidden: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Model Ratios')} />
      ),
      cell: ({ row }) => {
        const value = row.getValue('model_ratios') as string
        const truncated =
          value.length > 60 ? value.slice(0, 60) + '...' : value
        return (
          <div className='max-w-[300px] truncate font-mono text-xs text-muted-foreground'>
            {truncated}
          </div>
        )
      },
      enableSorting: false,
    },
    {
      accessorKey: 'created_time',
      meta: { label: t('Created'), mobileHidden: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Created')} />
      ),
      cell: ({ row }) => {
        return (
          <div className='min-w-[140px] font-mono text-sm'>
            {formatTimestampToDate(row.getValue('created_time'))}
          </div>
        )
      },
    },
    {
      accessorKey: 'updated_time',
      meta: { label: t('Updated'), mobileHidden: true },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Updated')} />
      ),
      cell: ({ row }) => {
        return (
          <div className='min-w-[140px] font-mono text-sm'>
            {formatTimestampToDate(row.getValue('updated_time'))}
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
