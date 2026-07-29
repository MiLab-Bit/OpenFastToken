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
 
'use client'

import {
  type ComponentProps,
  createContext,
  type HTMLAttributes,
  type KeyboardEventHandler,
  type ReactNode,
  useContext,
  useEffect,
  useRef,
  useState,
} from 'react'
import type { Element } from 'hast'
import { CheckIcon, CopyIcon } from 'lucide-react'
import {
  type BundledLanguage,
  codeToHtml,
  type ShikiTransformer,
} from 'shiki/bundle/web'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'

type CodeBlockProps = HTMLAttributes<HTMLDivElement> & {
  code: string
  language: BundledLanguage
  showLineNumbers?: boolean
  /** Number of lines to keep visible when the block is collapsed */
  collapsedLines?: number
  /** Whether the block should start collapsed */
  defaultCollapsed?: boolean
  /** Maximum number of lines expanded before truncation */
  maxExpandedLines?: number
  /** Whether to render the toolbar (e.g. copy button) */
  showToolbar?: boolean
}

type CodeBlockContextType = {
  code: string
}

const CodeBlockContext = createContext<CodeBlockContextType>({
  code: '',
})

const lineNumberTransformer: ShikiTransformer = {
  name: 'line-numbers',
  line(node: Element, line: number) {
    node.children.unshift({
      type: 'element',
      tagName: 'span',
      properties: {
        className: [
          'inline-block',
          'min-w-10',
          'mr-4',
          'text-right',
          'select-none',
          'text-muted-foreground',
        ],
      },
      children: [{ type: 'text', value: String(line) }],
    })
  },
}

export async function highlightCode(
  code: string,
  language: BundledLanguage,
  showLineNumbers = false
) {
  const transformers: ShikiTransformer[] = showLineNumbers
    ? [lineNumberTransformer]
    : []

  return codeToHtml(code, {
    lang: language,
    themes: {
      light: 'one-light',
      dark: 'one-dark-pro',
    },
    transformers,
  })
}

export const CodeBlock = ({
  code,
  language,
  showLineNumbers = false,
  className,
  children,
  ...props
}: CodeBlockProps) => {
  const [html, setHtml] = useState<string>('')

  useEffect(() => {
    let cancelled = false
    highlightCode(code, language, showLineNumbers).then((next) => {
      if (!cancelled) {
        setHtml(next)
      }
    })
    return () => {
      cancelled = true
    }
  }, [code, language, showLineNumbers])

  return (
    <CodeBlockContext.Provider value={{ code }}>
      <div
        className={cn(
          'group bg-background text-foreground relative w-full overflow-hidden rounded-md border',
          className
        )}
        {...props}
      >
        <div className='relative'>
          <div
            className='[&>pre]:bg-background! [&>pre]:text-foreground! overflow-hidden [&_code]:font-mono [&_code]:text-sm [&>pre]:m-0 [&>pre]:p-4 [&>pre]:text-sm'
            // biome-ignore lint/security/noDangerouslySetInnerHtml: "this is needed."
            dangerouslySetInnerHTML={{ __html: html }}
          />
          {children && (
            <div className='absolute top-2 right-2 flex items-center gap-2'>
              {children}
            </div>
          )}
        </div>
      </div>
    </CodeBlockContext.Provider>
  )
}

export type CodeBlockCopyButtonProps = ComponentProps<typeof Button> & {
  onCopy?: () => void
  onError?: (error: Error) => void
  timeout?: number
}

export const CodeBlockCopyButton = ({
  onCopy,
  onError,
  timeout = 2000,
  children,
  className,
  ...props
}: CodeBlockCopyButtonProps) => {
  const [isCopied, setIsCopied] = useState(false)
  const timerRef = useRef<NodeJS.Timeout | null>(null)
  const { code } = useContext(CodeBlockContext)

  const copyToClipboard = async () => {
    if (typeof window === 'undefined' || !navigator?.clipboard?.writeText) {
      onError?.(new Error('Clipboard API not available'))
      return
    }

    try {
      await navigator.clipboard.writeText(code)
      setIsCopied(true)
      onCopy?.()
      // Clear any existing timer
      if (timerRef.current) {
        clearTimeout(timerRef.current)
      }
      timerRef.current = setTimeout(() => {
        setIsCopied(false)
        timerRef.current = null
      }, timeout)
    } catch (error) {
      onError?.(error as Error)
    }
  }

  // Cleanup timer on unmount
  useEffect(() => {
    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current)
        timerRef.current = null
      }
    }
  }, [])

  const Icon = isCopied ? CheckIcon : CopyIcon

  return (
    <Button
      className={cn('shrink-0', className)}
      onClick={copyToClipboard}
      size='icon'
      variant='ghost'
      {...props}
    >
      {children ?? <Icon size={14} />}
    </Button>
  )
}

type CodeBlockEditorProps = {
  /** Current editor content */
  value: string
  /** Called whenever the content changes */
  onChange: (value: string) => void
  /** Language hint (informational only) */
  language?: string
  /** Disable editing */
  readOnly?: boolean
  /** Extra class names for the wrapper */
  className?: string
  /** Header title node */
  title?: ReactNode
  /** Header action nodes (e.g. save / cancel buttons) */
  actions?: ReactNode
  /** Accessible label for the textarea */
  ariaLabel?: string
  /** Keyboard handler forwarded to the underlying textarea */
  onKeyDown?: KeyboardEventHandler<HTMLTextAreaElement>
  /** Visible row count */
  rows?: number
}

/**
 * Editable code block used by the playground message editor.
 *
 * This is a lightweight, dependency-free textarea editor that mirrors the
 * `CodeBlockEditor` contract previously provided by the ai-elements package.
 */
export function CodeBlockEditor({
  value,
  onChange,
  language = 'markdown',
  readOnly = false,
  className,
  title,
  actions,
  ariaLabel,
  onKeyDown,
  rows = 8,
}: CodeBlockEditorProps) {
  return (
    <div
      className={cn(
        'bg-background text-foreground w-full overflow-hidden rounded-md border',
        className
      )}
    >
      {(title || actions) && (
        <div className='flex items-center justify-between border-b px-3 py-2'>
          <div className='text-sm font-medium'>{title}</div>
          <div className='flex items-center gap-2'>{actions}</div>
        </div>
      )}
      <textarea
        aria-label={ariaLabel}
        className='font-mono text-sm w-full resize-y bg-transparent p-4 outline-none'
        data-language={language}
        readOnly={readOnly}
        rows={rows}
        spellCheck={false}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        onKeyDown={onKeyDown}
      />
    </div>
  )
}
