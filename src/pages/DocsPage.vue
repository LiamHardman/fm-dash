<template>
    <q-page class="docs-page">
        <!-- Hero Section -->
        <PageHero
            icon="menu_book"
            badge="Documentation"
            title="Complete Guide to"
            highlight="FM-Dash"
            subtitle="Everything you need to master Football Manager data analysis, from getting started to advanced team management strategies."
            class="docs-hero"
        >
            <template #actions>
                <q-btn
                    unelevated
                    size="lg"
                    color="white"
                    text-color="primary"
                    label="Get Started"
                    icon="play_arrow"
                    @click="setActiveSection('getting-started')"
                    class="hero-cta"
                />
                <q-btn
                    outline
                    size="lg"
                    color="white"
                    label="API Reference"
                    icon="code"
                    @click="setActiveSection('api-reference')"
                    class="hero-secondary"
                />
                <q-btn
                    outline
                    size="lg"
                    color="white"
                    label="First Time Guide"
                    icon="school"
                    @click="showTutorial"
                    class="hero-tutorial"
                />
            </template>
            <template #side>
                <div class="hero-visual">
                    <div class="feature-grid">
                        <div
                            class="feature-card"
                            v-for="feature in heroFeatures"
                            :key="feature.id"
                        >
                            <q-icon
                                :name="feature.icon"
                                size="2rem"
                                class="feature-icon"
                            />
                            <div class="feature-text">{{ feature.title }}</div>
                        </div>
                    </div>
                </div>
            </template>
        </PageHero>

        <div class="docs-container">
            <div class="docs-sidebar">
                <div class="sidebar-header">
                    <h3>Navigation</h3>
                    <q-btn
                        flat
                        round
                        icon="close"
                        size="sm"
                        @click="mobileSidebarOpen = false"
                        class="mobile-close q-ml-auto"
                        v-if="$q.screen.lt.md"
                    />
                </div>

                <q-list class="nav-list">
                    <q-item
                        v-for="section in docSections"
                        :key="section.id"
                        clickable
                        v-ripple
                        :active="activeSection === section.id"
                        @click="setActiveSection(section.id)"
                        class="doc-nav-item"
                    >
                        <q-item-section avatar>
                            <q-icon :name="section.icon" size="1.2rem" />
                        </q-item-section>
                        <q-item-section>
                            <q-item-label class="nav-title">{{
                                section.title
                            }}</q-item-label>
                            <q-item-label caption class="nav-subtitle">{{
                                section.subtitle
                            }}</q-item-label>
                        </q-item-section>
                    </q-item>
                </q-list>

                <div class="sidebar-footer"></div>
            </div>

            <!-- Mobile Menu Button -->
            <q-btn
                v-if="$q.screen.lt.md"
                fab
                icon="menu"
                color="primary"
                @click="mobileSidebarOpen = true"
                class="mobile-menu-btn"
                size="md"
            />

            <div class="docs-content">
                <!-- Getting Started Section -->
                <div
                    v-if="activeSection === 'getting-started'"
                    class="content-section"
                >
                    <div class="section-header">
                        <div class="section-badge">
                            <q-icon name="hub" />
                            <span>Documentation Hub</span>
                        </div>
                        <h1 class="section-title">Welcome to FM-Dash</h1>
                        <p class="section-subtitle">
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
                                <h2>Use FM-Dash</h2>
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
                                    @click="setActiveSection('data-export')"
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
                                <h2>Host FM-Dash</h2>
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
                                        <span>Private & secure</span>
                                    </div>
                                </div>
                                <q-btn
                                    unelevated
                                    color="secondary"
                                    label="Deploy Now"
                                    icon="rocket_launch"
                                    @click="setActiveSection('local-deployment')"
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
                                <h2>Hack FM-Dash</h2>
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
                                    @click="setActiveSection('api-reference')"
                                    class="hub-btn"
                                    size="lg"
                                />
                            </q-card-section>
                        </q-card>
                    </div>

                </div>

                <!-- Data Export Guide Section -->
                <div
                    v-if="activeSection === 'data-export'"
                    class="content-section"
                >
                    <div class="section-header">
                        <div class="section-badge">
                            <q-icon name="upload" />
                            <span>Data Export</span>
                        </div>
                        <h1 class="section-title">Export Data from FM24</h1>
                        <p class="section-subtitle">
                            Learn how to export your Football Manager 24 player data 
                            for analysis in FM-Dash. This guide will walk you through 
                            the complete process from setup to export.
                        </p>
                    </div>

                    <!-- Overview Card -->
                    <q-card class="info-card">
                        <q-card-section>
                            <div class="card-header">
                                <q-icon
                                    name="info"
                                    size="1.5rem"
                                    color="primary"
                                />
                                <h3>What You'll Need</h3>
                            </div>
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
                        </q-card-section>
                    </q-card>

                    <!-- Step-by-Step Guide -->
                    <q-card class="info-card">
                        <q-card-section>
                            <div class="card-header">
                                <q-icon
                                    name="list_alt"
                                    size="1.5rem"
                                    color="primary"
                                />
                                <h3>Step-by-Step Export Process</h3>
                            </div>
                            <div class="export-steps">
                                <div
                                    class="export-step"
                                    v-for="(step, index) in exportSteps"
                                    :key="index"
                                >
                                    <div class="step-number">
                                        {{ index + 1 }}
                                    </div>
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
                        </q-card-section>
                    </q-card>

                    <!-- Performance Tips -->
                    <q-card class="info-card">
                        <q-card-section>
                            <div class="card-header">
                                <q-icon
                                    name="speed"
                                    size="1.5rem"
                                    color="orange"
                                />
                                <h3>Performance Tips</h3>
                            </div>
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
                        </q-card-section>
                    </q-card>

                    <!-- Troubleshooting -->
                    <q-card class="info-card">
                        <q-card-section>
                            <div class="card-header">
                                <q-icon
                                    name="help_outline"
                                    size="1.5rem"
                                    color="negative"
                                />
                                <h3>Troubleshooting</h3>
                            </div>
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
                        </q-card-section>
                    </q-card>

                    <!-- Next Steps -->
                    <q-card class="info-card success-card">
                        <q-card-section>
                            <div class="card-header">
                                <q-icon
                                    name="check_circle"
                                    size="1.5rem"
                                    color="positive"
                                />
                                <h3>Ready to Analyze!</h3>
                            </div>
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
                                    @click="setActiveSection('getting-started')"
                                    class="action-btn"
                                />
                            </div>
                        </q-card-section>
                    </q-card>
                </div>

                <!-- API Reference Section -->
                <div
                    v-if="activeSection === 'api-reference'"
                    class="content-section"
                >
                    <div class="section-header">
                        <div class="section-badge">
                            <q-icon name="code" />
                            <span>Developer Tools</span>
                        </div>
                        <h1 class="section-title">API Reference</h1>
                        <p class="section-subtitle">
                            Technical documentation for developers working with
                            FM-Dash's API endpoints and data structures.
                        </p>
                    </div>

                    <div class="content-cards">
                        <q-card class="info-card">
                            <q-card-section>
                                <div class="card-header">
                                    <q-icon
                                        name="api"
                                        size="1.5rem"
                                        color="primary"
                                    />
                                    <h3>Available Endpoints</h3>
                                </div>
                                <div class="api-endpoints">
                                    <div
                                        class="endpoint-item"
                                        v-for="endpoint in apiEndpoints"
                                        :key="endpoint.id"
                                    >
                                        <div
                                            class="endpoint-method"
                                            :class="
                                                endpoint.method.toLowerCase()
                                            "
                                        >
                                            {{ endpoint.method }}
                                        </div>
                                        <div class="endpoint-details">
                                            <code class="endpoint-path">{{
                                                endpoint.path
                                            }}</code>
                                            <p class="endpoint-description">
                                                {{ endpoint.description }}
                                            </p>
                                        </div>
                                    </div>
                                </div>
                            </q-card-section>
                        </q-card>
                    </div>
                </div>

                <!-- Local Deployment Section -->
                <div
                    v-if="activeSection === 'local-deployment'"
                    class="content-section"
                >
                    <div class="section-header">
                        <div class="section-badge">
                            <q-icon name="cloud_download" />
                            <span>Self-Hosting</span>
                        </div>
                        <h1 class="section-title">Local Deployment</h1>
                        <p class="section-subtitle">
                            Run FM-Dash on your own computer for personal use. 
                            Choose between Docker (easiest) or manual setup.
                        </p>
                    </div>

                    <!-- Prerequisites -->
                    <q-card class="info-card">
                        <q-card-section>
                            <div class="card-header">
                                <q-icon
                                    name="download"
                                    size="1.5rem"
                                    color="orange"
                                />
                                <h3>Prerequisites</h3>
                            </div>
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
                                    <a v-if="requirement.downloadUrl" :href="requirement.downloadUrl" target="_blank" class="download-link">
                                        Download {{ requirement.shortName }}
                                    </a>
                                </div>
                            </div>
                        </q-card-section>
                    </q-card>

                    <!-- Option 1: Docker -->
                    <q-card class="info-card method-card docker-method">
                        <q-card-section>
                            <div class="method-header">
                                <div class="method-badge recommended">
                                    <q-icon name="smart_toy" size="1.5rem" />
                                    <div>
                                        <h2>Option 1: Docker (Recommended)</h2>
                                        <p>The easiest way to run FM-Dash. Everything is pre-configured.</p>
                                    </div>
                                </div>
                            </div>
                            
                            <div class="method-content">
                                <div class="setup-steps">
                                    <div
                                        class="setup-step"
                                        v-for="(step, index) in dockerSteps"
                                        :key="index"
                                    >
                                        <div class="step-number">
                                            {{ index + 1 }}
                                        </div>
                                        <div class="step-content">
                                            <h4>{{ step.title }}</h4>
                                            <p>{{ step.description }}</p>
                                            <div v-if="step.commands" class="command-list">
                                                <div v-for="(command, cmdIndex) in step.commands" :key="cmdIndex" class="command-block">
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
                                    <h4>✅ Access Your Application</h4>
                                    <p>Once the container is running, open your web browser and go to:</p>
                                    <div class="access-url">
                                        <strong>http://localhost:3000</strong>
                                    </div>
                                    <p class="access-note">You should see the FM-Dash interface. You can now upload your Football Manager data and start analyzing!</p>
                                </div>
                            </div>
                        </q-card-section>
                    </q-card>

                    <!-- Option 2: Manual -->
                    <q-card class="info-card method-card manual-method">
                        <q-card-section>
                            <div class="method-header">
                                <div class="method-badge">
                                    <q-icon name="construction" size="1.5rem" />
                                    <div>
                                        <h2>Option 2: Manual Installation</h2>
                                        <p>For users who prefer to build and run the application manually.</p>
                                    </div>
                                </div>
                            </div>
                            
                            <div class="method-content">
                                <div class="setup-steps">
                                    <div
                                        class="setup-step"
                                        v-for="(step, index) in setupSteps"
                                        :key="index"
                                    >
                                        <div class="step-number">
                                            {{ index + 1 }}
                                        </div>
                                        <div class="step-content">
                                            <h4>{{ step.title }}</h4>
                                            <p>{{ step.description }}</p>
                                            <div v-if="step.commands" class="command-list">
                                                <div v-for="(command, cmdIndex) in step.commands" :key="cmdIndex" class="command-block">
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
                                    <h4>✅ Access Your Application</h4>
                                    <p>Once both servers are running, open your web browser and go to:</p>
                                    <div class="access-url">
                                        <strong>http://localhost:3000</strong>
                                    </div>
                                    <p class="access-note">You should see the FM-Dash interface. You can now upload your Football Manager data and start analyzing!</p>
                                </div>
                            </div>
                        </q-card-section>
                    </q-card>

                    <!-- Help Section -->
                    <q-card class="info-card">
                        <q-card-section>
                            <div class="card-header">
                                <q-icon
                                    name="help"
                                    size="1.5rem"
                                    color="primary"
                                />
                                <h3>Need More Help?</h3>
                            </div>
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
                        </q-card-section>
                    </q-card>
                </div>

                <!-- Rating Calculations Section -->
                <div
                    v-if="activeSection === 'rating-calculations'"
                    class="content-section"
                >
                    <div class="section-header">
                        <div class="section-badge">
                            <q-icon name="calculate" />
                            <span>Rating System</span>
                        </div>
                        <h1 class="section-title">How Ratings Are Calculated</h1>
                        <p class="section-subtitle">
                            Understanding how FM-Dash calculates different player ratings and what they mean for player analysis.
                        </p>
                    </div>

                    <!-- Overview -->
                    <q-card class="info-card">
                        <q-card-section>
                            <div class="card-header">
                                <q-icon
                                    name="info"
                                    size="1.5rem"
                                    color="primary"
                                />
                                <h3>Rating System Overview</h3>
                            </div>
                            <p class="card-description">
                                FM-Dash uses a sophisticated rating system that combines Football Manager's raw attributes 
                                into meaningful categories. Each rating type serves a different purpose in player analysis.
                            </p>
                        </q-card-section>
                    </q-card>

                    <!-- FIFA-Style Ratings -->
                    <q-card class="info-card">
                        <q-card-section>
                            <div class="card-header">
                                <q-icon
                                    name="sports_soccer"
                                    size="1.5rem"
                                    color="primary"
                                />
                                <h3>FIFA-Style Category Ratings</h3>
                            </div>
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
                        </q-card-section>
                    </q-card>

                    <!-- Overall Rating -->
                    <q-card class="info-card">
                        <q-card-section>
                            <div class="card-header">
                                <q-icon
                                    name="star"
                                    size="1.5rem"
                                    color="orange"
                                />
                                <h3>Overall Rating</h3>
                            </div>
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
                        </q-card-section>
                    </q-card>

                    <!-- Total Stats -->
                    <q-card class="info-card">
                        <q-card-section>
                            <div class="card-header">
                                <q-icon
                                    name="summarize"
                                    size="1.5rem"
                                    color="green"
                                />
                                <h3>Total Stats</h3>
                            </div>
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
                        </q-card-section>
                    </q-card>

                    <!-- Moneyball Rating -->
                    <q-card class="info-card">
                        <q-card-section>
                            <div class="card-header">
                                <q-icon
                                    name="trending_up"
                                    size="1.5rem"
                                    color="purple"
                                />
                                <h3>Moneyball Rating (MBR)</h3>
                            </div>
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
                        </q-card-section>
                    </q-card>

                    <!-- Scaling Information -->
                    <q-card class="info-card">
                        <q-card-section>
                            <div class="card-header">
                                <q-icon
                                    name="tune"
                                    size="1.5rem"
                                    color="blue"
                                />
                                <h3>Rating Scaling</h3>
                            </div>
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
                        </q-card-section>
                    </q-card>

                    <!-- Usage Tips -->
                    <q-card class="info-card success-card">
                        <q-card-section>
                            <div class="card-header">
                                <q-icon
                                    name="tips_and_updates"
                                    size="1.5rem"
                                    color="positive"
                                />
                                <h3>How to Use These Ratings</h3>
                            </div>
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
                        </q-card-section>
                    </q-card>
                </div>
            </div>
        </div>

        <!-- Mobile Sidebar Overlay -->
        <q-dialog
            v-model="mobileSidebarOpen"
            position="left"
            class="mobile-sidebar-dialog"
        >
            <div class="docs-sidebar mobile-sidebar">
                <div class="sidebar-header">
                    <h3>Navigation</h3>
                    <q-btn
                        flat
                        round
                        icon="close"
                        size="sm"
                        @click="mobileSidebarOpen = false"
                        class="mobile-close q-ml-auto"
                    />
                </div>

                <q-list class="nav-list">
                    <q-item
                        v-for="section in docSections"
                        :key="section.id"
                        clickable
                        v-ripple
                        :active="activeSection === section.id"
                        @click="
                            setActiveSection(section.id);
                            mobileSidebarOpen = false;
                        "
                        class="doc-nav-item"
                    >
                        <q-item-section avatar>
                            <q-icon :name="section.icon" size="1.2rem" />
                        </q-item-section>
                        <q-item-section>
                            <q-item-label class="nav-title">{{
                                section.title
                            }}</q-item-label>
                            <q-item-label caption class="nav-subtitle">{{
                                section.subtitle
                            }}</q-item-label>
                        </q-item-section>
                    </q-item>
                </q-list>
            </div>
        </q-dialog>
    </q-page>
