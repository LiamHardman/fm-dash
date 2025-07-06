<template>
    <q-card flat bordered class="stat-card full-height">
      <q-card-section class="stat-header">
        <div class="stat-name">{{ stat.name }}</div>
      </q-card-section>
      <q-card-section class="stat-players">
        <div v-if="players && players.length > 0">
          <q-list separator dense>
            <q-item
              v-for="(player, index) in players"
              :key="player.id || index"
              clickable
              @click="$emit('player-click', player)"
              class="player-item"
            >
              <q-item-section avatar>
                <div class="rank-badge">{{ index + 1 }}</div>
              </q-item-section>
              <q-item-section>
                <q-item-label class="player-name">{{ getPlayerName(player) }}</q-item-label>
                <q-item-label caption>{{ getPlayerClub(player) }}</q-item-label>
              </q-item-section>
              <q-item-section side>
                <div class="stat-value">{{ formatStatValue(player.attributes[stat.key], stat.key) }}</div>
              </q-item-section>
            </q-item>
          </q-list>
        </div>
        <div v-else class="no-data-message">
          No players match filters
        </div>
      </q-card-section>
    </q-card>
  </template>
  
  <script>
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'StatCard',
  props: {
    stat: {
      type: Object,
      required: true
    },
    players: {
      type: Array,
      default: () => []
    }
  },
  emits: ['player-click'],
  methods: {
    getPlayerName(player) {
      return player.name || player.Name || player.Player || 'Unknown Player'
    },
    getPlayerClub(player) {
      return player.club || player.Club || 'Unknown Club'
    },
    formatStatValue(value, statKey) {
      if (value === undefined || value === null || value === '-' || value === '') {
        return 'N/A'
      }
      const stringValue = String(value)
      const cleanValue = stringValue.replace(/,/g, '').replace(/%/g, '')
      const numValue = parseFloat(cleanValue)

      if (Number.isNaN(numValue)) {
        return stringValue
      }

      if (stringValue.includes('%') || ['Sv %', 'Conv %', 'Pas %', 'Tck R'].includes(statKey)) {
        return `${numValue.toFixed(1)}%`
      }

      if (numValue >= 1000 && Number.isInteger(numValue)) {
        return numValue.toLocaleString()
      }

      if (numValue % 1 === 0) {
        return numValue.toString()
      }

      return numValue.toFixed(2)
    }
  }
})
</script>
  
  <style lang="scss" scoped>
  .stat-card {
    border-radius: 12px;
    transition: all 0.3s ease;
    border: 1px solid #e0e0e0;
    background: #fff;
  
    .body--dark & {
      background: #1e1e1e;
      border: 1px solid #424242;
    }
  
    &:hover {
      transform: translateY(-4px);
      box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
    }
  }
  
  .stat-header {
    background: #f8f9fa;
    border-bottom: 1px solid #e0e0e0;
    padding: 12px 16px;
  
    .body--dark & {
      background: #2d2d2d;
      border-bottom: 1px solid #424242;
    }
  
    .stat-name {
      font-weight: 600;
      font-size: 0.95rem;
      color: #333;
  
      .body--dark & {
        color: #f5f5f5;
      }
    }
  }
  
  .stat-players {
    padding: 0;
  
    .body--dark .q-list--separator > .q-item-type + .q-item-type {
        border-top: 1px solid #424242;
    }
  }
  
  .player-item {
    transition: background-color 0.2s ease;
    padding: 12px 16px;
  
    &:hover {
      background: #f1f3f5;
      .body--dark & {
        background: #333;
      }
    }
  }
  
  .rank-badge {
    background: linear-gradient(135deg, #764ba2, #667eea);
    color: white;
    border-radius: 50%;
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.8rem;
    font-weight: 700;
    box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  }
  
  .player-name {
    font-weight: 600;
    color: #212529;
  
    .body--dark & {
      color: #f5f5f5;
    }
  }
  
  .q-item__label--caption {
      color: #6c757d;
  
      .body--dark & {
        color: #9e9e9e;
      }
  }
  
  .stat-value {
    font-weight: 700;
    color: #667eea;
    font-size: 1rem;
  }
  
  .no-data-message {
    padding: 24px 16px;
    text-align: center;
    color: #6c757d;
    font-style: italic;
    min-height: 100px;
    display: flex;
    align-items: center;
    justify-content: center;
  
    .body--dark & {
      color: #9e9e9e;
    }
  }
  </style>
  