/*
Copyright (C) 2023-2026 QuantumNous

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

/**
 * Sanitize an image source for safe rendering.
 *
 * Returns the original (relative or safe-protocol) source unchanged, or
 * `null` when the value is empty or uses a potentially dangerous protocol
 * (e.g. `javascript:`). This mirrors the behavior previously expected from
 * the `sanitizeImageSrc` helper in `stream-markdown-parser`.
 */
export function sanitizeImageSrc(src: string): string | null {
  if (!src) return null

  // Relative URLs and fragment-only references are safe to keep as-is.
  if (!/^[a-z][a-z0-9+.-]*:/i.test(src)) {
    return src
  }

  // Only allow data: URIs that actually carry image payloads.
  if (/^data:image\//i.test(src)) {
    return src
  }

  // Allow http(s) and blob: URLs.
  if (/^(https?:|blob:)/i.test(src)) {
    return src
  }

  return null
}

/**
 * Decide whether a link should open in a new browser tab.
 *
 * External `http(s)` links open in a new tab; same-origin / relative links
 * and anchors stay in place. Mirrors the previous
 * `shouldOpenLinkInNewTab` helper from `stream-markdown-parser`.
 */
export function shouldOpenLinkInNewTab(href: string): boolean {
  if (!href) return false
  return /^https?:\/\//i.test(href)
}
