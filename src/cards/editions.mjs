const standardLayout = { nameOffset: 0, nameMaxSize: 55, statLabelY: 1095, statValueY: 1165, identityY: 1245 };
const iconLayout = { ...standardLayout, nameOffset: 44, nameMaxSize: 48, statLabelY: 1157, statValueY: 1221, identityY: 1285, portraitCenterX: 610 };
const theme = (ink, rail, edge, tabInk, accentSurface) => ({ ink, rail, edge, tabInk, accentSurface });
const presentation = (id, artwork, palette, material, layout = standardLayout) => ({ id, artwork, theme: palette, material, layout });

/** Single source of truth for public identifiers, variants, art, layout and colour tokens. */
export const EDITION_MANIFEST = {
  'squad-player': { label: 'Squad Player', alias: true, defaultBase: 'gold', description: 'Uses a supplied Bronze, Silver or Gold base edition; defaults to non-rare Gold.' },
  bronze: { label: 'Bronze', variants: {
    nonrare: presentation('bronze-nonrare', new URL('../../assets/cards/bronze-nonrare.png', import.meta.url).href, theme('#3b2619', '#5c3825', '#f4c59d', '#fff1d8', '#c99766'), 'Brushed copper-bronze casting and a single bevel.'),
    rare: presentation('bronze-rare', new URL('../../assets/cards/bronze-rare.png', import.meta.url).href, theme('#3b2619', '#5c3825', '#f4c59d', '#fff1d8', '#d3a16e'), 'Polished copper ribs, hammered bronze recesses and a layered rim.'),
  } },
  silver: { label: 'Silver', variants: {
    nonrare: presentation('silver-nonrare', new URL('../../assets/cards/silver-nonrare.png', import.meta.url).href, theme('#26313b', '#536370', '#e0e9ee', '#ffffff', '#c7d0d6'), 'Satin sterling silver, shallow engraving and a single rail.'),
    rare: presentation('silver-rare', new URL('../../assets/cards/silver-rare.png', import.meta.url).href, theme('#232d37', '#d7e0e8', '#627789', '#232d37', '#d7e0e8'), 'Mirror platinum fans, diamond-cut fluting and stepped chrome rails.'),
  } },
  gold: { label: 'Gold', variants: {
    nonrare: presentation('gold-nonrare', new URL('../../assets/cards/gold-nonrare.png', import.meta.url).href, theme('#33200b', '#69512d', '#ffe6a3', '#fff5dc', '#e2c47f'), 'Satin gold, restrained shallow seams and a single bevel.'),
    rare: presentation('gold-rare', new URL('../../assets/cards/gold-rare.png', import.meta.url).href, theme('#33200b', '#493017', '#f8dfa4', '#fff5dc', '#e4bf70'), 'Accepted sculpted gold fans, mineral recesses and polished layered rails.'),
  } },
  bargain: { label: 'Bargain', variants: { standard: presentation('bargain', new URL('../../assets/cards/bargain.png', import.meta.url).href, theme('#f4fff8', '#0b4e3e', '#a3e8c5', '#f4fff8', '#0b4e3e'), 'Emerald glass and malachite gemstone strata with silver clasps.') } },
  wonderkid: { label: 'Wonderkid', variants: { standard: presentation('wonderkid', new URL('../../assets/cards/wonderkid.png', import.meta.url).href, theme('#fff7ff', '#422267', '#e8baff', '#fff7ff', '#422267'), 'Accepted violet crystal recesses and projecting iridescent chrome lattice.') } },
  totw: { label: 'Team of the Week', variants: { standard: presentation('totw', new URL('../../assets/cards/totw.png', import.meta.url).href, theme('#fff3d7', '#26241d', '#dfb861', '#fff3d7', '#2e2a20'), 'Black folded titanium, matte basalt and sharp gold blades.') } },
  motm: { label: 'Man of the Match', variants: { standard: presentation('motm', new URL('../../assets/cards/motm.png', import.meta.url).href, theme('#fff2e1', '#742331', '#f7bc82', '#fff2e1', '#9b372d'), 'Carmine enamel, ruby glass and copper floodlight rays.') } },
  hero: { label: 'Hero', variants: { standard: presentation('hero', new URL('../../assets/cards/hero.png', import.meta.url).href, theme('#fff3df', '#153653', '#ec9e63', '#fff3df', '#eaa266'), 'Navy enamel, copper and ivory impact rays with raised orange ribbons.') } },
  tots: { label: 'Team of the Season', variants: { standard: presentation('tots', new URL('../../assets/cards/tots.png', import.meta.url).href, theme('#fff5d9', '#163b82', '#e4c06c', '#fff5d9', '#183b8a'), 'Cobalt sapphire enamel, substantial gold laurel relief and a trophy halo.', { ...standardLayout, ratingOutline: { colour: '#112b69', width: 6 } }) } },
  icon: { label: 'Icon', variants: { standard: presentation('icon', new URL('../../assets/cards/icon.png', import.meta.url).href, theme('#392e20', '#eee7d8', '#98743c', '#392e20', '#eee7d8'), 'Accepted ivory carved marble, architectural gold supports and inset portrait arch.', iconLayout) } },
};

export function resolveEdition({ edition = 'squad-player', rare = false, baseEdition = 'gold' } = {}) {
  if (!Object.hasOwn(EDITION_MANIFEST, edition)) throw new TypeError(`Unknown card edition: ${String(edition)}`);
  if (typeof rare !== 'boolean') throw new TypeError('rare must be a boolean.');
  const requested = EDITION_MANIFEST[edition];
  const resolvedId = requested.alias ? baseEdition : edition;
  if (requested.alias && !['bronze', 'silver', 'gold'].includes(baseEdition)) throw new TypeError('Squad Player baseEdition must be bronze, silver or gold.');
  const definition = EDITION_MANIFEST[resolvedId];
  const supportsRare = Object.hasOwn(definition.variants, 'rare');
  if (rare && !supportsRare) throw new TypeError(`${definition.label} does not have a rare variant.`);
  const variant = supportsRare ? rare ? 'rare' : 'nonrare' : 'standard';
  const appearance = definition.variants[variant];
  const appearanceLabel = definition.label + (variant === 'rare' ? ' Rare' : variant === 'nonrare' ? ' Non-rare' : '');
  const label = requested.alias ? `${requested.label} · ${appearanceLabel}` : appearanceLabel;
  return { ...appearance, requestedEdition: edition, edition: resolvedId, variant, rare, label, appearanceLabel, theme: { ...appearance.theme, label } };
}

export const FAMILY_PRESENTATIONS = Object.entries(EDITION_MANIFEST).flatMap(([edition, definition]) => definition.alias ? [] : Object.keys(definition.variants).map(variant => ({ edition, rare: variant === 'rare' })));
