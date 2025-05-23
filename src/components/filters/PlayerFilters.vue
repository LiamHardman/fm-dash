# src/components/filters/PlayerFilters.vue
<template>
  <q-card
    class="q-mb-md filter-card"
    :class="$q.dark.isActive ? 'bg-grey-9' : 'bg-white'"
  >
    <q-card-section>
      <div class="text-subtitle1 q-mb-sm">
        Search Players (Using {{ currencySymbol }} for values)
      </div>
      <div class="row q-col-gutter-md items-end">
        <div class="col-12 col-sm-6 col-md-3 col-lg-2">
          <q-input
            v-model="filters.name"
            label="Player Name"
            dense
            outlined
            clearable
            @update:model-value="debouncedApplyFilters"
            :label-color="$q.dark.isActive ? 'grey-4' : ''"
            :input-class="$q.dark.isActive ? 'text-grey-3' : ''"
          />
        </div>
        <div class="col-12 col-sm-6 col-md-3 col-lg-2">
          <q-select
            v-model="filters.club"
            :options="clubOptions"
            label="Club"
            dense
            outlined
            clearable
            use-input
            hide-selected
            fill-input
            input-debounce="300"
            @filter="filterClubOptions"
            @update:model-value="applyFilters"
            :label-color="$q.dark.isActive ? 'grey-4' : ''"
            :popup-content-class="
              $q.dark.isActive ? 'bg-grey-8 text-white' : ''
            "
          >
            <template v-slot:no-option>
              <q-item>
                <q-item-section class="text-grey">
                  No results
                </q-item-section>
              </q-item>
            </template>
          </q-select>
        </div>
        <div class="col-12 col-sm-6 col-md-3 col-lg-2">
          <q-select
            v-model="filters.nationality"
            :options="nationalityOptions"
            label="Nationality"
            dense
            outlined
            clearable
            use-input
            hide-selected
            fill-input
            input-debounce="300"
            @filter="filterNationalityOptions"
            @update:model-value="applyFilters"
            :label-color="$q.dark.isActive ? 'grey-4' : ''"
            :popup-content-class="
              $q.dark.isActive ? 'bg-grey-8 text-white' : ''
            "
          >
            <template v-slot:no-option>
              <q-item>
                <q-item-section class="text-grey">
                  No results
                </q-item-section>
              </q-item>
            </template>
          </q-select>
        </div>
        <div class="col-12 col-sm-6 col-md-3 col-lg-2">
          <q-select
            v-model="filters.position"
            :options="positionOptions"
            label="Position"
            dense
            outlined
            clearable
            emit-value
            map-options
            @update:model-value="applyFilters"
            :label-color="$q.dark.isActive ? 'grey-4' : ''"
            :popup-content-class="
              $q.dark.isActive ? 'bg-grey-8 text-white' : ''
            "
          />
        </div>
        <div class="col-12 col-sm-6 col-md-3 col-lg-2">
          <q-select
            v-model="filters.mediaHandling"
            :options="mediaHandlingOptions"
            label="Media Handling"
            dense
            outlined
            multiple
            use-chips
            clearable
            emit-value
            map-options
            @update:model-value="applyFilters"
            :label-color="$q.dark.isActive ? 'grey-4' : ''"
            :popup-content-class="
              $q.dark.isActive ? 'bg-grey-8 text-white' : ''
            "
          />
        </div>
        <div class="col-12 col-sm-6 col-md-3 col-lg-2">
          <q-select
            v-model="filters.personality"
            :options="personalityOptions"
            label="Personality"
            dense
            outlined
            multiple
            use-chips
            clearable
            emit-value
            map-options
            @update:model-value="applyFilters"
            :label-color="$q.dark.isActive ? 'grey-4' : ''"
            :popup-content-class="
              $q.dark.isActive ? 'bg-grey-8 text-white' : ''
            "
          />
        </div>
        <div class="col-12 col-sm-4 col-md-2 col-lg-1">
          <q-input
            v-model.number="filters.minAge"
            type="number"
            label="Min Age"
            dense
            outlined
            clearable
            :min="0"
            @update:model-value="debouncedApplyFilters"
            :label-color="$q.dark.isActive ? 'grey-4' : ''"
            :input-class="$q.dark.isActive ? 'text-grey-3' : ''"
          />
        </div>
        <div class="col-12 col-sm-4 col-md-2 col-lg-1">
          <q-input
            v-model.number="filters.maxAge"
            type="number"
            label="Max Age"
            dense
            outlined
            clearable
            :min="0"
            @update:model-value="debouncedApplyFilters"
            :label-color="$q.dark.isActive ? 'grey-4' : ''"
            :input-class="$q.dark.isActive ? 'text-grey-3' : ''"
          />
        </div>

        <div class="col-12 col-md-6 col-lg-4">
          <div
            class="text-caption q-mb-xs"
            :class="
              $q.dark.isActive ? 'text-grey-4' : 'text-grey-7'
            "
          >
            Transfer Value ({{ currencySymbol }})
          </div>
          <div class="row items-center q-col-gutter-x-sm">
            <div class="col">
              <q-input
                v-model="transferValueTextInput"
                :label="`Enter Value (e.g., ${currencySymbol}1.5M, ${currencySymbol}500K)`"
                dense
                outlined
                clearable
                @update:model-value="debouncedUpdateNumericValueFromTextInput"
                :disable="!isDataAvailable"
                placeholder="Any"
                :label-color="$q.dark.isActive ? 'grey-4' : ''"
                :input-class="
                  $q.dark.isActive ? 'text-grey-3' : ''
                "
              />
            </div>
            <div class="col-auto">
              <q-btn-toggle
                v-model="filters.transferValueMode"
                @update:model-value="applyFilters"
                no-caps
                rounded
                unelevated
                toggle-color="primary"
                :color="$q.dark.isActive ? 'grey-7' : 'white'"
                :text-color="$q.dark.isActive ? 'white' : 'primary'"
                size="sm"
                padding="xs md"
                :options="[
                  {
                    label: '<',
                    value: 'less',
                    slot: 'less-than',
                  },
                  {
                    label: '>',
                    value: 'more',
                    slot: 'more-than',
                  },
                ]"
                :disable="
                  !isDataAvailable ||
                  filters.selectedTransferValue === null
                "
              >
                <template v-slot:less-than
                  ><q-tooltip>Less than selected value</q-tooltip></template
                >
                <template v-slot:more-than
                  ><q-tooltip>More than selected value</q-tooltip></template
                >
              </q-btn-toggle>
            </div>
          </div>
          <q-slider
            class="q-mt-sm"
            v-model="filters.selectedTransferValue"
            :min="transferValueRange.min"
            :max="transferValueRange.max"
            :step="transferValueSliderStep"
            label
            :label-value="formatSliderValueWithCurrency(filters.selectedTransferValue)"
            @update:model-value="applyFilters"
            :disable="
              !isDataAvailable ||
              transferValueRange.min >= transferValueRange.max
            "
            color="primary"
          />
          <div
            class="text-caption q-mt-xs"
            :class="$q.dark.isActive ? 'text-grey-5' : 'text-grey-7'"
            v-if="filters.selectedTransferValue !== null"
          >
            Current filter:
            {{ filters.transferValueMode === "less" ? "Less than" : "More than" }}
            {{ formatSliderValueWithCurrency(filters.selectedTransferValue) }}
          </div>
        </div>

        <div class="col-12 flex items-center q-mt-md">
          <q-btn
            color="grey"
            :text-color="$q.dark.isActive ? 'white' : 'dark'"
            label="Clear All Filters"
            class="full-width"
            @click="clearAllFilters"
            :disable="!hasActiveFilters"
            outline
          />
        </div>
      </div>
    </q-card-section>
  </q-card>
