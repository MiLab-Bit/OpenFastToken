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

/**
 * Typed structural cast helper.
 *
 * Use ONLY when TypeScript cannot prove that a value already satisfies the
 * target type at runtime, but you (the author) know it does — e.g. a synthetic
 * aggregate row that reuses a string as a numeric `id`, or a partial object
 * that the caller completes afterwards.
 *
 * This is preferred over `value as unknown as T`, which the lint rule bans via
 * `no-restricted-syntax`, because it keeps a single, intentional, traceable
 * assertion instead of a double cast.
 */
export function unsafeCast<T>(value: unknown): T {
  return value as T
}
