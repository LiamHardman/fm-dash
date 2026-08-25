<template>
    <q-page class="docs-page">
        <div class="page-container">
            <PageHeader
                title="Documentation"
                subtitle="Everything you need to master Football Manager data analysis, from getting started to advanced team management strategies."
                icon="menu_book"
            >
                <template #actions>
                    <q-btn
                        unelevated
                        color="primary"
                        label="Get Started"
                        icon="play_arrow"
                        @click="scrollToSection('getting-started')"
                    />
                    <q-btn
                        outline
                        color="primary"
                        label="API Reference"
                        icon="code"
                        @click="scrollToSection('api-reference')"
                    />
                    <q-btn
                        outline
                        color="primary"
                        label="First-Time Guide"
                        icon="school"
                        @click="showTutorial"
                    />
                </template>
            </PageHeader>

            <div class="docs-highlights">
                <div class="highlight-chip" v-for="feature in heroFeatures" :key="feature.id">
                    <q-icon :name="feature.icon" size="1.1rem" />
                    <span>{{ feature.title }}</span>
                </div>
            </div>

            <div class="docs-layout">
                <!-- Sticky table of contents (desktop) -->
                <nav class="docs-toc" aria-label="Table of contents">
                    <div class="toc-inner">
                        <div class="toc-title">On this page</div>
                        <ul class="toc-list">
                            <li v-for="section in docSections" :key="section.id" class="toc-item">
                                <a
                                    :href="`#${section.id}`"
                                    class="toc-link"
                                    :class="{ 'toc-link--active': activeAnchor === section.id }"
                                    @click.prevent="scrollToSection(section.id)"
                                >
                                    <q-icon :name="section.icon" size="16px" />
                                    <span>{{ section.title }}</span>
                                </a>
                                <ul v-if="section.subsections.length" class="toc-sublist">
                                    <li v-for="sub in section.subsections" :key="sub.id">
                                        <a
                                            :href="`#${sub.id}`"
                                            class="toc-sublink"
                                            :class="{ 'toc-sublink--active': activeAnchor === sub.id }"
                                            @click.prevent="scrollToSection(sub.id)"
                                        >{{ sub.title }}</a>
                                    </li>
                                </ul>
                            </li>
                        </ul>
                    </div>
                </nav>

                <!-- Mobile contents toggle -->
                <q-btn
                    v-if="$q.screen.lt.md"
                    fab
                    icon="menu_book"
                    color="primary"
                    class="mobile-toc-btn"
                    @click="mobileTocOpen = true"
                >
                    <q-tooltip>Contents</q-tooltip>
                </q-btn>

                <div class="docs-content">
                    <!-- Getting Started -->
                    <section id="getting-started" class="doc-section">
                        <div class="doc-section-header">
                            <div class="doc-section-badge">
                                <q-icon name="hub" />
                                <span>Documentation Hub</span>
                            </div>
                            <h2 class="doc-section-title">Welcome to FM-Dash</h2>
                            <p class="doc-section-subtitle">
                                Choose your path to get the most out of Football Manager data analysis.
                                Whether you're analyzing players, hosting your own instance, or building
                                integrations, we've got you covered.
                            </p>
                        </div>

                        <div class="hub-cards">
                            <q-card class="hub-card use-card">
                                <q-card-section class="hub-card-content">
                                    <div class="hub-icon">
                                        <q-icon name="sports_soccer" size="3rem" />
                                    </div>
                                    <h3>Use FM-Dash</h3>
                                    <p class="hub-description">
                                        Learn how to export your Football Manager 24 data and start
                                        analyzing players, teams, and formations with our powerful tools.
                                    </p>
                                    <div class="hub-features">
                                        <div class="feature-item">
                                            <q-icon name="upload" size="1rem" />
                                            <span>Export data from FM24</span>
                                        </div>
                                        <div class="feature-item">
                                            <q-icon name="analytics" size="1rem" />
                                            <span>Analyze player performance</span>
                                        </div>
                                        <div class="feature-item">
                                            <q-icon name="groups" size="1rem" />
                                            <span>Build optimal teams</span>
                                        </div>
                                    </div>
                                    <q-btn
                                        unelevated
                                        color="primary"
                                        label="Get Started"
                                        icon="play_arrow"
                                        @click="scrollToSection('data-export')"
                                        class="hub-btn"
                                        size="lg"
                                    />
                                </q-card-section>
                            </q-card>

                            <q-card class="hub-card host-card">
                                <q-card-section class="hub-card-content">
                                    <div class="hub-icon">
                                        <q-icon name="cloud_download" size="3rem" />
                                    </div>
                                    <h3>Host FM-Dash</h3>
                                    <p class="hub-description">
                                        Set up your own FM-Dash instance locally. Perfect for privacy,
                                        customization, or when you need full control over your data.
                                    </p>
                                    <div class="hub-features">
                                        <div class="feature-item">
                                            <q-icon name="smart_toy" size="1rem" />
                                            <span>Docker deployment</span>
                                        </div>
                                        <div class="feature-item">
                                            <q-icon name="construction" size="1rem" />
                                            <span>Manual installation</span>
                                        </div>
                                        <div class="feature-item">
                                            <q-icon name="security" size="1rem" />
                                            <span>Private &amp; secure</span>
                                        </div>
                                    </div>
                                    <q-btn
                                        unelevated
                                        color="secondary"
                                        label="Deploy Now"
                                        icon="rocket_launch"
                                        @click="scrollToSection('local-deployment')"
                                        class="hub-btn"
                                        size="lg"
                                    />
                                </q-card-section>
                            </q-card>

                            <q-card class="hub-card hack-card">
                                <q-card-section class="hub-card-content">
                                    <div class="hub-icon">
                                        <q-icon name="code" size="3rem" />
                                    </div>
                                    <h3>Hack FM-Dash</h3>
                                    <p class="hub-description">
                                        Build integrations, create custom tools, or contribute to the project.
                                        Explore our API and extend FM-Dash's capabilities.
                                    </p>
                                    <div class="hub-features">
                                        <div class="feature-item">
                                            <q-icon name="api" size="1rem" />
                                            <span>REST API endpoints</span>
                                        </div>
                                        <div class="feature-item">
                                            <q-icon name="integration_instructions" size="1rem" />
                                            <span>Integration guides</span>
                                        </div>
                                        <div class="feature-item">
                                            <q-icon name="source" size="1rem" />
                                            <span>Open source</span>
                                        </div>
                                    </div>
                                    <q-btn
                                        unelevated
                                        color="positive"
                                        label="Explore API"
                                        icon="terminal"
                                        @click="scrollToSection('api-reference')"
                                        class="hub-btn"
                                        size="lg"
                                    />
                                </q-card-section>
                            </q-card>
                        </div>

                        <div class="getting-started-grid">
                            <SectionCard title="Quick Start" icon="rocket_launch">
                                <ol class="quick-start-list">
                                    <li v-for="(step, index) in quickStartSteps" :key="step.title">
                                        <div class="quick-start-number">{{ index + 1 }}</div>
                                        <div>
                                            <h4>{{ step.title }}</h4>
                                            <p>{{ step.description }}</p>
                                        </div>
                                    </li>
                                </ol>
                            </SectionCard>

                            <SectionCard title="System Requirements" icon="checklist">
                                <ul class="requirements-list">
                                    <li v-for="requirement in systemRequirements" :key="requirement.id">
                                        <q-icon :name="requirement.icon" size="1.3rem" color="primary" />
                                        <div>
                                            <h4>{{ requirement.title }}</h4>
                                            <p>{{ requirement.description }}</p>
                                        </div>
                                    </li>
                                </ul>
                                <div class="format-chips">
                                    <span class="format-chip" v-for="format in dataFormats" :key="format">
                                        <q-icon name="description" size="0.9rem" />
                                        {{ format }}
                                    </span>
                                </div>
                            </SectionCard>
                        </div>
                    </section>

                    <!-- Data Export Guide -->
                    <section id="data-export" class="doc-section">
                        <div class="doc-section-header">
                            <div class="doc-section-badge">
                                <q-icon name="upload" />
                                <span>Data Export</span>
                            </div>
                            <h2 class="doc-section-title">Export Data from FM24</h2>
                            <p class="doc-section-subtitle">
                                Learn how to export your Football Manager 24 player data
                                for analysis in FM-Dash. This guide will walk you through
                                the complete process from setup to export.
                            </p>
                        </div>

                        <SectionCard
                            id="data-export-requirements"
                            title="What You'll Need"
                            icon="info"
                            class="q-mb-md"
                        >
                            <div class="requirement-grid">
                                <div class="requirement-item">
                                    <q-icon name="sports_soccer" size="1.5rem" color="primary" />
                                    <div>
                                        <h4>Football Manager 24</h4>
                                        <p>An active save game with players to analyze</p>
                                    </div>
                                </div>
                                <div class="requirement-item">
                                    <q-icon name="download" size="1.5rem" color="secondary" />
                                    <div>
                                        <h4>Steam Workshop Access</h4>
                                        <p>To download the FM Dash search view</p>
                                    </div>
                                </div>
                                <div class="requirement-item">
                                    <q-icon name="folder" size="1.5rem" color="positive" />
                                    <div>
                                        <h4>File Storage</h4>
                                        <p>A place to save your exported data</p>
                                    </div>
                                </div>
                            </div>
                        </SectionCard>

                        <SectionCard
                            id="data-export-steps"
                            title="Step-by-Step Export Process"
                            icon="list_alt"
                            class="q-mb-md"
                        >
                            <div class="export-steps">
                                <div class="export-step" v-for="(step, index) in exportSteps" :key="step.title">
                                    <div class="step-number">{{ index + 1 }}</div>
                                    <div class="step-content">
                                        <h4>{{ step.title }}</h4>
                                        <p>{{ step.description }}</p>
                                        <div v-if="step.note" class="step-note">
                                            <q-icon name="lightbulb" size="1rem" />
                                            {{ step.note }}
                                        </div>
                                        <div v-if="step.warning" class="step-warning">
                                            <q-icon name="warning" size="1rem" />
                                            {{ step.warning }}
                                        </div>
                                        <div v-if="step.link" class="step-link">
                                            <q-btn
                                                outline
                                                color="primary"
                                                :label="step.linkText"
                                                icon="open_in_new"
                                                :href="step.link"
                                                target="_blank"
                                                size="sm"
                                            />
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </SectionCard>

                        <SectionCard
                            id="data-export-performance"
                            title="Performance Tips"
                            icon="speed"
                            class="q-mb-md"
                        >
                            <div class="performance-tips">
                                <div class="tip-item">
                                    <q-icon name="filter_list" size="1.2rem" color="orange" />
                                    <div>
                                        <h4>Start Small</h4>
                                        <p>For your first export, limit to under 5,000 players to test the process quickly.</p>
                                    </div>
                                </div>
                                <div class="tip-item">
                                    <q-icon name="hourglass_empty" size="1.2rem" color="orange" />
                                    <div>
                                        <h4>Large Datasets</h4>
                                        <p>Exports with 10,000+ players can take 10+ seconds. Be patient and don't interact with the screen during export.</p>
                                    </div>
                                </div>
                                <div class="tip-item">
                                    <q-icon name="mouse" size="1.2rem" color="orange" />
                                    <div>
                                        <h4>Check Progress</h4>
                                        <p>You can test if the export is working by hovering over players. If nothing changes when you hover, it's working correctly. If you see hover effects, navigate away from scouting and back to try again.</p>
                                    </div>
                                </div>
                            </div>
                        </SectionCard>

                        <SectionCard
                            id="data-export-troubleshooting"
                            title="Troubleshooting"
                            icon="help_outline"
                            class="q-mb-md"
                        >
                            <div class="troubleshooting-items">
                                <div class="trouble-item">
                                    <h4>Export seems stuck or frozen</h4>
                                    <p>Large datasets can take 10+ seconds to export. You can test if it's working by hovering over players - if nothing changes, it's working correctly. If you see hover effects, navigate away from scouting and back to try again.</p>
                                </div>
                                <div class="trouble-item">
                                    <h4>Can't find the FM Dash Search view</h4>
                                    <p>Make sure you've subscribed to the Steam Workshop item and that Steam has downloaded it. Restart FM24 if the view doesn't appear.</p>
                                </div>
                                <div class="trouble-item">
                                    <h4>Export file is too large</h4>
                                    <p>Consider filtering your dataset further before export. Focus on specific leagues, age ranges, or positions to reduce file size.</p>
                                </div>
                            </div>
                        </SectionCard>

                        <SectionCard
                            id="data-export-next-steps"
                            title="Ready to Analyze!"
                            icon="check_circle"
                            class="success-card"
                        >
                            <p class="success-description">
                                Once you've successfully exported your data, you're ready to upload it to FM-Dash
                                and start analyzing your players. The exported HTML file contains all the player
                                data needed for comprehensive analysis.
                            </p>
                            <div class="next-actions">
                                <q-btn
                                    unelevated
                                    color="positive"
                                    label="Upload to FM-Dash"
                                    icon="cloud_upload"
                                    href="/"
                                    class="action-btn"
                                />
                                <q-btn
                                    outline
                                    color="primary"
                                    label="Getting Started Guide"
                                    icon="rocket_launch"
                                    @click="scrollToSection('getting-started')"
                                    class="action-btn"
                                />
                            </div>
                        </SectionCard>
                    </section>

                    <!-- API Reference -->
                    <section id="api-reference" class="doc-section">
                        <div class="doc-section-header">
                            <div class="doc-section-badge">
                                <q-icon name="code" />
                                <span>Developer Tools</span>
                            </div>
                            <h2 class="doc-section-title">API Reference</h2>
                            <p class="doc-section-subtitle">
                                Technical documentation for developers working with
                                FM-Dash's API endpoints and data structures.
                            </p>
                        </div>

                        <SectionCard title="Available Endpoints" icon="api">
                            <div class="api-endpoints">
                                <div class="endpoint-item" v-for="endpoint in apiEndpoints" :key="endpoint.id">
                                    <div class="endpoint-method" :class="endpoint.method.toLowerCase()">
                                        {{ endpoint.method }}
                                    </div>
                                    <div class="endpoint-details">
                                        <code class="endpoint-path">{{ endpoint.path }}</code>
                                        <p class="endpoint-description">{{ endpoint.description }}</p>
                                    </div>
                                </div>
                            </div>
                        </SectionCard>
                    </section>

                    <!-- Local Deployment -->
                    <section id="local-deployment" class="doc-section">
                        <div class="doc-section-header">
                            <div class="doc-section-badge">
                                <q-icon name="cloud_download" />
                                <span>Self-Hosting</span>
                            </div>
                            <h2 class="doc-section-title">Local Deployment</h2>
                            <p class="doc-section-subtitle">
                                Run FM-Dash on your own computer for personal use.
                                Choose between Docker (easiest) or manual setup.
                            </p>
                        </div>

                        <SectionCard
                            id="local-deployment-prerequisites"
                            title="Prerequisites"
                            icon="download"
                            class="q-mb-md"
                        >
                            <p class="card-description">
                                Choose your deployment method and install the required tools:
                            </p>
                            <div class="prerequisites-grid">
                                <div
                                    class="prerequisite-item"
                                    v-for="requirement in localRequirements"
                                    :key="requirement.id"
                                >
                                    <q-icon
                                        :name="requirement.icon"
                                        size="2rem"
                                        color="orange"
                                        class="prereq-icon"
                                    />
                                    <h4>{{ requirement.title }}</h4>
                                    <p>{{ requirement.description }}</p>
                                    <a
                                        v-if="requirement.downloadUrl"
                                        :href="requirement.downloadUrl"
                                        target="_blank"
                                        class="download-link"
                                    >
                                        Download {{ requirement.shortName }}
                                    </a>
                                </div>
                            </div>
                        </SectionCard>

                        <SectionCard
                            id="local-deployment-docker"
                            class="method-card docker-method q-mb-md"
                        >
                            <template #header>
                                <div class="method-badge recommended">
                                    <q-icon name="smart_toy" size="1.5rem" />
                                    <div>
                                        <h3>Option 1: Docker (Recommended)</h3>
                                        <p>The easiest way to run FM-Dash. Everything is pre-configured.</p>
                                    </div>
                                </div>
                            </template>

                            <div class="setup-steps">
                                <div class="setup-step" v-for="(step, index) in dockerSteps" :key="step.title">
                                    <div class="step-number">{{ index + 1 }}</div>
                                    <div class="step-content">
                                        <h4>{{ step.title }}</h4>
                                        <p>{{ step.description }}</p>
                                        <div v-if="step.commands" class="command-list">
                                            <div
                                                v-for="command in step.commands"
                                                :key="command"
                                                class="command-block"
                                            >
                                                <code>{{ command }}</code>
                                                <q-btn
                                                    flat
                                                    round
                                                    icon="content_copy"
                                                    size="sm"
                                                    @click="copyToClipboard(command)"
                                                    class="copy-btn"
                                                    dense
                                                />
                                            </div>
                                        </div>
                                        <div v-if="step.fileContent" class="file-content">
                                            <div class="file-header">
                                                <span class="file-name">{{ step.fileName }}</span>
                                                <q-btn
                                                    flat
                                                    round
                                                    icon="content_copy"
                                                    size="sm"
                                                    @click="copyToClipboard(step.fileContent)"
                                                    class="copy-btn"
                                                    dense
                                                />
                                            </div>
                                            <pre class="file-code">{{ step.fileContent }}</pre>
                                        </div>
                                        <div v-if="step.note" class="step-note">
                                            <q-icon name="info" size="1rem" />
                                            {{ step.note }}
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div class="final-step">
                                <h4><q-icon name="check_circle" color="positive" size="1.1rem" /> Access Your Application</h4>
                                <p>Once the container is running, open your web browser and go to:</p>
                                <div class="access-url"><strong>http://localhost:3000</strong></div>
                                <p class="access-note">You should see the FM-Dash interface. You can now upload your Football Manager data and start analyzing!</p>
                            </div>
                        </SectionCard>

                        <SectionCard
                            id="local-deployment-manual"
                            class="method-card manual-method q-mb-md"
                        >
                            <template #header>
                                <div class="method-badge">
                                    <q-icon name="construction" size="1.5rem" />
                                    <div>
                                        <h3>Option 2: Manual Installation</h3>
                                        <p>For users who prefer to build and run the application manually.</p>
                                    </div>
                                </div>
                            </template>

                            <div class="setup-steps">
                                <div class="setup-step" v-for="(step, index) in setupSteps" :key="step.title">
                                    <div class="step-number">{{ index + 1 }}</div>
                                    <div class="step-content">
                                        <h4>{{ step.title }}</h4>
                                        <p>{{ step.description }}</p>
                                        <div v-if="step.commands" class="command-list">
                                            <div
                                                v-for="command in step.commands"
                                                :key="command"
                                                class="command-block"
                                            >
                                                <code>{{ command }}</code>
                                                <q-btn
                                                    flat
                                                    round
                                                    icon="content_copy"
                                                    size="sm"
                                                    @click="copyToClipboard(command)"
                                                    class="copy-btn"
                                                    dense
                                                />
                                            </div>
                                        </div>
                                        <div v-if="step.note" class="step-note">
                                            <q-icon name="info" size="1rem" />
                                            {{ step.note }}
                                        </div>
                                    </div>
                                </div>
                            </div>

                            <div class="final-step">
                                <h4><q-icon name="check_circle" color="positive" size="1.1rem" /> Access Your Application</h4>
                                <p>Once both servers are running, open your web browser and go to:</p>
                                <div class="access-url"><strong>http://localhost:3000</strong></div>
                                <p class="access-note">You should see the FM-Dash interface. You can now upload your Football Manager data and start analyzing!</p>
                            </div>
                        </SectionCard>

                        <SectionCard id="local-deployment-help" title="Need More Help?" icon="help">
                            <div class="resource-links">
                                <q-btn
                                    outline
                                    color="primary"
                                    label="View Full Source Code"
                                    icon="code"
                                    href="https://github.com/LiamHardman/fmdash"
                                    target="_blank"
                                    class="resource-btn"
                                />
                                <q-btn
                                    outline
                                    color="green"
                                    label="Report Problems"
                                    icon="bug_report"
                                    href="https://github.com/LiamHardman/fmdash/issues"
                                    target="_blank"
                                    class="resource-btn"
                                />
                            </div>
                            <p class="resource-note">
                                If you're having issues, check the GitHub repository for detailed documentation and community support.
                            </p>
                        </SectionCard>
                    </section>

                    <!-- Rating Calculations -->
                    <section id="rating-calculations" class="doc-section">
                        <div class="doc-section-header">
                            <div class="doc-section-badge">
                                <q-icon name="calculate" />
                                <span>Rating System</span>
                            </div>
                            <h2 class="doc-section-title">How Ratings Are Calculated</h2>
                            <p class="doc-section-subtitle">
                                Understanding how FM-Dash calculates different player ratings and what they mean for player analysis.
                            </p>
                        </div>

                        <SectionCard id="rating-overview" title="Rating System Overview" icon="info" class="q-mb-md">
                            <p class="card-description">
                                FM-Dash uses a sophisticated rating system that combines Football Manager's raw attributes
                                into meaningful categories. Each rating type serves a different purpose in player analysis.
                            </p>
                        </SectionCard>

                        <SectionCard
                            id="rating-fifa"
                            title="FIFA-Style Category Ratings"
                            icon="sports_soccer"
                            class="q-mb-md"
                        >
                            <p class="card-description">
                                These ratings convert Football Manager's 1-20 attribute system into FIFA-style 0-99 ratings
                                that are easier to understand and compare.
                            </p>

                            <div class="rating-categories">
                                <div class="rating-category" v-for="category in fifaCategories" :key="category.name">
                                    <div class="category-header">
                                        <h4>{{ category.name }}</h4>
                                        <span class="category-description">{{ category.description }}</span>
                                    </div>
                                    <div class="weights-explanation">
                                        <p class="weights-note">
                                            <q-icon name="info" size="1rem" />
                                            <span>The numbers below are <strong>weights</strong> that determine how much each attribute contributes to the final rating. Higher weights = more important attributes.</span>
                                        </p>
                                    </div>
                                    <div class="attribute-weights">
                                        <div class="weight-item" v-for="weight in category.weights" :key="weight.attribute">
                                            <span class="attribute-name">{{ weight.attribute }}</span>
                                            <span class="weight-value">Weight: {{ weight.weight }}</span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </SectionCard>

                        <SectionCard id="rating-overall" title="Overall Rating" icon="star" class="q-mb-md">
                            <p class="card-description">
                                The Overall rating represents a player's best performance across all possible roles.
                                It's calculated by evaluating the player in every applicable position and taking the highest score.
                            </p>
                            <div class="overall-calculation">
                                <h4>How it's calculated:</h4>
                                <ol>
                                    <li>Evaluate the player in every applicable role (e.g., Central Midfielder, Winger, etc.)</li>
                                    <li>Each role uses specific attribute weights optimized for that position</li>
                                    <li>Take the highest score from all evaluated roles</li>
                                    <li>Apply non-linear scaling to compress lower ratings and maintain separation</li>
                                </ol>
                                <div class="calculation-note">
                                    <q-icon name="lightbulb" size="1rem" />
                                    <span>This ensures the Overall rating represents the player's peak potential rather than an average.</span>
                                </div>
                            </div>
                        </SectionCard>

                        <SectionCard id="rating-total-stats" title="Total Stats" icon="summarize" class="q-mb-md">
                            <p class="card-description">
                                Total Stats is the sum of all physical, mental, and technical attributes. This provides a raw
                                measure of a player's total attribute points.
                            </p>
                            <div class="total-stats-breakdown">
                                <h4>Included attributes:</h4>
                                <div class="stats-categories">
                                    <div class="stats-category">
                                        <h5>Physical (8 attributes)</h5>
                                        <p>Acceleration, Agility, Balance, Jumping, Natural Fitness, Pace, Stamina, Strength</p>
                                    </div>
                                    <div class="stats-category">
                                        <h5>Mental (14 attributes)</h5>
                                        <p>Aggression, Anticipation, Bravery, Composure, Concentration, Decisions, Determination, Flair, Leadership, Off The Ball, Positioning, Teamwork, Vision, Work Rate</p>
                                    </div>
                                    <div class="stats-category">
                                        <h5>Technical (14 attributes)</h5>
                                        <p>Corners, Crossing, Dribbling, Finishing, First Touch, Free Kicks, Heading, Long Shots, Long Throws, Marking, Passing, Penalty Taking, Tackling, Technique</p>
                                    </div>
                                </div>
                                <div class="calculation-note">
                                    <q-icon name="info" size="1rem" />
                                    <span>Only attributes with values greater than 0 are included in the total.</span>
                                </div>
                            </div>
                        </SectionCard>

                        <SectionCard id="rating-mbr" title="Moneyball Rating (MBR)" icon="trending_up" class="q-mb-md">
                            <p class="card-description">
                                The Moneyball Rating evaluates a player's value for money, considering their ability, age,
                                personality, and transfer value relative to market expectations.
                            </p>
                            <div class="mbr-calculation">
                                <h4>Formula components:</h4>
                                <div class="mbr-components">
                                    <div class="mbr-component">
                                        <h5>Base Rating</h5>
                                        <p>Player Overall ÷ 3</p>
                                    </div>
                                    <div class="mbr-component">
                                        <h5>Age Modifier</h5>
                                        <p>Bonuses for young players (16-25), penalties for older players (26+)</p>
                                    </div>
                                    <div class="mbr-component">
                                        <h5>Mentality Modifier</h5>
                                        <p>Bonuses for good personalities, penalties for poor ones</p>
                                    </div>
                                    <div class="mbr-component">
                                        <h5>Value Score</h5>
                                        <p>How much value you get per rating point compared to market expectations</p>
                                    </div>
                                    <div class="mbr-component">
                                        <h5>Transfer Value Penalty</h5>
                                        <p>Penalties for overpriced players, bonuses for undervalued ones</p>
                                    </div>
                                    <div class="mbr-component">
                                        <h5>Salary Penalty</h5>
                                        <p>Penalties for overpaid players, bonuses for underpaid ones</p>
                                    </div>
                                </div>
                                <div class="calculation-note">
                                    <q-icon name="trending_up" size="1rem" />
                                    <span>Higher MBR indicates better value for money. Perfect for finding undervalued players!</span>
                                </div>
                            </div>
                        </SectionCard>

                        <SectionCard id="rating-scaling" title="Rating Scaling" icon="tune" class="q-mb-md">
                            <p class="card-description">
                                FM-Dash uses sophisticated scaling to convert Football Manager's 1-20 attributes into more
                                intuitive 0-99 ratings while maintaining meaningful differentiation between players.
                            </p>
                            <div class="scaling-info">
                                <h4>Non-linear scaling:</h4>
                                <ul>
                                    <li><strong>Ratings 75+:</strong> Minimal compression to preserve elite player differentiation</li>
                                    <li><strong>Ratings below 75:</strong> Progressive compression using power curves to create better separation</li>
                                    <li><strong>Minimum progression:</strong> Ensures players with decent attributes don't cluster too low</li>
                                </ul>
                                <div class="calculation-note">
                                    <q-icon name="info" size="1rem" />
                                    <span>This scaling makes it easier to distinguish between players while maintaining realistic rating distributions.</span>
                                </div>
                            </div>
                        </SectionCard>

                        <SectionCard
                            id="rating-usage"
                            title="How to Use These Ratings"
                            icon="tips_and_updates"
                            class="success-card"
                        >
                            <div class="usage-tips">
                                <div class="tip-item">
                                    <h4>Overall Rating</h4>
                                    <p>Use for quick player comparisons and identifying the best players in your squad.</p>
                                </div>
                                <div class="tip-item">
                                    <h4>FIFA-Style Ratings</h4>
                                    <p>Use to understand a player's strengths and weaknesses in specific areas of the game.</p>
                                </div>
                                <div class="tip-item">
                                    <h4>Total Stats</h4>
                                    <p>Use to identify players with high raw attribute totals, regardless of how they're distributed.</p>
                                </div>
                                <div class="tip-item">
                                    <h4>Moneyball Rating</h4>
                                    <p>Use to find undervalued players and make smart transfer decisions based on value for money.</p>
                                </div>
                            </div>
                        </SectionCard>
                    </section>
                </div>
            </div>
        </div>

        <!-- Mobile table-of-contents dialog -->
        <q-dialog v-model="mobileTocOpen" position="left">
            <div class="mobile-toc">
                <div class="toc-inner">
                    <div class="mobile-toc-header">
                        <div class="toc-title">On this page</div>
                        <q-btn flat round dense icon="close" size="sm" @click="mobileTocOpen = false" />
                    </div>
                    <ul class="toc-list">
                        <li v-for="section in docSections" :key="section.id" class="toc-item">
                            <a
                                :href="`#${section.id}`"
                                class="toc-link"
                                :class="{ 'toc-link--active': activeAnchor === section.id }"
                                @click.prevent="scrollToSection(section.id)"
                            >
                                <q-icon :name="section.icon" size="16px" />
                                <span>{{ section.title }}</span>
                            </a>
                            <ul v-if="section.subsections.length" class="toc-sublist">
                                <li v-for="sub in section.subsections" :key="sub.id">
                                    <a
                                        :href="`#${sub.id}`"
                                        class="toc-sublink"
                                        :class="{ 'toc-sublink--active': activeAnchor === sub.id }"
                                        @click.prevent="scrollToSection(sub.id)"
                                    >{{ sub.title }}</a>
                                </li>
                            </ul>
                        </li>
                    </ul>
                </div>
            </div>
        </q-dialog>
    </q-page>