</template>

<script>
import { ref, computed, watch, defineComponent } from "vue";
import { useQuasar } from "quasar";
import { formatCurrency, parseCurrencyString } from "../../utils/currencyUtils";

// MODIFIED: Ordered short positions for filter dropdown
const orderedShortPositions = [
  "GK",
  "DR",
  "DC",
  "DL",
  "WBR",
  "WBL",
  "DM",
  "MR",
  "MC",
  "ML",
  "AMR",
  "AMC",
  "AML",
  "ST",
];

// MODIFIED: Mapping from short position codes (filter values) to standardized full names (for matching player.parsedPositions)
const shortToStandardizedLongPosMap = {
  GK: ["Goalkeeper"],
  DR: ["Right Back"],
  DC: ["Centre Back"],
  DL: ["Left Back"],
  WBR: ["Right Wing-Back"],
  WBL: ["Left Wing-Back"],
  DM: ["Centre Defensive Midfielder"], // Assuming DM primarily maps to Centre DM for filtering
  MR: ["Right Midfielder"],
  MC: ["Centre Midfielder"],
  ML: ["Left Midfielder"],
  AMR: ["Right Attacking Midfielder", "Right Winger"], // Include winger if it's a distinct parsedPosition
  AMC: ["Centre Attacking Midfielder"],
  AML: ["Left Attacking Midfielder", "Left Winger"], // Include winger
  ST: ["Striker"],
};

