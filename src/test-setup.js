import { config } from '@vue/test-utils'
import { vi } from 'vitest'

// Mock IntersectionObserver
global.IntersectionObserver = vi.fn(() => ({
  disconnect: vi.fn(),
  observe: vi.fn(),
  unobserve: vi.fn()
}))

// Mock ResizeObserver
global.ResizeObserver = vi.fn(() => ({
  disconnect: vi.fn(),
  observe: vi.fn(),
  unobserve: vi.fn()
}))

// Mock matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // deprecated
    removeListener: vi.fn(), // deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn()
  }))
})

// Mock window.scrollTo
global.scrollTo = vi.fn()

// Mock console methods to reduce noise in tests
global.console = {
  ...console,
  warn: vi.fn(),
  error: vi.fn()
}

// Set up global Vue Test Utils config
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

// Mock sessionStorage
const sessionStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn()
}
global.sessionStorage = sessionStorageMock

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn()
}
global.localStorage = localStorageMock
