// src/components/filters/PlayerFilters.vue
<template>
    <q-card
        class="q-mb-md filter-card"
        :class="quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'"
    >
        <q-card-section>
            <div class="text-subtitle1 q-mb-sm">
                Search Players (Using {{ currencySymbol }} for values)
            </div>

            <!-- First Row: Basic Filters -->
            <div class="row q-col-gutter-md q-mb-md">
                <div class="col-12 col-sm-6 col-md-3">
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
                <div class="col-12 col-sm-6 col-md-3">
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
                <div class="col-12 col-sm-6 col-md-3">
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
                <div class="col-12 col-sm-6 col-md-3">
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
            </div>

            <!-- Second Row: Advanced Filters -->
            <div class="row q-col-gutter-md q-mb-md">
                <div class="col-12 col-sm-6 col-md-3">
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
                <div class="col-12 col-sm-6 col-md-3">
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
                <div class="col-12 col-sm-6 col-md-3">
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
                <div class="col-12 col-sm-6 col-md-3">
                    <q-btn
                        color="primary"
                        :label="
                            'Set Minimum Stats' +
                            (hasActiveStatFilters ? ' (Active)' : '')
                        "
                        class="full-width"
                        @click="showMinimumStatsModal = true"
                        :disable="isLoading"
                        outline
                        icon="tune"
                        style="height: 40px;"
                    />
                </div>
            </div>

            <!-- Third Row: Range Sliders -->
            <div class="row q-col-gutter-md">
                <div class="col-12 col-md-3">
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

                <div class="col-12 col-md-3">
                    <div
                        class="text-caption q-mb-xs slider-label"
                        :class="
                            quasarInstance.dark.isActive
                                ? 'text-grey-4'
                                : 'text-grey-7'
                        "
                    >
                        Max Salary:
                        {{
                            filters.maxSalary === salarySliderMax
                                ? "Any"
                                : formatCurrency(
                                      filters.maxSalary,
                                      currencySymbol,
                                  )
                        }}
                    </div>
                    <q-slider
                        v-model="filters.maxSalary"
                        :min="salarySliderMin"
                        :max="salarySliderMax"
                        :step="salarySliderStep"
                        label-always
                        :label-value="
                            filters.maxSalary === salarySliderMax
                                ? 'Any'
                                : formatCurrency(
                                      filters.maxSalary,
                                      currencySymbol,
                                  )
                        "
                        @update:model-value="debouncedApplyFilters"
                        color="primary"
                        class="q-px-sm"
                        :disable="isLoading || !isDataAvailable"
                    />
                </div>

                <div class="col-12 col-md-4">
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
                        :min="currentSliderMin"
                        :max="currentSliderMax"
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
                            currentSliderMin >= currentSliderMax
                        "
                        color="primary"
                        class="q-px-sm"
                    />
                </div>

                <div class="col-12 col-md-2">
                    <q-btn
                        color="grey"
                        :text-color="
                            quasarInstance.dark.isActive ? 'white' : 'dark'
                        "
                        label="Clear All"
                        class="full-width"
                        @click="clearAllFilters"
                        :disable="isLoading || !hasActiveFilters"
                        outline
                        style="height: 40px;"
                    />
                </div>
            </div>

            <q-dialog v-model="showMinimumStatsModal" persistent maximized>
                <q-card
                    :class="
                        quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'
                    "
                >
                    <q-card-section>
                        <div class="text-h6">Set Minimum Stats</div>
                        <div class="text-subtitle2 text-grey-6 q-mt-xs">
                            Filter players by minimum stat values
                        </div>
                    </q-card-section>

                    <q-card-section class="q-pt-none modal-content">
                        <div class="row q-col-gutter-lg">
                            <div class="col-12 col-md-3">
                                <div class="attribute-group">
                                    <div
                                        class="text-h6 q-mb-sm attribute-group-title"
                                    >
                                        Stat Summaries
                                    </div>

                                    <div class="fifa-stat-item q-mb-md">
                                        <div
                                            class="text-caption q-mb-xs slider-label"
                                            :class="
                                                quasarInstance.dark.isActive
                                                    ? 'text-grey-4'
                                                    : 'text-grey-7'
                                            "
                                        >
                                            Min Overall:
                                            <span
                                                class="stat-value-badge q-ml-xs"
                                                :class="
                                                    getUnifiedRatingClass(
                                                        filters.minOverall,
                                                        100,
                                                    )
                                                "
                                            >
                                                {{ filters.minOverall || 0 }}
                                            </span>
                                        </div>
                                        <q-slider
                                            v-model="filters.minOverall"
                                            :min="0"
                                            :max="99"
                                            :step="1"
                                            color="primary"
                                            class="q-px-sm"
                                        />
                                    </div>

                                    <div class="fifa-stat-item q-mb-md">
                                        <div
                                            class="text-caption q-mb-xs slider-label"
                                            :class="
                                                quasarInstance.dark.isActive
                                                    ? 'text-grey-4'
                                                    : 'text-grey-7'
                                            "
                                        >
                                            Min PAC:
                                            <span
                                                class="stat-value-badge q-ml-xs"
                                                :class="
                                                    getUnifiedRatingClass(
                                                        filters.minPAC,
                                                        100,
                                                    )
                                                "
                                            >
                                                {{ filters.minPAC || 0 }}
                                            </span>
                                        </div>
                                        <q-slider
                                            v-model="filters.minPAC"
                                            :min="0"
                                            :max="99"
                                            :step="1"
                                            color="primary"
                                            class="q-px-sm"
                                        />
                                    </div>

                                    <div class="fifa-stat-item q-mb-md">
                                        <div
                                            class="text-caption q-mb-xs slider-label"
                                            :class="
                                                quasarInstance.dark.isActive
                                                    ? 'text-grey-4'
                                                    : 'text-grey-7'
                                            "
                                        >
                                            Min SHO:
                                            <span
                                                class="stat-value-badge q-ml-xs"
                                                :class="
                                                    getUnifiedRatingClass(
                                                        filters.minSHO,
                                                        100,
                                                    )
                                                "
                                            >
                                                {{ filters.minSHO || 0 }}
                                            </span>
                                        </div>
                                        <q-slider
                                            v-model="filters.minSHO"
                                            :min="0"
                                            :max="99"
                                            :step="1"
                                            color="primary"
                                            class="q-px-sm"
                                        />
                                    </div>

                                    <div class="fifa-stat-item q-mb-md">
                                        <div
                                            class="text-caption q-mb-xs slider-label"
                                            :class="
                                                quasarInstance.dark.isActive
                                                    ? 'text-grey-4'
                                                    : 'text-grey-7'
                                            "
                                        >
                                            Min PAS:
                                            <span
                                                class="stat-value-badge q-ml-xs"
                                                :class="
                                                    getUnifiedRatingClass(
                                                        filters.minPAS,
                                                        100,
                                                    )
                                                "
                                            >
                                                {{ filters.minPAS || 0 }}
                                            </span>
                                        </div>
                                        <q-slider
                                            v-model="filters.minPAS"
                                            :min="0"
                                            :max="99"
                                            :step="1"
                                            color="primary"
                                            class="q-px-sm"
                                        />
                                    </div>

                                    <div class="fifa-stat-item q-mb-md">
                                        <div
                                            class="text-caption q-mb-xs slider-label"
                                            :class="
                                                quasarInstance.dark.isActive
                                                    ? 'text-grey-4'
                                                    : 'text-grey-7'
                                            "
                                        >
                                            Min DRI:
                                            <span
                                                class="stat-value-badge q-ml-xs"
                                                :class="
                                                    getUnifiedRatingClass(
                                                        filters.minDRI,
                                                        100,
                                                    )
                                                "
                                            >
                                                {{ filters.minDRI || 0 }}
                                            </span>
                                        </div>
                                        <q-slider
                                            v-model="filters.minDRI"
                                            :min="0"
                                            :max="99"
                                            :step="1"
                                            color="primary"
                                            class="q-px-sm"
                                        />
                                    </div>

                                    <div class="fifa-stat-item q-mb-md">
                                        <div
                                            class="text-caption q-mb-xs slider-label"
                                            :class="
                                                quasarInstance.dark.isActive
                                                    ? 'text-grey-4'
                                                    : 'text-grey-7'
                                            "
                                        >
                                            Min DEF:
                                            <span
                                                class="stat-value-badge q-ml-xs"
                                                :class="
                                                    getUnifiedRatingClass(
                                                        filters.minDEF,
                                                        100,
                                                    )
                                                "
                                            >
                                                {{ filters.minDEF || 0 }}
                                            </span>
                                        </div>
                                        <q-slider
                                            v-model="filters.minDEF"
                                            :min="0"
                                            :max="99"
                                            :step="1"
                                            color="primary"
                                            class="q-px-sm"
                                        />
                                    </div>

                                    <div class="fifa-stat-item q-mb-md">
                                        <div
                                            class="text-caption q-mb-xs slider-label"
                                            :class="
                                                quasarInstance.dark.isActive
                                                    ? 'text-grey-4'
                                                    : 'text-grey-7'
                                            "
                                        >
                                            Min PHY:
                                            <span
                                                class="stat-value-badge q-ml-xs"
                                                :class="
                                                    getUnifiedRatingClass(
                                                        filters.minPHY,
                                                        100,
                                                    )
                                                "
                                            >
                                                {{ filters.minPHY || 0 }}
                                            </span>
                                        </div>
                                        <q-slider
                                            v-model="filters.minPHY"
                                            :min="0"
                                            :max="99"
                                            :step="1"
                                            color="primary"
                                            class="q-px-sm"
                                        />
                                    </div>

                                    <div class="fifa-stat-item q-mb-md">
                                        <div
                                            class="text-caption q-mb-xs slider-label"
                                            :class="
                                                quasarInstance.dark.isActive
                                                    ? 'text-grey-4'
                                                    : 'text-grey-7'
                                            "
                                        >
                                            Min GK:
                                            <span
                                                class="stat-value-badge q-ml-xs"
                                                :class="
                                                    getUnifiedRatingClass(
                                                        filters.minGK,
                                                        100,
                                                    )
                                                "
                                            >
                                                {{ filters.minGK || 0 }}
                                            </span>
                                        </div>
                                        <q-slider
                                            v-model="filters.minGK"
                                            :min="0"
                                            :max="99"
                                            :step="1"
                                            color="primary"
                                            class="q-px-sm"
                                        />
                                    </div>
                                </div>
                            </div>

                            <div class="col-12 col-md-9">
                                <div
                                    class="text-subtitle1 q-mb-sm attributes-section-title"
                                >
                                    FM Attributes (Min 0-20)
                                </div>
                                <div class="row q-col-gutter-md">
                                    <div class="col-12 col-lg-3 col-md-6">
                                        <q-card
                                            flat
                                            bordered
                                            class="attribute-category-card"
                                        >
                                            <q-card-section
                                                class="attribute-category-header-styled"
                                            >
                                                <div
                                                    class="text-subtitle1 text-weight-medium text-center"
                                                >
                                                    Technical
                                                </div>
                                            </q-card-section>
                                            <q-list
                                                separator
                                                dense
                                                class="attribute-list-column-styled"
                                            >
                                                <q-item
                                                    v-for="attr in technicalAttributeKeys"
                                                    :key="attr"
                                                    class="attribute-item-column-styled"
                                                >
                                                    <q-item-section
                                                        class="attribute-name-column-styled"
                                                    >
                                                        {{
                                                            formatAttrName(attr)
                                                        }}
                                                    </q-item-section>
                                                    <q-item-section
                                                        side
                                                        class="attribute-value-column-styled"
                                                    >
                                                        <span
                                                            v-if="
                                                                inlineEditingAttributeKey !==
                                                                attr
                                                            "
                                                            :class="
                                                                getUnifiedRatingClass(
                                                                    filters[
                                                                        `min${formatAttrKey(attr)}`
                                                                    ],
                                                                    20,
                                                                )
                                                            "
                                                            class="clickable-attribute attribute-badge-styled"
                                                            @click="
                                                                startInlineEdit(
                                                                    attr,
                                                                )
                                                            "
                                                        >
                                                            {{
                                                                filters[
                                                                    `min${formatAttrKey(attr)}`
                                                                ] || 0
                                                            }}
                                                        </span>
                                                        <q-input
                                                            v-else
                                                            :ref="
                                                                (el) =>
                                                                    (attributeInputRefs[
                                                                        attr
                                                                    ] = el)
                                                            "
                                                            v-model.number="
                                                                inlineEditingValue
                                                            "
                                                            type="number"
                                                            min="0"
                                                            max="20"
                                                            step="1"
                                                            dense
                                                            filled
                                                            autofocus
                                                            class="inline-attribute-input"
                                                            @keyup.enter="
                                                                finishInlineEdit
                                                            "
                                                            @blur="
                                                                finishInlineEdit
                                                            "
                                                        />
                                                    </q-item-section>
                                                </q-item>
                                            </q-list>
                                        </q-card>
                                    </div>
                                    <div class="col-12 col-lg-3 col-md-6">
                                        <q-card
                                            flat
                                            bordered
                                            class="attribute-category-card"
                                        >
                                            <q-card-section
                                                class="attribute-category-header-styled"
                                            >
                                                <div
                                                    class="text-subtitle1 text-weight-medium text-center"
                                                >
                                                    Mental
                                                </div>
                                            </q-card-section>
                                            <q-list
                                                separator
                                                dense
                                                class="attribute-list-column-styled"
                                            >
                                                <q-item
                                                    v-for="attr in mentalAttributeKeys"
                                                    :key="attr"
                                                    class="attribute-item-column-styled"
                                                >
                                                    <q-item-section
                                                        class="attribute-name-column-styled"
                                                    >
                                                        {{
                                                            formatAttrName(attr)
                                                        }}
                                                    </q-item-section>
                                                    <q-item-section
                                                        side
                                                        class="attribute-value-column-styled"
                                                    >
                                                        <span
                                                            v-if="
                                                                inlineEditingAttributeKey !==
                                                                attr
                                                            "
                                                            :class="
                                                                getUnifiedRatingClass(
                                                                    filters[
                                                                        `min${formatAttrKey(attr)}`
                                                                    ],
                                                                    20,
                                                                )
                                                            "
                                                            class="clickable-attribute attribute-badge-styled"
                                                            @click="
                                                                startInlineEdit(
                                                                    attr,
                                                                )
                                                            "
                                                        >
                                                            {{
                                                                filters[
                                                                    `min${formatAttrKey(attr)}`
                                                                ] || 0
                                                            }}
                                                        </span>
                                                        <q-input
                                                            v-else
                                                            :ref="
                                                                (el) =>
                                                                    (attributeInputRefs[
                                                                        attr
                                                                    ] = el)
                                                            "
                                                            v-model.number="
                                                                inlineEditingValue
                                                            "
                                                            type="number"
                                                            min="0"
                                                            max="20"
                                                            step="1"
                                                            dense
                                                            filled
                                                            autofocus
                                                            class="inline-attribute-input"
                                                            @keyup.enter="
                                                                finishInlineEdit
                                                            "
                                                            @blur="
                                                                finishInlineEdit
                                                            "
                                                        />
                                                    </q-item-section>
                                                </q-item>
                                            </q-list>
                                        </q-card>
                                    </div>
                                    <div class="col-12 col-lg-3 col-md-6">
                                        <q-card
                                            flat
                                            bordered
                                            class="attribute-category-card"
                                        >
                                            <q-card-section
                                                class="attribute-category-header-styled"
                                            >
                                                <div
                                                    class="text-subtitle1 text-weight-medium text-center"
                                                >
                                                    Physical
                                                </div>
                                            </q-card-section>
                                            <q-list
                                                separator
                                                dense
                                                class="attribute-list-column-styled"
                                            >
                                                <q-item
                                                    v-for="attr in physicalAttributeKeys"
                                                    :key="attr"
                                                    class="attribute-item-column-styled"
                                                >
                                                    <q-item-section
                                                        class="attribute-name-column-styled"
                                                    >
                                                        {{
                                                            formatAttrName(attr)
                                                        }}
                                                    </q-item-section>
                                                    <q-item-section
                                                        side
                                                        class="attribute-value-column-styled"
                                                    >
                                                        <span
                                                            v-if="
                                                                inlineEditingAttributeKey !==
                                                                attr
                                                            "
                                                            :class="
                                                                getUnifiedRatingClass(
                                                                    filters[
                                                                        `min${formatAttrKey(attr)}`
                                                                    ],
                                                                    20,
                                                                )
                                                            "
                                                            class="clickable-attribute attribute-badge-styled"
                                                            @click="
                                                                startInlineEdit(
                                                                    attr,
                                                                )
                                                            "
                                                        >
                                                            {{
                                                                filters[
                                                                    `min${formatAttrKey(attr)}`
                                                                ] || 0
                                                            }}
                                                        </span>
                                                        <q-input
                                                            v-else
                                                            :ref="
                                                                (el) =>
                                                                    (attributeInputRefs[
                                                                        attr
                                                                    ] = el)
                                                            "
                                                            v-model.number="
                                                                inlineEditingValue
                                                            "
                                                            type="number"
                                                            min="0"
                                                            max="20"
                                                            step="1"
                                                            dense
                                                            filled
                                                            autofocus
                                                            class="inline-attribute-input"
                                                            @keyup.enter="
                                                                finishInlineEdit
                                                            "
                                                            @blur="
                                                                finishInlineEdit
                                                            "
                                                        />
                                                    </q-item-section>
                                                </q-item>
                                            </q-list>
                                        </q-card>
                                    </div>
                                    <div class="col-12 col-lg-3 col-md-6">
                                        <q-card
                                            flat
                                            bordered
                                            class="attribute-category-card"
                                        >
                                            <q-card-section
                                                class="attribute-category-header-styled"
                                            >
                                                <div
                                                    class="text-subtitle1 text-weight-medium text-center"
                                                >
                                                    Goalkeeping
                                                </div>
                                            </q-card-section>
                                            <q-list
                                                separator
                                                dense
                                                class="attribute-list-column-styled"
                                            >
                                                <q-item
                                                    v-for="attr in goalkeeperAttributeKeys"
                                                    :key="attr"
                                                    class="attribute-item-column-styled"
                                                >
                                                    <q-item-section
                                                        class="attribute-name-column-styled"
                                                    >
                                                        {{
                                                            formatAttrName(attr)
                                                        }}
                                                    </q-item-section>
                                                    <q-item-section
                                                        side
                                                        class="attribute-value-column-styled"
                                                    >
                                                        <span
                                                            v-if="
                                                                inlineEditingAttributeKey !==
                                                                attr
                                                            "
                                                            :class="
                                                                getUnifiedRatingClass(
                                                                    filters[
                                                                        `min${formatAttrKey(attr)}`
                                                                    ],
                                                                    20,
                                                                )
                                                            "
                                                            class="clickable-attribute attribute-badge-styled"
                                                            @click="
                                                                startInlineEdit(
                                                                    attr,
                                                                )
                                                            "
                                                        >
                                                            {{
                                                                filters[
                                                                    `min${formatAttrKey(attr)}`
                                                                ] || 0
                                                            }}
                                                        </span>
                                                        <q-input
                                                            v-else
                                                            :ref="
                                                                (el) =>
                                                                    (attributeInputRefs[
                                                                        attr
                                                                    ] = el)
                                                            "
                                                            v-model.number="
                                                                inlineEditingValue
                                                            "
                                                            type="number"
                                                            min="0"
                                                            max="20"
                                                            step="1"
                                                            dense
                                                            filled
                                                            autofocus
                                                            class="inline-attribute-input"
                                                            @keyup.enter="
                                                                finishInlineEdit
                                                            "
                                                            @blur="
                                                                finishInlineEdit
                                                            "
                                                        />
                                                    </q-item-section>
                                                </q-item>
                                            </q-list>
                                        </q-card>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </q-card-section>

                    <q-card-actions align="right" class="q-pa-md">
                        <q-btn
                            flat
                            label="Reset All"
                            color="negative"
                            @click="resetMinimumStats"
                            :disable="!hasActiveStatFilters"
                        />
                        <q-btn
                            flat
                            label="Cancel"
                            color="grey"
                            @click="showMinimumStatsModal = false"
                        />
                        <q-btn
                            unelevated
                            label="Apply"
                            color="primary"
                            @click="applyMinimumStats"
                        />
                    </q-card-actions>
                </q-card>
            </q-dialog>
        </q-card-section>
    </q-card>
