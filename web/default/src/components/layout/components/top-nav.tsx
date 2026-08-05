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
import { Link } from '@tanstack/react-router'
import { Menu } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { type TopNavLink } from '../types'

type TopNavProps = React.HTMLAttributes<HTMLElement> & {
  links: TopNavLink[]
}

/**
 * 顶部导航栏组件
 * 在大屏幕显示水平导航，在小屏幕显示下拉菜单
 */
export function TopNav({ className, links, ...props }: TopNavProps) {
  // Normalize links so all optional fields have defaults
  const normalizedLinks = useMemo(
    () =>
      links.map((link) => ({
        isActive: false,
        disabled: false,
        external: false,
        requiresAuth: false,
        ...link,
      })),
    [links]
  )

  // Active link detection (parent-exposes via useRouterState if available)
  const [activeHref, setActiveHref] = useState<string | null>(null)

  return (
    <>
      {/* 移动端下拉菜单 */}
      <div className='lg:hidden'>
        <DropdownMenu modal={false}>
          <DropdownMenuTrigger
            render={<Button size='icon' variant='outline' className='size-7' />}
          >
            <Menu />
          </DropdownMenuTrigger>
          <DropdownMenuContent side='bottom' align='start'>
            {normalizedLinks.map(({ title, href, isActive, disabled, external }) => (
          <DropdownMenuItem
            key={`${title}-${href}`}
            render={
              external ? (
                <a
                  href={href}
                  target='_blank'
                  rel='noopener noreferrer'
                  className={!isActive && !activeHref ? 'text-muted-foreground' : ''}
                >
                  {title}
                </a>
              ) : (
                <Link
                  to={href}
                  onClick={() => setActiveHref(href)}
                  className={!isActive && !activeHref ? 'text-muted-foreground' : ''}
                  disabled={disabled}
                >
                  {title}
                </Link>
              )
            }
          />
        ))}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {/* 桌面端水平导航 */}
      <nav
        className={cn(
          'hidden items-center lg:flex lg:gap-0.5',
          className
        )}
        {...props}
      >
        {normalizedLinks.map(({ title, href, isActive, disabled, external }) =>
          external ? (
            <a
              key={`${title}-${href}`}
              href={href}
              target='_blank'
              rel='noopener noreferrer'
              className={`hover:text-foreground rounded-lg px-3 py-1.5 text-[13px] font-medium transition-colors duration-200 ${isActive || activeHref === href ? 'text-foreground' : 'text-muted-foreground'}`}
            >
              {title}
            </a>
          ) : (
            <Link
              key={`${title}-${href}`}
              to={href}
              onClick={() => setActiveHref(href)}
              disabled={disabled}
              className={`hover:text-foreground rounded-lg px-3 py-1.5 text-[13px] font-medium transition-colors duration-200 ${isActive || activeHref === href ? 'text-foreground' : 'text-muted-foreground'}`}
            >
              {title}
            </Link>
          )
        )}
      </nav>
    </>
  )
}