</template>

<script>
import { defineComponent, onBeforeUnmount, onMounted, ref } from 'vue'
import PageHeader from '@/components/layout/PageHeader.vue'
import SectionCard from '@/components/layout/SectionCard.vue'
import { useUiStore } from '@/stores/uiStore'

export default defineComponent({
  name: 'DocsPage',
  components: { PageHeader, SectionCard },
  setup() {
    const mobileTocOpen = ref(false)
    const activeAnchor = ref('getting-started')
    const uiStore = useUiStore()
    let observer = null

    const docSections = [
      {
        id: 'getting-started',
        title: 'Getting Started',
        subtitle: 'Setup and basics',
        icon: 'rocket_launch',
        subsections: [],
      },
      {
        id: 'data-export',
        title: 'Data Export Guide',
        subtitle: 'Export from FM24',
        icon: 'upload',
        subsections: [
          { id: 'data-export-requirements', title: "What You'll Need" },
          { id: 'data-export-steps', title: 'Step-by-Step Process' },
          { id: 'data-export-performance', title: 'Performance Tips' },
          { id: 'data-export-troubleshooting', title: 'Troubleshooting' },
          { id: 'data-export-next-steps', title: 'Ready to Analyze' },
        ],
      },
      {
        id: 'api-reference',
        title: 'API Reference',
        subtitle: 'Developer docs',
        icon: 'code',
        subsections: [],
      },
      {
        id: 'local-deployment',
        title: 'Local Deployment',
        subtitle: 'Self-hosting',
        icon: 'cloud_download',
        subsections: [
          { id: 'local-deployment-prerequisites', title: 'Prerequisites' },
          { id: 'local-deployment-docker', title: 'Docker Setup' },
          { id: 'local-deployment-manual', title: 'Manual Setup' },
          { id: 'local-deployment-help', title: 'Need Help?' },
        ],
      },
      {
        id: 'rating-calculations',
        title: 'Rating Calculations',
        subtitle: 'How ratings work',
        icon: 'calculate',
        subsections: [
          { id: 'rating-overview', title: 'Overview' },
          { id: 'rating-fifa', title: 'FIFA-Style Ratings' },
          { id: 'rating-overall', title: 'Overall Rating' },
          { id: 'rating-total-stats', title: 'Total Stats' },
          { id: 'rating-mbr', title: 'Moneyball Rating' },
          { id: 'rating-scaling', title: 'Rating Scaling' },
          { id: 'rating-usage', title: 'Usage Tips' },
        ],
      },
    ]

    const heroFeatures = [
      { id: 1, icon: 'analytics', title: 'Player Analysis' },
      { id: 2, icon: 'groups', title: 'Team Management' },
      { id: 3, icon: 'sports_soccer', title: 'Formation Tools' },
      { id: 4, icon: 'api', title: 'API Access' },
    ]

    const quickStartSteps = [
      {
        title: 'Upload Your Data',
        description: 'Import your Football Manager player data using our secure upload system.',
      },
      {
        title: 'Explore Analysis Tools',
        description: 'Use powerful analysis features to evaluate player performance and potential.',
      },
      {
        title: 'Manage Your Team',
        description: 'View formations, optimize lineups, and track team performance metrics.',
      },
    ]

    const systemRequirements = [
      {
        id: 1,
        icon: 'web',
        title: 'Modern Web Browser',
        description: 'Chrome, Firefox, Safari, or Edge (latest versions)',
      },
      {
        id: 2,
        icon: 'sports_soccer',
        title: 'Football Manager Data',
        description: 'Exported player data from FM (HTML only)',
      },
      {
        id: 3,
        icon: 'memory',
        title: '4GB+ RAM',
        description: 'For optimal performance with large datasets',
      },
    ]

    const apiEndpoints = [
      {
        id: 1,
        method: 'POST',
        path: '/upload',
        description:
          'Upload and process Football Manager player data files (HTML format). Returns a dataset ID (file hash) on success.',
      },
      {
        id: 2,
        method: 'GET',
        path: '/api/players/{dataset_id}',
        description:
          'Retrieve player information for a specific dataset. Supports query parameters for filtering (e.g., position, age, name) and pagination.',
      },
      {
        id: 3,
        method: 'GET',
        path: '/api/roles',
        description:
          'Get a list of available player roles and their associated attribute weights used for calculating player suitability and ratings.',
      },
      {
        id: 4,
        method: 'GET',
        path: '/api/leagues/{dataset_id}',
        description:
          'Retrieve all leagues present in a given dataset, along with team counts and aggregate quality metrics for each league.',
      },
      {
        id: 5,
        method: 'GET',
        path: '/api/teams/{dataset_id}?league={league_name}',
        description:
          'Get detailed team data for a specific league within a dataset. Includes player rosters, average ratings, and tactical information.',
      },
      {
        id: 6,
        method: 'POST',
        path: '/api/percentiles/{dataset_id}',
        description:
          'Calculate and retrieve player performance percentiles. Request body can specify player name for individual analysis, or division filters to compare against specific cohorts.',
      },
      {
        id: 7,
        method: 'GET',
        path: '/api/search/{dataset_id}?q={query}',
        description:
          'Perform a global search within a specific dataset for players, teams, leagues, or nations based on the provided query string.',
      },
      {
        id: 8,
        method: 'GET',
        path: '/api/config',
        description:
          'Retrieve application-level configuration, such as available player positions, attribute groups, UI settings, and version information.',
      },
      {
        id: 9,
        method: 'POST',
        path: '/api/bargain-hunter/{dataset_id}',
        description:
          'Analyze player data to find undervalued players (bargains). Request body includes criteria like max budget, max salary, min/max age, and minimum overall rating.',
      },
      {
        id: 10,
        method: 'GET',
        path: '/api/faces?id={face_id}',
        description:
          'Retrieve player face images by their unique face ID. Returns image data if available.',
      },
      {
        id: 11,
        method: 'GET',
        path: '/api/cache/nation-ratings/{dataset_id}',
        description:
          'Retrieves cached aggregated ratings (attack, midfield, defense, overall) for all nations represented in the specified dataset.',
      },
      {
        id: 12,
        method: 'POST',
        path: '/api/cache/nation-ratings/{dataset_id}',
        description:
          'Generates or updates the cached aggregated ratings for all nations in the specified dataset. (Primarily for internal use or administrative tasks).',
      },
    ]

    const dataFormats = ['HTML']

    const localRequirements = [
      {
        id: 1,
        icon: 'smart_toy',
        title: 'Docker & Docker Compose',
        description: 'For the easiest setup. Includes everything needed to run FM-Dash.',
        downloadUrl: 'https://www.docker.com/get-started',
        shortName: 'Docker',
      },
      {
        id: 2,
        icon: 'code',
        title: 'Node.js (version 18 or higher)',
        description: 'Required for manual setup. Download and install from the official website.',
        downloadUrl: 'https://nodejs.org/',
        shortName: 'Node.js',
      },
      {
        id: 3,
        icon: 'terminal',
        title: 'Go (version 1.24 or higher)',
        description: 'Required for manual setup. Needed to run the backend API server.',
        downloadUrl: 'https://golang.org/dl/',
        shortName: 'Go',
      },
      {
        id: 4,
        icon: 'source',
        title: 'Git',
        description: 'Required to download the source code from GitHub.',
        downloadUrl: 'https://git-scm.com/',
        shortName: 'Git',
      },
    ]

    const dockerSteps = [
      {
        title: 'Clone the Repository',
        description: 'Download the FM-Dash source code which includes the Docker configuration:',
        commands: ['git clone https://github.com/LiamHardman/fmdash.git', 'cd fmdash'],
        note: 'This downloads all the necessary files including docker-compose.yml',
      },
      {
        title: 'Create Docker Compose File',
        description: 'Create a docker-compose.yml file in the project directory with this content:',
        fileContent: `version: '3.8'

services:
  fmdash:
    build: .
    ports:
      - "3000:8080"
    environment:
      - PORT_GO_API=8091
      - PORT_NGINX=8080
      - ENABLE_METRICS=false
      - INSECURE_MODE=true
      - SERVICE_NAME=fmdash-local
      - MAX_UPLOAD_SIZE=50
      - SERVICE_VERSION=v1.0.0
      - ENVIRONMENT=local
      - DEPLOYMENT_ENV=docker
    volumes:
      - fmdash_data:/app/data
    restart: unless-stopped

volumes:
  fmdash_data:`,
        fileName: 'docker-compose.yml',
        note: 'This configuration runs FM-Dash without external dependencies like S3 storage',
      },
      {
        title: 'Start the Application',
        description: 'Build and start the FM-Dash containers:',
        commands: ['docker-compose up -d'],
        note: 'This builds the Docker image and starts the container. First run may take several minutes.',
      },
      {
        title: 'Verify Installation',
        description: 'Check that the application is running properly:',
        commands: ['docker-compose ps', 'docker-compose logs fmdash'],
        note: 'The first command shows running containers, the second shows application logs',
      },
    ]

    const setupSteps = [
      {
        title: 'Download the Source Code',
        description:
          'Open a terminal or command prompt and run these commands to download FM-Dash:',
        commands: ['git clone https://github.com/LiamHardman/fmdash.git', 'cd fmdash'],
        note: 'This creates a "fmdash" folder with all the necessary files',
      },
      {
        title: 'Install Frontend Dependencies',
        description: 'Install all the JavaScript packages needed for the frontend:',
        commands: ['npm install'],
        note: 'This step may take several minutes depending on your internet speed',
      },
      {
        title: 'Verify Go Installation',
        description: 'Make sure Go is properly installed and configured:',
        commands: ['go version'],
        note: 'You should see version 1.24 or higher. If not, install Go from the prerequisites above',
      },
      {
        title: 'Start the Application',
        description: 'Launch both the frontend and backend servers:',
        commands: ['./launch_dev.sh'],
        note: 'This script starts both servers automatically. Wait for both to fully start before proceeding.',
      },
    ]

    const fifaCategories = [
      {
        name: 'PAC (Pace)',
        description: 'How fast a player can move',
        weights: [
          { attribute: 'Acceleration', weight: 8 },
          { attribute: 'Pace', weight: 8 },
          { attribute: 'Agility', weight: 5 },
        ],
      },
      {
        name: 'SHO (Shooting)',
        description: "A player's ability to score goals",
        weights: [
          { attribute: 'Finishing', weight: 8 },
          { attribute: 'Long Shots', weight: 6 },
          { attribute: 'Penalty Taking', weight: 4 },
          { attribute: 'Heading', weight: 5 },
          { attribute: 'Composure', weight: 6 },
          { attribute: 'Technique', weight: 5 },
          { attribute: 'Anticipation', weight: 4 },
          { attribute: 'Decisions', weight: 4 },
          { attribute: 'Flair', weight: 3 },
        ],
      },
      {
        name: 'PAS (Passing)',
        description: "A player's ability to pass the ball effectively",
        weights: [
          { attribute: 'Passing', weight: 8 },
          { attribute: 'Crossing', weight: 6 },
          { attribute: 'Free Kicks', weight: 4 },
          { attribute: 'Vision', weight: 7 },
          { attribute: 'Technique', weight: 5 },
          { attribute: 'Teamwork', weight: 4 },
          { attribute: 'Decisions', weight: 4 },
          { attribute: 'Corners', weight: 3 },
          { attribute: 'First Touch', weight: 4 },
          { attribute: 'Off The Ball', weight: 3 },
        ],
      },
      {
        name: 'DRI (Dribbling)',
        description: "A player's ability to control the ball while moving",
        weights: [
          { attribute: 'Dribbling', weight: 8 },
          { attribute: 'First Touch', weight: 7 },
          { attribute: 'Technique', weight: 6 },
          { attribute: 'Flair', weight: 5 },
          { attribute: 'Composure', weight: 4 },
          { attribute: 'Off The Ball', weight: 3 },
        ],
      },
      {
        name: 'DEF (Defending)',
        description: "A player's defensive abilities",
        weights: [
          { attribute: 'Marking', weight: 8 },
          { attribute: 'Tackling', weight: 8 },
          { attribute: 'Heading', weight: 6 },
          { attribute: 'Anticipation', weight: 7 },
          { attribute: 'Concentration', weight: 6 },
          { attribute: 'Positioning', weight: 7 },
          { attribute: 'Decisions', weight: 5 },
          { attribute: 'Composure', weight: 4 },
          { attribute: 'Bravery', weight: 5 },
          { attribute: 'Aggression', weight: 4 },
          { attribute: 'Work Rate', weight: 4 },
        ],
      },
      {
        name: 'PHY (Physical)',
        description: "A player's physical attributes",
        weights: [
          { attribute: 'Strength', weight: 8 },
          { attribute: 'Stamina', weight: 7 },
          { attribute: 'Natural Fitness', weight: 6 },
          { attribute: 'Jumping', weight: 5 },
          { attribute: 'Balance', weight: 4 },
          { attribute: 'Aggression', weight: 5 },
          { attribute: 'Bravery', weight: 4 },
          { attribute: 'Work Rate', weight: 4 },
        ],
      },
    ]

    const exportSteps = [
      {
        title: 'Download the FM Dash Search View',
        description:
          'First, you need to download a custom search view from the Steam Workshop that contains all the player attributes FM-Dash needs for analysis.',
        link: 'https://steamcommunity.com/sharedfiles/filedetails/?id=3498467200',
        linkText: 'Download FM Dash Search View',
        note: "Make sure you're logged into Steam and subscribed to the workshop item.",
      },
      {
        title: 'Import the View in FM24',
        description:
          'Open Football Manager 24, navigate to Scouting, then click "Overview" (next to the "X Players Filtered" text). Select "Custom" → "Import View" and choose "FM Dash Search".',
        note: "If you don't see the view, restart FM24 and make sure Steam has downloaded the workshop item.",
      },
      {
        title: 'Filter Your Dataset',
        description:
          "Use FM24's filtering options to narrow down your player selection. Consider filtering by league, position, age, or other criteria to focus on the players you want to analyze.",
        note: 'Start with under 5,000 players for your first export to test the process quickly.',
      },
      {
        title: 'Select All Players',
        description:
          'Once you have your filtered list, select all players using Ctrl+A (or Cmd+A on Mac). This will highlight all visible players in the current view.',
        warning: 'Make sure all players are selected before proceeding to the export step.',
      },
      {
        title: 'Export as Web Page',
        description:
          'With all players selected, press Ctrl+P (or Cmd+P on Mac) to open the print dialog, then choose "Web Page" as the format. This creates an HTML file with all the player data.',
        warning:
          "This process can be slow for large datasets (10,000+ players). Expect 10+ seconds and don't interact with the screen during export.",
      },
      {
        title: 'Save Your Export File',
        description:
          "Choose a memorable location to save your HTML export file. You'll need to upload this file to FM-Dash for analysis.",
        note: 'Consider naming the file with the date and dataset description for easy identification later.',
      },
    ]

    const copyToClipboard = async (text) => {
      try {
        await navigator.clipboard.writeText(text)
        // You could add a toast notification here if desired
      } catch (_err) {}
    }

    const scrollToSection = (sectionId) => {
      mobileTocOpen.value = false
      const el = document.getElementById(sectionId)
      if (el) {
        el.scrollIntoView({ behavior: 'smooth', block: 'start' })
      }
    }

    const showTutorial = () => {
      uiStore.showTutorial()
    }

    // Scroll-spy: highlight the TOC entry for whichever anchor is currently
    // nearest the top of the viewport. rootMargin trims the observation zone
    // to the top ~20% of the viewport (below the fixed app header) so only
    // one heading is "current" at a time as the page is scrolled.
    onMounted(() => {
      const allAnchorIds = docSections.flatMap((section) => [
        section.id,
        ...section.subsections.map((sub) => sub.id),
      ])
      const targets = allAnchorIds
        .map((id) => document.getElementById(id))
        .filter((el) => el !== null)

      if (targets.length === 0) return

      observer = new IntersectionObserver(
        (entries) => {
          for (const entry of entries) {
            if (entry.isIntersecting) {
              activeAnchor.value = entry.target.id
            }
          }
        },
        { rootMargin: '-96px 0px -75% 0px', threshold: 0 }
      )

      for (const target of targets) {
        observer.observe(target)
      }
    })

    onBeforeUnmount(() => {
      if (observer) {
        observer.disconnect()
        observer = null
      }
    })

    return {
      mobileTocOpen,
      activeAnchor,
      docSections,
      heroFeatures,
      quickStartSteps,
      systemRequirements,
      apiEndpoints,
      dataFormats,
      localRequirements,
      dockerSteps,
      setupSteps,
      exportSteps,
      fifaCategories,
      copyToClipboard,
      scrollToSection,
      showTutorial,
    }
  },
})
</script>