</template>

<script>
import { memoizedComputed } from '@/composables/useMemoization'
import { usePlayerStore } from '@/stores/playerStore'
import { formatCurrency } from '@/utils/currencyUtils'
import { useQuasar } from 'quasar'
import { computed, defineComponent, nextTick, onMounted, ref, watch } from 'vue'

// Define attribute keys (ensure these match keys in player.attributes)
// These are the raw keys from the data.
const rawTechnicalAttributeKeys = [
  'Cor',
  'Cro',
  'Dri',
  'Fin',
  'Fir',
  'Fre',
  'Hea',
  'Lon',
  'L Th',
  'Mar',
  'Pas',
  'Pen',
  'Tck',
  'Tec'
]
const rawMentalAttributeKeys = [
  'Agg',
  'Ant',
  'Bra',
  'Cmp',
  'Cnt',
  'Dec',
  'Det',
  'Fla',
  'Ldr',
  'OtB',
  'Pos',
  'Tea',
  'Vis',
  'Wor'
]
const rawPhysicalAttributeKeys = ['Acc', 'Agi', 'Bal', 'Jum', 'Nat', 'Pac', 'Sta', 'Str']
const rawGoalkeeperAttributeKeys = [
  'Aer',
  'Cmd',
  'Com',
  'Ecc',
  'Han',
  'Kic',
  '1v1',
  'Pun',
  'Ref',
  'TRO',
  'Thr'
]

