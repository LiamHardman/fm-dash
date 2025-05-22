// src/utils/formations.js
// This utility file defines formation structures.

export const formations = {
  442: {
    name: "4-4-2",
    layout: [
      // Attackers (Top Row on display after reversal)
      {
        count: 2,
        positions: [
          { id: "LST", role: "ST" }, // Left Striker
          { id: "RST", role: "ST" }, // Right Striker
        ],
      },
      // Midfielders
      {
        count: 4,
        positions: [
          { id: "LM", role: "LM" }, // Left Midfielder
          { id: "LCM", role: "LCM" }, // Left Central Midfielder
          { id: "RCM", role: "RCM" }, // Right Central Midfielder
          { id: "RM", role: "RM" }, // Right Midfielder
        ],
      },
      // Defenders
      {
        count: 4,
        positions: [
          { id: "LB", role: "LB" }, // Left Back
          { id: "LCB", role: "LCB" }, // Left Centre Back
          { id: "RCB", role: "RCB" }, // Right Centre Back
          { id: "RB", role: "RB" }, // Right Back
        ],
      },
      // Goalkeeper (Bottom Row on display after reversal)
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
    ],
  },
  "433_attacking": {
    name: "4-3-3 Attacking",
    layout: [
      // Attackers
      {
        count: 3,
        positions: [
          { id: "LW", role: "LW" }, // Left Winger
          { id: "ST", role: "ST" }, // Striker
          { id: "RW", role: "RW" }, // Right Winger
        ],
      },
      // Midfielders
      {
        count: 3,
        positions: [
          { id: "LCM", role: "LCM" }, // Left Central Midfielder (Attacking)
          { id: "CM", role: "CM" }, // Central Midfielder
          { id: "RCM", role: "RCM" }, // Right Central Midfielder (Attacking)
        ],
      },
      // Defenders
      {
        count: 4,
        positions: [
          { id: "LB", role: "LB" },
          { id: "LCB", role: "LCB" },
          { id: "RCB", role: "RCB" },
          { id: "RB", role: "RB" },
        ],
      },
      // Goalkeeper
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
    ],
  },
  "433_defensive": {
    name: "4-3-3 Defensive (DM)",
    layout: [
      // Attackers
      {
        count: 3,
        positions: [
          { id: "LW", role: "LW" },
          { id: "ST", role: "ST" },
          { id: "RW", role: "RW" },
        ],
      },
      // Central Midfielders
      {
        count: 2,
        positions: [
          { id: "LCM", role: "CM" }, // Left Central Midfielder
          { id: "RCM", role: "CM" }, // Right Central Midfielder
        ],
      },
      // Defensive Midfielder
      { count: 1, positions: [{ id: "DM", role: "DM" }] },
      // Defenders
      {
        count: 4,
        positions: [
          { id: "LB", role: "LB" },
          { id: "LCB", role: "LCB" },
          { id: "RCB", role: "RCB" },
          { id: "RB", role: "RB" },
        ],
      },
      // Goalkeeper
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
    ],
  },
  352: {
    name: "3-5-2",
    layout: [
      // Attackers
      {
        count: 2,
        positions: [
          { id: "LST", role: "ST" }, // Left Striker
          { id: "RST", role: "ST" }, // Right Striker
        ],
      },
      // Midfielders (Wide players first, then central)
      {
        count: 5, // LWB, LCM, CM, RCM, RWB
        positions: [
          { id: "LWB", role: "LWB" }, // Left Wing-Back
          { id: "LCM", role: "LCM" }, // Left Central Midfielder
          { id: "CM", role: "CM" }, // Central Midfielder
          { id: "RCM", role: "RCM" }, // Right Central Midfielder
          { id: "RWB", role: "RWB" }, // Right Wing-Back
        ],
      },
      // Defenders
      {
        count: 3,
        positions: [
          { id: "LCB", role: "LCB" }, // Left Centre Back
          { id: "CB", role: "CB" }, // Central Centre Back
          { id: "RCB", role: "RCB" }, // Right Centre Back
        ],
      },
      // Goalkeeper
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
    ],
  },
  "4231_wide": {
    name: "4-2-3-1 Wide",
    layout: [
      // Striker
      { count: 1, positions: [{ id: "ST", role: "ST" }] },
      // Attacking Midfielders (Wide first, then central)
      {
        count: 3,
        positions: [
          { id: "LAM", role: "AML" }, // Left Attacking Midfielder (Winger)
          { id: "CAM", role: "AMC" }, // Central Attacking Midfielder
          { id: "RAM", role: "AMR" }, // Right Attacking Midfielder (Winger)
        ],
      },
      // Defensive Midfielders
      {
        count: 2,
        positions: [
          { id: "LDM", role: "LDM" }, // Left Defensive Midfielder
          { id: "RDM", role: "RDM" }, // Right Defensive Midfielder
        ],
      },
      // Defenders
      {
        count: 4,
        positions: [
          { id: "LB", role: "LB" },
          { id: "LCB", role: "LCB" },
          { id: "RCB", role: "RCB" },
          { id: "RB", role: "RB" },
        ],
      },
      // Goalkeeper
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
    ],
  },
  "532_wb": {
    // This is effectively the same as 3-5-2 but named differently, ensuring roles are correct.
    name: "5-3-2 WB",
    layout: [
      // Attackers
      {
        count: 2,
        positions: [
          { id: "LST", role: "ST" },
          { id: "RST", role: "ST" },
        ],
      },
      // Midfielders (Central)
      {
        count: 3,
        positions: [
          { id: "LCM", role: "LCM" },
          { id: "CM", role: "CM" },
          { id: "RCM", role: "RCM" },
        ],
      },
      // Wing-Backs and Defenders
      // For display, this row will be above the CBs after TeamViewPage reverses the whole layout.
      // PitchDisplay will render LWB on left, CBs in middle, RWB on right.
      {
        count: 5, // LWB, LCB, CB, RCB, RWB
        positions: [
          { id: "LWB", role: "LWB" },
          { id: "LCB", role: "LCB" },
          { id: "CB", role: "CB" },
          { id: "RCB", role: "RCB" },
          { id: "RWB", role: "RWB" },
        ],
      },
      // Goalkeeper
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
    ],
  },
  // Example: 3-4-3
  "343_diamond": {
    name: "3-4-3 Diamond",
    layout: [
      // Attackers
      {
        count: 3,
        positions: [
          { id: "LW", role: "LW" },
          { id: "ST", role: "ST" },
          { id: "RW", role: "RW" },
        ],
      },
      // Midfielders (Diamond)
      { count: 1, positions: [{ id: "CAM", role: "AMC" }] }, // Attacking Mid
      {
        count: 2,
        positions: [
          { id: "LM", role: "LM" }, // Left Mid (or LWB if you prefer that role name)
          { id: "RM", role: "RM" }, // Right Mid (or RWB)
        ],
      },
      { count: 1, positions: [{ id: "DM", role: "DM" }] }, // Defensive Mid

      // Defenders
      {
        count: 3,
        positions: [
          { id: "LCB", role: "LCB" },
          { id: "CB", role: "CB" },
          { id: "RCB", role: "RCB" },
        ],
      },
      // Goalkeeper
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
    ],
  },
};

/**
 * Retrieves the layout for a given formation key.
 * The layout is structured for display, typically with attackers at the top.
 * TeamViewPage will reverse this array for pitch display (GK at bottom).
 * @param {string} key - The formation key (e.g., "442").
 * @returns {Array} The layout array for the formation, or an empty array if not found.
 */
export function getFormationLayout(key) {
  return formations[key] ? formations[key].layout : [];
}
