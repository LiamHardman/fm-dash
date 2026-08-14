import { computed } from 'vue'

export function usePlayerTableColumns(isGoalkeeperView, showValueScore, showCA) {
  // Column styles
  const nameColumnStyle = 'width: 220px; min-width: 220px; max-width: 220px; overflow: hidden;'
  const ageColumnStyle =
    'width: 60px; min-width: 60px; max-width: 60px; text-align: center; white-space: nowrap;'
  const positionColumnStyle = 'width: 150px; min-width: 150px; max-width: 150px; overflow: hidden;'
  const clubColumnStyle =
    'width: 180px; min-width: 180px; max-width: 180px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;'
  const moneyColumnStyle =
    'width: 110px; min-width: 110px; max-width: 110px; text-align: right; white-space: nowrap;'
  const overallColumnStyle =
    'width: 70px; min-width: 70px; max-width: 70px; text-align: center; white-space: nowrap;'
  const fifaStatColumnStyle =
    'width: 60px; min-width: 60px; max-width: 60px; text-align: center; white-space: nowrap;'
  const widerFifaStatColumnStyle =
    'width: 80px; min-width: 80px; max-width: 80px; text-align: center; white-space: nowrap;'
  const textColumnStyle =
    'width: 120px; min-width: 120px; max-width: 120px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;'
  const nationalityColumnStyle =
    'width: 150px; min-width: 150px; max-width: 150px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;'

  const baseColumnDefinitions = {
    name: {
      name: 'name',
      label: 'Name',
      field: 'name',
      sortable: true,
      align: 'left',
      style: nameColumnStyle,
      headerStyle: nameColumnStyle,
    },
    age: {
      name: 'age',
      label: 'Age',
      field: 'age',
      sortable: true,
      align: 'center',
      style: ageColumnStyle,
      headerStyle: ageColumnStyle,
    },
    position: {
      name: 'position',
      label: 'Position',
      field: 'position',
      sortable: true,
      align: 'left',
      style: positionColumnStyle,
      headerStyle: positionColumnStyle,
    },
    club: {
      name: 'club',
      label: 'Club',
      field: 'club',
      sortable: true,
      align: 'left',
      style: clubColumnStyle,
      headerStyle: clubColumnStyle,
    },
    transfer_value: {
      name: 'transfer_value',
      label: 'Value',
      field: 'transfer_value',
      sortable: true,
      align: 'right',
      sortField: 'transferValueAmount',
      style: moneyColumnStyle,
      headerStyle: moneyColumnStyle,
    },
    wage: {
      name: 'wage',
      label: 'Salary',
      field: 'wage',
      sortable: true,
      align: 'right',
      sortField: 'wageAmount',
      style: moneyColumnStyle,
      headerStyle: moneyColumnStyle,
    },
    Overall: {
      name: 'Overall',
      label: 'Overall',
      field: 'Overall',
      sortable: true,
      align: 'center',
      isOverallStat: true,
      style: overallColumnStyle,
      headerStyle: overallColumnStyle,
    },
    CA: {
      name: 'CA',
      label: 'CA',
      field: 'ca',
      sortable: true,
      align: 'center',
      isOverallStat: true,
      gaugeMax: 200,
      style: overallColumnStyle,
      headerStyle: overallColumnStyle,
    },
    valueScore: {
      name: 'valueScore',
      label: 'Value Score',
      field: 'valueScore',
      sortable: true,
      align: 'center',
      isValueScore: true,
      style: overallColumnStyle,
      headerStyle: overallColumnStyle,
    },
    personality: {
      name: 'personality',
      label: 'Personality',
      field: 'personality',
      sortable: true,
      align: 'left',
      style: textColumnStyle,
      headerStyle: textColumnStyle,
    },
    media_handling: {
      name: 'media_handling',
      label: 'Media Desc.',
      field: 'media_handling',
      sortable: true,
      align: 'left',
      style: textColumnStyle,
      headerStyle: textColumnStyle,
    },
    nationality_display: {
      name: 'nationality_display',
      label: 'Nationality',
      field: 'nationality',
      sortable: true,
      align: 'left',
      style: nationalityColumnStyle,
      headerStyle: nationalityColumnStyle,
    },
  }

  const allFifaStatDefinitions = {
    GK: {
      name: 'GK',
      label: 'GK',
      field: 'gk',
      sortable: true,
      align: 'center',
      isFifaStat: true,
      style: fifaStatColumnStyle,
      headerStyle: fifaStatColumnStyle,
    },
    DIV: {
      name: 'DIV',
      label: 'DIV',
      field: 'div',
      sortable: true,
      align: 'center',
      isFifaStat: true,
      style: fifaStatColumnStyle,
      headerStyle: fifaStatColumnStyle,
    },
    HAN: {
      name: 'HAN',
      label: 'HAN',
      field: 'han',
      sortable: true,
      align: 'center',
      isFifaStat: true,
      style: fifaStatColumnStyle,
      headerStyle: fifaStatColumnStyle,
    },
    REF: {
      name: 'REF',
      label: 'REF',
      field: 'ref',
      sortable: true,
      align: 'center',
      isFifaStat: true,
      style: fifaStatColumnStyle,
      headerStyle: fifaStatColumnStyle,
    },
    KIC: {
      name: 'KIC',
      label: 'KIC',
      field: 'kic',
      sortable: true,
      align: 'center',
      isFifaStat: true,
      style: fifaStatColumnStyle,
      headerStyle: fifaStatColumnStyle,
    },
    SPD: {
      name: 'SPD',
      label: 'SPD',
      field: 'spd',
      sortable: true,
      align: 'center',
      isFifaStat: true,
      style: fifaStatColumnStyle,
      headerStyle: fifaStatColumnStyle,
    },
    POS: {
      name: 'POS',
      label: 'POS',
      field: 'pos',
      sortable: true,
      align: 'center',
      isFifaStat: true,
      style: fifaStatColumnStyle,
      headerStyle: fifaStatColumnStyle,
    },
    PAC: {
      name: 'PAC',
      label: 'PAC',
      field: 'pac',
      sortable: true,
      align: 'center',
      isFifaStat: true,
      style: fifaStatColumnStyle,
      headerStyle: fifaStatColumnStyle,
    },
    SHO: {
      name: 'SHO',
      label: 'SHO',
      field: 'sho',
      sortable: true,
      align: 'center',
      isFifaStat: true,
      style: fifaStatColumnStyle,
      headerStyle: fifaStatColumnStyle,
    },
    PAS: {
      name: 'PAS',
      label: 'PAS',
      field: 'pas',
      sortable: true,
      align: 'center',
      isFifaStat: true,
      style: fifaStatColumnStyle,
      headerStyle: fifaStatColumnStyle,
    },
    DRI: {
      name: 'DRI',
      label: 'DRI',
      field: 'dri',
      sortable: true,
      align: 'center',
      isFifaStat: true,
      style: fifaStatColumnStyle,
      headerStyle: fifaStatColumnStyle,
    },
    DEF: {
      name: 'DEF',
      label: 'DEF',
      field: 'def',
      sortable: true,
      align: 'center',
      isFifaStat: true,
      style: fifaStatColumnStyle,
      headerStyle: fifaStatColumnStyle,
    },
    PHY: {
      name: 'PHY',
      label: 'PHY',
      field: 'phy',
      sortable: true,
      align: 'center',
      isFifaStat: true,
      style: fifaStatColumnStyle,
      headerStyle: fifaStatColumnStyle,
    },
    TotalStats: {
      name: 'TotalStats',
      label: 'Total',
      field: 'totalStats',
      sortable: true,
      align: 'center',
      isFifaStat: true,
      style: widerFifaStatColumnStyle,
      headerStyle: widerFifaStatColumnStyle,
    },
    MBR: {
      name: 'MBR',
      label: 'MBR',
      field: 'mbr',
      sortable: true,
      align: 'center',
      isFifaStat: true,
      gaugeMax: 100,
      style: widerFifaStatColumnStyle,
      headerStyle: widerFifaStatColumnStyle,
    },
  }

  const currentColumns = computed(() => {
    const newOrderBase = [
      baseColumnDefinitions.name,
      baseColumnDefinitions.age,
      baseColumnDefinitions.position,
      baseColumnDefinitions.club,
      baseColumnDefinitions.transfer_value,
      baseColumnDefinitions.wage,
      baseColumnDefinitions.Overall,
    ]

    // Add value score column if enabled
    if (showValueScore) {
      newOrderBase.push(baseColumnDefinitions.valueScore)
    }

    // Add CA column if enabled in settings
    const caEnabled = typeof showCA === 'object' && showCA !== null ? showCA.value : showCA
    if (caEnabled) {
      newOrderBase.push(baseColumnDefinitions.CA)
    }

    const fifaColumnsInOrder = isGoalkeeperView
      ? [
          allFifaStatDefinitions.DIV,
          allFifaStatDefinitions.HAN,
          allFifaStatDefinitions.REF,
          allFifaStatDefinitions.KIC,
          allFifaStatDefinitions.SPD,
          allFifaStatDefinitions.POS,
        ]
      : [
          allFifaStatDefinitions.PAC,
          allFifaStatDefinitions.SHO,
          allFifaStatDefinitions.PAS,
          allFifaStatDefinitions.DRI,
          allFifaStatDefinitions.DEF,
          allFifaStatDefinitions.PHY,
          allFifaStatDefinitions.TotalStats,
          allFifaStatDefinitions.MBR,
        ]

    return [...newOrderBase, ...fifaColumnsInOrder]
  })

  const getColumnLabel = (fieldName) => {
    const col = currentColumns.value.find((c) => c.name === fieldName)
    return col ? col.label : fieldName
  }

  const getSortFieldKey = (colName) => {
    const colDef = currentColumns.value.find((c) => c.name === colName)
    return colDef?.sortField || colDef?.field || colName
  }

  return {
    currentColumns,
    getColumnLabel,
    getSortFieldKey,
    baseColumnDefinitions,
    allFifaStatDefinitions,
  }
}