</template>

<script>
import { defineComponent, ref } from 'vue'
import PageHero from '@/components/PageHero.vue'
import { useUiStore } from '@/stores/uiStore'

export default defineComponent({
  name: 'DocsPage',
  components: { PageHero },
  setup() {
    const activeSection = ref('getting-started')
    const mobileSidebarOpen = ref(false)
    const uiStore = useUiStore()

    const docSections = [
      {
        id: 'getting-started',
        title: 'Getting Started',
        subtitle: 'Setup and basics',
        icon: 'rocket_launch',
      },
      {
        id: 'data-export',
        title: 'Data Export Guide',
        subtitle: 'Export from FM24',
        icon: 'upload',
      },
      {
        id: 'api-reference',
        title: 'API Reference',
        subtitle: 'Developer docs',
        icon: 'code',
      },
      {
        id: 'local-deployment',
        title: 'Local Deployment',
        subtitle: 'Self-hosting',
        icon: 'cloud_download',
      },
      {
        id: 'rating-calculations',
        title: 'Rating Calculations',
        subtitle: 'How ratings work',
        icon: 'calculate',
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

    const setActiveSection = (sectionId) => {
      activeSection.value = sectionId
    }

    const showTutorial = () => {
      uiStore.showTutorial()
    }

    return {
      activeSection,
      mobileSidebarOpen,
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
      setActiveSection,
      showTutorial,
    }
  },
})
</script>

<style lang="scss" scoped>
.docs-page {
    padding: 0;
    background: var(--q-page);
    min-height: 100vh;
}

// Hero layout/colors come from the shared PageHero component; the styles
// below only cover the slotted content (buttons + feature grid).
.docs-hero {
    // Let .docs-container's negative margin overlap the hero gradient.
    margin-bottom: 0;
}

.hero-cta {
    padding: 0.75rem 2rem;
    border-radius: 2rem;
    font-weight: 600;
    text-transform: none;
    box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);
    transition: all 0.3s ease;

    &:hover {
        transform: translateY(-2px);
        box-shadow: 0 6px 20px rgba(0, 0, 0, 0.3);
    }
}

.hero-secondary,
.hero-tutorial {
    padding: 0.75rem 2rem;
    border-radius: 2rem;
    font-weight: 600;
    text-transform: none;
    backdrop-filter: blur(10px);
    transition: all 0.3s ease;

    &:hover {
        background: rgba(255, 255, 255, 0.1);
        transform: translateY(-2px);
    }
}

.hero-visual {
    .feature-grid {
        display: grid;
        grid-template-columns: repeat(2, 1fr);
        gap: 1.5rem;

        @media (max-width: 768px) {
            grid-template-columns: repeat(2, 1fr);
            gap: 1rem;
        }
    }

    .feature-card {
        background: rgba(255, 255, 255, 0.1);
        border-radius: 1rem;
        padding: 1.5rem;
        text-align: center;
        backdrop-filter: blur(10px);
        border: 1px solid rgba(255, 255, 255, 0.2);
        transition: all 0.3s ease;

        &:hover {
            transform: translateY(-5px);
            box-shadow: 0 10px 25px rgba(0, 0, 0, 0.15);
        }

        .feature-icon {
            color: #ffd700;
            margin-bottom: 0.5rem;
        }

        .feature-text {
            font-weight: 600;
            font-size: 0.875rem;
        }
    }
}

// Main Content Area
.docs-container {
    display: flex;
    min-height: calc(100vh - 400px);
    background: var(--q-page);
    border-radius: 2rem 2rem 0 0;
    margin-top: -2rem;
    position: relative;
    z-index: 2;
    box-shadow: 0 -10px 30px rgba(0, 0, 0, 0.1);

    @media (max-width: 768px) {
        flex-direction: column;
        margin-top: -1rem;
        border-radius: 1rem 1rem 0 0;
    }
}

// Sidebar
.docs-sidebar {
    width: 320px;
    background: var(--q-dark-page);
    border-radius: 2rem 0 0 0;
    border-right: 1px solid var(--q-separator-color);
    display: flex;
    flex-direction: column;

    .body--light & {
        background: #f8fafc;
    }

    &.mobile-sidebar {
        width: 280px;
        border-radius: 0;
    }

    @media (max-width: 768px) {
        display: none;
    }
}

.sidebar-header {
    padding: 2rem 1.5rem 1rem;
    display: flex;
    align-items: center;
    border-bottom: 1px solid var(--q-separator-color);

    h3 {
        margin: 0;
        color: var(--q-primary);
        font-weight: 700;
        font-size: 1.25rem;
    }

    .mobile-close {
        margin-left: auto;
    }
}

.nav-list {
    flex: 1;
    padding: 1rem;
}

.doc-nav-item {
    border-radius: 12px;
    margin-bottom: 0.5rem;
    transition: all 0.3s ease;

    &:hover {
        background: rgba(25, 118, 210, 0.08);
        transform: translateX(4px);
    }

    &.q-item--active {
        background: linear-gradient(135deg, var(--q-primary), #3949ab);
        color: white;
        box-shadow: 0 4px 12px rgba(25, 118, 210, 0.3);

        .nav-title,
        .nav-subtitle {
            color: white;
        }
    }

    .nav-title {
        font-weight: 600;
        font-size: 0.95rem;
    }

    .nav-subtitle {
        font-size: 0.8rem;
        opacity: 0.7;
    }
}

.sidebar-footer {
    padding: 1rem;
    border-top: 1px solid var(--q-separator-color);

    .help-card {
        background: linear-gradient(135deg, var(--q-primary), #3949ab);
        color: white;
        border-radius: 12px;
        box-shadow: 0 4px 15px rgba(25, 118, 210, 0.2);
    }
}

// Mobile Menu
.mobile-menu-btn {
    position: fixed;
    bottom: 2rem;
    right: 2rem;
    z-index: 1000;
    box-shadow: 0 4px 15px rgba(0, 0, 0, 0.2);

    @media (min-width: 768px) {
        display: none !important;
    }
}

.mobile-sidebar-dialog {
    .q-dialog__inner {
        padding: 0;
    }
}

// Content Section
.docs-content {
    flex: 1;
    padding: 2rem;
    max-width: none;
    overflow-y: auto;

    @media (max-width: 768px) {
        padding: 1rem;
    }
}

.content-section {
    max-width: 900px;
    margin: 0 auto;
}

.section-header {
    margin-bottom: 3rem;

    .section-badge {
        display: inline-flex;
        align-items: center;
        gap: 0.5rem;
        background: rgba(25, 118, 210, 0.1);
        color: var(--q-primary);
        padding: 0.5rem 1rem;
        border-radius: 2rem;
        font-size: 0.875rem;
        font-weight: 600;
        margin-bottom: 1rem;
    }

    .section-title {
        font-size: 2.5rem;
        font-weight: 700;
        color: var(--accent);
        margin-bottom: 1rem;
        line-height: 1.2;

        .body--dark & {
            color: #64b5f6;
        }

        @media (max-width: 768px) {
            font-size: 2rem;
        }
    }

    .section-subtitle {
        font-size: 1.25rem;
        line-height: 1.6;
        color: #424242;
        margin: 0;

        .body--dark & {
            color: #b0b0b0;
        }
    }
}

.content-cards {
    display: grid;
    gap: 2rem;

    @media (min-width: 768px) {
        grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
    }
}

.info-card {
    border-radius: 16px;
    border: 1px solid var(--q-separator-color);
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
    transition: all 0.3s ease;
    background: var(--q-card);

    &:hover {
        transform: translateY(-5px);
        box-shadow: 0 8px 30px rgba(0, 0, 0, 0.12);
    }

    .card-header {
        display: flex;
        align-items: center;
        gap: 1rem;
        margin-bottom: 1.5rem;

        h3 {
            margin: 0;
            font-size: 1.25rem;
            font-weight: 700;
            color: var(--accent);

            .body--dark & {
                color: #64b5f6;
            }
        }
    }

    .card-description {
        line-height: 1.6;
        color: var(--q-secondary);
        margin: 0;
    }
}

// Step List
.step-list {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
}

.step-item {
    display: flex;
    align-items: flex-start;
    gap: 1rem;

    .step-number {
        background: linear-gradient(135deg, var(--q-primary), #3949ab);
        color: white;
        width: 2rem;
        height: 2rem;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: 700;
        font-size: 0.875rem;
        flex-shrink: 0;
        margin-top: 0.25rem;
    }

    .step-content {
        h4 {
            margin: 0 0 0.5rem 0;
            font-weight: 600;
            color: var(--accent);

            .body--dark & {
                color: #64b5f6;
            }
        }

        p {
            margin: 0;
            line-height: 1.5;
            color: var(--q-secondary);
        }
    }
}

// Feature Grid Content
.feature-grid-content {
    display: grid;
    gap: 1.5rem;

    @media (min-width: 768px) {
        grid-template-columns: repeat(2, 1fr);
    }
}

.feature-item {
    display: flex;
    align-items: flex-start;
    gap: 1rem;

    h4 {
        margin: 0 0 0.5rem 0;
        font-weight: 600;
        color: var(--accent);
        font-size: 1rem;

        .body--dark & {
            color: #64b5f6;
        }
    }

    p {
        margin: 0;
        line-height: 1.5;
        color: var(--q-secondary);
        font-size: 0.875rem;
    }
}

// API Endpoints
.api-endpoints {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.endpoint-item {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    padding: 1rem;
    background: rgba(25, 118, 210, 0.05);
    border-radius: 8px;
    border-left: 3px solid var(--q-primary);

    .endpoint-method {
        padding: 0.25rem 0.75rem;
        border-radius: 4px;
        font-weight: 700;
        font-size: 0.75rem;
        text-transform: uppercase;
        color: white;
        flex-shrink: 0;

        &.post {
            background: #4caf50;
        }

        &.get {
            background: var(--accent);
        }

        &.put {
            background: #ff9800;
        }

        &.delete {
            background: #f44336;
        }
    }

    .endpoint-details {
        .endpoint-path {
            font-family: "Courier New", monospace;
            background: var(--q-dark-page);
            padding: 0.25rem 0.5rem;
            border-radius: 4px;
            font-size: 0.875rem;
            display: inline-block;
            margin-bottom: 0.5rem;

            .body--light & {
                background: #f0f0f0;
            }
        }

        .endpoint-description {
            margin: 0;
            line-height: 1.5;
            color: var(--q-secondary);
            font-size: 0.875rem;
        }
    }
}

// Format Badges
.format-badges {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-top: 1rem;
}

// Requirement List
.requirement-list {
    .q-item {
        padding: 0.75rem 0;
        border-bottom: 1px solid rgba(0, 0, 0, 0.05);

        &:last-child {
            border-bottom: none;
        }
    }
}

@media (max-width: 768px) {
    .hero-container {
        padding: 0 1rem;
    }

    .content-cards {
        grid-template-columns: 1fr;
    }

    .feature-grid-content {
        grid-template-columns: 1fr;
    }

    .endpoint-item {
        flex-direction: column;
        gap: 0.5rem;
    }
}

// Local Deployment Styles
.benefit-list,
.requirement-list {
    display: grid;
    gap: 1rem;

    @media (min-width: 768px) {
        grid-template-columns: 1fr;
    }
}

.benefit-item,
.requirement-item {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    padding: 1rem;
    background: rgba(25, 118, 210, 0.05);
    border-radius: 8px;
    border-left: 3px solid var(--q-primary);

    h4 {
        margin: 0 0 0.5rem 0;
        font-weight: 600;
        color: var(--accent);
        font-size: 1rem;

        .body--dark & {
            color: #64b5f6;
        }
    }

    p {
        margin: 0;
        line-height: 1.5;
        color: var(--q-secondary);
        font-size: 0.875rem;
    }
}

.setup-steps {
    display: flex;
    flex-direction: column;
    gap: 2rem;
}

.setup-step {
    display: flex;
    align-items: flex-start;
    gap: 1rem;

    .step-number {
        background: linear-gradient(135deg, var(--q-primary), #3949ab);
        color: white;
        width: 2.5rem;
        height: 2.5rem;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: 700;
        font-size: 1rem;
        flex-shrink: 0;
        margin-top: 0.25rem;
    }

    .step-content {
        flex: 1;

        h4 {
            margin: 0 0 0.5rem 0;
            font-weight: 600;
            color: var(--accent);
            font-size: 1.1rem;

            .body--dark & {
                color: #64b5f6;
            }
        }

        p {
            margin: 0 0 1rem 0;
            line-height: 1.5;
            color: var(--q-secondary);
        }
    }
}

.command-block {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: var(--q-dark-page);
    border: 1px solid var(--q-separator-color);
    border-radius: 6px;
    padding: 0.75rem 1rem;
    margin: 0.5rem 0;
    font-family: 'Courier New', monospace;
    position: relative;

    .body--light & {
        background: #f8f9fa;
    }

    code {
        flex: 1;
        font-size: 0.875rem;
        color: var(--q-primary);
        font-weight: 500;
    }

    .copy-btn {
        opacity: 0.7;
        transition: opacity 0.2s ease;

        &:hover {
            opacity: 1;
        }
    }
}

.step-note {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: rgba(76, 175, 80, 0.1);
    border: 1px solid rgba(76, 175, 80, 0.3);
    border-radius: 6px;
    padding: 0.5rem 0.75rem;
    font-size: 0.875rem;
    color: #2e7d32;
    margin: 0.5rem 0;

    .body--dark & {
        color: #81c784;
    }
}

.launch-options {
    display: grid;
    gap: 1.5rem;
    margin-top: 1rem;

    @media (min-width: 768px) {
        grid-template-columns: repeat(2, 1fr);
    }
}

.launch-option {
    padding: 1.5rem;
    background: rgba(25, 118, 210, 0.05);
    border-radius: 12px;
    border: 1px solid rgba(25, 118, 210, 0.2);

    h4 {
        margin: 0 0 1rem 0;
        font-weight: 600;
        color: var(--accent);
        font-size: 1rem;

        .body--dark & {
            color: #64b5f6;
        }
    }

    p {
        margin: 0.5rem 0 0 0;
        line-height: 1.5;
        color: var(--q-secondary);
        font-size: 0.875rem;
    }
}

.troubleshooting-list {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
}

.trouble-item {
    padding: 1.5rem;
    background: rgba(244, 67, 54, 0.05);
    border-radius: 8px;
    border-left: 3px solid #f44336;

    h4 {
        margin: 0 0 0.5rem 0;
        font-weight: 600;
        color: #d32f2f;
        font-size: 1rem;

        .body--dark & {
            color: #f48fb1;
        }
    }

    p {
        margin: 0 0 1rem 0;
        line-height: 1.5;
        color: var(--q-secondary);
        font-size: 0.875rem;
    }
}

.resource-links {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    margin-bottom: 1rem;

    .resource-btn {
        flex: 1;
        min-width: 150px;
        text-transform: none;
        border-radius: 8px;
        padding: 0.75rem 1.5rem;
        font-weight: 500;
    }

    @media (max-width: 768px) {
        flex-direction: column;

        .resource-btn {
            width: 100%;
        }
    }
}

.resource-note {
    margin: 0;
    line-height: 1.6;
    color: var(--q-secondary);
    font-size: 0.875rem;
    text-align: center;
    padding: 1rem;
    background: rgba(25, 118, 210, 0.05);
    border-radius: 8px;
}

.download-link {
    display: inline-flex;
    align-items: center;
    color: var(--q-primary);
    text-decoration: none;
    font-weight: 500;
    font-size: 0.875rem;
    margin-top: 0.5rem;
    
    &:hover {
        text-decoration: underline;
    }
}

.command-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin: 0.5rem 0;
}

.manual-steps {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    margin-top: 1rem;
}

.manual-step {
    .step-label {
        font-weight: 600;
        color: var(--q-primary);
        font-size: 0.875rem;
        display: block;
        margin-bottom: 0.5rem;
    }
}

.launch-option {
    &.recommended {
        border: 2px solid var(--q-primary);
        position: relative;
        
        &::before {
            content: "RECOMMENDED";
            position: absolute;
            top: -8px;
            right: 1rem;
            background: var(--q-primary);
            color: white;
            padding: 0.25rem 0.75rem;
            border-radius: 4px;
            font-size: 0.7rem;
            font-weight: 700;
        }
    }
}

.option-note {
    background: rgba(25, 118, 210, 0.1);
    border-radius: 6px;
    padding: 0.75rem;
    margin-top: 0.5rem;
    font-size: 0.875rem;
    border-left: 3px solid var(--q-primary);
}

.prerequisites-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 1.5rem;
    margin-top: 1rem;

    @media (max-width: 768px) {
        grid-template-columns: repeat(2, 1fr);
        gap: 1rem;
    }

    @media (max-width: 480px) {
        grid-template-columns: 1fr;
    }
}

.prerequisite-item {
    background: rgba(25, 118, 210, 0.05);
    border-radius: 12px;
    padding: 1.5rem 1rem;
    text-align: center;
    border: 1px solid rgba(25, 118, 210, 0.2);
    transition: all 0.3s ease;

    &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 15px rgba(25, 118, 210, 0.15);
    }

    .prereq-icon {
        margin-bottom: 1rem;
        display: block;
    }

    h4 {
        margin: 0 0 0.75rem 0;
        font-weight: 600;
        color: var(--accent);
        font-size: 1rem;
        line-height: 1.3;

        .body--dark & {
            color: #64b5f6;
        }
    }

    p {
        margin: 0 0 1rem 0;
        line-height: 1.4;
        color: var(--q-secondary);
        font-size: 0.875rem;
    }
}

.file-content {
    margin: 1rem 0;
    border: 1px solid var(--q-separator-color);
    border-radius: 8px;
    overflow: hidden;
    background: var(--q-dark-page);

    .body--light & {
        background: #f8f9fa;
    }
}

.file-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem 1rem;
    background: rgba(25, 118, 210, 0.1);
    border-bottom: 1px solid var(--q-separator-color);

    .file-name {
        font-weight: 600;
        color: var(--q-primary);
        font-size: 0.875rem;
    }

    .copy-btn {
        opacity: 0.7;
        transition: opacity 0.2s ease;

        &:hover {
            opacity: 1;
        }
    }
}

.file-code {
    margin: 0;
    padding: 1rem;
    font-family: 'Monaco', 'Courier New', monospace;
    font-size: 0.8rem;
    line-height: 1.5;
    color: var(--q-primary);
    background: transparent;
    overflow-x: auto;
    white-space: pre-wrap;
    word-wrap: break-word;
}

.method-card {
    margin: 2rem 0;
    border: 2px solid transparent;
    
    &.docker-method {
        border-color: var(--accent);
        background: linear-gradient(135deg, rgba(33, 150, 243, 0.05), rgba(33, 150, 243, 0.02));
    }
    
    &.manual-method {
        border-color: #4caf50;
        background: linear-gradient(135deg, rgba(76, 175, 80, 0.05), rgba(76, 175, 80, 0.02));
    }
}

.method-header {
    margin-bottom: 2rem;
    
    .method-badge {
        display: flex;
        align-items: center;
        gap: 1rem;
        padding: 1.5rem;
        border-radius: 12px;
        position: relative;
        
        &.recommended {
            background: linear-gradient(135deg, rgba(33, 150, 243, 0.1), rgba(33, 150, 243, 0.05));
            border: 2px solid var(--accent);
            
            &::after {
                content: "RECOMMENDED";
                position: absolute;
                top: -8px;
                right: 1rem;
                background: var(--accent);
                color: white;
                padding: 0.25rem 0.75rem;
                border-radius: 4px;
                font-size: 0.7rem;
                font-weight: 700;
            }
        }
        
        &:not(.recommended) {
            background: linear-gradient(135deg, rgba(76, 175, 80, 0.1), rgba(76, 175, 80, 0.05));
            border: 2px solid #4caf50;
        }
        
        .q-icon {
            color: inherit;
        }
        
        h2 {
            margin: 0 0 0.5rem 0;
            font-size: 1.5rem;
            font-weight: 700;
            color: var(--accent);
            
            .body--dark & {
                color: #64b5f6;
            }
            
            .manual-method & {
                color: #2e7d32;
                
                .body--dark & {
                    color: #81c784;
                }
            }
        }
        
        p {
            margin: 0;
            color: var(--q-secondary);
            font-size: 1rem;
        }
    }
}

.method-content {
    padding: 0 1rem;
}

.final-step {
    margin-top: 2rem;
    padding: 1.5rem;
    background: linear-gradient(135deg, rgba(76, 175, 80, 0.1), rgba(76, 175, 80, 0.05));
    border-radius: 12px;
    border: 1px solid rgba(76, 175, 80, 0.3);
    
    h4 {
        margin: 0 0 1rem 0;
        font-size: 1.1rem;
        font-weight: 600;
        color: #2e7d32;
        display: flex;
        align-items: center;
        gap: 0.5rem;
        
        .body--dark & {
            color: #81c784;
        }
    }
    
    p {
        margin: 0 0 1rem 0;
        line-height: 1.6;
        color: var(--q-secondary);
        
        &.access-note {
            margin: 1rem 0 0 0;
            font-size: 0.9rem;
            font-style: italic;
        }
    }
}

.access-url {
    background: var(--q-dark-page);
    border: 1px solid var(--q-separator-color);
    border-radius: 8px;
    padding: 1rem;
    text-align: center;
    font-family: 'Monaco', 'Courier New', monospace;
    font-size: 1.1rem;
    margin: 1rem 0;
    
    .body--light & {
        background: #f8f9fa;
    }
    
    strong {
        color: var(--q-primary);
    }
}

// Data Export Guide Styles
.requirement-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 1.5rem;
    margin-top: 1rem;

    @media (max-width: 768px) {
        grid-template-columns: 1fr;
        gap: 1rem;
    }
}

.requirement-item {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    padding: 1rem;
    background: rgba(25, 118, 210, 0.05);
    border-radius: 8px;
    border-left: 3px solid var(--q-primary);
    transition: all 0.3s ease;

    &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 15px rgba(25, 118, 210, 0.15);
    }

    h4 {
        margin: 0 0 0.5rem 0;
        font-weight: 600;
        color: var(--accent);
        font-size: 1rem;

        .body--dark & {
            color: #64b5f6;
        }
    }

    p {
        margin: 0;
        line-height: 1.4;
        color: var(--q-secondary);
        font-size: 0.875rem;
    }
}