function debounce(fn, delay) {
  let timeoutID = null;
  return function (...args) {
    clearTimeout(timeoutID);
    timeoutID = setTimeout(() => {
      fn.apply(this, args);
    }, delay);
  };
}

export default defineComponent({
  name: "PlayerFilters",
  props: {
    players: {
      type: Array,
      required: true,
    },
    currencySymbol: {
      type: String,
      default: "$",
    },
    transferValueRange: {
      type: Object,
      default: () => ({ min: 0, max: 100000000 }),
    },
    uniqueClubs: {
      type: Array,
      default: () => [],
    },
    uniqueNationalities: {
      type: Array,
      default: () => [],
    },
    uniqueMediaHandlings: {
      type: Array,
      default: () => [],
    },
    uniquePersonalities: {
      type: Array,
      default: () => [],
    },
  },
  emits: ["update:filters", "filter-changed"],
  setup(props, { emit }) {
    const $q = useQuasar();

    const isDataAvailable = computed(() => props.players && props.players.length > 0);

    // Initialize with default values
    const filters = ref({
      name: "",
      club: null,
      selectedTransferValue: null,
      transferValueMode: "less",
      position: null,
      nationality: null,
      mediaHandling: [],
      personality: [],
      minAge: null,
      maxAge: null,
    });

    const transferValueTextInput = ref("");

    const hasActiveFilters = computed(
      () =>
        filters.value.name !== "" ||
        filters.value.club !== null ||
        filters.value.selectedTransferValue !== null ||
        filters.value.position !== null ||
        filters.value.nationality !== null ||
        (Array.isArray(filters.value.mediaHandling) && filters.value.mediaHandling.length > 0) ||
        (Array.isArray(filters.value.personality) && filters.value.personality.length > 0) ||
        filters.value.minAge !== null ||
        filters.value.maxAge !== null
    );

    // Computed properties for dropdown options
    const clubOptions = ref([]);
    const nationalityOptions = ref([]);
    
    const positionOptions = computed(() => {
      const options = [{ label: "Any Position", value: null }];
      orderedShortPositions.forEach((shortPos) => {
        options.push({ label: shortPos, value: shortPos });
      });
      return options;
    });

    const mediaHandlingOptions = computed(() => 
      props.uniqueMediaHandlings.map((mh) => ({ label: mh, value: mh }))
    );

    const personalityOptions = computed(() => 
      props.uniquePersonalities.map((p) => ({ label: p, value: p }))
    );

    const transferValueSliderStep = computed(() => {
      const range = props.transferValueRange.max - props.transferValueRange.min;
      if (range <= 0) return 10000;
      if (range < 50000) return 1000;
      if (range < 250000) return 5000;
      if (range < 1000000) return 10000;
      if (range < 10000000) return 50000;
      if (range < 50000000) return 100000;
      return 250000;
    });

    const formatSliderValueWithCurrency = (value) => {
      if (value === null || value === undefined) return "";
      return formatCurrency(value, props.currencySymbol);
    };

    // Watch for changes in props
    watch(
      () => props.uniqueClubs,
      (newClubs) => {
        clubOptions.value = newClubs;
      },
      { immediate: true }
    );

    watch(
      () => props.uniqueNationalities,
      (newNationalities) => {
        nationalityOptions.value = newNationalities;
      },
      { immediate: true }
    );

    watch(
      () => props.transferValueRange,
      (newRange) => {
        if (filters.value.selectedTransferValue === null) {
          filters.value.selectedTransferValue = newRange.max;
          transferValueTextInput.value = formatSliderValueWithCurrency(newRange.max);
        } else {
          // Make sure the current value is within the new range
          filters.value.selectedTransferValue = Math.max(
            newRange.min,
            Math.min(filters.value.selectedTransferValue, newRange.max)
          );
        }
      },
      { immediate: true }
    );

    // Methods for filter handling
    const applyFilters = () => {
      emit("filter-changed", { ...filters.value });
    };

    const debouncedApplyFilters = debounce(applyFilters, 300);

    const updateNumericValueFromTextInput = () => {
      const numericValue = parseCurrencyString(transferValueTextInput.value);
      if (numericValue !== null) {
        const clampedValue = Math.max(
          props.transferValueRange.min,
          Math.min(numericValue, props.transferValueRange.max)
        );
        if (filters.value.selectedTransferValue !== clampedValue) {
          filters.value.selectedTransferValue = clampedValue;
        }
      } else if (transferValueTextInput.value.trim() === "") {
        if (filters.value.selectedTransferValue !== props.transferValueRange.max) {
          filters.value.selectedTransferValue = props.transferValueRange.max;
        }
      }
      applyFilters();
    };
    
    const debouncedUpdateNumericValueFromTextInput = debounce(updateNumericValueFromTextInput, 400);

    watch(
      () => filters.value.selectedTransferValue,
      (newValue) => {
        const currentTextParsed = parseCurrencyString(transferValueTextInput.value);
        if (currentTextParsed !== newValue || newValue === props.transferValueRange.max) {
          transferValueTextInput.value =
            newValue === props.transferValueRange.max && newValue !== null
              ? ""
              : formatSliderValueWithCurrency(newValue);
        }
      }
    );

    const clearAllFilters = () => {
      filters.value = {
        name: "",
        club: null,
        selectedTransferValue: props.transferValueRange.max,
        transferValueMode: "less",
        position: null,
        nationality: null,
        mediaHandling: [],
        personality: [],
        minAge: null,
        maxAge: null,
      };
      transferValueTextInput.value = "";
      applyFilters();
    };

    const filterClubOptions = (val, update) => {
      if (val === "") {
        update(() => {
          clubOptions.value = props.uniqueClubs;
        });
        return;
      }
      update(() => {
        const needle = val.toLowerCase();
        clubOptions.value = props.uniqueClubs.filter(
          (v) => v.toLowerCase().indexOf(needle) > -1
        );
      });
    };

    const filterNationalityOptions = (val, update) => {
      if (val === "") {
        update(() => {
          nationalityOptions.value = props.uniqueNationalities;
        });
        return;
      }
      update(() => {
        const needle = val.toLowerCase();
        nationalityOptions.value = props.uniqueNationalities.filter(
          (v) => v.toLowerCase().indexOf(needle) > -1
        );
      });
    };

    return {
      filters,
      hasActiveFilters,
      transferValueTextInput,
      clubOptions,
      nationalityOptions,
      positionOptions,
      mediaHandlingOptions,
      personalityOptions,
      transferValueSliderStep,
      isDataAvailable,
      applyFilters,
      debouncedApplyFilters,
      clearAllFilters,
      formatSliderValueWithCurrency,
      debouncedUpdateNumericValueFromTextInput,
      filterClubOptions,
      filterNationalityOptions,
      transferValueRange: props.transferValueRange,
    };
  },
});
</script>

<style lang="scss" scoped>
.filter-card {
  border-radius: $generic-border-radius;
}
</style>