<style lang="scss" scoped>
.docs-page {
    min-height: 100vh;
    background: var(--surface-page);
}

.page-container {
    max-width: var(--content-max-width);
    margin: 0 auto;
    padding: var(--page-gutter);

    @media (max-width: 768px) {
        padding: var(--page-gutter-sm);
    }
}

// ── Highlight chip row (preserves the original hero feature callouts) ─────
.docs-highlights {
    display: flex;
    flex-wrap: wrap;
    gap: 0.6rem;
    margin-bottom: var(--section-gap);
}

.highlight-chip {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.4rem 0.9rem;
    border-radius: 999px;
    background: var(--accent-soft);
    color: var(--accent);
    font-size: 0.825rem;
    font-weight: 600;
}

// ── Layout: sticky TOC + content column ────────────────────────────────
.docs-layout {
    display: flex;
    align-items: flex-start;
    gap: 2rem;

    @media (max-width: 1024px) {
        gap: 1rem;
    }
}

.docs-toc {
    flex: 0 0 240px;
    position: sticky;
    top: calc(60px + var(--page-gutter));
    max-height: calc(100vh - 60px - 2 * var(--page-gutter));
    overflow-y: auto;

    @media (max-width: 1024px) {
        display: none;
    }
}

.toc-inner {
    background: var(--surface-card);
    border: 1px solid var(--surface-border);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-1);
    padding: 1rem;
}

