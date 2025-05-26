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

                <div class="col-12 col-sm-6 col-md-4 col-lg-3 filter-item-container">
                    <div
                        class="text-caption q-mb-xs slider-label"
                        :class="
                            quasarInstance.dark.isActive
                                ? 'text-grey-4'
                                : 'text-grey-7'
                        "
                    >
                        Max Salary: {{
                            filters.maxSalary === salarySliderMax
                                ? "Any"
                                : formatCurrency(filters.maxSalary, currencySymbol)
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
                                : formatCurrency(filters.maxSalary, currencySymbol)
                        "
                        @update:model-value="debouncedApplyFilters"
                        color="primary"
                        class="q-px-sm"
                        :disable="isLoading || !isDataAvailable"
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

            <!-- Set Minimum Stats Button -->
            <div class="row q-col-gutter-x-md q-col-gutter-y-sm items-start q-mt-sm">
                <div
                    class="col-12 col-md-4 col-lg-2 filter-item-container self-end"
                >
                    <q-btn
                        color="primary"
                        :label="'Set Minimum Stats' + (hasActiveStatFilters ? ' (Active)' : '')"
                        class="full-width"
                        @click="showMinimumStatsModal = true"
                        :disable="isLoading"
                        outline
                        dense
                        icon="tune"
                    />
                </div>
            </div>

            <!-- Minimum Stats Modal -->
            <q-dialog v-model="showMinimumStatsModal" persistent maximized>
                <q-card
                    :class="quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-white'"
                >
                    <q-card-section>
                        <div class="text-h6">Set Minimum Stats</div>
                        <div class="text-subtitle2 text-grey-6 q-mt-xs">
                            Filter players by minimum stat values
                        </div>
                    </q-card-section>

                    <q-card-section class="q-pt-none modal-content">
                        <div class="row q-col-gutter-lg">
                            <!-- Left Column: FIFA-Style Stats -->
                            <div class="col-12 col-md-3">
                                <div class="attribute-group">
                                    <div class="text-h6 q-mb-sm attribute-group-title">
                                        FIFA-Style Stats
                                    </div>
                                    
                                    <!-- Overall Slider -->
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
                                                :class="getStatColorClass(filters.minOverall)"
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

                                    <!-- PHY Slider -->
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
                                                :class="getStatColorClass(filters.minPHY)"
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

                                    <!-- SHO Slider -->
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
                                                :class="getStatColorClass(filters.minSHO)"
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

                                    <!-- PAS Slider -->
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
                                                :class="getStatColorClass(filters.minPAS)"
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

                                    <!-- DRI Slider -->
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
                                                :class="getStatColorClass(filters.minDRI)"
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

                                    <!-- DEF Slider -->
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
                                                :class="getStatColorClass(filters.minDEF)"
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

                                    <!-- MEN Slider -->
                                    <div class="fifa-stat-item q-mb-md">
                                        <div
                                            class="text-caption q-mb-xs slider-label"
                                            :class="
                                                quasarInstance.dark.isActive
                                                    ? 'text-grey-4'
                                                    : 'text-grey-7'
                                            "
                                        >
                                            Min MEN:
                                            <span 
                                                class="stat-value-badge q-ml-xs"
                                                :class="getStatColorClass(filters.minMEN)"
                                            >
                                                {{ filters.minMEN || 0 }}
                                            </span>
                                        </div>
                                        <q-slider
                                            v-model="filters.minMEN"
                                            :min="0"
                                            :max="99"
                                            :step="1"
                                            color="primary"
                                            class="q-px-sm"
                                        />
                                    </div>

                                    <!-- GK Slider -->
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
                                                :class="getStatColorClass(filters.minGK)"
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

                            <!-- Right Column: FM Attributes -->
                            <div class="col-12 col-md-9">
                                <q-card flat bordered :class="quasarInstance.dark.isActive ? 'bg-grey-9' : 'bg-blue-grey-1'">
                                    <q-card-section class="attributes-section">
                                        <div class="text-subtitle1 q-mb-xs attributes-section-title">
                                            FM Attributes
                                        </div>
                                        
                                        <!-- Technical Attributes -->
                                        <div class="attribute-group q-mb-md">
                                            <div class="text-caption text-bold q-mb-xs attribute-group-title">
                                                Technical
                                            </div>
                                            <div class="row attribute-list">
                                                <template v-for="attr in technicalAttributeKeys" :key="attr">
                                                    <div class="col-6 col-sm-4 col-md-3 attribute-item">
                                                        <div class="row no-wrap items-baseline q-gutter-x-xs">
                                                            <div class="col-auto attribute-name">{{ formatAttrName(attr) }}:</div>
                                                            <div class="col-auto text-weight-bold attribute-value">
                                                                <span 
                                                                    :class="getAttributeColorClass(filters[`min${formatAttrKey(attr)}`])"
                                                                    class="clickable-attribute"
                                                                    @click="openAttributeEditor(attr)"
                                                                >
                                                                    {{ filters[`min${formatAttrKey(attr)}`] || 0 }}
                                                                </span>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </template>
                                            </div>
                                        </div>
                                        
                                        <!-- Mental Attributes -->
                                        <div class="attribute-group q-mb-md">
                                            <div class="text-caption text-bold q-mb-xs attribute-group-title">
                                                Mental
                                            </div>
                                            <div class="row attribute-list">
                                                <template v-for="attr in mentalAttributeKeys" :key="attr">
                                                    <div class="col-6 col-sm-4 col-md-3 attribute-item">
                                                        <div class="row no-wrap items-baseline q-gutter-x-xs">
                                                            <div class="col-auto attribute-name">{{ formatAttrName(attr) }}:</div>
                                                            <div class="col-auto text-weight-bold attribute-value">
                                                                <span 
                                                                    :class="getAttributeColorClass(filters[`min${formatAttrKey(attr)}`])"
                                                                    class="clickable-attribute"
                                                                    @click="openAttributeEditor(attr)"
                                                                >
                                                                    {{ filters[`min${formatAttrKey(attr)}`] || 0 }}
                                                                </span>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </template>
                                            </div>
                                        </div>
                                        
                                        <!-- Physical Attributes -->
                                        <div class="attribute-group q-mb-md">
                                            <div class="text-caption text-bold q-mb-xs attribute-group-title">
                                                Physical
                                            </div>
                                            <div class="row attribute-list">
                                                <template v-for="attr in physicalAttributeKeys" :key="attr">
                                                    <div class="col-6 col-sm-4 col-md-3 attribute-item">
                                                        <div class="row no-wrap items-baseline q-gutter-x-xs">
                                                            <div class="col-auto attribute-name">{{ formatAttrName(attr) }}:</div>
                                                            <div class="col-auto text-weight-bold attribute-value">
                                                                <span 
                                                                    :class="getAttributeColorClass(filters[`min${formatAttrKey(attr)}`])"
                                                                    class="clickable-attribute"
                                                                    @click="openAttributeEditor(attr)"
                                                                >
                                                                    {{ filters[`min${formatAttrKey(attr)}`] || 0 }}
                                                                </span>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </template>
                                            </div>
                                        </div>
                                        
                                        <!-- Goalkeeping Attributes -->
                                        <div class="attribute-group">
                                            <div class="text-caption text-bold q-mb-xs attribute-group-title">
                                                Goalkeeping
                                            </div>
                                            <div class="row attribute-list">
                                                <template v-for="attr in goalkeeperAttributeKeys" :key="attr">
                                                    <div class="col-6 col-sm-4 col-md-3 attribute-item">
                                                        <div class="row no-wrap items-baseline q-gutter-x-xs">
                                                            <div class="col-auto attribute-name">{{ formatAttrName(attr) }}:</div>
                                                            <div class="col-auto text-weight-bold attribute-value">
                                                                <span 
                                                                    :class="getAttributeColorClass(filters[`min${formatAttrKey(attr)}`])"
                                                                    class="clickable-attribute"
                                                                    @click="openAttributeEditor(attr)"
                                                                >
                                                                    {{ filters[`min${formatAttrKey(attr)}`] || 0 }}
                                                                </span>
                                                            </div>
                                                        </div>
                                                    </div>
                                                </template>
                                            </div>
                                        </div>
                                    </q-card-section>
                                </q-card>
                            </div>
                        </div>

                        <!-- Attribute Editor Dialog -->
                        <q-dialog v-model="showAttributeEditor" persistent>
                            <q-card>
                                <q-card-section>
                                    <div class="text-h6">Set Minimum {{ formatAttrName(editingAttribute) }}</div>
                                </q-card-section>

                                <q-card-section class="q-pt-none">
                                    <q-input
                                        v-model.number="editingValue"
                                        type="number"
                                        :min="0"
                                        :max="20"
                                        label="Minimum value (0-20)"
                                        filled
                                        autofocus
                                        @keyup.enter="saveAttributeValue"
                                    />
                                </q-card-section>

                                <q-card-actions align="right">
                                    <q-btn flat label="Cancel" color="grey" @click="cancelAttributeEdit" />
                                    <q-btn flat label="Clear" color="negative" @click="clearAttributeValue" />
                                    <q-btn unelevated label="Save" color="primary" @click="saveAttributeValue" />
                                </q-card-actions>
                            </q-card>
                        </q-dialog>
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
import { ref, computed, watch, defineComponent, onMounted } from "vue";
import { useQuasar } from "quasar";
import { usePlayerStore } from "@/stores/playerStore";
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
const SALARY_SLIDER_MIN = 0;
const SALARY_SLIDER_MAX = 1000000; // Default max, will be updated based on data

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
            // Range of the *currently filtered* data in playerStore
            type: Object,
            default: () => ({ min: 0, max: 100000000 }),
        },
        initialDatasetRange: {
            // New prop: True global range of the *entire dataset*
            type: Object,
            default: () => ({ min: 0, max: 100000000 }),
        },
        salaryRange: {
            // Range of salary values in the dataset
            type: Object,
            default: () => ({ min: 0, max: 1000000 }),
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
                // This will hold the slider's current values
                min: props.transferValueRange.min,
                max: props.transferValueRange.max,
            },
            maxSalary: SALARY_SLIDER_MAX,
            // FIFA-style stat minimum filters
            minOverall: 0,
            minPHY: 0,
            minSHO: 0,
            minPAS: 0,
            minDRI: 0,
            minDEF: 0,
            minMEN: 0,
            minGK: 0,
        });

        const showMinimumStatsModal = ref(false);
        const showAttributeEditor = ref(false);
        const editingAttribute = ref('');
        const editingValue = ref(0);

        // Helper functions for attributes
        const formatAttrName = (attr) => {
            return attr
                .replace(/_/g, ' ')
                .split(' ')
                .map(word => word.charAt(0).toUpperCase() + word.slice(1))
                .join(' ');
        };

        const formatAttrKey = (attr) => {
            return attr
                .split('_')
                .map((word, index) => index === 0 ? word.charAt(0).toUpperCase() + word.slice(1) : word.charAt(0).toUpperCase() + word.slice(1))
                .join('');
        };

        const getAttributeColorClass = (value) => {
            const numValue = parseInt(value, 10) || 0;
            if (numValue >= 15) return 'text-green-10';
            if (numValue >= 13) return 'text-green-8';
            if (numValue >= 10) return 'text-amber-8';
            if (numValue >= 7) return 'text-orange';
            if (numValue > 0) return 'text-red';
            return 'rating-na'; // Grey - N/A
        };

        // Attribute keys for FM attributes (same as PlayerAttributesSection)
        const technicalAttributeKeys = [
            'crossing', 'dribbling', 'finishing', 'first_touch', 'free_kick_taking',
            'heading', 'long_shots', 'long_throws', 'marking', 'passing', 'penalty_taking',
            'tackling', 'technique', 'corners'
        ];

        const mentalAttributeKeys = [
            'aggression', 'anticipation', 'bravery', 'composure', 'concentration',
            'decisions', 'determination', 'flair', 'leadership', 'off_the_ball',
            'positioning', 'teamwork', 'vision', 'work_rate'
        ];

        const physicalAttributeKeys = [
            'acceleration', 'agility', 'balance', 'jumping_reach', 'natural_fitness',
            'pace', 'stamina', 'strength'
        ];

        const goalkeeperAttributeKeys = [
            'aerial_reach', 'command_of_area', 'communication', 'eccentricity',
            'handling', 'kicking', 'one_on_ones', 'punching', 'reflexes',
            'rushing_out', 'tendency_to_punch', 'throwing'
        ];

        // Add all attribute filters to the filters object
        const allAttributeKeys = [
            ...technicalAttributeKeys,
            ...mentalAttributeKeys,
            ...physicalAttributeKeys,
            ...goalkeeperAttributeKeys
        ];

        // Initialize all attribute filters
        allAttributeKeys.forEach(attr => {
            const filterKey = `min${formatAttrKey(attr)}`;
            if (!filters.value[filterKey]) {
                filters.value[filterKey] = 0;
            }
        });

        const clubOptions = ref([]);
        const nationalityOptions = ref([]);

        // These computed properties define the slider's operational min/max.
        // They should react to props.transferValueRange (the range of currently filtered data).
        const currentSliderMin = computed(() => props.transferValueRange.min);
        const currentSliderMax = computed(() => props.transferValueRange.max);

        const isDataAvailable = computed(
            () => playerStore.allPlayers && playerStore.allPlayers.length > 0,
        );

        const salarySliderMin = computed(() => props.salaryRange?.min || 0);
        const salarySliderMax = computed(() => props.salaryRange?.max || 1000000);

        const salarySliderStep = computed(() => {
            const range = salarySliderMax.value - salarySliderMin.value;
            if (range <= 0) return 1000;
            if (range < 50000) return 500;
            if (range < 250000) return 2500;
            if (range < 1000000) return 5000;
            if (range < 10000000) return 25000;
            return 50000;
        });

        const hasActiveFilters = computed(() => {
            // Use props.initialDatasetRange for transfer value default comparison
            const defValMin = props.initialDatasetRange.min;
            const defValMax = props.initialDatasetRange.max;

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
                filters.value.ageRange.min !== AGE_SLIDER_MIN ||
                filters.value.ageRange.max !== AGE_SLIDER_MAX ||
                filters.value.transferValueRangeLocal.min !== defValMin ||
                filters.value.transferValueRangeLocal.max !== defValMax ||
                filters.value.maxSalary !== salarySliderMax.value ||
                hasActiveStatFilters.value
            );
        });

        const hasActiveStatFilters = computed(() => {
            const hasActiveFifaStats = (
                filters.value.minOverall > 0 ||
                filters.value.minPHY > 0 ||
                filters.value.minSHO > 0 ||
                filters.value.minPAS > 0 ||
                filters.value.minDRI > 0 ||
                filters.value.minDEF > 0 ||
                filters.value.minMEN > 0 ||
                filters.value.minGK > 0
            );
            
            const hasActiveAttributeFilters = allAttributeKeys.some(attr => {
                const filterKey = `min${formatAttrKey(attr)}`;
                return filters.value[filterKey] > 0;
            });
            
            return hasActiveFifaStats || hasActiveAttributeFilters;
        });

        const getStatColorClass = (value) => {
            const numValue = parseInt(value, 10) || 0;
            if (numValue >= 90) return 'rating-tier-6'; // Purple - Elite
            if (numValue >= 80) return 'rating-tier-5'; // Teal - Excellent
            if (numValue >= 70) return 'rating-tier-4'; // Green - Good
            if (numValue >= 55) return 'rating-tier-3'; // Light Blue - Average
            if (numValue >= 40) return 'rating-tier-2'; // Orange - Below Average
            if (numValue > 0) return 'rating-tier-1'; // Red - Poor
            return 'rating-na'; // Grey - N/A
        };

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
            // Step calculation should be based on the slider's current operational range
            const range = currentSliderMax.value - currentSliderMin.value;
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
            // "Any" logic now uses the static initialDatasetRange from props
            if (isMaxBoundary) {
                if (
                    props.initialDatasetRange &&
                    typeof props.initialDatasetRange.max === "number" &&
                    value === props.initialDatasetRange.max
                ) {
                    return "Any";
                }
            } else {
                // Min boundary
                if (
                    props.initialDatasetRange &&
                    typeof props.initialDatasetRange.min === "number" &&
                    value === props.initialDatasetRange.min
                ) {
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

        // Watch the dynamic transferValueRange prop to update the local slider values
        // if they fall outside the new dynamic range from the parent.
        watch(
            () => props.transferValueRange,
            (newDynamicRange) => {
                if (
                    newDynamicRange &&
                    typeof newDynamicRange.min === "number" &&
                    typeof newDynamicRange.max === "number"
                ) {
                    // Update local slider values only if they are outside the new dynamic range
                    // or if they were uninitialized (null).
                    let changed = false;
                    if (
                        filters.value.transferValueRangeLocal.min === null ||
                        filters.value.transferValueRangeLocal.min <
                            newDynamicRange.min
                    ) {
                        filters.value.transferValueRangeLocal.min =
                            newDynamicRange.min;
                        changed = true;
                    }
                    if (
                        filters.value.transferValueRangeLocal.max === null ||
                        filters.value.transferValueRangeLocal.max >
                            newDynamicRange.max
                    ) {
                        filters.value.transferValueRangeLocal.max =
                            newDynamicRange.max;
                        changed = true;
                    }
                    // If values were clamped, emit filter change
                    if (changed) {
                        // applyFilters(); // Or debouncedApplyFilters if preferred, but direct might be better for clamping
                    }
                }
            },
            { deep: true, immediate: true },
        );
        // Also watch initialDatasetRange to set the initial state of transferValueRangeLocal correctly
        watch(
            () => props.initialDatasetRange,
            (newInitialRange) => {
                if (
                    newInitialRange &&
                    typeof newInitialRange.min === "number" &&
                    typeof newInitialRange.max === "number"
                ) {
                    // Set the initial local filter range to match the full dataset range
                    filters.value.transferValueRangeLocal = {
                        min: newInitialRange.min,
                        max: newInitialRange.max,
                    };
                }
            },
            { deep: true, immediate: true },
        );

        const applyFilters = () => {
            if (props.isLoading) return;
            // Ensure the emitted filter range is clamped by the current slider's operational min/max
            // This should ideally not be necessary if v-model works correctly with q-range's :min and :max
            const clampedMin = Math.max(
                filters.value.transferValueRangeLocal.min,
                currentSliderMin.value,
            );
            const clampedMax = Math.min(
                filters.value.transferValueRangeLocal.max,
                currentSliderMax.value,
            );

            emit("filter-changed", {
                ...filters.value,
                transferValueRangeLocal: { min: clampedMin, max: clampedMax },
            });
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
                    // Reset to the true initial dataset range
                    min: props.initialDatasetRange
                        ? props.initialDatasetRange.min
                        : 0,
                    max: props.initialDatasetRange
                        ? props.initialDatasetRange.max
                        : 100000000,
                },
                maxSalary: salarySliderMax.value,
                // Reset FIFA-style stat minimum filters
                minOverall: 0,
                minPHY: 0,
                minSHO: 0,
                minPAS: 0,
                minDRI: 0,
                minDEF: 0,
                minMEN: 0,
                minGK: 0,
            };
            
            // Reset all attribute filters
            allAttributeKeys.forEach(attr => {
                const filterKey = `min${formatAttrKey(attr)}`;
                filters.value[filterKey] = 0;
            });
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
            // Set initial values from props if they are valid
            if (props.initialDatasetRange) {
                filters.value.transferValueRangeLocal = {
                    min: props.initialDatasetRange.min,
                    max: props.initialDatasetRange.max,
                };
            }
            if (props.salaryRange?.max) {
                filters.value.maxSalary = props.salaryRange.max;
            } else {
                filters.value.maxSalary = SALARY_SLIDER_MAX;
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
                // When dataset changes, reset filters, including transferValueRangeLocal to the new initial range
                if (newId && props.initialDatasetRange) {
                    filters.value.transferValueRangeLocal = {
                        min: props.initialDatasetRange.min,
                        max: props.initialDatasetRange.max,
                    };
                }
            },
        );

        const resetMinimumStats = () => {
            // Reset FIFA-style stats
            filters.value.minOverall = 0;
            filters.value.minPHY = 0;
            filters.value.minSHO = 0;
            filters.value.minPAS = 0;
            filters.value.minDRI = 0;
            filters.value.minDEF = 0;
            filters.value.minMEN = 0;
            filters.value.minGK = 0;
            
            // Reset all attribute filters
            allAttributeKeys.forEach(attr => {
                const filterKey = `min${formatAttrKey(attr)}`;
                filters.value[filterKey] = 0;
            });
        };

        const applyMinimumStats = () => {
            showMinimumStatsModal.value = false;
            applyFilters();
        };

        const openAttributeEditor = (attr) => {
            editingAttribute.value = attr;
            const filterKey = `min${formatAttrKey(attr)}`;
            editingValue.value = filters.value[filterKey] || 0;
            showAttributeEditor.value = true;
        };

        const saveAttributeValue = () => {
            const filterKey = `min${formatAttrKey(editingAttribute.value)}`;
            filters.value[filterKey] = Math.max(0, Math.min(20, editingValue.value || 0));
            showAttributeEditor.value = false;
        };

        const clearAttributeValue = () => {
            const filterKey = `min${formatAttrKey(editingAttribute.value)}`;
            filters.value[filterKey] = 0;
            showAttributeEditor.value = false;
        };

        const cancelAttributeEdit = () => {
            showAttributeEditor.value = false;
        };

        return {
            quasarInstance,
            filters,
            hasActiveFilters,
            hasActiveStatFilters,
            showMinimumStatsModal,
            showAttributeEditor,
            editingAttribute,
            editingValue,
            getStatColorClass,
            getAttributeColorClass,
            formatAttrName,
            formatAttrKey,
            technicalAttributeKeys,
            mentalAttributeKeys,
            physicalAttributeKeys,
            goalkeeperAttributeKeys,
            resetMinimumStats,
            applyMinimumStats,
            openAttributeEditor,
            saveAttributeValue,
            clearAttributeValue,
            cancelAttributeEdit,
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
            currentSliderMin, // For q-range :min
            currentSliderMax, // For q-range :max
            salarySliderMin,
            salarySliderMax,
            salarySliderStep,
            formatCurrency,
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
    margin-bottom: 0px;
    line-height: 1.2;
}
.filter-item-container {
    // Ensure consistent vertical alignment if items wrap
}
.stat-value-badge {
    display: inline-block;
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 600;
    min-width: 24px;
    text-align: center;
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
    
    .body--dark & {
        color: lighten($primary, 20%);
        border-bottom-color: lighten($primary, 20%);
    }
}

.fifa-stat-item {
    width: 100%;
}

.clickable-attribute {
    cursor: pointer;
    padding: 2px 4px;
    border-radius: 3px;
    transition: background-color 0.2s ease;
    
    &:hover {
        background-color: rgba(0, 0, 0, 0.1);
        
        .body--dark & {
            background-color: rgba(255, 255, 255, 0.1);
        }
    }
}

.attributes-section {
    padding: 12px;
}

.attributes-section-title {
    font-weight: 600;
}

.attribute-item {
    margin-bottom: 6px;
    padding-right: 8px;
}

.attribute-name {
    font-size: 0.85rem;
    color: rgba(0, 0, 0, 0.7);
    
    .body--dark & {
        color: rgba(255, 255, 255, 0.7);
    }
}

.attribute-value {
    font-size: 0.9rem;
}
</style>