.export-steps {
    display: flex;
    flex-direction: column;
    gap: 2rem;
}

.export-step {
    display: flex;
    align-items: flex-start;
    gap: 1rem;

    .step-number {
        background: linear-gradient(135deg, var(--q-primary), #3949ab);
        color: white;
        width: 2.5rem;
        height: 2.5rem;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: 700;
        font-size: 1rem;
        flex-shrink: 0;
        margin-top: 0.25rem;
    }

    .step-content {
        flex: 1;

        h4 {
            margin: 0 0 0.5rem 0;
            font-weight: 600;
            color: var(--accent);
            font-size: 1.1rem;

            .body--dark & {
                color: #64b5f6;
            }
        }

        p {
            margin: 0 0 1rem 0;
            line-height: 1.5;
            color: var(--q-secondary);
        }
    }
}

.step-warning {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: rgba(255, 152, 0, 0.1);
    border: 1px solid rgba(255, 152, 0, 0.3);
    border-radius: 6px;
    padding: 0.5rem 0.75rem;
    font-size: 0.875rem;
    color: #ef6c00;
    margin: 0.5rem 0;

    .body--dark & {
        color: #ffb74d;
    }
}

.step-link {
    margin-top: 0.75rem;
}

.performance-tips {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.tip-item {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    padding: 1rem;
    background: rgba(255, 152, 0, 0.05);
    border-radius: 8px;
    border-left: 3px solid #ff9800;

    h4 {
        margin: 0 0 0.5rem 0;
        font-weight: 600;
        color: #ef6c00;
        font-size: 1rem;

        .body--dark & {
            color: #ffb74d;
        }
    }

    p {
        margin: 0;
        line-height: 1.4;
        color: var(--q-secondary);
        font-size: 0.875rem;
    }
}

.troubleshooting-items {
    display: flex;
    flex-direction: column;
    gap: 1rem;
}

.success-card {
    background: linear-gradient(135deg, rgba(76, 175, 80, 0.1), rgba(76, 175, 80, 0.05));
    border: 1px solid rgba(76, 175, 80, 0.3);
}

.success-description {
    line-height: 1.6;
    color: var(--q-secondary);
    margin-bottom: 1.5rem;
}

.next-actions {
    display: flex;
    gap: 1rem;
    flex-wrap: wrap;

    @media (max-width: 768px) {
        flex-direction: column;
    }

    .action-btn {
        flex: 1;
        min-width: 150px;
        text-transform: none;
        border-radius: 8px;
        padding: 0.75rem 1.5rem;
        font-weight: 500;
    }
}

// Documentation Hub Styles
.hub-cards {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    align-items: stretch;
    gap: 2rem;
    margin-bottom: 3rem;

    @media (max-width: 1024px) {
        grid-template-columns: 1fr;
        gap: 1.5rem;
    }
}

.hub-card {
    border-radius: 16px;
    border: 2px solid transparent;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
    transition: all 0.3s ease;
    background: var(--q-card);
    overflow: hidden;
    position: relative;
    display: flex;
    flex-direction: column;

    &:hover {
        transform: translateY(-5px);
        box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15);
    }

    &.use-card {
        border-color: var(--accent);
        background: linear-gradient(135deg, rgba(25, 118, 210, 0.05), rgba(25, 118, 210, 0.02));

        &:hover {
            border-color: var(--accent);
            box-shadow: 0 12px 40px rgba(25, 118, 210, 0.2);
        }

        .hub-icon {
            color: var(--accent);
        }
    }

    &.host-card {
        border-color: #9c27b0;
        background: linear-gradient(135deg, rgba(156, 39, 176, 0.05), rgba(156, 39, 176, 0.02));

        &:hover {
            border-color: #9c27b0;
            box-shadow: 0 12px 40px rgba(156, 39, 176, 0.2);
        }

        .hub-icon {
            color: #9c27b0;
        }
    }

    &.hack-card {
        border-color: #4caf50;
        background: linear-gradient(135deg, rgba(76, 175, 80, 0.05), rgba(76, 175, 80, 0.02));

        &:hover {
            border-color: #4caf50;
            box-shadow: 0 12px 40px rgba(76, 175, 80, 0.2);
        }

        .hub-icon {
            color: #4caf50;
        }
    }
}

.hub-card-content {
    padding: 2rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    height: 100%;
}

.hub-icon {
    margin-bottom: 1.5rem;
    opacity: 0.9;
}

.hub-card h2 {
    margin: 0 0 1rem 0;
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--accent);

    .body--dark & {
        color: #64b5f6;
    }

    .host-card & {
        color: #9c27b0;

        .body--dark & {
            color: #ce93d8;
        }
    }

    .hack-card & {
        color: #4caf50;

        .body--dark & {
            color: #81c784;
        }
    }
}

