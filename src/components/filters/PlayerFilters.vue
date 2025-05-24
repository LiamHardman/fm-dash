/ src/components/filters/PlayerFilters.vue
<template>
    <q-card
        class="q-mb-md filter-card"
        :class="quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'"
    >
        <q-card-section>
            <div class="text-subtitle1 q-mb-sm">
                Search Players (Using {{ currencySymbol }} for values)
            </div>

            <div class="row q-col-gutter-x-md q-col-gutter-y-sm items-end">
                <div class="col-12 col-sm-6 col-md-4 col-lg-2">
                    <q-input
                        v-model="filters.name"
                        label="Player Name"
                        dense
                        filled
                        clearable
                        @update:model-value="debouncedApplyFilters"
                        :label-color="
                            quasarInstance.dark.isActive ? 'grey-4' : ''
                        "
                        :input-class="
                            quasarInstance.dark.isActive ? 'text-grey-3' : ''
                        "
                        :disable="isLoading"
                    />
                </div>
                <div class="col-12 col-sm-6 col-md-4 col-lg-2">
                    <q-select
                        v-model="filters.club"
                        :options="clubOptions"
                        label="Club"
                        dense
                        filled
                        clearable
                        use-input
                        hide-selected
                        fill-input
                        input-debounce="300"
                        @filter="filterClubOptions"
                        @update:model-value="applyFilters"
                        :label-color="
                            quasarInstance.dark.isActive ? 'grey-4' : ''
                        "
                        :popup-content-class="
                            quasarInstance.dark.isActive
                                ? 'bg-grey-8 text-white'
                                : ''
                        "
                        behavior="menu"
                        :disable="isLoading"
                    >
                        <template v-slot:no-option>
                            <q-item
                                ><q-item-section class="text-grey"
                                    >No results</q-item-section
                                ></q-item
                            >
                        </template>
                    </q-select>
                </div>
                <div class="col-12 col-sm-6 col-md-4 col-lg-2">
                    <q-select
                        v-model="filters.nationality"
                        :options="nationalityOptions"
                        label="Nationality"
                        dense
                        filled
                        clearable
                        use-input
                        hide-selected
                        fill-input
                        input-debounce="300"
                        @filter="filterNationalityOptions"
                        @update:model-value="applyFilters"
                        :label-color="
                            quasarInstance.dark.isActive ? 'grey-4' : ''
                        "
                        :popup-content-class="
                            quasarInstance.dark.isActive
                                ? 'bg-grey-8 text-white'
                                : ''
                        "
                        behavior="menu"
                        :disable="isLoading"
                    >
                        <template v-slot:no-option>
                            <q-item
                                ><q-item-section class="text-grey"
                                    >No results</q-item-section
                                ></q-item
                            >
                        </template>
                    </q-select>
                </div>
                <div class="col-12 col-sm-6 col-md-3 col-lg-2">
                    <q-select
                        v-model="filters.position"
                        :options="positionOptions"
                        label="Position"
                        dense
                        filled
                        clearable
                        emit-value
                        map-options
                        @update:model-value="onPositionChange"
                        :label-color="
                            quasarInstance.dark.isActive ? 'grey-4' : ''
                        "
                        :popup-content-class="
                            quasarInstance.dark.isActive
                                ? 'bg-grey-8 text-white'
                                : ''
                        "
                        behavior="menu"
                        :disable="isLoading"
                    />
                </div>
                <div class="col-12 col-sm-6 col-md-3 col-lg-2">
                    <q-select
                        v-model="filters.role"
                        :options="roleFilterOptions"
                        label="Role"
                        dense
                        filled
                        clearable
                        emit-value
                        map-options
                        @update:model-value="applyFilters"
                        :disable="
                            isLoading ||
                            !filters.position ||
                            roleFilterOptions.length === 0 ||
                            (roleFilterOptions.length === 1 &&
                                roleFilterOptions[0].value === null)
                        "
                        :label-color="
                            quasarInstance.dark.isActive ? 'grey-4' : ''
                        "
                        :popup-content-class="
                            quasarInstance.dark.isActive
                                ? 'bg-grey-8 text-white'
                                : ''
                        "
                        behavior="menu"
                    >
                        <template v-slot:no-option>
                            <q-item>
                                <q-item-section class="text-grey">
                                    {{
                                        filters.position
                                            ? "No specific roles for this position"
                                            : "Select position first"
                                    }}
                                </q-item-section>
                            </q-item>
                        </template>
                    </q-select>
                </div>
                <div class="col-12 col-sm-6 col-md-3 col-lg-2">
                    <q-select
                        v-model="filters.mediaHandling"
                        :options="mediaHandlingOptions"
                        label="Media Handling"
                        dense
                        filled
                        multiple
                        use-chips
                        clearable
                        emit-value
                        map-options
                        @update:model-value="applyFilters"
                        :label-color="
                            quasarInstance.dark.isActive ? 'grey-4' : ''
                        "
                        :popup-content-class="
                            quasarInstance.dark.isActive
                                ? 'bg-grey-8 text-white'
                                : ''
                        "
                        behavior="menu"
                        :disable="isLoading"
                    />
                </div>
            </div>

            <div
                class="row q-col-gutter-x-md q-col-gutter-y-sm items-start q-mt-sm"
            >
                <div
                    class="col-12 col-sm-6 col-md-4 col-lg-3 filter-item-container"
                >
                    <q-select
                        v-model="filters.personality"
                        :options="personalityOptions"
                        label="Personality"
                        dense
                        filled
                        multiple
                        use-chips
                        clearable
                        emit-value
                        map-options
                        @update:model-value="applyFilters"
                        :label-color="
                            quasarInstance.dark.isActive ? 'grey-4' : ''
                        "
                        :popup-content-class="
                            quasarInstance.dark.isActive
                                ? 'bg-grey-8 text-white'
                                : ''
                        "
                        behavior="menu"
                        :disable="isLoading"
                    />
                </div>

                <div
                    class="col-12 col-sm-6 col-md-4 col-lg-3 filter-item-container"
                >
                    <div
                        class="text-caption q-mb-xs slider-label"
                        :class="
                            quasarInstance.dark.isActive
                                ? 'text-grey-4'
                                : 'text-grey-7'
                        "
                    >
                        Age Range: {{ filters.ageRange.min }} -
                        {{
                            filters.ageRange.max === ageSliderMax
                                ? ageSliderMax + "+"
                                : filters.ageRange.max
                        }}
                    </div>
                    <q-range
                        v-model="filters.ageRange"
                        :min="ageSliderMin"
                        :max="ageSliderMax"
                        :step="1"
                        label-always
                        :left-label-value="filters.ageRange.min + ' yrs'"
                        :right-label-value="
                            filters.ageRange.max +
                            (filters.ageRange.max === ageSliderMax ? '+' : '') +
                            ' yrs'
                        "
                        @update:model-value="debouncedApplyFilters"
                        color="primary"
                        class="q-px-sm"
                        :disable="isLoading"
                    />
                </div>

                <div class="col-12 col-md-8 col-lg-4 filter-item-container">
                    <div
                        class="text-caption q-mb-xs slider-label"
                        :class="
                            quasarInstance.dark.isActive
                                ? 'text-grey-4'
                                : 'text-grey-7'
                        "
                    >
                        Transfer Value ({{ currencySymbol }})
                    </div>
                    <q-range
                        v-model="filters.transferValueRangeLocal"
                        :min="transferValueRange.min"
                        :max="transferValueRange.max"
                        :step="transferValueSliderStep"
                        label-always
                        :left-label-value="
                            formatRangeLabel(
                                filters.transferValueRangeLocal.min,
                                false,
                            )
                        "
                        :right-label-value="
                            formatRangeLabel(
                                filters.transferValueRangeLocal.max,
                                true,
                            )
                        "
                        @update:model-value="debouncedApplyFilters"
                        :disable="
                            isLoading ||
                            !isDataAvailable ||
                            transferValueRange.min >= transferValueRange.max
                        "
                        color="primary"
                        class="q-px-sm"
                    />
                </div>
                <div
                    class="col-12 col-md-4 col-lg-2 filter-item-container self-end"
                >
                    <q-btn
                        color="grey"
                        :text-color="
                            quasarInstance.dark.isActive ? 'white' : 'dark'
                        "
                        label="Clear All Filters"
                        class="full-width"
                        @click="clearAllFilters"
                        :disable="isLoading || !hasActiveFilters"
                        outline
                        dense
                    />
                </div>
            </div>
        </q-card-section>
    </q-card>
