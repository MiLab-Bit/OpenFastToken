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
*/
import { useState } from 'react'
import { Download } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth-store'
import { exportCsv } from '@/lib/exportCsv'

interface ExportCsvButtonProps {
  url: string
  params?: Record<string, unknown>
  filename?: string
  /** Minimum role required to see the button. Omit to show to all logged-in users. */
  requireRole?: number
  label?: string
  loadingLabel?: string
  className?: string
}

export function ExportCsvButton({
  url,
  params,
  filename,
  requireRole,
  label,
  loadingLabel,
  className,
}: ExportCsvButtonProps) {
  const { t } = useTranslation()
  const [loading, setLoading] = useState(false)
  const role = useAuthStore((state) => state.auth.user?.role ?? 0)

  if (requireRole !== undefined && role < requireRole) {
    return null
  }

  const handleClick = async () => {
    setLoading(true)
    try {
      await exportCsv(url, { params, filename })
    } finally {
      setLoading(false)
    }
  }

  return (
    <Button
      variant='outline'
      size='sm'
      onClick={handleClick}
      disabled={loading}
      className={className}
    >
      <Download className='h-4 w-4' />
      {loading ? (loadingLabel ?? t('Exporting...')) : (label ?? t('Export CSV'))}
    </Button>
  )
}