.hub-description {
    line-height: 1.6;
    color: var(--q-secondary);
    margin-bottom: 1.5rem;
    flex-grow: 1;
}

.hub-features {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    margin-bottom: 2rem;
    width: 100%;

    .feature-item {
        display: flex;
        align-items: center;
        gap: 0.75rem;
        font-size: 0.875rem;
        color: var(--q-secondary);

        .q-icon {
            opacity: 0.7;
        }
    }
}

.hub-btn {
    text-transform: none;
    border-radius: 12px;
    padding: 0.75rem 2rem;
    font-weight: 600;
    width: 100%;
    max-width: 200px;
}

// Stats Card
.stats-card {
    background: linear-gradient(135deg, rgba(25, 118, 210, 0.05), rgba(25, 118, 210, 0.02));
    border: 1px solid rgba(25, 118, 210, 0.2);
}

.stats-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 2rem;

    @media (max-width: 768px) {
        grid-template-columns: repeat(2, 1fr);
        gap: 1rem;
    }

    @media (max-width: 480px) {
        grid-template-columns: 1fr;
    }
}

.stat-item {
    text-align: center;

    .stat-number {
        font-size: 2rem;
        font-weight: 700;
        color: var(--q-primary);
        margin-bottom: 0.5rem;
    }

    .stat-label {
        font-size: 0.875rem;
        color: var(--q-secondary);
        font-weight: 500;
    }
}

