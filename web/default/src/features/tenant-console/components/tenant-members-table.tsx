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
import { useMemo, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  type ColumnDef,
  type PaginationState,
  getCoreRowModel,
  useReactTable,
} from '@tanstack/react-table'
import { useTranslation } from 'react-i18next'
import { DataTableColumnHeader, DataTablePage } from '@/components/data-table'
import { StatusBadge, type StatusVariant } from '@/components/status-badge'
import { TableId } from '@/components/table-id'
import { getTenantMembers } from '../api'
import { type TenantMember } from '../types'

// ============================================================================
// Tenant Members Table
//
// Read-only, server-paginated roster of the current tenant. Pagination lives in
// local component state rather than URL state: the console route hosts no other
// list, so keeping it local avoids adding a search schema to the route while
// still reusing the shared DataTablePage shell.
//
// Sorting is disabled on purpose — the backend paginates, so sorting a single
// page client-side would be misleading.
// ============================================================================

const MEMBER_ROLE_VARIANTS: Record<string, StatusVariant> = {
  admin: 'warning',
  member: 'success',
}

const MEMBER_STATUS_VARIANTS: Record<string, StatusVariant> = {
  active: 'success',
  inactive: 'neutral',
}

export function TenantMembersTable() {
  const { t } = useTranslation()
  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 20,
  })

  const { data, isLoading, isFetching } = useQuery({
    queryKey: ['tenant-members', pagination.pageIndex + 1, pagination.pageSize],
    queryFn: async () => {
      const result = await getTenantMembers(
        pagination.pageIndex + 1,
        pagination.pageSize
      )
      return {
        items: result.data?.items ?? [],
        total: result.data?.total ?? 0,
      }
    },
    placeholderData: (previousData) => previousData,
  })

  const members = useMemo<TenantMember[]>(() => data?.items ?? [], [data])
  const columns = useTenantMembersColumns()

  const table = useReactTable({
    data: members,
    columns,
    state: { pagination },
    enableSorting: false,
    enableRowSelection: false,
    onPaginationChange: setPagination,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    pageCount: Math.max(
      1,
      Math.ceil((data?.total ?? 0) / Math.max(1, pagination.pageSize))
    ),
  })

  return (
    <DataTablePage
      table={table}
      columns={columns}
      isLoading={isLoading}
      isFetching={isFetching}
      emptyTitle={t('No Members Found')}
      emptyDescription={t('This enterprise has no members yet.')}
      skeletonKeyPrefix='tenant-members-skeleton'
      toolbarProps={null}
      paginationInFooter={false}
    />
  )
}

function useTenantMembersColumns(): ColumnDef<TenantMember>[] {
  const { t } = useTranslation()

  // Literal keys keep the i18n extraction script able to discover them.
  const roleLabels: Record<string, string> = {
    admin: t('Admin'),
    member: t('Member'),
  }
  const statusLabels: Record<string, string> = {
    active: t('Active'),
    inactive: t('Inactive'),
  }

  return [
    {
      accessorKey: 'id',
      meta: { label: t('ID') },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('ID')} />
      ),
      cell: ({ row }) => (
        <TableId value={row.getValue('id') as number} className='w-[60px]' />
      ),
    },
    {
      accessorKey: 'username',
      meta: { label: t('Username') },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Username')} />
      ),
      cell: ({ row }) => (
        <div className='max-w-[180px] truncate font-medium'>
          {(row.getValue('username') as string) || '-'}
        </div>
      ),
    },
    {
      accessorKey: 'display_name',
      meta: { label: t('Display Name') },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Display Name')} />
      ),
      cell: ({ row }) => (
        <div className='max-w-[180px] truncate'>
          {(row.getValue('display_name') as string) || '-'}
        </div>
      ),
    },
    {
      accessorKey: 'role',
      meta: { label: t('Role') },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Role')} />
      ),
      cell: ({ row }) => {
        const role = (row.getValue('role') as string) || ''
        if (!role) {
          return <span className='text-sm'>-</span>
        }
        return (
          <StatusBadge
            label={roleLabels[role] ?? role}
            variant={MEMBER_ROLE_VARIANTS[role] ?? 'neutral'}
            copyable={false}
          />
        )
      },
    },
    {
      accessorKey: 'status',
      meta: { label: t('Status') },
      header: ({ column }) => (
        <DataTableColumnHeader column={column} title={t('Status')} />
      ),
      cell: ({ row }) => {
        const status = (row.getValue('status') as string) || ''
        if (!status) {
          return <span className='text-sm'>-</span>
        }
        return (
          <StatusBadge
            label={statusLabels[status] ?? status}
            variant={MEMBER_STATUS_VARIANTS[status] ?? 'neutral'}
            copyable={false}
          />
        )
      },
    },
  ]
}
