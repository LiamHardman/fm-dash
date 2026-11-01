// FM eligibility from playstyleplus.md. Groups specify an average; separate
// groups must each pass. Priority is the minimum total required by the rule.
const average = (minimum, ...keys) => ({ minimum, keys });
const rule = (id, groups, extra = {}) => ({ id, groups, ...extra });

const OUTFIELD_RULES = [
  rule("acrobatic", [average(17, "Agi", "Tec", "Fin")]),
  rule("chip-shot", [
    average(18, "Fin"),
    average(15, "Ant"),
    average(15, "Vis"),
  ]),
  rule("dead-ball", [average(17, "Fre")]),
  rule("finesse-shot", [
    average(17, "Fin"),
    average(17, "Tec"),
    average(15, "Lon"),
  ]),
  rule("game-changer", [average(17, "Fin", "Fla")]),
  rule("low-driven-shot", [average(18, "Fin"), average(18, "Tec")], {
    maximum: ["Lon", 14],
  }),
  rule("power-shot", [average(17, "Fin", "Lon")]),
  rule("precision-header", [average(16, "Hea", "Jum")]),
  rule("incisive-pass", [average(18, "Pas")]),
  rule("inventive", [average(16, "Pas", "Tec", "Fla")]),
  rule("long-ball-pass", [average(16, "Pas", "Vis")]),
  rule("pinged-pass", [average(16, "Pas", "Tec")]),
  rule("tiki-taka", [average(15, "OtB", "Cmp", "Pas")]),
  rule("whipped-pass", [average(17, "Cro")]),
  rule("first-touch", [average(17, "Fir")]),
  rule("press-proven", [average(16, "Dri", "Agi", "Pas", "Bal")]),
  rule("rapid", [average(18, "Pac")]),
  rule("technical", [average(17, "Agi", "Dri")]),
  rule("trickster", [average(17, "Dri", "Fla")]),
  rule("aerial-fortress", [average(16, "Jum", "Str")]),
  rule("anticipate", [average(17, "Cmp", "Tck")]),
  rule("block", [average(16, "Ant", "Pos", "Mar")]),
  rule("intercept", [average(17, "Mar", "Pos")]),
  rule("jockey", [average(17, "Mar")]),
  rule("slide-tackle", [average(17, "Tck")]),
  rule("bruiser", [average(17, "Str")], {
    positions: ["CB", "RB", "LB", "RWB", "LWB"],
  }),
  rule("enforcer", [average(17, "Str")], { positions: ["ST"] }),
  rule("long-throw", [average(16, "L Th")], { last: true }),
  rule("quick-step", [average(18, "Acc")]),
  rule("relentless", [average(17, "Wor"), average(17, "Sta")]),
];

const GOALKEEPER_RULES = [
  rule("rush-out", [average(16, "TRO", "1v1")]),
  rule("cross-claimer", [average(17, "Pun")]),
  rule("deflector", [
    average(14, "Aer"),
    average(16, "Han"),
    average(14, "Ref"),
  ]),
  rule("far-reach", [average(17, "Aer"), average(14, "Jum")]),
  rule("far-throw", [average(16, "Thr")]),
  rule("footwork", [average(15, "1v1", "Ref")]),
  rule("pinged-pass", [average(14, "Pas", "Tec")]),
  rule("long-ball-pass", [average(14, "Pas", "Vis")]),
];

function attributeReader(player) {
  // Only read FM attribute maps: top-level pac/pas/etc. are card stats (0–99).
  const attributes = {
    ...player.attributes,
    ...player.numeric_attributes,
    ...player.numericAttributes,
  };
  return (key) => {
    const raw = attributes[key];
    if (
      raw == null ||
      (typeof raw !== "number" && typeof raw !== "string") ||
      String(raw).trim() === ""
    )
      return undefined;
    const value = Number(raw);
    return Number.isFinite(value) && value >= 1 && value <= 20
      ? value
      : undefined;
  };
}

export function selectPlaystylePlus(
  player,
  { goalkeeper = false, positions = [] } = {},
) {
  const read = attributeReader(player);
  const rules = goalkeeper ? GOALKEEPER_RULES : OUTFIELD_RULES;
  let best;
  let bestPriority = -1;
  for (const candidate of rules) {
    if (
      candidate.positions &&
      !candidate.positions.some((position) => positions.includes(position))
    )
      continue;
    if (
      !candidate.groups.every(({ minimum, keys }) => {
        const values = keys.map(read);
        return (
          values.every((value) => value !== undefined) &&
          values.reduce((sum, value) => sum + value, 0) >= minimum * keys.length
        );
      })
    )
      continue;
    if (candidate.maximum) {
      const value = read(candidate.maximum[0]);
      if (value === undefined || value > candidate.maximum[1]) continue;
    }
    const priority = candidate.last
      ? 0
      : candidate.groups.reduce(
          (sum, { minimum, keys }) => sum + minimum * keys.length,
          0,
        );
    // Equal requirements retain the definition order for a stable result.
    if (priority > bestPriority) {
      best = candidate.id;
      bestPriority = priority;
    }
  }
  return best;
}
