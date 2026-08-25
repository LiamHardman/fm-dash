<template>
    <div class="star-rating" :title="`${stars} / 5`">
        <q-icon
            v-for="i in 5"
            :key="i"
            :name="starIcon(i)"
            size="16px"
            :class="starIcon(i) === 'star_border' ? 'star-empty' : 'star-filled'"
        />
        <span v-if="showLabel && label" class="star-label">{{ label }}</span>
    </div>
</template>

<script>
import { defineComponent } from 'vue'

export default defineComponent({
  name: 'StarRating',
  props: {
    stars: { type: Number, required: true },
    label: { type: String, default: '' },
    showLabel: { type: Boolean, default: true },
  },
  setup(props) {
    function starIcon(position) {
      const filledThreshold = position - 0.5
      if (props.stars >= position) return 'star'
      if (props.stars >= filledThreshold) return 'star_half'
      return 'star_border'
    }
    return { starIcon }
  },
})
</script>

<style scoped>
.star-rating {
    display: flex;
    align-items: center;
    gap: 1px;
}
.star-filled {
    color: #f5a623;
}
.star-empty {
    color: var(--text-muted);
}
.star-label {
    margin-left: 6px;
    font-size: 11px;
    color: var(--text-secondary);
    white-space: nowrap;
}
</style>