.toc-title {
    font-size: 0.75rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--text-muted);
    margin-bottom: 0.75rem;
    padding: 0 0.5rem;
}

.toc-list {
    list-style: none;
    margin: 0;
    padding: 0;
}

.toc-item {
    margin-bottom: 0.2rem;
}

.toc-link {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.45rem 0.5rem;
    border-radius: var(--radius-sm);
    color: var(--text-secondary);
    text-decoration: none;
    font-size: 0.875rem;
    font-weight: 500;
    transition:
        background 0.15s ease,
        color 0.15s ease;

    .q-icon {
        color: var(--text-muted);
        flex-shrink: 0;
    }

    &:hover {
        background: var(--accent-soft);
        color: var(--accent);
    }

    &--active {
        background: var(--accent-soft-strong);
        color: var(--accent);
        font-weight: 700;

        .q-icon {
            color: var(--accent);
        }
    }
}

.toc-sublist {
    list-style: none;
    margin: 0.1rem 0 0.4rem 1.75rem;
    padding: 0;
    border-left: 1px solid var(--surface-border);
}

.toc-sublink {
    display: block;
    padding: 0.3rem 0 0.3rem 0.75rem;
    color: var(--text-muted);
    text-decoration: none;
    font-size: 0.8rem;
    transition: color 0.15s ease;

    &:hover {
        color: var(--accent);
    }

    &--active {
        color: var(--accent);
        font-weight: 600;
    }
}

