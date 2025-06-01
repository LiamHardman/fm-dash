// Formation layouts using more Football Manager-centric role descriptors.
// Layouts are defined with attackers first (top of pitch) and GK last (bottom of pitch).

export const formations = {
  '41212_narrow_fm': {
    name: '4-1-2-1-2 Narrow (Diamond FM)',
    layout: [
      {
        count: 2,
        positions: [
          { id: 'STCL_41212N', role: 'ST (C)' },
          { id: 'STCR_41212N', role: 'ST (C)' }
        ]
      },
      { count: 1, positions: [{ id: 'AMC_41212N', role: 'AM (C)' }] },
      {
        count: 2, // Two Central Midfielders
        positions: [
          { id: 'MCL_41212N', role: 'M (C)' },
          { id: 'MCR_41212N', role: 'M (C)' }
        ]
      },
      { count: 1, positions: [{ id: 'DMC_41212N', role: 'DM (C)' }] },
      {
        count: 4,
        positions: [
          { id: 'DL_41212N', role: 'D (L)' },
          { id: 'DCL_41212N', role: 'D (C)' },
          { id: 'DCR_41212N', role: 'D (C)' },
          { id: 'DR_41212N', role: 'D (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_41212N', role: 'GK' }] }
    ]
  },
  '4132_dm_flat_mids': {
    name: '4-1-3-2 (DM, Flat Mids)',
    layout: [
      {
        count: 2,
        positions: [
          { id: 'STCL_4132', role: 'ST (C)' },
          { id: 'STCR_4132', role: 'ST (C)' }
        ]
      },
      {
        count: 3, // Flat line of 3 midfielders
        positions: [
          { id: 'ML_4132', role: 'M (L)' },
          { id: 'MC_4132', role: 'M (C)' },
          { id: 'MR_4132', role: 'M (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'DMC_4132', role: 'DM (C)' }] },
      {
        count: 4,
        positions: [
          { id: 'DL_4132', role: 'D (L)' },
          { id: 'DCL_4132', role: 'D (C)' },
          { id: 'DCR_4132', role: 'D (C)' },
          { id: 'DR_4132', role: 'D (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_4132', role: 'GK' }] }
    ]
  },
  '4141_flat': {
    name: '4-1-4-1 Flat',
    layout: [
      { count: 1, positions: [{ id: 'ST_4141', role: 'ST (C)' }] },
      {
        count: 4,
        positions: [
          { id: 'ML_4141', role: 'M (L)' },
          { id: 'MCL_4141', role: 'M (C)' },
          { id: 'MCR_4141', role: 'M (C)' },
          { id: 'MR_4141', role: 'M (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'DMC_4141', role: 'DM (C)' }] },
      {
        count: 4,
        positions: [
          { id: 'DL_4141', role: 'D (L)' },
          { id: 'DCL_4141', role: 'D (C)' },
          { id: 'DCR_4141', role: 'D (C)' },
          { id: 'DR_4141', role: 'D (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_4141', role: 'GK' }] }
    ]
  },
  '433_dm_wide': {
    // 4-1-2-3 DM Wide in FM terms (DM is the line above defenders, then 2 CMs)
    name: '4-1-2-3 DM Wide (4-3-3 DM)',
    layout: [
      {
        count: 3,
        positions: [
          { id: 'AML_433DM', role: 'AM (L)' },
          { id: 'STC_433DM', role: 'ST (C)' },
          { id: 'AMR_433DM', role: 'AM (R)' }
        ]
      },
      {
        // Two CMs ahead of DM
        count: 2,
        positions: [
          { id: 'MCL_433DM', role: 'M (C)' },
          { id: 'MCR_433DM', role: 'M (C)' }
        ]
      },
      { count: 1, positions: [{ id: 'DM_433DM', role: 'DM (C)' }] }, // Defensive Midfielder (Centre) - This is the line above defenders.
      {
        count: 4,
        positions: [
          { id: 'DL_433DM', role: 'D (L)' },
          { id: 'DCL_433DM', role: 'D (C)' },
          { id: 'DCR_433DM', role: 'D (C)' },
          { id: 'DR_433DM', role: 'D (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_433DM', role: 'GK' }] }
    ]
  },
  '4222_dual_cam_dm': {
    name: '4-2-2-2 (Dual CAMs, Dual DMs)',
    layout: [
      {
        count: 2,
        positions: [
          { id: 'STCL_4222', role: 'ST (C)' },
          { id: 'STCR_4222', role: 'ST (C)' }
        ]
      },
      {
        count: 2, // Dual Attacking Midfielders (Central)
        positions: [
          { id: 'AMCL_4222', role: 'AM (C)' },
          { id: 'AMCR_4222', role: 'AM (C)' }
        ]
      },
      {
        count: 2, // Dual Defensive Midfielders - This is the line above defenders.
        positions: [
          { id: 'DMCL_4222', role: 'DM (C)' },
          { id: 'DMCR_4222', role: 'DM (C)' }
        ]
      },
      {
        count: 4,
        positions: [
          { id: 'DL_4222', role: 'D (L)' },
          { id: 'DCL_4222', role: 'D (C)' },
          { id: 'DCR_4222', role: 'D (C)' },
          { id: 'DR_4222', role: 'D (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_4222', role: 'GK' }] }
    ]
  },
  '4231_dm_am_wide': {
    name: '4-2-3-1 DM AM Wide',
    layout: [
      { count: 1, positions: [{ id: 'STC_4231W', role: 'ST (C)' }] },
      {
        // Attacking Midfield Trio
        count: 3,
        positions: [
          { id: 'AML_4231W', role: 'AM (L)' },
          { id: 'AMC_4231W', role: 'AM (C)' },
          { id: 'AMR_4231W', role: 'AM (R)' }
        ]
      },
      {
        // Two Defensive Midfielders - This is the line above defenders.
        count: 2,
        positions: [
          { id: 'DMCL_4231W', role: 'DM (C)' },
          { id: 'DMCR_4231W', role: 'DM (C)' }
        ]
      },
      {
        count: 4,
        positions: [
          { id: 'DL_4231W', role: 'D (L)' },
          { id: 'DCL_4231W', role: 'D (C)' },
          { id: 'DCR_4231W', role: 'D (C)' },
          { id: 'DR_4231W', role: 'D (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_4231W', role: 'GK' }] }
    ]
  },
  '4231_narrow_dm': {
    name: '4-2-3-1 Narrow (3 AMCs, 2 DMs)',
    layout: [
      { count: 1, positions: [{ id: 'STC_4231N', role: 'ST (C)' }] },
      {
        count: 3, // Three central AMs
        positions: [
          { id: 'AMCL_4231N', role: 'AM (C)' },
          { id: 'AMC_4231N', role: 'AM (C)' },
          { id: 'AMCR_4231N', role: 'AM (C)' }
        ]
      },
      {
        count: 2, // Two Defensive Midfielders - This is the line above defenders.
        positions: [
          { id: 'DMCL_4231N', role: 'DM (C)' },
          { id: 'DMCR_4231N', role: 'DM (C)' }
        ]
      },
      {
        count: 4,
        positions: [
          { id: 'DL_4231N', role: 'D (L)' },
          { id: 'DCL_4231N', role: 'D (C)' },
          { id: 'DCR_4231N', role: 'D (C)' },
          { id: 'DR_4231N', role: 'D (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_4231N', role: 'GK' }] }
    ]
  },
  '424_flat_mc': {
    name: '4-2-4 (Flat MCs, AM L/R Wingers)',
    layout: [
      {
        count: 4, // Four attackers: two wingers, two strikers
        positions: [
          { id: 'AML_424', role: 'AM (L)' },
          { id: 'STCL_424', role: 'ST (C)' },
          { id: 'STCR_424', role: 'ST (C)' },
          { id: 'AMR_424', role: 'AM (R)' }
        ]
      },
      {
        count: 2, // Two Central Midfielders - This is the line above defenders.
        positions: [
          { id: 'MCL_424', role: 'M (C)' },
          { id: 'MCR_424', role: 'M (C)' }
        ]
      },
      {
        count: 4,
        positions: [
          { id: 'DL_424', role: 'D (L)' },
          { id: 'DCL_424', role: 'D (C)' },
          { id: 'DCR_424', role: 'D (C)' },
          { id: 'DR_424', role: 'D (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_424', role: 'GK' }] }
    ]
  },
  '4321_christmas_tree': {
    name: '4-3-2-1 (Christmas Tree)',
    layout: [
      { count: 1, positions: [{ id: 'ST_4321', role: 'ST (C)' }] },
      {
        count: 2, // Two AMs behind striker (LF/RF roles)
        positions: [
          { id: 'AML_4321', role: 'AM (L)' },
          { id: 'AMR_4321', role: 'AM (R)' }
        ]
      },
      {
        count: 3, // Three Central Midfielders - This is the line above defenders.
        positions: [
          { id: 'MCL_4321', role: 'M (C)' },
          { id: 'MC_4321', role: 'M (C)' },
          { id: 'MCR_4321', role: 'M (C)' }
        ]
      },
      {
        count: 4,
        positions: [
          { id: 'DL_4321', role: 'D (L)' },
          { id: 'DCL_4321', role: 'D (C)' },
          { id: 'DCR_4321', role: 'D (C)' },
          { id: 'DR_4321', role: 'D (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_4321', role: 'GK' }] }
    ]
  },
  '433_false_nine': {
    name: '4-3-3 False Nine',
    layout: [
      {
        count: 3, // Attacking line, central player is AM(C) acting as F9
        positions: [
          { id: 'AML_433F9', role: 'AM (L)' },
          { id: 'AMC_433F9', role: 'AM (C)' },
          { id: 'AMR_433F9', role: 'AM (R)' }
        ]
      },
      {
        count: 3, // Three Central Midfielders - This is the line above defenders.
        positions: [
          { id: 'MCL_433F9', role: 'M (C)' },
          { id: 'MC_433F9', role: 'M (C)' },
          { id: 'MCR_433F9', role: 'M (C)' }
        ]
      },
      {
        count: 4,
        positions: [
          { id: 'DL_433F9', role: 'D (L)' },
          { id: 'DCL_433F9', role: 'D (C)' },
          { id: 'DCR_433F9', role: 'D (C)' },
          { id: 'DR_433F9', role: 'D (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_433F9', role: 'GK' }] }
    ]
  },
  '433_wide': {
    name: '4-3-3 Wide',
    layout: [
      {
        count: 3,
        positions: [
          { id: 'AML_433W', role: 'AM (L)' },
          { id: 'STC_433W', role: 'ST (C)' },
          { id: 'AMR_433W', role: 'AM (R)' }
        ]
      },
      {
        count: 3, // Three Central Midfielders - This is the line above defenders.
        positions: [
          { id: 'MCL_433W', role: 'M (C)' },
          { id: 'MC_433W', role: 'M (C)' },
          { id: 'MCR_433W', role: 'M (C)' }
        ]
      },
      {
        count: 4,
        positions: [
          { id: 'DL_433W', role: 'D (L)' },
          { id: 'DCL_433W', role: 'D (C)' },
          { id: 'DCR_433W', role: 'D (C)' },
          { id: 'DR_433W', role: 'D (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_433W', role: 'GK' }] }
    ]
  },
  '4411_cf_behind_st': {
    name: '4-4-1-1 (CF behind ST)',
    layout: [
      { count: 1, positions: [{ id: 'ST_4411', role: 'ST (C)' }] },
      { count: 1, positions: [{ id: 'AMC_4411', role: 'AM (C)' }] },
      {
        count: 4, // Midfield line - This is the line above defenders.
        positions: [
          { id: 'ML_4411', role: 'M (L)' },
          { id: 'MCL_4411', role: 'M (C)' },
          { id: 'MCR_4411', role: 'M (C)' },
          { id: 'MR_4411', role: 'M (R)' }
        ]
      },
      {
        count: 4,
        positions: [
          { id: 'DL_4411', role: 'D (L)' },
          { id: 'DCL_4411', role: 'D (C)' },
          { id: 'DCR_4411', role: 'D (C)' },
          { id: 'DR_4411', role: 'D (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_4411', role: 'GK' }] }
    ]
  },
  '442_classic': {
    name: '4-4-2 Classic',
    layout: [
      {
        count: 2,
        positions: [
          { id: 'LST_442C', role: 'ST (C)' },
          { id: 'RST_442C', role: 'ST (C)' }
        ]
      },
      {
        count: 4, // Midfield line - This is the line above defenders.
        positions: [
          { id: 'ML_442C', role: 'M (L)' },
          { id: 'MCL_442C', role: 'M (C)' },
          { id: 'MCR_442C', role: 'M (C)' },
          { id: 'MR_442C', role: 'M (R)' }
        ]
      },
      {
        count: 4,
        positions: [
          { id: 'DL_442C', role: 'D (L)' },
          { id: 'DCL_442C', role: 'D (C)' },
          { id: 'DCR_442C', role: 'D (C)' },
          { id: 'DR_442C', role: 'D (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_442C', role: 'GK' }] }
    ]
  },
  '451_flat': {
    name: '4-5-1 Flat',
    layout: [
      { count: 1, positions: [{ id: 'ST_451F', role: 'ST (C)' }] },
      {
        count: 5, // Midfield line - This is the line above defenders.
        positions: [
          { id: 'ML_451F', role: 'M (L)' },
          { id: 'MCL_451F', role: 'M (C)' },
          { id: 'MC_451F', role: 'M (C)' },
          { id: 'MCR_451F', role: 'M (C)' },
          { id: 'MR_451F', role: 'M (R)' }
        ]
      },
      {
        count: 4,
        positions: [
          { id: 'DL_451F', role: 'D (L)' },
          { id: 'DCL_451F', role: 'D (C)' },
          { id: 'DCR_451F', role: 'D (C)' },
          { id: 'DR_451F', role: 'D (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_451F', role: 'GK' }] }
    ]
  },
  '5212_wb': {
    name: '5-2-1-2 WB',
    layout: [
      {
        count: 2,
        positions: [
          { id: 'STCL_5212', role: 'ST (C)' },
          { id: 'STCR_5212', role: 'ST (C)' }
        ]
      },
      { count: 1, positions: [{ id: 'AMC_5212', role: 'AM (C)' }] },
      {
        count: 2, // Two Central Midfielders - This is the line above defenders.
        positions: [
          { id: 'MCL_5212', role: 'M (C)' },
          { id: 'MCR_5212', role: 'M (C)' }
        ]
      },
      {
        count: 5, // 3 CBs and 2 WBs
        positions: [
          { id: 'WBL_5212', role: 'WB (L)' },
          { id: 'DCL_5212', role: 'D (C)' },
          { id: 'DC_5212', role: 'D (C)' },
          { id: 'DCR_5212', role: 'D (C)' },
          { id: 'WBR_5212', role: 'WB (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_5212', role: 'GK' }] }
    ]
  },
  '541_flat_wb': {
    name: '5-4-1 Flat WB',
    layout: [
      { count: 1, positions: [{ id: 'ST_541F', role: 'ST (C)' }] },
      {
        count: 4, // Midfield line - This is the line above defenders.
        positions: [
          { id: 'ML_541F', role: 'M (L)' },
          { id: 'MCL_541F', role: 'M (C)' },
          { id: 'MCR_541F', role: 'M (C)' },
          { id: 'MR_541F', role: 'M (R)' }
        ]
      },
      {
        count: 5, // 3 CBs and 2 WBs
        positions: [
          { id: 'WBL_541F', role: 'WB (L)' },
          { id: 'DCL_541F', role: 'D (C)' },
          { id: 'DC_541F', role: 'D (C)' },
          { id: 'DCR_541F', role: 'D (C)' },
          { id: 'WBR_541F', role: 'WB (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_541F', role: 'GK' }] }
    ]
  },
  '3142_dm_wb': {
    name: '3-1-4-2 DM WB',
    layout: [
      {
        count: 2,
        positions: [
          { id: 'STCL_3142', role: 'ST (C)' },
          { id: 'STCR_3142', role: 'ST (C)' }
        ]
      },
      {
        count: 4, // Wide midfielders/wing-backs and two CMs
        positions: [
          { id: 'WBL_3142', role: 'WB (L)' },
          { id: 'MCL_3142', role: 'M (C)' },
          { id: 'MCR_3142', role: 'M (C)' },
          { id: 'WBR_3142', role: 'WB (R)' }
        ]
      },
      { count: 1, positions: [{ id: 'DMC_3142', role: 'DM (C)' }] }, // This is the line above defenders.
      {
        count: 3, // Three Centre-Backs
        positions: [
          { id: 'DCL_3142', role: 'D (C)' },
          { id: 'DC_3142', role: 'D (C)' },
          { id: 'DCR_3142', role: 'D (C)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_3142', role: 'GK' }] }
    ]
  },
  '3412_wb': {
    name: '3-4-1-2 WB',
    layout: [
      {
        count: 2,
        positions: [
          { id: 'STCL_3412', role: 'ST (C)' },
          { id: 'STCR_3412', role: 'ST (C)' }
        ]
      },
      { count: 1, positions: [{ id: 'AMC_3412', role: 'AM (C)' }] },
      {
        count: 4, // Midfield line - This is the line above defenders.
        positions: [
          { id: 'WBL_3412', role: 'WB (L)' },
          { id: 'MCL_3412', role: 'M (C)' },
          { id: 'MCR_3412', role: 'M (C)' },
          { id: 'WBR_3412', role: 'WB (R)' }
        ]
      },
      {
        count: 3,
        positions: [
          { id: 'DCL_3412', role: 'D (C)' },
          { id: 'DC_3412', role: 'D (C)' },
          { id: 'DCR_3412', role: 'D (C)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_3412', role: 'GK' }] }
    ]
  },
  '3421_wb': {
    name: '3-4-2-1 WB (Dual AMs)',
    layout: [
      { count: 1, positions: [{ id: 'ST_3421', role: 'ST (C)' }] },
      {
        count: 2,
        positions: [
          { id: 'AML_3421', role: 'AM (L)' },
          { id: 'AMR_3421', role: 'AM (R)' }
        ]
      },
      {
        count: 4, // Midfield line - This is the line above defenders.
        positions: [
          { id: 'WBL_3421', role: 'WB (L)' },
          { id: 'MCL_3421', role: 'M (C)' },
          { id: 'MCR_3421', role: 'M (C)' },
          { id: 'WBR_3421', role: 'WB (R)' }
        ]
      },
      {
        count: 3,
        positions: [
          { id: 'DCL_3421', role: 'D (C)' },
          { id: 'DC_3421', role: 'D (C)' },
          { id: 'DCR_3421', role: 'D (C)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_3421', role: 'GK' }] }
    ]
  },
  '343_fm': {
    name: '3-4-3 FM',
    layout: [
      {
        count: 3,
        positions: [
          { id: 'AML_343', role: 'AM (L)' },
          { id: 'STC_343', role: 'ST (C)' },
          { id: 'AMR_343', role: 'AM (R)' }
        ]
      },
      {
        count: 4, // Midfield line - This is the line above defenders.
        positions: [
          { id: 'WBL_343', role: 'WB (L)' },
          { id: 'MCL_343', role: 'M (C)' },
          { id: 'MCR_343', role: 'M (C)' },
          { id: 'WBR_343', role: 'WB (R)' }
        ]
      },
      {
        count: 3,
        positions: [
          { id: 'DCL_343', role: 'D (C)' },
          { id: 'DC_343', role: 'D (C)' },
          { id: 'DCR_343', role: 'D (C)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_343', role: 'GK' }] }
    ]
  },
  '352_fm': {
    name: '3-5-2 / 5-3-2 WB',
    layout: [
      {
        count: 2,
        positions: [
          { id: 'STCL_352', role: 'ST (C)' },
          { id: 'STCR_352', role: 'ST (C)' }
        ]
      },
      {
        count: 5, // Midfield line - This is the line above defenders.
        positions: [
          { id: 'WBL_352', role: 'WB (L)' },
          { id: 'MCL_352', role: 'M (C)' },
          { id: 'MC_352', role: 'M (C)' },
          { id: 'MCR_352', role: 'M (C)' },
          { id: 'WBR_352', role: 'WB (R)' }
        ]
      },
      {
        count: 3,
        positions: [
          { id: 'DCL_352', role: 'D (C)' },
          { id: 'DC_352', role: 'D (C)' },
          { id: 'DCR_352', role: 'D (C)' }
        ]
      },
      { count: 1, positions: [{ id: 'GK_352', role: 'GK' }] }
    ]
  }
}

/**
 * Retrieves the layout for a given formation key.
 * @param {string} key - The formation key (e.g., "442_classic").
 * @returns {Array} The layout array for the formation, or an empty array if not found.
 */
export function getFormationLayout(key) {
  return formations[key] ? formations[key].layout : []
}
