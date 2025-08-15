export const rawTechnicalAttributeKeysConst = [
  'Cor',
  'Cro',
  'Dri',
  'Fin',
  'Fir',
  'Fre',
  'Hea',
  'Lon',
  'L Th',
  'Mar',
  'Pas',
  'Pen',
  'Tck',
  'Tec',
]

export const rawMentalAttributeKeysConst = [
  'Agg',
  'Ant',
  'Bra',
  'Cmp',
  'Cnt',
  'Dec',
  'Det',
  'Fla',
  'Ldr',
  'OtB',
  'Pos',
  'Tea',
  'Vis',
  'Wor',
]

export const rawPhysicalAttributeKeysConst = [
  'Acc',
  'Agi',
  'Bal',
  'Jum',
  'Nat',
  'Pac',
  'Sta',
  'Str',
]

export const rawGoalkeeperAttributeKeysConst = [
  'Aer',
  'Cmd',
  'Com',
  'Ecc',
  'Han',
  'Kic',
  '1v1',
  'Pun',
  'Ref',
  'TRO',
  'Thr',
]

export const allRawFmAttributeKeys = [
  ...rawTechnicalAttributeKeysConst,
  ...rawMentalAttributeKeysConst,
  ...rawPhysicalAttributeKeysConst,
  ...rawGoalkeeperAttributeKeysConst,
]

export const formatFilterKeyPrefix = (attrKey) => {
  return attrKey.replace(/\s+/g, '').replace(/\(|\)/g, '')
}

export const parseTransferValueRange = (valueString) => {
  if (!valueString) return { min: 0, max: 0 }

  const rangeMatch = valueString.match(/£([\d.]+)M\s*-\s*£([\d.]+)M/)
  if (rangeMatch) {
    const minValue = Number.parseFloat(rangeMatch[1]) * 1000000
    const maxValue = Number.parseFloat(rangeMatch[2]) * 1000000
    return { min: minValue, max: maxValue }
  }

  const rangeKMatch = valueString.match(/£([\d.]+)K\s*-\s*£([\d.]+)K/i)
  if (rangeKMatch) {
    const minValue = Number.parseFloat(rangeMatch[1]) * 1000
    const maxValue = Number.parseFloat(rangeMatch[2]) * 1000
    return { min: minValue, max: maxValue }
  }

  const singleMatch = valueString.match(/£([\d.]+)M/)
  if (singleMatch) {
    const value = Number.parseFloat(singleMatch[1]) * 1000000
    return { min: value, max: value }
  }

  const kMatch = valueString.match(/£([\d.]+)k/i)
  if (kMatch) {
    const value = Number.parseFloat(kMatch[1]) * 1000
    return { min: value, max: value }
  }

  const kMatchUpper = valueString.match(/£([\d.]+)K/)
  if (kMatchUpper) {
    const value = Number.parseFloat(kMatchUpper[1]) * 1000
    return { min: value, max: value }
  }

  const singleValueMatch = valueString.match(/£([\d,]+)/)
  if (singleValueMatch) {
    const number = Number.parseInt(singleValueMatch[1].replace(/,/g, ''))
    let value = number
    if (number >= 1000000) {
      value = (number / 1000000) * 1000000
    } else if (number >= 1000) {
      value = (number / 1000) * 1000
    }
    return { min: value, max: value }
  }

  return { min: 0, max: 0 }
}

export const formatNumber = (num) => {
  if (num >= 1000000) {
    return `${(num / 1000000).toFixed(1).replace(/\.0$/, '')}M`
  }
  if (num >= 1000) {
    return `${(num / 1000).toFixed(1).replace(/\.0$/, '')}K`
  }
  return num?.toString() || '0'
}