</template>

<script>
import { ref, computed, watch, defineComponent, onMounted } from "vue";
import { useQuasar } from "quasar";
import { usePlayerStore } from "@/stores/playerStore"; // Corrected Import Path
import { formatCurrency } from "@/utils/currencyUtils";

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

const AGE_SLIDER_MIN = 15;
const AGE_SLIDER_MAX = 50;

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
        currencySymbol: { type: String, default: "$" },
        transferValueRange: {
            type: Object,
            default: () => ({ min: 0, max: 100000000 }),
        },
        uniqueClubs: { type: Array, default: () => [] },
        uniqueNationalities: { type: Array, default: () => [] },
        uniqueMediaHandlings: { type: Array, default: () => [] },
        uniquePersonalities: { type: Array, default: () => [] },
        isLoading: { type: Boolean, default: false },
    },
    emits: ["filter-changed"],
    setup(props, { emit }) {
        const quasarInstance = useQuasar();
        const playerStore = usePlayerStore();

        const filters = ref({
            name: "",
            club: null,
            position: null,
            role: null,
            nationality: null,
            mediaHandling: [],
            personality: [],
            ageRange: { min: AGE_SLIDER_MIN, max: AGE_SLIDER_MAX },
            transferValueRangeLocal: {
                min: props.transferValueRange.min,
                max: props.transferValueRange.max,
            },
        });

        const clubOptions = ref([]);
        const nationalityOptions = ref([]);

        const absoluteMinTransferValue = ref(props.transferValueRange.min);
        const absoluteMaxTransferValue = ref(props.transferValueRange.max);

        const isDataAvailable = computed(
            () => playerStore.allPlayers && playerStore.allPlayers.length > 0,
        );

        const hasActiveFilters = computed(() => {
            const defAgeMin = AGE_SLIDER_MIN;
            const defAgeMax = AGE_SLIDER_MAX;
            const defValMin = absoluteMinTransferValue.value;
            const defValMax = absoluteMaxTransferValue.value;

            return (
                filters.value.name !== "" ||
                filters.value.club !== null ||
                filters.value.position !== null ||
                filters.value.role !== null ||
                filters.value.nationality !== null ||
                (Array.isArray(filters.value.mediaHandling) &&
                    filters.value.mediaHandling.length > 0) ||
                (Array.isArray(filters.value.personality) &&
                    filters.value.personality.length > 0) ||
                filters.value.ageRange.min !== defAgeMin ||
                filters.value.ageRange.max !== defAgeMax ||
                filters.value.transferValueRangeLocal.min !== defValMin ||
                filters.value.transferValueRangeLocal.max !== defValMax
            );
        });

        const positionOptions = computed(() => {
            const options = [{ label: "Any Position", value: null }];
            orderedShortPositions.forEach((shortPos) => {
                options.push({ label: shortPos, value: shortPos });
            });
            return options;
        });

        const roleFilterOptions = computed(() => {
            if (
                !filters.value.position ||
                !playerStore.allAvailableRoles ||
                playerStore.allAvailableRoles.length === 0
            ) {
                return [{ label: "Any Role", value: null }];
            }
            const selectedPosShortCode = filters.value.position;
            const filtered = playerStore.allAvailableRoles
                .filter((roleFullName) =>
                    roleFullName.startsWith(selectedPosShortCode + " - "),
                )
                .map((roleFullName) => ({
                    label: roleFullName,
                    value: roleFullName,
                }))
                .sort((a, b) => a.label.localeCompare(b.label));
            return [{ label: "Any Role", value: null }, ...filtered];
        });

        const mediaHandlingOptions = computed(() =>
            props.uniqueMediaHandlings.map((mh) => ({ label: mh, value: mh })),
        );

        const personalityOptions = computed(() =>
            props.uniquePersonalities.map((p) => ({ label: p, value: p })),
        );

        const transferValueSliderStep = computed(() => {
            const range =
                props.transferValueRange.max - props.transferValueRange.min;
            if (range <= 0) return 10000;
            if (range < 50000) return 1000;
            if (range < 250000) return 5000;
            if (range < 1000000) return 10000;
            if (range < 10000000) return 50000;
            if (range < 50000000) return 100000;
            return 250000;
        });

        const formatRangeLabel = (value, isMaxBoundary = false) => {
            if (value === null || value === undefined) return "N/A";
            if (isMaxBoundary) {
                if (
                    absoluteMaxTransferValue.value !== null &&
                    value === absoluteMaxTransferValue.value
                ) {
                    return "Any";
                }
            } else {
                if (
                    absoluteMinTransferValue.value !== null &&
                    value === absoluteMinTransferValue.value
                ) {
                    // For min, just show the formatted value, not "Min" or "Any"
                    return formatCurrency(value, props.currencySymbol) || "0";
                }
            }
            return formatCurrency(value, props.currencySymbol);
        };

        watch(
            () => props.uniqueClubs,
            (newClubs) => {
                clubOptions.value = newClubs;
            },
            { immediate: true },
        );
        watch(
            () => props.uniqueNationalities,
            (newNats) => {
                nationalityOptions.value = newNats;
            },
            { immediate: true },
        );

        watch(
            () => props.transferValueRange,
            (newRange) => {
                if (
                    newRange &&
                    typeof newRange.min === "number" &&
                    typeof newRange.max === "number"
                ) {
                    if (
                        absoluteMinTransferValue.value === null ||
                        newRange.min !== absoluteMinTransferValue.value
                    ) {
                        absoluteMinTransferValue.value = newRange.min;
                    }
                    if (
                        absoluteMaxTransferValue.value === null ||
                        newRange.max !== absoluteMaxTransferValue.value
                    ) {
                        absoluteMaxTransferValue.value = newRange.max;
                    }
                    // Only reset local filter if it's outside the new absolute bounds or was never set
                    if (
                        filters.value.transferValueRangeLocal.min <
                            newRange.min ||
                        filters.value.transferValueRangeLocal.min >
                            newRange.max ||
                        filters.value.transferValueRangeLocal.max <
                            newRange.min ||
                        filters.value.transferValueRangeLocal.max >
                            newRange.max ||
                        filters.value.transferValueRangeLocal.min === null ||
                        filters.value.transferValueRangeLocal.max === null
                    ) {
                        filters.value.transferValueRangeLocal = {
                            min: newRange.min,
                            max: newRange.max,
                        };
                    }
                }
            },
            { deep: true, immediate: true },
        );

        const applyFilters = () => {
            if (props.isLoading) return;
            emit("filter-changed", { ...filters.value });
        };
        const debouncedApplyFilters = debounce(applyFilters, 400);

        const onPositionChange = () => {
            filters.value.role = null;
            applyFilters();
        };

        const clearAllFilters = () => {
            filters.value = {
                name: "",
                club: null,
                position: null,
                role: null,
                nationality: null,
                mediaHandling: [],
                personality: [],
                ageRange: { min: AGE_SLIDER_MIN, max: AGE_SLIDER_MAX },
                transferValueRangeLocal: {
                    min:
                        absoluteMinTransferValue.value !== null
                            ? absoluteMinTransferValue.value
                            : props.transferValueRange.min,
                    max:
                        absoluteMaxTransferValue.value !== null
                            ? absoluteMaxTransferValue.value
                            : props.transferValueRange.max,
                },
            };
            applyFilters();
        };

        const filterClubOptions = (val, update, abort) => {
            if (val.length < 1 && val !== "") {
                abort();
                return;
            }
            update(() => {
                const needle = val.toLowerCase();
                clubOptions.value = props.uniqueClubs.filter(
                    (v) => v.toLowerCase().indexOf(needle) > -1,
                );
            });
        };

        const filterNationalityOptions = (val, update, abort) => {
            if (val.length < 1 && val !== "") {
                abort();
                return;
            }
            update(() => {
                const needle = val.toLowerCase();
                nationalityOptions.value = props.uniqueNationalities.filter(
                    (v) => v.toLowerCase().indexOf(needle) > -1,
                );
            });
        };

        onMounted(async () => {
            if (
                playerStore.allAvailableRoles.length === 0 &&
                playerStore.currentDatasetId
            ) {
                await playerStore.fetchAllAvailableRoles();
            }
            if (props.transferValueRange) {
                absoluteMinTransferValue.value = props.transferValueRange.min;
                absoluteMaxTransferValue.value = props.transferValueRange.max;
                filters.value.transferValueRangeLocal = {
                    min: props.transferValueRange.min,
                    max: props.transferValueRange.max,
                };
            }
            filters.value.ageRange = {
                min: AGE_SLIDER_MIN,
                max: AGE_SLIDER_MAX,
            };
        });

        watch(
            () => playerStore.currentDatasetId,
            async (newId) => {
                if (newId && playerStore.allAvailableRoles.length === 0) {
                    await playerStore.fetchAllAvailableRoles();
                }
                if (newId && props.transferValueRange) {
                    absoluteMinTransferValue.value =
                        props.transferValueRange.min;
                    absoluteMaxTransferValue.value =
                        props.transferValueRange.max;
                    filters.value.transferValueRangeLocal = {
                        min: props.transferValueRange.min,
                        max: props.transferValueRange.max,
                    };
                }
            },
        );

        return {
            quasarInstance,
            filters,
            hasActiveFilters,
            clubOptions,
            nationalityOptions,
            positionOptions,
            roleFilterOptions,
            mediaHandlingOptions,
            personalityOptions,
            transferValueSliderStep,
            isDataAvailable,
            applyFilters,
            debouncedApplyFilters,
            clearAllFilters,
            formatRangeLabel,
            filterClubOptions,
            filterNationalityOptions,
            onPositionChange,
            ageSliderMin: AGE_SLIDER_MIN,
            ageSliderMax: AGE_SLIDER_MAX,
            transferValueRange: computed(() => props.transferValueRange),
        };
    },
});
</script>

