<template>
    <q-card
        class="q-mb-md filter-card"
        :class="quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'"
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
                        :label-color="
                            quasarInstance.dark.isActive ? 'grey-4' : ''
                        "
                        :input-class="
                            quasarInstance.dark.isActive ? 'text-grey-3' : ''
                        "
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
                        outlined
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
                    />
                </div>
                <div class="col-12 col-sm-6 col-md-3 col-lg-3">
                    <q-select
                        v-model="filters.role"
                        :options="roleFilterOptions"
                        label="Role"
                        dense
                        outlined
                        clearable
                        emit-value
                        map-options
                        @update:model-value="applyFilters"
                        :disable="
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
                        outlined
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
                        :label-color="
                            quasarInstance.dark.isActive ? 'grey-4' : ''
                        "
                        :popup-content-class="
                            quasarInstance.dark.isActive
                                ? 'bg-grey-8 text-white'
                                : ''
                        "
                        behavior="menu"
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
                        :label-color="
                            quasarInstance.dark.isActive ? 'grey-4' : ''
                        "
                        :input-class="
                            quasarInstance.dark.isActive ? 'text-grey-3' : ''
                        "
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
                        :label-color="
                            quasarInstance.dark.isActive ? 'grey-4' : ''
                        "
                        :input-class="
                            quasarInstance.dark.isActive ? 'text-grey-3' : ''
                        "
                    />
                </div>
                <div class="col-12 col-md-6 col-lg-4">
                    <div
                        class="text-caption q-mb-xs"
                        :class="
                            quasarInstance.dark.isActive
                                ? 'text-grey-4'
                                : 'text-grey-7'
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
                                @update:model-value="
                                    debouncedUpdateNumericValueFromTextInput
                                "
                                :disable="!isDataAvailable"
                                placeholder="Any"
                                :label-color="
                                    quasarInstance.dark.isActive ? 'grey-4' : ''
                                "
                                :input-class="
                                    quasarInstance.dark.isActive
                                        ? 'text-grey-3'
                                        : ''
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
                                :color="
                                    quasarInstance.dark.isActive
                                        ? 'grey-7'
                                        : 'white'
                                "
                                :text-color="
                                    quasarInstance.dark.isActive
                                        ? 'white'
                                        : 'primary'
                                "
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
                                    filters.selectedTransferValue === null ||
                                    filters.selectedTransferValue ===
                                        transferValueRange.max
                                "
                            >
                                <template v-slot:less-than
                                    ><q-tooltip
                                        >Less than selected value</q-tooltip
                                    ></template
                                >
                                <template v-slot:more-than
                                    ><q-tooltip
                                        >More than selected value</q-tooltip
                                    ></template
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
                        :label-value="
                            formatSliderValueWithCurrency(
                                filters.selectedTransferValue,
                            )
                        "
                        @change="applyFilters"
                        :disable="
                            !isDataAvailable ||
                            transferValueRange.min >= transferValueRange.max
                        "
                        color="primary"
                    />
                    <div
                        class="text-caption q-mt-xs"
                        :class="
                            quasarInstance.dark.isActive
                                ? 'text-grey-5'
                                : 'text-grey-7'
                        "
                        v-if="
                            filters.selectedTransferValue !== null &&
                            filters.selectedTransferValue <
                                transferValueRange.max
                        "
                    >
                        Current filter:
                        {{
                            filters.transferValueMode === "less"
                                ? "Less than"
                                : "More than"
                        }}
                        {{
                            formatSliderValueWithCurrency(
                                filters.selectedTransferValue,
                            )
                        }}
                    </div>
                </div>
                <div class="col-12 flex items-center q-mt-md">
                    <q-btn
                        color="grey"
                        :text-color="
                            quasarInstance.dark.isActive ? 'white' : 'dark'
                        "
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
import { ref, computed, watch, defineComponent, onMounted } from "vue";
import { useQuasar } from "quasar";
import { usePlayerStore } from "../../stores/playerStore";
import { formatCurrency, parseCurrencyString } from "../../utils/currencyUtils";

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
        players: { type: Array, required: true },
        currencySymbol: { type: String, default: "$" },
        transferValueRange: {
            type: Object,
            default: () => ({ min: 0, max: 100000000 }),
        },
        uniqueClubs: { type: Array, default: () => [] },
        uniqueNationalities: { type: Array, default: () => [] },
        uniqueMediaHandlings: { type: Array, default: () => [] },
        uniquePersonalities: { type: Array, default: () => [] },
    },
    emits: ["filter-changed"],
    setup(props, { emit }) {
        const quasarInstance = useQuasar();
        const playerStore = usePlayerStore();

        const filters = ref({
            name: "",
            club: null,
            selectedTransferValue: props.transferValueRange.max,
            transferValueMode: "less",
            position: null,
            role: null,
            nationality: null,
            mediaHandling: [],
            personality: [],
            minAge: null,
            maxAge: null,
        });

        const transferValueTextInput = ref("");
        const clubOptions = ref([]);
        const nationalityOptions = ref([]);

        const isDataAvailable = computed(
            () => playerStore.allPlayers && playerStore.allPlayers.length > 0,
        );

        const hasActiveFilters = computed(
            () =>
                filters.value.name !== "" ||
                filters.value.club !== null ||
                (filters.value.selectedTransferValue !== null &&
                    filters.value.selectedTransferValue <
                        props.transferValueRange.max) ||
                filters.value.position !== null ||
                filters.value.role !== null ||
                filters.value.nationality !== null ||
                (Array.isArray(filters.value.mediaHandling) &&
                    filters.value.mediaHandling.length > 0) ||
                (Array.isArray(filters.value.personality) &&
                    filters.value.personality.length > 0) ||
                filters.value.minAge !== null ||
                filters.value.maxAge !== null,
        );

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

        const formatSliderValueWithCurrency = (value) => {
            if (value === null || value === undefined) return "Any";
            if (value === props.transferValueRange.max) return "Any";
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
            (newRange, oldRange) => {
                if (
                    newRange &&
                    (newRange.min !== oldRange?.min ||
                        newRange.max !== oldRange?.max)
                ) {
                    filters.value.selectedTransferValue = newRange.max;
                    if (
                        document.activeElement !==
                        transferValueTextInput.value?.$el?.querySelector(
                            "input",
                        )
                    ) {
                        transferValueTextInput.value = "";
                    }
                } else if (
                    filters.value.selectedTransferValue === null &&
                    newRange
                ) {
                    filters.value.selectedTransferValue = newRange.max;
                    if (
                        document.activeElement !==
                        transferValueTextInput.value?.$el?.querySelector(
                            "input",
                        )
                    ) {
                        transferValueTextInput.value = "";
                    }
                }
            },
            { immediate: true, deep: true },
        );

        const applyFilters = () => {
            console.log(
                "PlayerFilters: applyFilters called with:",
                JSON.parse(JSON.stringify(filters.value)),
            );
            emit("filter-changed", { ...filters.value });
        };
        const debouncedApplyFilters = debounce(applyFilters, 350);

        const onPositionChange = () => {
            filters.value.role = null;
            applyFilters();
        };

        const updateNumericValueFromTextInput = () => {
            const numericValue = parseCurrencyString(
                transferValueTextInput.value,
            );
            if (numericValue !== null) {
                filters.value.selectedTransferValue = Math.max(
                    props.transferValueRange.min,
                    Math.min(numericValue, props.transferValueRange.max),
                );
            } else if (transferValueTextInput.value.trim() === "") {
                filters.value.selectedTransferValue =
                    props.transferValueRange.max;
            }
            debouncedApplyFilters();
        };
        const debouncedUpdateNumericValueFromTextInput = debounce(
            updateNumericValueFromTextInput,
            450,
        );

        watch(
            () => filters.value.selectedTransferValue,
            (newValue) => {
                const currentTextParsed = parseCurrencyString(
                    transferValueTextInput.value,
                );
                if (
                    newValue === props.transferValueRange.max &&
                    newValue !== null
                ) {
                    if (transferValueTextInput.value !== "") {
                        if (
                            document.activeElement !==
                            transferValueTextInput.value?.$el?.querySelector(
                                "input",
                            )
                        ) {
                            transferValueTextInput.value = "";
                        }
                    }
                } else if (currentTextParsed !== newValue) {
                    if (
                        document.activeElement !==
                        transferValueTextInput.value?.$el?.querySelector(
                            "input",
                        )
                    ) {
                        transferValueTextInput.value =
                            formatSliderValueWithCurrency(newValue);
                    }
                }
            },
        );

        const clearAllFilters = () => {
            filters.value = {
                name: "",
                club: null,
                selectedTransferValue: props.transferValueRange.max,
                transferValueMode: "less",
                position: null,
                role: null,
                nationality: null,
                mediaHandling: [],
                personality: [],
                minAge: null,
                maxAge: null,
            };
            transferValueTextInput.value = "";
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
            filters.value.selectedTransferValue = props.transferValueRange.max;
            transferValueTextInput.value = "";
        });

        watch(
            () => playerStore.currentDatasetId,
            async (newId) => {
                if (newId && playerStore.allAvailableRoles.length === 0) {
                    await playerStore.fetchAllAvailableRoles();
                }
            },
        );
        watch(
            () => playerStore.allAvailableRoles,
            () => {
                // This watcher ensures roleFilterOptions updates if roles load after position is set
            },
        );

        return {
            quasarInstance,
            filters,
            hasActiveFilters,
            transferValueTextInput,
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
            formatSliderValueWithCurrency,
            debouncedUpdateNumericValueFromTextInput,
            filterClubOptions,
            filterNationalityOptions,
            onPositionChange,
            transferValueRange: computed(() => props.transferValueRange),
        };
    },
});
</script>

<style lang="scss" scoped>
.filter-card {
    border-radius: 8px;
}
.body--dark .q-field--outlined .q-field__control:before {
    border-color: rgba(255, 255, 255, 0.24);
}
.body--dark .q-field__label {
    color: rgba(255, 255, 255, 0.6);
}
.body--dark .q-select .q-field__input,
.body--dark .q-input .q-field__input {
    color: rgba(255, 255, 255, 0.87);
}
</style>
