import { PLAYSTYLE_PLUS_CATALOG } from './playstyle-plus-catalog.mjs';

const badgeEscape = value => String(value).replace(/[&<>"']/g, char => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&apos;' })[char]);
function solidHex(value) {
  if (typeof value !== 'string' || !/^#(?:[0-9a-f]{3}|[0-9a-f]{6})$/i.test(value)) throw new TypeError('Badge theme colours must be opaque #RGB or #RRGGBB values.');
  return (value.length === 4 ? '#' + [...value.slice(1)].map(char => char + char).join('') : value).toLowerCase();
}
function luminance(value) {
  const hex = solidHex(value).slice(1);
  const rgb = [0, 2, 4].map(offset => parseInt(hex.slice(offset, offset + 2), 16) / 255).map(channel => channel <= .04045 ? channel / 12.92 : ((channel + .055) / 1.055) ** 2.4);
  return rgb[0] * .2126 + rgb[1] * .7152 + rgb[2] * .0722;
}
export function colourContrast(first, second) {
  const values = [luminance(first), luminance(second)].sort((a, b) => b - a);
  return (values[0] + .05) / (values[1] + .05);
}

/** One policy for every icon and edition; no sampling of artwork or per-icon colours. */
export function resolvePlaystylePalette(theme) {
  const surface = solidHex(theme.accentSurface ?? theme.rail);
  const preferred = solidHex(theme.ink);
  const preferredContrast = colourContrast(surface, preferred);
  const ink = preferredContrast >= 4.5 ? preferred : ['#000000', '#ffffff'].sort((a, b) => colourContrast(surface, b) - colourContrast(surface, a))[0];
  return { surface, ink, rim: ink, guard: surface, contrast: colourContrast(surface, ink), usedFallback: ink !== preferred };
}

export function playstylePlusLabel(id) {
  if (id == null || id === '') return '';
  if (!Object.hasOwn(PLAYSTYLE_PLUS_CATALOG, id)) throw new TypeError(`Unknown Playstyle+ identifier: ${String(id)}`);
  return `${PLAYSTYLE_PLUS_CATALOG[id].label} Playstyle+`;
}

/** Exact supplied vector paths with resolved presentation attributes, ready for SVG/PNG. */
export function renderPlaystylePlus({ id, theme, x = 8, y = 744, size = 174 }) {
  const label = playstylePlusLabel(id);
  if (!label) return '';
  if (![x, y, size].every(Number.isFinite) || size <= 0) throw new TypeError('Badge placement must use finite coordinates and a positive size.');
  const entry = PLAYSTYLE_PLUS_CATALOG[id];
  const palette = resolvePlaystylePalette(theme);
  return `<svg x="${x}" y="${y}" width="${size}" height="${size}" viewBox="-12 -12 584 584" role="img" aria-label="${badgeEscape(label)}" data-playstyle-plus="${badgeEscape(id)}"><title>${badgeEscape(label)}</title><path d="${entry.hull}" fill="${palette.guard}" stroke="${palette.guard}" stroke-width="16" stroke-linejoin="round"/>${entry.paths.map(path => `<path d="${path.d}" fill="${path.role === 'surface' ? palette.surface : path.role === 'rim' ? palette.rim : palette.ink}"${path.fillRule ? ` fill-rule="${path.fillRule}"` : ''}${path.clipRule ? ` clip-rule="${path.clipRule}"` : ''}/>`).join('')}</svg>`;
}