<style lang="scss" scoped>
.filter-card {
    border-radius: 8px;
}
.body--dark .q-field--filled .q-field__control {
    background-color: rgba(255, 255, 255, 0.07);
    &:before {
        border-bottom-color: rgba(255, 255, 255, 0.24);
    }
    &:hover:before {
        border-bottom-color: rgba(255, 255, 255, 0.5);
    }
}
.body--dark .q-field--filled.q-field--focused .q-field__control:after {
    border-bottom-color: $primary;
}
.body--dark .q-field__label {
    color: rgba(255, 255, 255, 0.6);
}
.body--dark .q-select .q-field__input,
.body--dark .q-input .q-field__input {
    color: rgba(255, 255, 255, 0.87);
}
.body--light .q-field--filled .q-field__control {
    background-color: rgba(0, 0, 0, 0.04);
    &:before {
        border-bottom-color: rgba(0, 0, 0, 0.12);
    }
    &:hover:before {
        border-bottom-color: rgba(0, 0, 0, 0.32);
    }
}
.body--light .q-field--filled.q-field--focused .q-field__control:after {
    border-bottom-color: $primary;
}
.q-mt-sm {
    margin-top: 12px;
}
.q-px-sm {
    padding-left: 4px;
    padding-right: 4px;
}
.slider-label {
    padding-left: 4px;
    margin-bottom: 0px; // Reduced from q-mb-xs default for tighter spacing
    line-height: 1.2; // Ensure consistent line height
}
.filter-item-container {
    // Ensure consistent vertical alignment if items wrap
}
</style>