// Full names for display
const attributeFullNameMap = {
  Cor: 'Corners',
  Cro: 'Crossing',
  Dri: 'Dribbling',
  Fin: 'Finishing',
  Fir: 'First Touch',
  Fre: 'Free Kick Taking',
  Hea: 'Heading',
  Lon: 'Long Shots',
  'L Th': 'Long Throws',
  Mar: 'Marking',
  Pas: 'Passing',
  Pen: 'Penalty Taking',
  Tck: 'Tackling',
  Tec: 'Technique',
  Agg: 'Aggression',
  Ant: 'Anticipation',
  Bra: 'Bravery',
  Cmp: 'Composure',
  Cnt: 'Concentration',
  Dec: 'Decisions',
  Det: 'Determination',
  Fla: 'Flair',
  Ldr: 'Leadership',
  OtB: 'Off the Ball',
  Pos: 'Positioning',
  Tea: 'Teamwork',
  Vis: 'Vision',
  Wor: 'Work Rate',
  Acc: 'Acceleration',
  Agi: 'Agility',
  Bal: 'Balance',
  Jum: 'Jumping Reach',
  Nat: 'Natural Fitness',
  Pac: 'Pace',
  Sta: 'Stamina',
  Str: 'Strength',
  Aer: 'Aerial Reach',
  Cmd: 'Command of Area',
  Com: 'Communication',
  Ecc: 'Eccentricity',
  Han: 'Handling',
  Kic: 'Kicking',
  '1v1': 'One on Ones',
  Pun: 'Punching (Tendency)',
  Ref: 'Reflexes',
  TRO: 'Rushing Out (Tendency)',
  Thr: 'Throwing'
}