// ── Mobile TOC ──────────────────────────────────────────────────────────
.mobile-toc-btn {
    display: none;
    position: fixed;
    bottom: 1.5rem;
    right: 1.5rem;
    z-index: 1000;
    box-shadow: var(--shadow-3);

    @media (max-width: 1024px) {
        display: inline-flex;
    }
}

.mobile-toc {
    width: 280px;
    max-width: 85vw;
    height: 100vh;
    background: var(--surface-card);
    padding: 1rem;
    overflow-y: auto;

    .toc-inner {
        border: none;
        box-shadow: none;
        padding: 0;
    }
}

.mobile-toc-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 0.5rem;

    .toc-title {
        margin-bottom: 0;
    }
}

// ── Content column ──────────────────────────────────────────────────────
.docs-content {
    flex: 1;
    min-width: 0;
    max-width: 900px;
    display: flex;
    flex-direction: column;
    gap: calc(var(--section-gap) * 2);
}

.doc-section {
    scroll-margin-top: 96px;
}

.docs-content :deep(.section-card) {
    scroll-margin-top: 96px;
}

.doc-section-header {
    margin-bottom: var(--section-gap);

    .doc-section-badge {
        display: inline-flex;
        align-items: center;
        gap: 0.5rem;
        background: var(--accent-soft);
        color: var(--accent);
        padding: 0.4rem 0.9rem;
        border-radius: 999px;
        font-size: 0.8rem;
        font-weight: 700;
        margin-bottom: 0.85rem;
    }

    .doc-section-title {
        font-size: 2rem;
        font-weight: 700;
        color: var(--text-primary);
        margin: 0 0 0.6rem 0;
        line-height: 1.2;

        @media (max-width: 768px) {
            font-size: 1.6rem;
        }
    }

    .doc-section-subtitle {
        font-size: 1.05rem;
        line-height: 1.6;
        color: var(--text-secondary);
        margin: 0;
    }
}

