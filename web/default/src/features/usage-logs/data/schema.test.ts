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
import { describe, it, expect } from 'vitest'
import { usageLogSchema } from './schema'

describe('usageLogSchema', () => {
  it('applies defaults for optional fields', () => {
    const res = usageLogSchema.safeParse({
      id: 1,
      user_id: 2,
      created_at: 3,
      type: 4,
      content: 'hello',
    })
    expect(res.success).toBe(true)
    if (res.success) {
      expect(res.data.username).toBe('')
      expect(res.data.token_name).toBe('')
      expect(res.data.model_name).toBe('')
      expect(res.data.quota).toBe(0)
      expect(res.data.prompt_tokens).toBe(0)
      expect(res.data.completion_tokens).toBe(0)
      expect(res.data.channel_name).toBe('')
    }
  })

  it('fails when required fields are missing', () => {
    const res = usageLogSchema.safeParse({ id: 1 })
    expect(res.success).toBe(false)
  })

  it('parses a fully populated record', () => {
    const res = usageLogSchema.safeParse({
      id: 10,
      user_id: 20,
      created_at: 1700000000,
      type: 1,
      content: 'chat completion',
      username: 'bob',
      token_name: 'gpt-4',
      model_name: 'gpt-4',
      quota: 500,
      prompt_tokens: 10,
      completion_tokens: 20,
      use_time: 1234,
      is_stream: true,
      channel: 5,
      channel_name: 'openai',
      token_id: 7,
      group: 'team',
      ip: '127.0.0.1',
      other: '{}',
      request_id: 'req-1',
      upstream_request_id: 'up-1',
    })
    expect(res.success).toBe(true)
    if (res.success) {
      expect(res.data.username).toBe('bob')
      expect(res.data.quota).toBe(500)
      expect(res.data.channel_name).toBe('openai')
    }
  })
})