const orderedShortPositions = [
  'GK',
  'DR',
  'DC',
  'DL',
  'WBR',
  'WBL',
  'DM',
  'MR',
  'MC',
  'ML',
  'AMR',
  'AMC',
  'AML',
  'ST'
]

const AGE_SLIDER_MIN = 15
const AGE_SLIDER_MAX = 50
const _SALARY_SLIDER_MIN = 0
const SALARY_SLIDER_MAX = 1000000

function debounce(fn, delay) {
  let timeoutID = null
  return function (...args) {
    clearTimeout(timeoutID)
    timeoutID = setTimeout(() => {
      fn.apply(this, args)
    }, delay)
  }
}

export default defineComponent({
  name: 'PlayerFilters',
  props: {
    currencySymbol: { type: String, default: '$' },
    transferValueRange: {
      type: Object,
      default: () => ({ min: 0, max: 100000000 })
    },
    initialDatasetRange: {
      type: Object,
      default: () => ({ min: 0, max: 100000000 })
    },
    salaryRange: {
      type: Object,
      default: () => ({ min: 0, max: 1000000 })
    },
    uniqueClubs: { type: Array, default: () => [] },
    uniqueNationalities: { type: Array, default: () => [] },
    uniqueMediaHandlings: { type: Array, default: () => [] },
    uniquePersonalities: { type: Array, default: () => [] },
    isLoading: { type: Boolean, default: false }
  },
  emits: ['filter-changed'],
  setup(props, { emit }) {
    const quasarInstance = useQuasar()
    const playerStore = usePlayerStore()
    const filters = ref({
      name: '',
      club: null,
      position: null,
      role: null,
      nationality: null,
      mediaHandling: [],
      personality: [],
      ageRange: { min: AGE_SLIDER_MIN, max: AGE_SLIDER_MAX },
      transferValueRangeLocal: {
        min: 0,
        max: 100000000
      },
      maxSalary: SALARY_SLIDER_MAX,
      minOverall: 0,
      minPHY: 0,
      minSHO: 0,
      minPAS: 0,
      minDRI: 0,
      minDEF: 0,
      minMEN: 0,
      minGK: 0
    })

    const showMinimumStatsModal = ref(false)

    // State for inline editing
    const inlineEditingAttributeKey = ref(null) // e.g., "Cor", "Agg"
    const inlineEditingValue = ref(0)
    const attributeInputRefs = ref({}) // To store refs to q-input components

    const formatAttrName = attrKey => attributeFullNameMap[attrKey] || attrKey

    const formatAttrKey = attrKey => {
      return attrKey.replace(/\s+/g, '').replace(/\(|\)/g, '')
    }

    const getUnifiedRatingClass = (value, maxScale = 20) => {
      const numValue = Number.parseInt(value, 10)
      if (Number.isNaN(numValue) || value === null || value === undefined || value === '-')
        return 'rating-na'

      const percentage = (numValue / maxScale) * 100

      if (maxScale === 20) {
        if (numValue >= 18) return 'rating-tier-6'
        if (numValue >= 15) return 'rating-tier-5'
        if (numValue >= 13) return 'rating-tier-4'
        if (numValue >= 10) return 'rating-tier-3'
        if (numValue >= 7) return 'rating-tier-2'
        if (numValue >= 1) return 'rating-tier-1'
      } else {
        if (percentage >= 90) return 'rating-tier-6'
        if (percentage >= 80) return 'rating-tier-5'
        if (percentage >= 70) return 'rating-tier-4'
        if (percentage >= 55) return 'rating-tier-3'
        if (percentage >= 40) return 'rating-tier-2'
        if (percentage > 0) return 'rating-tier-1'
      }
      return 'rating-na'
    }

    const technicalAttributeKeys = rawTechnicalAttributeKeys
    const mentalAttributeKeys = rawMentalAttributeKeys
    const physicalAttributeKeys = rawPhysicalAttributeKeys
    const goalkeeperAttributeKeys = rawGoalkeeperAttributeKeys

    const allAttributeKeys = [
      ...technicalAttributeKeys,
      ...mentalAttributeKeys,
      ...physicalAttributeKeys,
      ...goalkeeperAttributeKeys
    ]

    for (const attr of allAttributeKeys) {
      const filterKey = `min${formatAttrKey(attr)}`
      if (!filters.value[filterKey]) {
        filters.value[filterKey] = 0
      }
    }

    const clubOptions = ref([])
    const nationalityOptions = ref([])
    const currentSliderMin = computed(() => props.transferValueRange.min)
    const currentSliderMax = computed(() => props.transferValueRange.max)
    const isDataAvailable = computed(
      () => playerStore.allPlayers && playerStore.allPlayers.length > 0
    )
    const salarySliderMin = computed(() => props.salaryRange?.min || 0)
    const salarySliderMax = computed(() => props.salaryRange?.max || SALARY_SLIDER_MAX)

    const salarySliderStep = computed(() => {
      const range = salarySliderMax.value - salarySliderMin.value
      if (range <= 0) return 1000
      if (range < 50000) return 500
      if (range < 250000) return 2500
      if (range < 1000000) return 5000
      if (range < 10000000) return 25000
      return 50000
    })

    const hasActiveStatFilters = computed(() => {
      const hasActiveFifaStats =
        filters.value.minOverall > 0 ||
        filters.value.minPHY > 0 ||
        filters.value.minSHO > 0 ||
        filters.value.minPAS > 0 ||
        filters.value.minDRI > 0 ||
        filters.value.minDEF > 0 ||
        filters.value.minMEN > 0 ||
        filters.value.minGK > 0
      const hasActiveAttributeFilters = allAttributeKeys.some(attr => {
        const filterKey = `min${formatAttrKey(attr)}`
        return filters.value[filterKey] > 0
      })
      return hasActiveFifaStats || hasActiveAttributeFilters
    })

    const hasActiveFilters = computed(() => {
      const defValMin = props.initialDatasetRange.min
      const defValMax = props.initialDatasetRange.max
      return (
        filters.value.name !== '' ||
        filters.value.club !== null ||
        filters.value.position !== null ||
        filters.value.role !== null ||
        filters.value.nationality !== null ||
        (Array.isArray(filters.value.mediaHandling) && filters.value.mediaHandling.length > 0) ||
        (Array.isArray(filters.value.personality) && filters.value.personality.length > 0) ||
        filters.value.ageRange.min !== AGE_SLIDER_MIN ||
        filters.value.ageRange.max !== AGE_SLIDER_MAX ||
        filters.value.transferValueRangeLocal.min !== defValMin ||
        filters.value.transferValueRangeLocal.max !== defValMax ||
        filters.value.maxSalary !== salarySliderMax.value ||
        hasActiveStatFilters.value
      )
    })

    const positionOptions = computed(() => {
      const options = [{ label: 'Any Position', value: null }]
      for (const shortPos of orderedShortPositions) {
        options.push({ label: shortPos, value: shortPos })
      }
      return options
    })

    const roleFilterOptions = computed(() => {
      if (
        !filters.value.position ||
        !playerStore.allAvailableRoles ||
        playerStore.allAvailableRoles.length === 0
      ) {
        return [{ label: 'Any Role', value: null }]
      }
      const selectedPosShortCode = filters.value.position
      const filtered = playerStore.allAvailableRoles
        .filter(roleFullName => roleFullName.startsWith(`${selectedPosShortCode} - `))
        .map(roleFullName => ({
          label: roleFullName,
          value: roleFullName
        }))
        .sort((a, b) => a.label.localeCompare(b.label))
      return [{ label: 'Any Role', value: null }, ...filtered]
    })

    const mediaHandlingOptions = computed(() =>
      props.uniqueMediaHandlings.map(mh => ({ label: mh, value: mh }))
    )
    const personalityOptions = computed(() =>
      props.uniquePersonalities.map(p => ({ label: p, value: p }))
    )

    const transferValueSliderStep = computed(() => {
      const range = currentSliderMax.value - currentSliderMin.value
      if (range <= 0) return 10000
      if (range < 50000) return 1000
      if (range < 250000) return 5000
      if (range < 1000000) return 10000
      if (range < 10000000) return 50000
      if (range < 50000000) return 100000
      return 250000
    })

    const formatRangeLabel = (value, isMaxBoundary = false) => {
      if (value === null || value === undefined) return 'N/A'
      if (isMaxBoundary) {
        if (
          props.initialDatasetRange &&
          typeof props.initialDatasetRange.max === 'number' &&
          value === props.initialDatasetRange.max
        ) {
          return 'Any'
        }
      } else {
        if (
          props.initialDatasetRange &&
          typeof props.initialDatasetRange.min === 'number' &&
          value === props.initialDatasetRange.min
        ) {
          return formatCurrency(value, props.currencySymbol) || '0'
        }
      }
      return formatCurrency(value, props.currencySymbol)
    }

    watch(
      () => props.uniqueClubs,
      newClubs => {
        clubOptions.value = newClubs
      },
      { immediate: true }
    )
    watch(
      () => props.uniqueNationalities,
      newNats => {
        nationalityOptions.value = newNats
      },
      { immediate: true }
    )
    watch(
      () => props.transferValueRange,
      newDynamicRange => {
        if (
          newDynamicRange &&
          typeof newDynamicRange.min === 'number' &&
          typeof newDynamicRange.max === 'number'
        ) {
          // Only update if we don't have valid values yet or if the new range is different
          const currentMin = filters.value.transferValueRangeLocal.min
          const currentMax = filters.value.transferValueRangeLocal.max

          if (currentMin === 0 && currentMax === 100000000) {
            // We still have default values, so update with real data
            filters.value.transferValueRangeLocal = {
              min: newDynamicRange.min,
              max: newDynamicRange.max
            }
          } else {
            // Clamp existing values to new valid range
            let _changed = false
            if (currentMin < newDynamicRange.min) {
              filters.value.transferValueRangeLocal.min = newDynamicRange.min
              _changed = true
            }
            if (currentMax > newDynamicRange.max) {
              filters.value.transferValueRangeLocal.max = newDynamicRange.max
              _changed = true
            }
          }
        }
      },
      { deep: true, immediate: true }
    )
    watch(
      () => props.initialDatasetRange,
      newInitialRange => {
        if (
          newInitialRange &&
          typeof newInitialRange.min === 'number' &&
          typeof newInitialRange.max === 'number'
        ) {
          // Only update if we still have default values
          const currentMin = filters.value.transferValueRangeLocal.min
          const currentMax = filters.value.transferValueRangeLocal.max

          if (currentMin === 0 && currentMax === 100000000) {
            filters.value.transferValueRangeLocal = {
              min: newInitialRange.min,
              max: newInitialRange.max
            }
          }
        }
      },
      { deep: true, immediate: true }
    )
    watch(
      () => props.salaryRange,
      newSalaryRange => {
        if (newSalaryRange?.max && typeof newSalaryRange.max === 'number') {
          // Only update if we still have the default value
          if (filters.value.maxSalary === SALARY_SLIDER_MAX) {
            filters.value.maxSalary = newSalaryRange.max
          }
        }
      },
      { deep: true, immediate: true }
    )

    const applyFilters = () => {
      if (props.isLoading) return
      const clampedMin = Math.max(filters.value.transferValueRangeLocal.min, currentSliderMin.value)
      const clampedMax = Math.min(filters.value.transferValueRangeLocal.max, currentSliderMax.value)
      emit('filter-changed', {
        ...filters.value,
        transferValueRangeLocal: { min: clampedMin, max: clampedMax }
      })
    }
    const debouncedApplyFilters = debounce(applyFilters, 400)

    const onPositionChange = () => {
      filters.value.role = null
      applyFilters()
    }

    const clearAllFilters = () => {
      filters.value = {
        name: '',
        club: null,
        position: null,
        role: null,
        nationality: null,
        mediaHandling: [],
        personality: [],
        ageRange: { min: AGE_SLIDER_MIN, max: AGE_SLIDER_MAX },
        transferValueRangeLocal: {
          min: props.initialDatasetRange ? props.initialDatasetRange.min : 0,
          max: props.initialDatasetRange ? props.initialDatasetRange.max : 100000000
        },
        maxSalary: salarySliderMax.value,
        minOverall: 0,
        minPHY: 0,
        minSHO: 0,
        minPAS: 0,
        minDRI: 0,
        minDEF: 0,
        minMEN: 0,
        minGK: 0
      }
      for (const attr of allAttributeKeys) {
        filters.value[`min${formatAttrKey(attr)}`] = 0
      }
      applyFilters()
    }

    const filterClubOptions = (val, update, abort) => {
      if (val.length < 1 && val !== '') {
        abort()
        return
      }
      update(() => {
        const needle = val.toLowerCase()
        clubOptions.value = props.uniqueClubs.filter(v => v.toLowerCase().indexOf(needle) > -1)
      })
    }
    const filterNationalityOptions = (val, update, abort) => {
      if (val.length < 1 && val !== '') {
        abort()
        return
      }
      update(() => {
        const needle = val.toLowerCase()
        nationalityOptions.value = props.uniqueNationalities.filter(
          v => v.toLowerCase().indexOf(needle) > -1
        )
      })
    }

    onMounted(async () => {
      if (playerStore.allAvailableRoles.length === 0 && playerStore.currentDatasetId) {
        await playerStore.fetchAllAvailableRoles()
      }

      // Initialize transfer value range from the correct prop
      if (
        props.initialDatasetRange &&
        typeof props.initialDatasetRange.min === 'number' &&
        typeof props.initialDatasetRange.max === 'number'
      ) {
        filters.value.transferValueRangeLocal = {
          min: props.initialDatasetRange.min,
          max: props.initialDatasetRange.max
        }
      } else if (
        props.transferValueRange &&
        typeof props.transferValueRange.min === 'number' &&
        typeof props.transferValueRange.max === 'number'
      ) {
        filters.value.transferValueRangeLocal = {
          min: props.transferValueRange.min,
          max: props.transferValueRange.max
        }
      }

      // Initialize max salary from salary range prop
      if (props.salaryRange?.max && typeof props.salaryRange.max === 'number') {
        filters.value.maxSalary = props.salaryRange.max
      }

      filters.value.ageRange = {
        min: AGE_SLIDER_MIN,
        max: AGE_SLIDER_MAX
      }
    })

    watch(
      () => playerStore.currentDatasetId,
      async newId => {
        if (newId && playerStore.allAvailableRoles.length === 0) {
          await playerStore.fetchAllAvailableRoles()
        }
        if (newId && props.initialDatasetRange) {
          filters.value.transferValueRangeLocal = {
            min: props.initialDatasetRange.min,
            max: props.initialDatasetRange.max
          }
        }
      }
    )

    const resetMinimumStats = () => {
      filters.value.minOverall = 0
      filters.value.minPHY = 0
      filters.value.minSHO = 0
      filters.value.minPAS = 0
      filters.value.minDRI = 0
      filters.value.minDEF = 0
      filters.value.minMEN = 0
      filters.value.minGK = 0
      for (const attr of allAttributeKeys) {
        filters.value[`min${formatAttrKey(attr)}`] = 0
      }
    }

    const applyMinimumStats = () => {
      finishInlineEdit() // Ensure any pending inline edit is saved
      showMinimumStatsModal.value = false
      applyFilters()
    }

    const startInlineEdit = attrKey => {
      finishInlineEdit() // Save any previous edit first
      inlineEditingAttributeKey.value = attrKey
      const filterKey = `min${formatAttrKey(attrKey)}`
      inlineEditingValue.value = filters.value[filterKey] || 0
      nextTick(() => {
        const inputEl = attributeInputRefs.value[attrKey]
        if (inputEl?.focus) {
          setTimeout(() => inputEl.focus(), 0) // Timeout to ensure focus works after render
        }
      })
    }

    const finishInlineEdit = () => {
      if (inlineEditingAttributeKey.value) {
        const attrKey = inlineEditingAttributeKey.value
        const filterKey = `min${formatAttrKey(attrKey)}`
        let val = Number.parseInt(inlineEditingValue.value, 10)
        if (Number.isNaN(val) || val < 0) val = 0
        if (val > 20) val = 20
        filters.value[filterKey] = val
        inlineEditingAttributeKey.value = null
      }
    }

    return {
      quasarInstance,
      filters,
      hasActiveFilters,
      hasActiveStatFilters,
      showMinimumStatsModal,
      inlineEditingAttributeKey,
      inlineEditingValue,
      attributeInputRefs, // For inline editing
      getUnifiedRatingClass,
      formatAttrName,
      formatAttrKey,
      technicalAttributeKeys,
      mentalAttributeKeys,
      physicalAttributeKeys,
      goalkeeperAttributeKeys,
      resetMinimumStats,
      applyMinimumStats,
      startInlineEdit,
      finishInlineEdit, // Inline editing methods
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
      currentSliderMin,
      currentSliderMax,
      salarySliderMin,
      salarySliderMax,
      salarySliderStep,
      formatCurrency
    }
  }
})
</script>