.card-description {
    line-height: 1.6;
    color: var(--text-secondary);
    margin: 0;
}

// ── Getting Started: hub cards ──────────────────────────────────────────
.hub-cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
    gap: 1.5rem;
    margin-bottom: var(--section-gap);
}

.hub-card {
    border-radius: var(--radius-lg);
    border: 1px solid var(--surface-border);
    background: var(--surface-card);
    box-shadow: var(--shadow-1);
    transition:
        transform 0.2s ease,
        box-shadow 0.2s ease;

    &:hover {
        transform: var(--lift-md);
        box-shadow: var(--shadow-2);
    }

    .hub-card-content {
        text-align: center;
        padding: 1.75rem 1.5rem;
    }

    .hub-icon {
        color: var(--accent);
        margin-bottom: 0.75rem;
    }

    h3 {
        margin: 0 0 0.5rem 0;
        font-size: 1.25rem;
        font-weight: 700;
        color: var(--text-primary);
    }

    .hub-description {
        color: var(--text-secondary);
        line-height: 1.5;
        margin: 0 0 1.1rem 0;
        font-size: 0.9rem;
    }

    .hub-features {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
        margin-bottom: 1.25rem;
        text-align: left;
    }

    .feature-item {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        color: var(--text-secondary);
        font-size: 0.85rem;

        .q-icon {
            color: var(--accent);
            flex-shrink: 0;
        }
    }

    .hub-btn {
        width: 100%;
        border-radius: var(--radius-md);
        text-transform: none;
        font-weight: 600;
    }
}

