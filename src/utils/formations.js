// src/utils/formations.js
// This is a new utility file to define formation structures.
// Create this file in `src/utils/formations.js`

export const formations = {
  442: {
    name: "4-4-2",
    layout: [
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
      {
        count: 4,
        positions: [
          { id: "RB", role: "RB" },
          { id: "RCB", role: "RCB" },
          { id: "LCB", role: "LCB" },
          { id: "LB", role: "LB" },
        ],
      },
      {
        count: 4,
        positions: [
          { id: "RM", role: "RM" },
          { id: "RCM", role: "RCM" },
          { id: "LCM", role: "LCM" },
          { id: "LM", role: "LM" },
        ],
      },
      {
        count: 2,
        positions: [
          { id: "RST", role: "ST" },
          { id: "LST", role: "ST" },
        ],
      },
    ],
  },
  "433_attacking": {
    name: "4-3-3 Attacking",
    layout: [
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
      {
        count: 4,
        positions: [
          { id: "RB", role: "RB" },
          { id: "RCB", role: "RCB" },
          { id: "LCB", role: "LCB" },
          { id: "LB", role: "LB" },
        ],
      },
      {
        count: 3,
        positions: [
          { id: "RCM", role: "RCM" },
          { id: "CM", role: "CM" },
          { id: "LCM", role: "LCM" },
        ],
      },
      {
        count: 3,
        positions: [
          { id: "RW", role: "RW" },
          { id: "ST", role: "ST" },
          { id: "LW", role: "LW" },
        ],
      },
    ],
  },
  "433_defensive": {
    name: "4-3-3 Defensive (DM)",
    layout: [
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
      {
        count: 4,
        positions: [
          { id: "RB", role: "RB" },
          { id: "RCB", role: "RCB" },
          { id: "LCB", role: "LCB" },
          { id: "LB", role: "LB" },
        ],
      },
      { count: 1, positions: [{ id: "DM", role: "DM" }] },
      {
        count: 2,
        positions: [
          { id: "RCM", role: "RCM" },
          { id: "LCM", role: "LCM" },
        ],
      },
      {
        count: 3,
        positions: [
          { id: "RW", role: "RW" },
          { id: "ST", role: "ST" },
          { id: "LW", role: "LW" },
        ],
      },
    ],
  },
  352: {
    name: "3-5-2",
    layout: [
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
      {
        count: 3,
        positions: [
          { id: "RCB", role: "CB" },
          { id: "CB", role: "CB" },
          { id: "LCB", role: "CB" },
        ],
      },
      { count: 1, positions: [{ id: "RWB", role: "RWB" }] },
      {
        count: 3,
        positions: [
          { id: "RCM", role: "CM" },
          { id: "CM", role: "CM" },
          { id: "LCM", role: "CM" },
        ],
      },
      { count: 1, positions: [{ id: "LWB", role: "LWB" }] },
      {
        count: 2,
        positions: [
          { id: "RST", role: "ST" },
          { id: "LST", role: "ST" },
        ],
      },
    ],
  },
  "4231_wide": {
    name: "4-2-3-1 Wide",
    layout: [
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
      {
        count: 4,
        positions: [
          { id: "RB", role: "RB" },
          { id: "RCB", role: "CB" },
          { id: "LCB", role: "CB" },
          { id: "LB", role: "LB" },
        ],
      },
      {
        count: 2,
        positions: [
          { id: "RDM", role: "DM" },
          { id: "LDM", role: "DM" },
        ],
      },
      {
        count: 3,
        positions: [
          { id: "RAM", role: "AMR" },
          { id: "CAM", role: "AMC" },
          { id: "LAM", role: "AML" },
        ],
      },
      { count: 1, positions: [{ id: "ST", role: "ST" }] },
    ],
  },
  "532_wb": {
    name: "5-3-2 WB",
    layout: [
      { count: 1, positions: [{ id: "GK", role: "GK" }] },
      { count: 1, positions: [{ id: "RWB", role: "RWB" }] },
      {
        count: 3,
        positions: [
          { id: "RCB", role: "CB" },
          { id: "CB", role: "CB" },
          { id: "LCB", role: "CB" },
        ],
      },
      { count: 1, positions: [{ id: "LWB", role: "LWB" }] },
      {
        count: 3,
        positions: [
          { id: "RCM", role: "CM" },
          { id: "CM", role: "CM" },
          { id: "LCM", role: "CM" },
        ],
      },
      {
        count: 2,
        positions: [
          { id: "RST", role: "ST" },
          { id: "LST", role: "ST" },
        ],
      },
    ],
  },
  // Add more formations as needed
};

export function getFormationLayout(key) {
  return formations[key] ? formations[key].layout : [];
}