<style lang="scss" scoped>
// SCSS variables from quasar-variables.scss are available here
// e.g. $primary, $grey-1, etc.

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
    margin-bottom: 0px;
    line-height: 1.2;
}

.stat-value-badge {
    display: inline-block;
    padding: 3px 8px;
    border-radius: 4px;
    font-size: 0.8rem;
    font-weight: 700;
    min-width: 30px;
    text-align: center;
    color: white;
    line-height: 1;
    vertical-align: middle;
}

.modal-content {
    max-height: 80vh;
    overflow-y: auto;
}

.attribute-group-title {
    color: $primary;
    border-bottom: 1px solid $primary;
    padding-bottom: 4px;
    margin-bottom: 12px;
    font-weight: 500;

    .body--dark & {
        color: lighten($primary, 25%);
        border-bottom-color: lighten($primary, 25%);
    }
}

.fifa-stat-item {
    width: 100%;
}

.attributes-section {
    padding: 0;
}
.attributes-section-title {
    font-weight: 600;
    margin-bottom: 12px;
}

.attribute-category-card {
    margin-bottom: 16px;
    height: 100%;
    display: flex;
    flex-direction: column;
    background-color: transparent !important;
    border: 1px solid rgba(0, 0, 0, 0.08);
    .body--dark & {
        border: 1px solid rgba(255, 255, 255, 0.12);
    }
}