.getting-started-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 1.5rem;
}

.quick-start-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 1rem;

    li {
        display: flex;
        align-items: flex-start;
        gap: 0.85rem;
    }

    h4 {
        margin: 0 0 0.2rem 0;
        font-size: 0.9rem;
        font-weight: 700;
        color: var(--text-primary);
    }

    p {
        margin: 0;
        font-size: 0.85rem;
        color: var(--text-secondary);
        line-height: 1.5;
    }
}

.quick-start-number {
    flex-shrink: 0;
    width: 1.8rem;
    height: 1.8rem;
    border-radius: 50%;
    background: var(--accent);
    color: var(--text-on-brand);
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    font-size: 0.85rem;
}

.requirements-list {
    list-style: none;
    margin: 0 0 1rem 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.85rem;

    li {
        display: flex;
        align-items: flex-start;
        gap: 0.75rem;
    }

    h4 {
        margin: 0 0 0.15rem 0;
        font-size: 0.875rem;
        font-weight: 700;
        color: var(--text-primary);
    }

    p {
        margin: 0;
        font-size: 0.8rem;
        color: var(--text-secondary);
    }
}

.format-chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    padding-top: 0.75rem;
    border-top: 1px solid var(--surface-border);
}

.format-chip {
    display: inline-flex;
    align-items: center;
    gap: 0.3rem;
    padding: 0.25rem 0.65rem;
    border-radius: 999px;
    background: var(--surface-raised);
    border: 1px solid var(--surface-border);
    color: var(--text-secondary);
    font-size: 0.75rem;
    font-weight: 600;
}

// ── Shared building blocks reused across sections ──────────────────────
.requirement-grid,
.prerequisites-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 1.25rem;
}

.requirement-item {
    display: flex;
    align-items: flex-start;
    gap: 0.85rem;

    h4 {
        margin: 0 0 0.2rem 0;
        font-size: 0.9rem;
        font-weight: 700;
        color: var(--text-primary);
    }

    p {
        margin: 0;
        font-size: 0.8rem;
        color: var(--text-secondary);
    }
}

.prerequisite-item {
    text-align: center;
    padding: 1.1rem;
    border-radius: var(--radius-md);
    background: var(--surface-raised);
    border: 1px solid var(--surface-border);

    .prereq-icon {
        margin-bottom: 0.5rem;
    }

    h4 {
        margin: 0 0 0.35rem 0;
        font-size: 0.9rem;
        font-weight: 700;
        color: var(--text-primary);
    }

    p {
        margin: 0 0 0.6rem 0;
        font-size: 0.8rem;
        color: var(--text-secondary);
    }

    .download-link {
        color: var(--accent);
        font-size: 0.8rem;
        font-weight: 600;
        text-decoration: none;

        &:hover {
            text-decoration: underline;
        }
    }
}

.export-steps,
.setup-steps {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
}

.export-step,
.setup-step {
    display: flex;
    gap: 1rem;
}

.step-number {
    flex-shrink: 0;
    width: 2rem;
    height: 2rem;
    border-radius: 50%;
    background: var(--accent);
    color: var(--text-on-brand);
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    font-size: 0.9rem;
}

.step-content {
    flex: 1;
    min-width: 0;

    h4 {
        margin: 0 0 0.3rem 0;
        font-size: 0.95rem;
        font-weight: 700;
        color: var(--text-primary);
    }

    p {
        margin: 0 0 0.5rem 0;
        color: var(--text-secondary);
        font-size: 0.875rem;
        line-height: 1.55;
    }
}

.step-note,
.step-warning,
.calculation-note {
    display: flex;
    align-items: flex-start;
    gap: 0.4rem;
    padding: 0.55rem 0.75rem;
    border-radius: var(--radius-sm);
    font-size: 0.8rem;
    line-height: 1.4;
    margin-top: 0.5rem;
}

.step-note,
.calculation-note {
    background: var(--accent-soft);
    color: var(--text-primary);

    .q-icon {
        color: var(--accent);
        flex-shrink: 0;
    }
}

.step-warning {
    background: color-mix(in srgb, var(--q-warning, orange) 12%, transparent);
    color: var(--text-primary);

    .q-icon {
        color: var(--q-warning, orange);
        flex-shrink: 0;
    }
}

.step-link {
    margin-top: 0.6rem;
}

.command-list {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    margin-top: 0.5rem;
}

