<template>
    <div class="fifa-card" @click="handleCardClick">
        <div class="card-content">
            <!-- Header: Rating, Position, Name -->
            <div class="card-top">
                <div class="player-details">
                    <div class="player-rating">{{ player.overall }}</div>
                    <div class="player-position">{{ player.position }}</div>
                    <div class="player-name">{{ player.name }}</div>
                </div>
            </div>

            <!-- Middle: Photo, Nation, Club -->
            <div class="card-middle">
                <div class="nation-placeholder">
                    <!-- Simple box for nation flag -->
                </div>
                <div class="club-placeholder">
                    <!-- Simple circle for club logo -->
                </div>
                <div class="player-photo">
                    <q-avatar
                        size="100px"
                        :color="$q.dark.isActive ? 'grey-8' : 'grey-5'"
                        :text-color="$q.dark.isActive ? 'grey-5' : 'grey-8'"
                    >
                        <q-icon name="person" size="60px" />
                    </q-avatar>
                </div>
            </div>

            <!-- Footer: Stats -->
            <div class="card-bottom">
                <div class="stats-grid">
                    <div class="stat-item"><span>{{ player.pac }}</span> PAC</div>
                    <div class="stat-item"><span>{{ player.dri }}</span> DRI</div>
                    <div class="stat-item"><span>{{ player.sho }}</span> SHO</div>
                    <div class="stat-item"><span>{{ player.def }}</span> DEF</div>
                    <div class="stat-item"><span>{{ player.pas }}</span> PAS</div>
                    <div class="stat-item"><span>{{ player.phy }}</span> PHY</div>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
import { useQuasar } from 'quasar'
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'PlayerCards',
  props: {
    player: {
      type: Object,
      required: true
    },
    currencySymbol: {
      type: String,
      default: '$'
    }
  },
  emits: ['click'],
  setup(props, { emit }) {
    const qInstance = useQuasar()

    const handleCardClick = () => {
      emit('click')
    }

    return {
      qInstance,
      handleCardClick
    }
  }
})
</script>

<style lang="scss" scoped>
// Define variables for easy theme changes
$card-bg: #2a2a2a;
$gold-accent: #c89b3c;
$gold-gradient: linear-gradient(45deg, #d4af37, #c89b3c);
$text-light: #e0e0e0;
$text-dark: #333;
$border-color: #444;

// New FIFA Card Design
.fifa-card {
    width: 280px;
    height: 420px;
    background: $card-bg;
    border-radius: 12px;
    position: relative;
    overflow: hidden;
    color: $text-light;
    font-family: 'Roboto', 'Helvetica Neue', 'Arial', sans-serif;
    border: 2px solid $gold-accent;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
    transition: transform 0.3s ease, box-shadow 0.3s ease;
    cursor: pointer;

    &:hover {
        transform: translateY(-10px) scale(1.03);
        box-shadow: 0 20px 40px rgba($gold-accent, 0.3);
    }

    // This pseudo-element creates the large diagonal background shape
    &::before {
        content: '';
        position: absolute;
        top: -50px;
        left: 0;
        width: 100%;
        height: 220px;
        background: #1c1c1c;
        transform: skewY(-8deg);
        z-index: 1;
    }

    // This pseudo-element creates the gold diagonal line
    &::after {
        content: '';
        position: absolute;
        top: 140px;
        left: -10%;
        width: 120%;
        height: 4px;
        background: $gold-gradient;
        transform: skewY(-8deg);
        z-index: 3;
        box-shadow: 0 0 10px rgba($gold-accent, 0.7);
    }
}

.card-content {
    position: relative;
    z-index: 2;
    height: 100%;
    display: flex;
    flex-direction: column;
    padding: 1rem;
}

// Top section of the card
.card-top {
    position: relative;
    z-index: 4;
    padding: 0.5rem 1rem;
    .player-details {
        display: flex;
        align-items: baseline;
        gap: 0.75rem;
    }

    .player-rating {
        font-size: 2.5rem;
        font-weight: 900;
        color: $gold-accent;
        line-height: 1;
    }

    .player-position {
        font-size: 1.25rem;
        font-weight: 700;
        line-height: 1;
    }

    .player-name {
        font-size: 1.25rem;
        font-weight: 500;
        margin-left: auto; // Pushes the name to the right
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
}

// Middle section with placeholders and photo
.card-middle {
    position: relative;
    flex-grow: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2; // Below the gold line but above the bg

    .player-photo {
        .q-avatar {
            border: 3px solid rgba(255,255,255,0.1);
        }
    }

    .nation-placeholder, .club-placeholder {
        position: absolute;
        background-color: rgba(255, 255, 255, 0.05);
        border: 1px solid rgba(255, 255, 255, 0.1);
    }

    .nation-placeholder {
        top: 25px;
        left: 15px;
        width: 50px;
        height: 35px;
        border-radius: 4px;
    }

    .club-placeholder {
        top: 70px;
        left: 15px;
        width: 50px;
        height: 50px;
        border-radius: 50%;
    }
}

// Bottom section with stats
.card-bottom {
    padding: 1rem;
    position: relative;
    z-index: 2;

    .stats-grid {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        grid-template-rows: repeat(3, 1fr);
        gap: 0.5rem 1.5rem;
    }

    .stat-item {
        display: flex;
        align-items: baseline;
        font-size: 1.1rem;
        font-weight: 500;

        span {
            color: $gold-accent;
            font-weight: 700;
            font-size: 1.2rem;
            margin-right: 0.5rem;
            width: 30px; // Aligns the stat labels
        }
    }
}

// Responsive adjustments
@media (max-width: 600px) {
    .fifa-card {
        width: 260px;
        height: 400px;
    }
}
</style>
