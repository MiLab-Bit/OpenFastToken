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
import {
  beginPasskeyVerification,
  finishPasskeyVerification,
  getPasskeyStatus,
} from '../passkey'
import {
  buildAssertionResult,
  prepareCredentialRequestOptions,
  type PasskeyServerResponse,
} from '@/lib/passkey'
import { api } from '@/lib/api'
import type { VerificationMethod, VerificationMethods } from './types'

/**
 * Fetch available verification methods for the current user.
 */

/**
 * Detect if the current environment supports Passkeys.
 */
async function detectPasskeySupport(): Promise<boolean> {
  try {
    return (
      typeof navigator !== 'undefined' &&
      typeof navigator.credentials !== 'undefined' &&
      typeof PublicKeyCredential !== 'undefined'
    )
  } catch {
    return false
  }
}

export async function checkVerificationMethods(): Promise<VerificationMethods> {
  try {
    const [passkeyResponse, passkeySupported] =
      await Promise.all([
        getPasskeyStatus(),
        detectPasskeySupport(),
      ])

    const hasPasskey =
      Boolean(passkeyResponse?.success) &&
      Boolean(passkeyResponse?.data?.enabled)

    return {
      has2FA: false, // 2FA feature removed
      hasPasskey,
      passkeySupported,
    }
  } catch (error) {
     
    console.error('[Secure Verification] Failed to check methods', error)
    return {
      has2FA: false,
      hasPasskey: false,
      passkeySupported: false,
    }
  }
}

/**
 * Execute a verification flow based on the method type.
 */
export async function verify(
  method: VerificationMethod,
  _code?: string
): Promise<void> {
  switch (method) {
    case 'passkey':
      return verifyPasskey()
    default:
      throw new Error(`Unsupported verification method: ${method}`)
  }
}

/**
 * Perform Passkey verification flow.
 */
async function verifyPasskey(): Promise<void> {
  if (typeof navigator === 'undefined' || !navigator.credentials) {
    throw new Error('Passkey verification is not supported in this environment')
  }

  try {
    const beginResponse = await beginPasskeyVerification()
    if (!beginResponse.success) {
      throw new Error(beginResponse.message || 'Failed to start verification')
    }

    const publicKey = prepareCredentialRequestOptions(
      (beginResponse.data?.options ?? beginResponse.data) as PasskeyServerResponse
    )

    const credential = (await navigator.credentials.get({
      publicKey,
    })) as PublicKeyCredential | null

    if (!credential) {
      throw new Error('Passkey verification was cancelled')
    }

    const assertion = buildAssertionResult(credential)
    if (!assertion) {
      throw new Error('Unable to build Passkey assertion')
    }

    const finishResponse = await finishPasskeyVerification(assertion)
    if (!finishResponse.success) {
      throw new Error(finishResponse.message || 'Passkey verification failed')
    }

    const verifyResponse = await api.post('/api/verify', {
      method: 'passkey',
    })

    if (!verifyResponse.data?.success) {
      throw new Error(
        verifyResponse.data?.message || 'Failed to complete verification'
      )
    }
  } catch (error: unknown) {
    if (error instanceof DOMException && error.name === 'NotAllowedError') {
      throw new Error('Passkey verification was cancelled or timed out', {
        cause: error,
      })
    }
    if (error instanceof DOMException && error.name === 'InvalidStateError') {
      throw new Error(
        'Passkey verification is not available in the current state',
        { cause: error }
      )
    }
    throw error
  }
}
