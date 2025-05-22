// src/utils/formations.js
// Formation layouts using more Football Manager-centric role descriptors.
// Layouts are defined with attackers first (top of pitch) and GK last (bottom of pitch).

export const formations = {
  "442_classic": {
    // Renamed for clarity, classic 4-4-2
    name: "4-4-2 Classic",
    layout: [
      {
        count: 2,
        positions: [
          { id: "LST", role: "ST (C)" }, // Striker (Centre)
          { id: "RST", role: "ST (C)" }, // Striker (Centre)
        ],
      },
      {
        count: 4,
        positions: [
          { id: "ML", role: "M (L)" }, // Midfielder (Left)
          { id: "MCL", role: "M (C)" }, // Midfielder (Centre)
          { id: "MCR", role: "M (C)" }, // Midfielder (Centre)
          { id: "MR", role: "M (R)" }, // Midfielder (Right)
        ],
      },
      {
        count: 4,
        positions: [
          { id: "DL", role: "D (L)" }, // Defender (Left)
          { id: "DCL", role: "D (C)" }, // Defender (Centre Left)
          { id: "DCR", role: "D (C)" }, // Defender (Centre Right)
          { id: "DR", role: "D (R)" }, // Defender (Right)
        ],
      },
      { count: 1, positions: [{ id: "GK", role: "GK" }] }, // Goalkeeper
    ],
  },
  "433_wide": {
    // Common 4-3-3 with wingers
    name: "4-3-3 Wide",
    layout: [
      {
        count: 3,
        positions: [
          { id: "AML", role: "AM (L)" }, // Attacking Midfielder (Left) / Winger
          { id: "STC", role: "ST (C)" }, // Striker (Centre)
          { id: "AMR", role: "AM (R)" }, // Attacking Midfielder (Right) / Winger
        ],
      },
      {
        count: 3, // Central Midfielders
        positions: [
          { id: "MCL", role: "M (C)" },
          { id: "MC", role: "M (C)" },
          { id: "MCR", role: "M (C)" },
        ],
      },
      {
        count: 4,
        positions: [
          { id: "DL", role: "D (L)" },
          { id: "DCL", role: "D (C)" },
          { id: "DCR", role: "D (C)" },
          { id: "DR", role: "D (R)" },
        ],
      },
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
    ],
  },
  "433_dm_wide": {
    // 4-1-2-3 DM Wide in FM terms
    name: "4-1-2-3 DM Wide (4-3-3 DM)",
    layout: [
      {
        count: 3,
        positions: [
          { id: "AML", role: "AM (L)" },
          { id: "STC", role: "ST (C)" },
          { id: "AMR", role: "AM (R)" },
        ],
      },
      {
        // Two CMs ahead of DM
        count: 2,
        positions: [
          { id: "MCL", role: "M (C)" },
          { id: "MCR", role: "M (C)" },
        ],
      },
      { count: 1, positions: [{ id: "DM", role: "DM (C)" }] }, // Defensive Midfielder (Centre)
      {
        count: 4,
        positions: [
          { id: "DL", role: "D (L)" },
          { id: "DCL", role: "D (C)" },
          { id: "DCR", role: "D (C)" },
          { id: "DR", role: "D (R)" },
        ],
      },
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
    ],
  },
  "352_fm": {
    // Common 3-5-2 / 5-3-2 WB in FM
    name: "3-5-2 / 5-3-2 WB",
    layout: [
      {
        count: 2,
        positions: [
          { id: "STCL", role: "ST (C)" },
          { id: "STCR", role: "ST (C)" },
        ],
      },
      {
        // Midfield line including Wing-Backs
        count: 5,
        positions: [
          { id: "WBL", role: "WB (L)" }, // Wing-Back (Left)
          { id: "MCL", role: "M (C)" },
          { id: "MC", role: "M (C)" },
          { id: "MCR", role: "M (C)" },
          { id: "WBR", role: "WB (R)" }, // Wing-Back (Right)
        ],
      },
      {
        // Three Centre-Backs
        count: 3,
        positions: [
          { id: "DCL", role: "D (C)" },
          { id: "DC", role: "D (C)" },
          { id: "DCR", role: "D (C)" },
        ],
      },
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
    ],
  },
  "4231_dm_am_wide": {
    // Common 4-2-3-1 with DMs and AMs
    name: "4-2-3-1 DM AM Wide",
    layout: [
      { count: 1, positions: [{ id: "STC", role: "ST (C)" }] },
      {
        // Attacking Midfield Trio
        count: 3,
        positions: [
          { id: "AML", role: "AM (L)" },
          { id: "AMC", role: "AM (C)" },
          { id: "AMR", role: "AM (R)" },
        ],
      },
      {
        // Two Defensive Midfielders
        count: 2,
        positions: [
          { id: "DMCL", role: "DM (C)" },
          { id: "DMCR", role: "DM (C)" }, // Or DML/DMR if your data supports it
        ],
      },
      {
        count: 4,
        positions: [
          { id: "DL", role: "D (L)" },
          { id: "DCL", role: "D (C)" },
          { id: "DCR", role: "D (C)" },
          { id: "DR", role: "D (R)" },
        ],
      },
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
    ],
  },
  "41212_narrow_fm": {
    name: "4-1-2-1-2 Narrow (Diamond FM)",
    layout: [
      {
        count: 2,
        positions: [
          { id: "STCL", role: "ST (C)" },
          { id: "STCR", role: "ST (C)" },
        ],
      },
      { count: 1, positions: [{ id: "AMC", role: "AM (C)" }] },
      {
        count: 2, // Two Central Midfielders
        positions: [
          { id: "MCL", role: "M (C)" },
          { id: "MCR", role: "M (C)" },
        ],
      },
      { count: 1, positions: [{ id: "DMC", role: "DM (C)" }] },
      {
        count: 4,
        positions: [
          { id: "DL", role: "D (L)" },
          { id: "DCL", role: "D (C)" },
          { id: "DCR", role: "D (C)" },
          { id: "DR", role: "D (R)" },
        ],
      },
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
    ],
  },
  "343_fm": {
    // Common 3-4-3 with Wingers and two CMs
    name: "3-4-3 FM",
    layout: [
      {
        // Front Three
        count: 3,
        positions: [
          { id: "AML", role: "AM (L)" },
          { id: "STC", role: "ST (C)" },
          { id: "AMR", role: "AM (R)" },
        ],
      },
      {
        // Midfield Four (could be M(L), M(C), M(C), M(R) or WB(L), M(C), M(C), WB(R))
        // Using Wing-Backs for typical 3-4-3 width
        count: 4,
        positions: [
          { id: "WBL", role: "WB (L)" },
          { id: "MCL", role: "M (C)" },
          { id: "MCR", role: "M (C)" },
          { id: "WBR", role: "WB (R)" },
        ],
      },
      {
        // Three Centre-Backs
        count: 3,
        positions: [
          { id: "DCL", role: "D (C)" },
          { id: "DC", role: "D (C)" },
          { id: "DCR", role: "D (C)" },
        ],
      },
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
    ],
  },
  // Add more formations as needed, using FM-style roles like "D (RLC)", "M (C)", "AM (R)" etc.
};

/**
 * Retrieves the layout for a given formation key.
 * @param {string} key - The formation key (e.g., "442_classic").
 * @returns {Array} The layout array for the formation, or an empty array if not found.
 */
export function getFormationLayout(key) {
  return formations[key] ? formations[key].layout : [];
}