.attribute-category-header-styled {
    background-color: rgba(0, 0, 0, 0.03);
    padding: 8px 12px;
    border-bottom: 1px solid rgba(0, 0, 0, 0.08);
    text-align: center;

    .body--dark & {
        background-color: rgba(255, 255, 255, 0.05);
        border-bottom: 1px solid rgba(255, 255, 255, 0.12);
    }
    .text-subtitle1 {
        font-size: 0.9rem;
        font-weight: 600;
    }
}

.attribute-list-column-styled {
    padding: 8px 4px;
    flex-grow: 1;
    overflow-y: auto;
}

.attribute-item-column-styled {
    padding: 4px 8px;
    min-height: 32px;
    display: flex;
    justify-content: space-between;
    align-items: center;

    &.q-item--dark {
    }
    &:not(:last-child):after {
        background-color: rgba(0, 0, 0, 0.08) !important;
        .body--dark & {
            background-color: rgba(255, 255, 255, 0.1) !important;
        }
    }
}

.attribute-name-column-styled {
    font-size: 0.75rem;
    color: $grey-8;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    padding-right: 6px;

    .body--dark & {
        color: $grey-4;
    }
}

.attribute-value-column-styled {
    // Styles for the container of the badge/input
    min-width: 40px; // Ensure enough space for the input
    display: flex;
    justify-content: flex-end;
}

