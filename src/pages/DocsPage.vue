<template>
    <q-page class="docs-page">
        <!-- Hero Section -->
        <div class="hero-section">
            <div class="hero-container">
                <div class="hero-content">
                    <div class="hero-badge">
                        <q-icon name="menu_book" size="1.2rem" />
                        <span>Documentation</span>
                    </div>
                    <h1 class="hero-title">
                        Complete Guide to
                        <span class="gradient-text">FMDB</span>
                    </h1>
                    <p class="hero-subtitle">
                        Everything you need to master Football Manager data
                        analysis, from getting started to advanced team
                        management strategies.
                    </p>
                    <div class="hero-actions">
                        <q-btn
                            unelevated
                            size="lg"
                            color="primary"
                            label="Get Started"
                            icon="play_arrow"
                            @click="setActiveSection('getting-started')"
                            class="hero-cta"
                        />
                        <q-btn
                            outline
                            size="lg"
                            color="primary"
                            label="API Reference"
                            icon="code"
                            @click="setActiveSection('api-reference')"
                            class="hero-secondary"
                        />
                    </div>
                </div>
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
            </div>
        </div>

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
                            <q-icon name="rocket_launch" />
                            <span>Quick Start</span>
                        </div>
                        <h1 class="section-title">Getting Started</h1>
                        <p class="section-subtitle">
                            Welcome to FMDB! This guide will help you get
                            started with using the Football Manager Database and
                            Player Analysis Tool.
                        </p>
                    </div>

                    <div class="content-cards">
                        <q-card class="info-card">
                            <q-card-section>
                                <div class="card-header">
                                    <q-icon
                                        name="speed"
                                        size="1.5rem"
                                        color="primary"
                                    />
                                    <h3>Quick Start Guide</h3>
                                </div>
                                <div class="step-list">
                                    <div
                                        class="step-item"
                                        v-for="(step, index) in quickStartSteps"
                                        :key="index"
                                    >
                                        <div class="step-number">
                                            {{ index + 1 }}
                                        </div>
                                        <div class="step-content">
                                            <h4>{{ step.title }}</h4>
                                            <p>{{ step.description }}</p>
                                        </div>
                                    </div>
                                </div>
                            </q-card-section>
                        </q-card>

                        <q-card class="info-card">
                            <q-card-section>
                                <div class="card-header">
                                    <q-icon
                                        name="computer"
                                        size="1.5rem"
                                        color="secondary"
                                    />
                                    <h3>System Requirements</h3>
                                </div>
                                <q-list class="requirement-list">
                                    <q-item
                                        v-for="req in systemRequirements"
                                        :key="req.id"
                                    >
                                        <q-item-section avatar>
                                            <q-icon
                                                :name="req.icon"
                                                color="positive"
                                            />
                                        </q-item-section>
                                        <q-item-section>
                                            <q-item-label>{{
                                                req.title
                                            }}</q-item-label>
                                            <q-item-label caption>{{
                                                req.description
                                            }}</q-item-label>
                                        </q-item-section>
                                    </q-item>
                                </q-list>
                            </q-card-section>
                        </q-card>
                    </div>
                </div>

                <!-- Player Analysis Section -->
                <div
                    v-if="activeSection === 'player-analysis'"
                    class="content-section"
                >
                    <div class="section-header">
                        <div class="section-badge">
                            <q-icon name="analytics" />
                            <span>Analysis Tools</span>
                        </div>
                        <h1 class="section-title">Player Analysis</h1>
                        <p class="section-subtitle">
                            Learn how to analyze player performance, attributes,
                            and potential using FMDB's powerful analysis tools.
                        </p>
                    </div>

                    <div class="content-cards">
                        <q-card class="info-card">
                            <q-card-section>
                                <div class="card-header">
                                    <q-icon
                                        name="star"
                                        size="1.5rem"
                                        color="amber"
                                    />
                                    <h3>Key Features</h3>
                                </div>
                                <div class="feature-grid-content">
                                    <div
                                        class="feature-item"
                                        v-for="feature in analysisFeatures"
                                        :key="feature.id"
                                    >
                                        <q-icon
                                            :name="feature.icon"
                                            size="1.2rem"
                                            :color="feature.color"
                                        />
                                        <div>
                                            <h4>{{ feature.title }}</h4>
                                            <p>{{ feature.description }}</p>
                                        </div>
                                    </div>
                                </div>
                            </q-card-section>
                        </q-card>

                        <q-card class="info-card">
                            <q-card-section>
                                <div class="card-header">
                                    <q-icon
                                        name="psychology"
                                        size="1.5rem"
                                        color="purple"
                                    />
                                    <h3>Understanding Player Ratings</h3>
                                </div>
                                <p class="card-description">
                                    Player ratings are calculated based on
                                    role-specific attribute weights and current
                                    performance metrics. Our algorithm considers
                                    position requirements, tactical fit, and
                                    statistical performance to provide accurate
                                    assessments.
                                </p>
                                <q-btn
                                    color="purple"
                                    label="Learn More"
                                    icon="arrow_forward"
                                    flat
                                    class="q-mt-md"
                                />
                            </q-card-section>
                        </q-card>
                    </div>
                </div>

                <!-- Team Management Section -->
                <div
                    v-if="activeSection === 'team-management'"
                    class="content-section"
                >
                    <div class="section-header">
                        <div class="section-badge">
                            <q-icon name="groups" />
                            <span>Team Tools</span>
                        </div>
                        <h1 class="section-title">Team Management</h1>
                        <p class="section-subtitle">
                            Organize and manage your team data effectively with
                            FMDB's comprehensive team management features.
                        </p>
                    </div>

                    <div class="content-cards">
                        <q-card class="info-card">
                            <q-card-section>
                                <div class="card-header">
                                    <q-icon
                                        name="sports_soccer"
                                        size="1.5rem"
                                        color="green"
                                    />
                                    <h3>Team View Features</h3>
                                </div>
                                <div class="feature-grid-content">
                                    <div
                                        class="feature-item"
                                        v-for="feature in teamFeatures"
                                        :key="feature.id"
                                    >
                                        <q-icon
                                            :name="feature.icon"
                                            size="1.2rem"
                                            color="green"
                                        />
                                        <div>
                                            <h4>{{ feature.title }}</h4>
                                            <p>{{ feature.description }}</p>
                                        </div>
                                    </div>
                                </div>
                            </q-card-section>
                        </q-card>

                        <q-card class="info-card">
                            <q-card-section>
                                <div class="card-header">
                                    <q-icon
                                        name="strategy"
                                        size="1.5rem"
                                        color="orange"
                                    />
                                    <h3>Formation Analysis</h3>
                                </div>
                                <p class="card-description">
                                    Analyze different formations and how your
                                    players fit into various tactical setups.
                                    Understand positional chemistry, tactical
                                    balance, and optimize your team's
                                    performance.
                                </p>
                                <q-btn
                                    color="orange"
                                    label="View Formations"
                                    icon="sports"
                                    flat
                                    class="q-mt-md"
                                />
                            </q-card-section>
                        </q-card>
                    </div>
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
                            FMDB's API endpoints and data structures.
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
import { defineComponent, ref } from "vue";

