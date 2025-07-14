import { config } from '@vue/test-utils'
import { vi } from 'vitest'

global.IntersectionObserver = vi.fn(() => ({
  disconnect: vi.fn(),
  observe: vi.fn(),
  unobserve: vi.fn()
}))

global.ResizeObserver = vi.fn(() => ({
  disconnect: vi.fn(),
  observe: vi.fn(),
  unobserve: vi.fn()
}))

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn()
  }))
})

global.scrollTo = vi.fn()

global.console = {
  ...console,
  warn: vi.fn(),
  error: vi.fn()
}

config.global.stubs = {
  // Stub Quasar components by default
  QBtn: true,
  QInput: true,
  QIcon: true,
  QCard: true,
  QCardSection: true,
  QList: true,
  QItem: true,
  QItemSection: true,
  QItemLabel: true,
  QChip: true,
  QSpinner: true,
  QDialog: true,
  QTable: true,
  QTh: true,
  QTd: true,
  QTr: true,
  QCheckbox: true,
  QSelect: true,
  QSlider: true,
  QRange: true,
  QToggle: true,
  QTab: true,
  QTabs: true,
  QTabPanel: true,
  QTabPanels: true
}

const sessionStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn()
}
global.sessionStorage = sessionStorageMock

const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn()
}
global.localStorage = localStorageMock
