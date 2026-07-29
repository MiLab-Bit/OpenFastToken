import { describe, it, expect, beforeEach } from 'vitest'
import { useNotificationStore } from './notification-store'

describe('notification-store', () => {
  beforeEach(() => {
    // Reset to a clean initial state before each test.
    useNotificationStore.setState({
      lastReadNotice: '',
      readAnnouncementKeys: [],
      closedUntilDate: null,
      // actions are preserved by zustand when only data fields are set
    } as never)
    // Re-bind actions (setState with partial data keeps existing actions).
    useNotificationStore.getState().markNoticeRead('')
    useNotificationStore.getState().markAnnouncementsRead([])
    useNotificationStore.getState().setClosedUntilDate(null)
  })

  it('initializes with empty read state', () => {
    const s = useNotificationStore.getState()
    expect(s.lastReadNotice).toBe('')
    expect(s.readAnnouncementKeys).toEqual([])
    expect(s.closedUntilDate).toBeNull()
  })

  it('markNoticeRead trims and stores content', () => {
    useNotificationStore.getState().markNoticeRead('  hello world  ')
    expect(useNotificationStore.getState().lastReadNotice).toBe('hello world')
  })

  it('markAnnouncementsRead dedupes keys', () => {
    useNotificationStore.getState().markAnnouncementsRead(['a', 'a', 'b'])
    expect(useNotificationStore.getState().readAnnouncementKeys).toEqual(['a', 'b'])
    expect(useNotificationStore.getState().isAnnouncementRead('a')).toBe(true)
    expect(useNotificationStore.getState().isAnnouncementRead('z')).toBe(false)
  })

  it('isAnnouncementRead reflects marked keys', () => {
    expect(useNotificationStore.getState().isAnnouncementRead('x')).toBe(false)
    useNotificationStore.getState().markAnnouncementsRead(['x'])
    expect(useNotificationStore.getState().isAnnouncementRead('x')).toBe(true)
  })

  it('isNoticeClosed compares closedUntilDate to today', () => {
    expect(useNotificationStore.getState().isNoticeClosed()).toBe(false)
    useNotificationStore.getState().setClosedUntilDate('not-today')
    expect(useNotificationStore.getState().isNoticeClosed()).toBe(false)
    useNotificationStore.getState().setClosedUntilDate(new Date().toDateString())
    expect(useNotificationStore.getState().isNoticeClosed()).toBe(true)
  })
})