export default defineComponent({
    name: "DocsPage",
    setup() {
        const activeSection = ref("getting-started");
        const mobileSidebarOpen = ref(false);

        const docSections = [
            {
                id: "getting-started",
                title: "Getting Started",
                subtitle: "Setup and basics",
                icon: "rocket_launch",
            },
            {
                id: "player-analysis",
                title: "Player Analysis",
                subtitle: "Analyze performance",
                icon: "analytics",
            },
            {
                id: "team-management",
                title: "Team Management",
                subtitle: "Organize teams",
                icon: "groups",
            },
            {
                id: "api-reference",
                title: "API Reference",
                subtitle: "Developer docs",
                icon: "code",
            },
        ];

        const heroFeatures = [
            { id: 1, icon: "analytics", title: "Player Analysis" },
            { id: 2, icon: "groups", title: "Team Management" },
            { id: 3, icon: "sports_soccer", title: "Formation Tools" },
            { id: 4, icon: "api", title: "API Access" },
        ];

        const quickStartSteps = [
            {
                title: "Upload Your Data",
                description:
                    "Import your Football Manager player data using our secure upload system.",
            },
            {
                title: "Explore Analysis Tools",
                description:
                    "Use powerful analysis features to evaluate player performance and potential.",
            },
            {
                title: "Manage Your Team",
                description:
                    "View formations, optimize lineups, and track team performance metrics.",
            },
        ];

        const systemRequirements = [
            {
                id: 1,
                icon: "web",
                title: "Modern Web Browser",
                description:
                    "Chrome, Firefox, Safari, or Edge (latest versions)",
            },
            {
                id: 2,
                icon: "sports_soccer",
                title: "Football Manager Data",
                description: "Exported player data from FM (HTML only)",
            },
            {
                id: 3,
                icon: "memory",
                title: "4GB+ RAM",
                description: "For optimal performance with large datasets",
            },
        ];

        const analysisFeatures = [
            {
                id: 1,
                icon: "star_rate",
                color: "amber",
                title: "Attribute Weighting",
                description: "Role-specific attribute importance calculation",
            },
            {
                id: 2,
                icon: "trending_up",
                color: "green",
                title: "Performance Tracking",
                description: "Monitor player development and statistics",
            },
            {
                id: 3,
                icon: "compare_arrows",
                color: "blue",
                title: "Player Comparison",
                description: "Side-by-side analysis of multiple players",
            },
            {
                id: 4,
                icon: "location_on",
                color: "purple",
                title: "Position Analysis",
                description: "Suitability ratings for different positions",
            },
        ];

        const teamFeatures = [
            {
                id: 1,
                icon: "visibility",
                title: "Visual Formation Display",
                description: "See your team's tactical setup at a glance",
            },
            {
                id: 2,
                icon: "place",
                title: "Player Positioning",
                description: "Optimal positioning based on attributes",
            },
            {
                id: 3,
                icon: "group",
                title: "Squad Overview",
                description: "Complete squad analysis and depth charts",
            },
        ];

        const apiEndpoints = [
            {
                id: 1,
                method: "POST",
                path: "/api/upload",
                description:
                    "Upload and process Football Manager player data files",
            },
            {
                id: 2,
                method: "GET",
                path: "/api/players",
                description:
                    "Retrieve player information with filtering and pagination",
            },
            {
                id: 3,
                method: "GET",
                path: "/api/teams",
                description:
                    "Get team data including formations and statistics",
            },
            {
                id: 4,
                method: "GET",
                path: "/api/analysis",
                description: "Access advanced analysis results and metrics",
            },
        ];

        const dataFormats = ["HTML"];

        const setActiveSection = (sectionId) => {
            activeSection.value = sectionId;
        };

        return {
            activeSection,
            mobileSidebarOpen,
            docSections,
            heroFeatures,
            quickStartSteps,
            systemRequirements,
            analysisFeatures,
            teamFeatures,
            apiEndpoints,
            dataFormats,
            setActiveSection,
        };
    },
});
</script>