.attribute-badge-styled {
    cursor: pointer;
    padding: 3px 7px;
    border-radius: 4px;
    transition:
        background-color 0.2s ease,
        box-shadow 0.2s ease;
    display: inline-block;
    min-width: 28px;
    text-align: center;
    font-weight: 600;
    line-height: 1.2;
    font-size: 0.8rem;
    color: white;

    &:hover {
        box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
        .body--dark & {
            box-shadow: 0 1px 3px rgba(255, 255, 255, 0.2);
        }
    }
}

.inline-attribute-input {
    width: 55px; // Adjust width as needed
    :deep(.q-field__control) {
        height: 28px; // Make it compact
        padding: 0 4px;
        min-height: 28px !important;
    }
    :deep(.q-field__native) {
        font-size: 0.8rem;
        text-align: center;
        font-weight: 600;
        padding: 0;
    }
    // Remove spinner buttons for number input for a cleaner look
    :deep(input[type="number"]::-webkit-inner-spin-button),
    :deep(input[type="number"]::-webkit-outer-spin-button) {
        -webkit-appearance: none;
        margin: 0;
    }
    :deep(input[type="number"]) {
        -moz-appearance: textfield; // Firefox
    }
}

// Global rating tier classes from app.scss (ensure these are consistent)
.rating-tier-6 {
    background-color: #7e57c2;
    color: white !important;
    .body--dark & {
        background-color: #9575cd;
    }
}
.rating-tier-5 {
    background-color: #26a69a;
    color: white !important;
    .body--dark & {
        background-color: #00897b;
    }
}
.rating-tier-4 {
    background-color: #66bb6a;
    color: white !important;
    .body--dark & {
        background-color: #4caf50;
    }
}
.rating-tier-3 {
    background-color: #42a5f5;
    color: white !important;
    .body--dark & {
        background-color: #2196f3;
    }
}
.rating-tier-2 {
    background-color: #ffa726;
    color: #333333 !important;
    .body--dark & {
        background-color: #fb8c00;
        color: white !important;
    }
}
.rating-tier-1 {
    background-color: #ef5350;
    color: white !important;
    .body--dark & {
        background-color: #e53935;
    }
}
.rating-na {
    background-color: #bdbdbd;
    color: #424242 !important;
    .body--dark & {
        background-color: #424242;
        color: #bdbdbd !important;
    }
}

.rating-tier-1,
.rating-tier-2,
.rating-tier-3,
.rating-tier-4,
.rating-tier-5,
.rating-tier-6,
.rating-na {
    border: none !important;
}
</style>
