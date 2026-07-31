import { describe, it, expect, beforeEach } from 'vitest'
import {
  useSystemConfigStore,
  getSystemName,
  getLogo,
  getFooterHtml,
  DEFAULT_CURRENCY_CONFIG,
} from './system-config-store'

describe('system-config-store', () => {
  beforeEach(() => {
    useSystemConfigStore.setState({
      config: {
        systemName: 'Default',
        logo: 'default-logo',
        currency: { ...DEFAULT_CURRENCY_CONFIG },
      },
      loading: true,
      loadedLogoUrl: 'default-logo',
    } as never)
  })

  it('exposes selector helpers', () => {
    expect(getSystemName()).toBe('Default')
    expect(getLogo()).toBe('default-logo')
    expect(getFooterHtml()).toBeUndefined()
  })

  it('setConfig merges top-level fields', () => {
    useSystemConfigStore.getState().setConfig({ systemName: 'Acme' })
    expect(getSystemName()).toBe('Acme')
    // logo untouched
    expect(getLogo()).toBe('default-logo')
  })

  it('setConfig deep-merges the currency sub-object', () => {
    useSystemConfigStore.getState().setConfig({
      currency: { displayInCurrency: false } as never,
    })
    const { currency } = useSystemConfigStore.getState().config
    expect(currency.displayInCurrency).toBe(false)
    // other currency fields preserved
    expect(currency.quotaDisplayType).toBe(
      DEFAULT_CURRENCY_CONFIG.quotaDisplayType
    )
  })

  it('setConfig can set footerHtml', () => {
    useSystemConfigStore.getState().setConfig({ footerHtml: '<p>hi</p>' })
    expect(getFooterHtml()).toBe('<p>hi</p>')
  })

  it('setLoading updates loading flag', () => {
    useSystemConfigStore.getState().setLoading(false)
    expect(useSystemConfigStore.getState().loading).toBe(false)
  })
})