// Rating Calculations Styles
.weights-explanation {
    margin-bottom: 1rem;
}

.weights-note {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: rgba(255, 152, 0, 0.1);
    border: 1px solid rgba(255, 152, 0, 0.3);
    border-radius: 6px;
    padding: 0.75rem;
    font-size: 0.875rem;
    color: #ef6c00;
    margin: 0;

    .body--dark & {
        color: #ffb74d;
    }

    strong {
        font-weight: 600;
    }
}

.rating-categories {
    display: grid;
    gap: 1.5rem;
    margin-top: 1.5rem;
}

.rating-category {
    background: rgba(25, 118, 210, 0.05);
    border-radius: 12px;
    padding: 1.5rem;
    border: 1px solid rgba(25, 118, 210, 0.2);
    transition: all 0.3s ease;

    &:hover {
        transform: translateY(-2px);
        box-shadow: 0 4px 15px rgba(25, 118, 210, 0.15);
    }
}

.category-header {
    margin-bottom: 1rem;

    h4 {
        margin: 0 0 0.5rem 0;
        font-weight: 700;
        color: var(--accent);
        font-size: 1.1rem;

        .body--dark & {
            color: #64b5f6;
        }
    }

    .category-description {
        color: var(--q-secondary);
        font-size: 0.875rem;
        font-style: italic;
    }
}