<style lang="scss" scoped>
.docs-page {
    padding: 0;
    background: var(--q-page);
    min-height: 100vh;
}

// Hero Section
.hero-section {
    background: linear-gradient(
        135deg,
        rgba(25, 118, 210, 0.95) 0%,
        rgba(33, 150, 243, 0.95) 100%
    );
    color: white;
    padding: 4rem 0;
    position: relative;
    overflow: hidden;

    .body--light & {
        background: linear-gradient(
            135deg,
            rgba(25, 118, 210, 0.95) 0%,
            rgba(33, 150, 243, 0.95) 100%
        );
    }

    &::before {
        content: "";
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='60' height='60' viewBox='0 0 60 60'%3E%3Cg fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='0.05'%3E%3Cpath d='m36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E");
        opacity: 0.3;
    }
}

.hero-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 3rem;
    align-items: center;
    position: relative;
    z-index: 1;

    @media (max-width: 768px) {
        grid-template-columns: 1fr;
        gap: 2rem;
        text-align: center;
    }
}

.hero-content {
    .hero-badge {
        display: inline-flex;
        align-items: center;
        gap: 0.5rem;
        background: rgba(255, 255, 255, 0.15);
        padding: 0.5rem 1rem;
        border-radius: 2rem;
        font-size: 0.875rem;
        font-weight: 500;
        margin-bottom: 1.5rem;
        backdrop-filter: blur(10px);
    }

    .hero-title {
        font-size: clamp(2.5rem, 4vw, 3.5rem);
        font-weight: 700;
        line-height: 1.1;
        margin-bottom: 1rem;

        .gradient-text {
            background: linear-gradient(135deg, #ffd700, #ff6b35);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
    }

    .hero-subtitle {
        font-size: 1.25rem;
        line-height: 1.6;
        margin-bottom: 2rem;
        opacity: 0.9;
    }

    .hero-actions {
        display: flex;
        gap: 1rem;
        flex-wrap: wrap;

        @media (max-width: 768px) {
            justify-content: center;
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

        .hero-secondary {
            padding: 0.75rem 2rem;
            border-radius: 2rem;
            font-weight: 600;
            text-transform: none;
            border: 2px solid rgba(255, 255, 255, 0.3);
            backdrop-filter: blur(10px);
            transition: all 0.3s ease;

            &:hover {
                background: rgba(255, 255, 255, 0.1);
                transform: translateY(-2px);
            }
        }
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
        background: linear-gradient(135deg, var(--q-primary), #1976d2);
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
        background: linear-gradient(135deg, var(--q-primary), #1976d2);
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
        color: #1976d2;
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
            color: #1976d2;

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
        background: linear-gradient(135deg, var(--q-primary), #1976d2);
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
            color: #1976d2;

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
        color: #1976d2;
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
            background: #2196f3;
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
</style>