.command-block {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    background: var(--surface-raised);
    border: 1px solid var(--surface-border);
    border-radius: var(--radius-sm);
    padding: 0.4rem 0.4rem 0.4rem 0.75rem;

    code {
        font-family: 'Courier New', monospace;
        font-size: 0.8rem;
        color: var(--text-primary);
        overflow-x: auto;
        white-space: nowrap;
    }
}

.file-content {
    margin-top: 0.5rem;
    border: 1px solid var(--surface-border);
    border-radius: var(--radius-sm);
    overflow: hidden;
}

.file-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.4rem 0.75rem;
    background: var(--surface-raised);
    border-bottom: 1px solid var(--surface-border);

    .file-name {
        font-size: 0.8rem;
        font-weight: 600;
        color: var(--text-primary);
        font-family: 'Courier New', monospace;
    }
}

.file-code {
    margin: 0;
    padding: 0.85rem;
    font-family: 'Courier New', monospace;
    font-size: 0.75rem;
    line-height: 1.5;
    color: var(--text-primary);
    background: var(--surface-card);
    overflow-x: auto;
    white-space: pre;
}

.performance-tips,
.usage-tips {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.tip-item {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;

    h4 {
        margin: 0 0 0.2rem 0;
        font-size: 0.9rem;
        font-weight: 700;
        color: var(--text-primary);
    }

    p {
        margin: 0;
        font-size: 0.85rem;
        color: var(--text-secondary);
        line-height: 1.5;
    }
}

.troubleshooting-items {
    display: flex;
    flex-direction: column;
    gap: 1.1rem;
}

.trouble-item {
    padding: 0.9rem 1rem;
    background: var(--surface-raised);
    border-left: 3px solid var(--accent);
    border-radius: var(--radius-sm);

    h4 {
        margin: 0 0 0.3rem 0;
        font-size: 0.9rem;
        font-weight: 700;
        color: var(--text-primary);
    }

    p {
        margin: 0;
        font-size: 0.85rem;
        color: var(--text-secondary);
        line-height: 1.5;
    }
}

.success-card {
    :deep(.section-card__icon),
    :deep(.section-card__title) {
        color: var(--q-positive, #21ba45);
    }
}

.success-description {
    color: var(--text-secondary);
    line-height: 1.6;
    margin: 0 0 1.25rem 0;
}

.next-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
}

.action-btn {
    border-radius: var(--radius-md);
    text-transform: none;
    font-weight: 600;
}

// ── API Reference ────────────────────────────────────────────────────────
.api-endpoints {
    display: flex;
    flex-direction: column;
    gap: 0.9rem;
}

.endpoint-item {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    padding: 0.85rem;
    border-radius: var(--radius-sm);
    background: var(--surface-raised);
    border: 1px solid var(--surface-border);
}

.endpoint-method {
    flex-shrink: 0;
    padding: 0.25rem 0.6rem;
    border-radius: var(--radius-sm);
    font-size: 0.75rem;
    font-weight: 700;
    font-family: 'Courier New', monospace;
    min-width: 3.6rem;
    text-align: center;

    &.get {
        background: color-mix(in srgb, var(--q-positive, #21ba45) 15%, transparent);
        color: var(--q-positive, #21ba45);
    }

    &.post {
        background: var(--accent-soft-strong);
        color: var(--accent);
    }
}

.endpoint-details {
    min-width: 0;

    .endpoint-path {
        font-family: 'Courier New', monospace;
        font-size: 0.85rem;
        font-weight: 700;
        color: var(--text-primary);
    }

    .endpoint-description {
        margin: 0.3rem 0 0 0;
        font-size: 0.8rem;
        color: var(--text-secondary);
        line-height: 1.5;
    }
}

// ── Local Deployment: method cards ──────────────────────────────────────
.method-badge {
    display: flex;
    align-items: center;
    gap: 0.85rem;
    width: 100%;

    .q-icon {
        color: var(--accent);
        flex-shrink: 0;
    }

    h3 {
        margin: 0;
        font-size: 1.1rem;
        font-weight: 700;
        color: var(--text-primary);
    }

    p {
        margin: 0.15rem 0 0 0;
        font-size: 0.825rem;
        color: var(--text-secondary);
    }

    &.recommended {
        .q-icon {
            color: var(--q-positive, #21ba45);
        }
    }
}

.final-step {
    margin-top: 1.25rem;
    padding: 1rem;
    background: var(--accent-soft);
    border-radius: var(--radius-md);

    h4 {
        display: flex;
        align-items: center;
        gap: 0.4rem;
        margin: 0 0 0.4rem 0;
        font-size: 0.95rem;
        font-weight: 700;
        color: var(--text-primary);
    }

    p {
        margin: 0 0 0.5rem 0;
        color: var(--text-secondary);
        font-size: 0.85rem;
    }

    .access-note {
        margin-bottom: 0;
    }
}

.access-url {
    padding: 0.6rem 0.9rem;
    background: var(--surface-card);
    border: 1px solid var(--surface-border);
    border-radius: var(--radius-sm);
    margin-bottom: 0.5rem;
    font-family: 'Courier New', monospace;
    color: var(--accent);
}

.resource-links {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
    margin-bottom: 0.85rem;
}

.resource-btn {
    border-radius: var(--radius-md);
    text-transform: none;
    font-weight: 600;
}

.resource-note {
    margin: 0;
    color: var(--text-secondary);
    font-size: 0.85rem;
}

// ── Rating Calculations ──────────────────────────────────────────────────
.rating-categories {
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
}

.rating-category {
    padding: 1rem;
    border-radius: var(--radius-md);
    background: var(--surface-raised);
    border: 1px solid var(--surface-border);
}

.category-header {
    display: flex;
    flex-wrap: wrap;
    align-items: baseline;
    gap: 0.5rem;
    margin-bottom: 0.6rem;

    h4 {
        margin: 0;
        font-size: 1rem;
        font-weight: 700;
        color: var(--accent);
    }

    .category-description {
        font-size: 0.8rem;
        color: var(--text-secondary);
    }
}

.weights-explanation {
    margin-bottom: 0.75rem;
}

.weights-note {
    display: flex;
    align-items: flex-start;
    gap: 0.4rem;
    margin: 0;
    font-size: 0.8rem;
    color: var(--text-secondary);
    line-height: 1.5;

    .q-icon {
        color: var(--accent);
        flex-shrink: 0;
        margin-top: 0.1rem;
    }
}

.attribute-weights {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
}

.weight-item {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    padding: 0.4rem 0.65rem;
    background: var(--surface-card);
    border: 1px solid var(--surface-border);
    border-radius: var(--radius-sm);
    font-size: 0.75rem;

    .attribute-name {
        font-weight: 700;
        color: var(--text-primary);
    }

    .weight-value {
        color: var(--text-secondary);
    }
}

.overall-calculation,
.total-stats-breakdown,
.mbr-calculation,
.scaling-info {
    h4 {
        margin: 1rem 0 0.5rem 0;
        font-size: 0.9rem;
        font-weight: 700;
        color: var(--text-primary);
    }

    ol,
    ul {
        margin: 0;
        padding-left: 1.25rem;
        color: var(--text-secondary);
        font-size: 0.875rem;
        line-height: 1.7;
    }
}

.stats-categories {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
    gap: 1rem;
}

.stats-category {
    padding: 0.85rem;
    background: var(--surface-raised);
    border: 1px solid var(--surface-border);
    border-radius: var(--radius-sm);

    h5 {
        margin: 0 0 0.35rem 0;
        font-size: 0.825rem;
        font-weight: 700;
        color: var(--accent);
    }

    p {
        margin: 0;
        font-size: 0.775rem;
        color: var(--text-secondary);
        line-height: 1.5;
    }
}

.mbr-components {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 0.85rem;
}

.mbr-component {
    padding: 0.75rem;
    background: var(--surface-raised);
    border: 1px solid var(--surface-border);
    border-radius: var(--radius-sm);

    h5 {
        margin: 0 0 0.25rem 0;
        font-size: 0.8rem;
        font-weight: 700;
        color: var(--text-primary);
    }

    p {
        margin: 0;
        font-size: 0.775rem;
        color: var(--text-secondary);
        line-height: 1.5;
    }
}
</style>
