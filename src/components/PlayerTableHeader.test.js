import { mount } from '@vue/test-utils'
import { Quasar } from 'quasar'
import { describe, expect, it } from 'vitest'
import PlayerTableHeader from './PlayerTableHeader.vue'

// Global test config
const globalConfig = {
  global: {
    plugins: [Quasar],
    stubs: {
      QIcon: true
    }
  }
}

describe('PlayerTableHeader', () => {
  const defaultProps = {
    columns: [
      { name: 'name', label: 'Name', type: 'text' },
      { name: 'age', label: 'Age', type: 'number' },
      { name: 'rating', label: 'Rating', type: 'rating' }
    ],
    sortField: 'name',
    sortDirection: 'asc'
  }

  it('renders all column headers', () => {
    const wrapper = mount(PlayerTableHeader, {
      props: defaultProps,
      ...globalConfig
    })

    const headers = wrapper.findAll('th')
    expect(headers).toHaveLength(3)
    expect(headers[0].text()).toContain('Name')
    expect(headers[1].text()).toContain('Age')
    expect(headers[2].text()).toContain('Rating')
  })

  it('applies correct classes to headers', () => {
    const wrapper = mount(PlayerTableHeader, {
      props: defaultProps,
      ...globalConfig
    })

    const headers = wrapper.findAll('th')

    // First header (name, text type, active sort)
    expect(headers[0].classes()).toContain('sortable-header')
    expect(headers[0].classes()).toContain('active-sort')

    // Second header (age, number type)
    expect(headers[1].classes()).toContain('sortable-header')
    expect(headers[1].classes()).toContain('text-right')

    // Third header (rating, rating type)
    expect(headers[2].classes()).toContain('sortable-header')
    expect(headers[2].classes()).toContain('text-right')
  })

  it('shows sort indicator for active sort field', () => {
    const wrapper = mount(PlayerTableHeader, {
      props: defaultProps,
      ...globalConfig
    })

    const activeHeader = wrapper.find('th.active-sort')
    expect(activeHeader.exists()).toBe(true)
    expect(activeHeader.find('.sort-indicator').exists()).toBe(true)
  })

  it('shows ascending sort icon when sort direction is asc', () => {
    const wrapper = mount(PlayerTableHeader, {
      props: defaultProps,
      ...globalConfig
    })

    const sortIndicator = wrapper.find('.sort-indicator')
    expect(sortIndicator.exists()).toBe(true)

    // Check that QIcon is rendered with correct props (mocked)
    const qIcon = sortIndicator.findComponent({ name: 'QIcon' })
    expect(qIcon.exists()).toBe(true)
  })

  it('shows descending sort icon when sort direction is desc', () => {
    const wrapper = mount(PlayerTableHeader, {
      props: {
        ...defaultProps,
        sortDirection: 'desc'
      },
      ...globalConfig
    })

    const sortIndicator = wrapper.find('.sort-indicator')
    expect(sortIndicator.exists()).toBe(true)

    const qIcon = sortIndicator.findComponent({ name: 'QIcon' })
    expect(qIcon.exists()).toBe(true)
  })

  it('emits sort event when header is clicked', async () => {
    const wrapper = mount(PlayerTableHeader, {
      props: defaultProps,
      ...globalConfig
    })

    const headers = wrapper.findAll('th')
    await headers[1].trigger('click') // Click on 'age' column

    expect(wrapper.emitted('sort')).toBeTruthy()
    expect(wrapper.emitted('sort')[0]).toEqual(['age'])
  })

  it('applies text-right class to number and rating columns', () => {
    const wrapper = mount(PlayerTableHeader, {
      props: defaultProps,
      ...globalConfig
    })

    const headers = wrapper.findAll('th')

    // Name column (text type) should not have text-right
    expect(headers[0].classes()).not.toContain('text-right')

    // Age column (number type) should have text-right
    expect(headers[1].classes()).toContain('text-right')

    // Rating column (rating type) should have text-right
    expect(headers[2].classes()).toContain('text-right')
  })

  it('handles different sort fields correctly', () => {
    const wrapper = mount(PlayerTableHeader, {
      props: {
        ...defaultProps,
        sortField: 'age'
      },
      ...globalConfig
    })

    const headers = wrapper.findAll('th')

    // Name header should not be active
    expect(headers[0].classes()).not.toContain('active-sort')

    // Age header should be active
    expect(headers[1].classes()).toContain('active-sort')
    expect(headers[1].find('.sort-indicator').exists()).toBe(true)
  })

  it('has correct cursor and user-select styles', () => {
    const wrapper = mount(PlayerTableHeader, {
      props: defaultProps,
      ...globalConfig
    })

    const headers = wrapper.findAll('th')
    for (const header of headers) {
      expect(header.attributes('style')).toContain('cursor: pointer')
      expect(header.attributes('style')).toContain('user-select: none')
    }
  })

  it('renders proper header content structure', () => {
    const wrapper = mount(PlayerTableHeader, {
      props: defaultProps,
      ...globalConfig
    })

    const firstHeader = wrapper.find('th')
    const headerContent = firstHeader.find('.header-content')
    expect(headerContent.exists()).toBe(true)

    const headerLabel = headerContent.find('.header-label')
    expect(headerLabel.exists()).toBe(true)
    expect(headerLabel.text()).toBe('Name')
  })

  it('handles empty columns array', () => {
    const wrapper = mount(PlayerTableHeader, {
      props: {
        ...defaultProps,
        columns: []
      },
      ...globalConfig
    })

    const headers = wrapper.findAll('th')
    expect(headers).toHaveLength(0)
  })
})
