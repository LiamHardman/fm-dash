import { shallowReactive } from 'vue'

export const COLOUR_SCALES = {
  fm: { label: 'FM attributes', min: 1, max: 20 },
  fifa: { label: 'FIFA-style stats', min: 0, max: 100 },
}

const defaultColours = ['#ef5350', '#ffa726', '#42a5f5', '#66bb6a', '#26a69a', '#7e57c2']
export const defaultBands = (scale) =>
  (scale === 'fm' ? [1, 8, 11, 14, 16, 18] : [0, 40, 55, 70, 80, 90]).map((min, i) => ({
    min,
    color: defaultColours[i],
  }))

// Shared with class helpers used outside component setup. Replacing a palette
// invalidates Vue renders without memoizing stale band classifications.
export const activeColourBands = shallowReactive({
  fm: defaultBands('fm'),
  fifa: defaultBands('fifa'),
})

export function validateBands(bands, scale) {
  const range = COLOUR_SCALES[scale]
  if (!Array.isArray(bands) || bands.length < 1 || bands.length > 20) {
    return 'Choose between 1 and 20 bands.'
  }
  if (bands[0].min !== range.min) return `The first band must start at ${range.min}.`
  let previous = range.min - 1
  for (const band of bands) {
    if (!Number.isInteger(band.min) || band.min <= previous || band.min > range.max) {
      return `Band starts must be increasing whole numbers from ${range.min} to ${range.max}.`
    }
    if (!/^#[0-9a-f]{6}$/i.test(band.color)) return 'Enter each colour as #RRGGBB.'
    previous = band.min
  }
  return ''
}

export function bandIndex(value, bands) {
  for (let i = bands.length - 1; i >= 0; i--) {
    if (value >= bands[i].min) return i
  }
  return -1
}

function resolveBand(value, maxScale) {
  if (value === null || value === undefined || String(value).trim() === '') return null
  const numeric = Number(value)
  if (!Number.isFinite(numeric) || !Number.isFinite(maxScale) || maxScale <= 0) return null
  const scale = maxScale === 20 ? 'fm' : 'fifa'
  const scaled = maxScale === 20 ? numeric : (numeric / maxScale) * 100
  const index = bandIndex(scaled, activeColourBands[scale])
  return index < 0 ? null : { scale, index, band: activeColourBands[scale][index] }
}

export function getUnifiedRatingClass(value, maxScale = 100) {
  const match = resolveBand(value, maxScale)
  return match ? `attribute-colour attribute-colour-${match.scale}-${match.index}` : 'rating-na'
}

export function getRatingColour(value, maxScale = 100) {
  return resolveBand(value, maxScale)?.band.color || '#9ca3af'
}

// Pick whichever text colour has the higher WCAG contrast against the chosen background.
export function colourText(color) {
  const rgb = [1, 3, 5].map((offset) => {
    const component = Number.parseInt(color.slice(offset, offset + 2), 16) / 255
    return component <= 0.04045 ? component / 12.92 : ((component + 0.055) / 1.055) ** 2.4
  })
  const luminance = rgb[0] * 0.2126 + rgb[1] * 0.7152 + rgb[2] * 0.0722
  return (luminance + 0.05) / 0.05 >= 1.05 / (luminance + 0.05) ? '#000000' : '#ffffff'
}

export function applyColourVariables(style) {
  for (const [scale, bands] of Object.entries(activeColourBands)) {
    bands.forEach((band, index) => {
      style.setProperty(`--attribute-${scale}-${index}`, band.color)
      style.setProperty(`--attribute-${scale}-${index}-text`, colourText(band.color))
    })
  }
}