.attribute-weights {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 0.75rem;
}

.weight-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem 0.75rem;
    background: rgba(255, 255, 255, 0.5);
    border-radius: 6px;
    font-size: 0.875rem;

    .body--dark & {
        background: rgba(255, 255, 255, 0.1);
    }

    .attribute-name {
        color: var(--q-secondary);
    }

    .weight-value {
        font-weight: 600;
        color: var(--q-primary);
        background: rgba(25, 118, 210, 0.1);
        padding: 0.25rem 0.5rem;
        border-radius: 4px;
        min-width: 2rem;
        text-align: center;
    }
}

.overall-calculation,
.total-stats-breakdown,
.mbr-calculation,
.scaling-info {
    margin-top: 1.5rem;

    h4 {
        margin: 0 0 1rem 0;
        font-weight: 600;
        color: var(--accent);
        font-size: 1rem;

        .body--dark & {
            color: #64b5f6;
        }
    }

    ol, ul {
        margin: 0 0 1rem 0;
        padding-left: 1.5rem;

        li {
            margin-bottom: 0.5rem;
            line-height: 1.5;
            color: var(--q-secondary);
        }
    }
}

.calculation-note {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: rgba(76, 175, 80, 0.1);
    border: 1px solid rgba(76, 175, 80, 0.3);
    border-radius: 6px;
    padding: 0.75rem;
    font-size: 0.875rem;
    color: #2e7d32;
    margin-top: 1rem;

    .body--dark & {
        color: #81c784;
    }
}

