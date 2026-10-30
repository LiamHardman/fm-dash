import { renderPlaystylePlus, playstylePlusLabel } from './playstyle-plus.mjs';
import { resolveEdition } from './editions.mjs';

export const OUTFIELD_LABELS = ['PAC', 'SHO', 'PAS', 'DRI', 'DEF', 'PHY'];
export const GOALKEEPER_LABELS = ['DIV', 'HAN', 'KIC', 'REF', 'SPD', 'POS'];
export const CREST_EDITIONS = Object.fromEntries(['gold', 'wonderkid', 'icon'].map(edition => [edition, resolveEdition({ edition, rare: edition === 'gold' }).theme]));

const escapeXml = value => String(value ?? '').replace(/[&<>"']/g, char => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&apos;' })[char]);
const text = (x, y, value, size, attrs = '') => `<text x="${x}" y="${y}" font-size="${size}" ${attrs}>${escapeXml(value)}</text>`;
const imageSource = source => {
  if (!source) return '';
  if (typeof source !== 'string') throw new TypeError('Image sources must be URL strings.');
  const value = source.trim();
  const hasScheme = /^[a-z][a-z0-9+.-]*:/i.test(value);
  if (!hasScheme || /^(?:data:image\/(?:png|jpeg|webp|svg\+xml)[;,]|https?:\/\/|file:\/\/\/|blob:)/i.test(value)) return escapeXml(value);
  throw new TypeError('Image sources must be image data, HTTP(S), blob or local URL paths.');
};

export function normalisePositions(primaryPosition, alternatePositions = []) {
  if (!Array.isArray(alternatePositions)) throw new TypeError('alternatePositions must be an array.');
  const primary = String(primaryPosition ?? '').trim().toUpperCase();
  const alternates = [...new Set(alternatePositions.map(position => String(position).trim().toUpperCase()).filter(position => position && position !== primary))];
  if (alternates.length > 8) throw new RangeError('This card layout accommodates up to eight alternate positions.');
  return { primary, alternates };
}

function playerName(value) {
  const name = String(value ?? '').trim().toUpperCase();
  const words = name.split(/\s+/);
  if (name.length <= 16 || words.length === 1) return [{ value: name, y: 1040, size: Math.min(66, 700 / Math.max(Array.from(name).length * .98, 1)) }];
  let split = 1;
  for (let index = 2; index < words.length; index++) {
    if (Math.abs(words.slice(0, index).join(' ').length - words.slice(index).join(' ').length) < Math.abs(words.slice(0, split).join(' ').length - words.slice(split).join(' ').length)) split = index;
  }
  const lines = [words.slice(0, split).join(' '), words.slice(split).join(' ')];
  const size = Math.min(55, 720 / (Math.max(...lines.map(line => Array.from(line).length)) * .98));
  return lines.map((line, index) => ({ value: line, y: 1006 + index * 57, size }));
}

function identityMarks(player, theme, centreY = 1262) {
  const marks = [
    player.nation && { ...player.nation, kind: 'nation', width: 120, height: 80 },
    player.club && { ...player.club, kind: 'club', width: 96, height: 96 },
  ].filter(Boolean);
  if (!marks.length) return '';
  const gap = 44, total = marks.reduce((sum, mark) => sum + mark.width, 0) + (marks.length - 1) * gap;
  let x = 512 - total / 2;
  return marks.map(mark => {
    const currentX = x, y = centreY - mark.height / 2;
    x += mark.width + gap;
    const label = escapeXml(`${mark.kind === 'nation' ? 'Nation' : 'Club'}: ${mark.label ?? mark.code ?? 'not supplied'}`);
    return `<g role="img" aria-label="${label}"><title>${label}</title>${mark.src ? `<image href="${imageSource(mark.src)}" x="${currentX}" y="${y}" width="${mark.width}" height="${mark.height}" preserveAspectRatio="xMidYMid meet"/>` : `<rect x="${currentX}" y="${y}" width="${mark.width}" height="${mark.height}" rx="5" fill="none" stroke="${theme.ink}" stroke-width="2"/>${text(currentX + mark.width / 2, centreY + 10, mark.code ?? '—', 40, 'text-anchor="middle" font-weight="700"')}`}</g>`;
  }).join('');
}

function positionRail(alternates, theme) {
  if (!alternates.length) return '';
  const x = 909, width = 169, top = 448;
  const rowHeight = Math.min(99, 610 / alternates.length), bottom = top + rowHeight * alternates.length;
  return `<g aria-label="Alternate positions: ${escapeXml(alternates.join(', '))}"><path d="M${x} ${top}l${width / 2} -24 ${width / 2} 24V${bottom}l-${width / 2} 24-${width / 2}-24Z" fill="${theme.rail}" stroke="${theme.edge}" stroke-width="5"/>${alternates.map((position, index) => `${index ? `<path d="M${x + 2} ${top + rowHeight * index}h${width - 4}" stroke="${theme.edge}" stroke-width="2"/>` : ''}${text(x + width / 2, top + rowHeight * (index + .5) + 16, position, Math.min(49, rowHeight * .59, 139 / Math.max(position.length * .7, 1)), `fill="${theme.tabInk}" text-anchor="middle" font-weight="900"`)}`).join('')}</g>`;
}

function supportingInfo(player, theme) {
  const values = [['AGE', player.age], ['VALUE', player.transferValue], ['WAGE / WEEK', player.weeklyWage]];
  return `<g aria-label="Supporting information"><path d="M95 1508H929l26 155-26 155H95l-26-155Z" fill="${theme.rail}" stroke="${theme.edge}" stroke-width="3"/>${values.map(([label, value], index) => {
    const y = 1580 + index * 94;
    const display = String(value ?? '\u2014');
    const units = [...display].reduce((sum, char) => sum + (/[MW@]/.test(char) ? 1 : /[A-Z]/.test(char) ? .85 : /[0-9]/.test(char) ? .66 : /\s/.test(char) ? .4 : .7), 0);
    const size = Math.min(64, 420 / Math.max(units, 1));
    return text(140, y, label, 48, `fill="${theme.tabInk}" font-weight="600"`) + text(885, y, display, size, `fill="${theme.tabInk}" text-anchor="end" font-weight="700"`);
  }).join('')}</g>`;
}

/** Pure SVG renderer. Ratings, attribute values and positions are supplied by the host. */
export function renderPlayerCard({ player, edition = 'squad-player', rare = false, baseEdition = 'gold', artwork, portrait, portraitMode = 'cutout', showSupportingInfo = false, idPrefix = 'crest' }) {
  if (!player || typeof player !== 'object') throw new TypeError('player is required.');
  const appearance = resolveEdition({ edition, rare, baseEdition });
  const { theme, layout } = appearance;
  const artSource = artwork || appearance.artwork;
  // Standard artwork is centred on x=512; Icon has its own inset portrait window.
  const portraitCentre = layout.portraitCenterX ?? 512;
  const portraitBounds = portraitMode === 'photo'
    ? { x: portraitCentre - 282, y: 300, width: 564, height: 675 }
    : { x: portraitCentre - 340, y: 250, width: 680, height: 725 };
  const portraitWindow = `M${portraitBounds.x} ${portraitBounds.y}h${portraitBounds.width}V975H${portraitBounds.x}Z`;

  if (!['cutout', 'photo'].includes(portraitMode)) throw new TypeError('portraitMode must be cutout or photo.');
  if (!/^[a-zA-Z][\w-]*$/.test(idPrefix)) throw new TypeError('idPrefix must be a valid unique SVG identifier prefix.');
  const { primary, alternates } = normalisePositions(player.primaryPosition, player.alternatePositions);
  const labels = player.goalkeeper ? GOALKEEPER_LABELS : OUTFIELD_LABELS;
  const rating = player.rating ?? '—';
  const plusName = playstylePlusLabel(player.playstylePlus);
  const description = labels.map(label => `${label}: ${player.attributes?.[label] ?? 'not supplied'}`).join(', ');
  const identities = [player.nation && `Nation: ${player.nation.label ?? player.nation.code ?? 'not supplied'}`, player.club && `Club: ${player.club.label ?? player.club.code ?? 'not supplied'}`].filter(Boolean).join(', ');
  const support = showSupportingInfo ? `, Age: ${player.age ?? 'not supplied'}, Value: ${player.transferValue ?? 'not supplied'}, Weekly wage: ${player.weeklyWage ?? 'not supplied'}` : '';
  const title = `${player.name ?? 'Unnamed player'}, ${rating}, ${primary || 'position not supplied'}, ${theme.label}${alternates.length ? ', alternate positions ' + alternates.join(', ') : ''}${plusName ? ', ' + plusName : ''}, ${description}${identities ? ', ' + identities : ''}${support}`;
  // Reserve the rail footprint even when empty, so changing positions does not resize the card.
  const width = 1100;
  const height = showSupportingInfo ? 1870 : 1536;
  const ratingOutline = layout.ratingOutline ? ` stroke="${layout.ratingOutline.colour}" stroke-width="${layout.ratingOutline.width}" stroke-linejoin="round" paint-order="stroke"` : '';
  const stats = labels.map((label, index) => {
    const x = 188 + index * 130;
    return `<g aria-label="${label}: ${escapeXml(player.attributes?.[label] ?? 'not supplied')}">${text(x, layout.statLabelY, label, 35, 'text-anchor="middle" font-weight="700" letter-spacing="1"')}${text(x, layout.statValueY, player.attributes?.[label] ?? '—', 71, 'text-anchor="middle" font-weight="900"')}</g>`;
  }).join('');
  return `<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}" role="img" aria-labelledby="${idPrefix}-title" style="font-family:Arial,sans-serif;letter-spacing:0;fill:${theme.ink}"><title id="${idPrefix}-title">${escapeXml(title)}</title><defs><clipPath id="${idPrefix}-art-bounds"><rect width="1024" height="1536" rx="8"/></clipPath><clipPath id="${idPrefix}-portrait-window"><path d="${portraitWindow}"/></clipPath><linearGradient id="${idPrefix}-fade" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="white"/><stop offset=".86" stop-color="white"/><stop offset="1" stop-color="black"/></linearGradient><mask id="${idPrefix}-portrait-mask"><rect width="1024" height="975" fill="url(#${idPrefix}-fade)"/></mask></defs>${layout.portraitBackdrop ? `<path d="${layout.portraitBackdrop.d}" fill="${layout.portraitBackdrop.fill}"/>` : ''}<image href="${imageSource(artSource)}" width="1024" height="1536" clip-path="url(#${idPrefix}-art-bounds)"/>${portrait ? `<g clip-path="url(#${idPrefix}-portrait-window)" mask="url(#${idPrefix}-portrait-mask)"><image href="${imageSource(portrait)}" x="${portraitBounds.x}" y="${portraitBounds.y}" width="${portraitBounds.width}" height="${portraitBounds.height}" preserveAspectRatio="xMidYMax meet"/></g>` : `<path d="M${portraitCentre - 110} 950v-145q0-85 110-85t110 85v145z" fill="${theme.ink}" opacity=".13"/><circle cx="${portraitCentre}" cy="615" r="74" fill="${theme.ink}" opacity=".13"/>`}<g aria-label="Rating and primary position" style="text-anchor:middle;letter-spacing:0;font-kerning:none">${text(208, 364, rating, 130, 'font-weight="900" text-anchor="middle"' + ratingOutline)}${text(208, 425, primary || '—', Math.min(55, 185 / Math.max(primary.length * .7, 1)), 'font-weight="700" text-anchor="middle"')}</g>${playerName(player.name).map(line => text(512, line.y + layout.nameOffset, line.value, Math.min(line.size, layout.nameMaxSize), 'font-weight="900" text-anchor="middle"')).join('')}${stats}${identityMarks(player, theme, layout.identityY)}${renderPlaystylePlus({ id: player.playstylePlus, theme })}${positionRail(alternates, theme)}${showSupportingInfo ? supportingInfo(player, theme) : ''}</svg>`;
}

/** Compatibility entry point for the original three-edition studies. New hosts use renderPlayerCard. */
export function renderForgedCrest(options) {
  const edition = options.edition ?? 'gold';
  return renderPlayerCard({ ...options, edition, rare: options.rare ?? edition === 'gold' });
}