.stats-categories {
    display: grid;
    gap: 1rem;
    margin: 1rem 0;

    @media (min-width: 768px) {
        grid-template-columns: repeat(3, 1fr);
    }
}

.stats-category {
    background: rgba(25, 118, 210, 0.05);
    border-radius: 8px;
    padding: 1rem;
    border-left: 3px solid var(--q-primary);

    h5 {
        margin: 0 0 0.5rem 0;
        font-weight: 600;
        color: var(--accent);
        font-size: 0.95rem;

        .body--dark & {
            color: #64b5f6;
        }
    }

    p {
        margin: 0;
        line-height: 1.4;
        color: var(--q-secondary);
        font-size: 0.8rem;
    }
}

.mbr-components {
    display: grid;
    gap: 1rem;
    margin: 1rem 0;

    @media (min-width: 768px) {
        grid-template-columns: repeat(2, 1fr);
    }
}

.mbr-component {
    background: rgba(156, 39, 176, 0.05);
    border-radius: 8px;
    padding: 1rem;
    border-left: 3px solid #9c27b0;

    h5 {
        margin: 0 0 0.5rem 0;
        font-weight: 600;
        color: #9c27b0;
        font-size: 0.95rem;

        .body--dark & {
            color: #ce93d8;
        }
    }

    p {
        margin: 0;
        line-height: 1.4;
        color: var(--q-secondary);
        font-size: 0.875rem;
    }
}

.usage-tips {
    display: grid;
    gap: 1rem;
    margin-top: 1rem;

    @media (min-width: 768px) {
        grid-template-columns: repeat(2, 1fr);
    }
}

.usage-tips .tip-item {
    background: rgba(76, 175, 80, 0.05);
    border-radius: 8px;
    padding: 1rem;
    border-left: 3px solid #4caf50;

    h4 {
        margin: 0 0 0.5rem 0;
        font-weight: 600;
        color: #2e7d32;
        font-size: 0.95rem;

        .body--dark & {
            color: #81c784;
        }
    }

    p {
        margin: 0;
        line-height: 1.4;
        color: var(--q-secondary);
        font-size: 0.875rem;
    }
}
</style>
