<template>
    <q-page class="docs-page">
        <div class="docs-container">
            <div class="docs-sidebar">
                <h3>Documentation</h3>
                <q-list>
                    <q-item
                        v-for="section in docSections"
                        :key="section.id"
                        clickable
                        v-ripple
                        :active="activeSection === section.id"
                        @click="setActiveSection(section.id)"
                        class="doc-nav-item"
                    >
                        <q-item-section>
                            <q-item-label>{{ section.title }}</q-item-label>
                        </q-item-section>
                    </q-item>
                </q-list>
            </div>
            
            <div class="docs-content">
                <div v-if="activeSection === 'getting-started'">
                    <h1>Getting Started</h1>
                    <p>Welcome to FMDB! This guide will help you get started with using the Football Manager Database and Player Analysis Tool.</p>
                    
                    <h2>Quick Start</h2>
                    <ol>
                        <li>Upload your Football Manager player data</li>
                        <li>Use the analysis tools to evaluate players</li>
                        <li>View team formations and player positions</li>
                    </ol>
                    
                    <h2>System Requirements</h2>
                    <ul>
                        <li>Modern web browser (Chrome, Firefox, Safari, Edge)</li>
                        <li>Football Manager save files or exported data</li>
                    </ul>
                </div>
                
                <div v-if="activeSection === 'player-analysis'">
                    <h1>Player Analysis</h1>
                    <p>Learn how to analyze player performance, attributes, and potential using FMDB's powerful analysis tools.</p>
                    
                    <h2>Key Features</h2>
                    <ul>
                        <li>Attribute weighting and role-specific ratings</li>
                        <li>Performance statistics tracking</li>
                        <li>Player comparison tools</li>
                        <li>Position suitability analysis</li>
                    </ul>
                    
                    <h2>Understanding Player Ratings</h2>
                    <p>Player ratings are calculated based on role-specific attribute weights and current performance metrics.</p>
                </div>
                
                <div v-if="activeSection === 'team-management'">
                    <h1>Team Management</h1>
                    <p>Organize and manage your team data effectively with FMDB's team management features.</p>
                    
                    <h2>Team View</h2>
                    <ul>
                        <li>Visual formation display</li>
                        <li>Player positioning</li>
                        <li>Squad overview</li>
                    </ul>
                    
                    <h2>Formation Analysis</h2>
                    <p>Analyze different formations and how your players fit into various tactical setups.</p>
                </div>
                
                <div v-if="activeSection === 'api-reference'">
                    <h1>API Reference</h1>
                    <p>Technical documentation for developers working with FMDB's API endpoints.</p>
                    
                    <h2>Available Endpoints</h2>
                    <ul>
                        <li><code>POST /api/upload</code> - Upload player data</li>
                        <li><code>GET /api/players</code> - Retrieve player information</li>
                        <li><code>GET /api/teams</code> - Get team data</li>
                    </ul>
                    
                    <h2>Data Formats</h2>
                    <p>FMDB supports various Football Manager export formats including CSV and XML.</p>
                </div>
            </div>
        </div>
    </q-page>
</template>

<script>
import { defineComponent, ref } from "vue";

export default defineComponent({
    name: "DocsPage",
    setup() {
        const activeSection = ref("getting-started");
        
        const docSections = [
            { id: "getting-started", title: "Getting Started" },
            { id: "player-analysis", title: "Player Analysis" },
            { id: "team-management", title: "Team Management" },
            { id: "api-reference", title: "API Reference" },
        ];
        
        const setActiveSection = (sectionId) => {
            activeSection.value = sectionId;
        };
        
        return {
            activeSection,
            docSections,
            setActiveSection,
        };
    },
});
</script>

<style lang="scss" scoped>
.docs-page {
    padding: 0;
}

.docs-container {
    display: flex;
    min-height: calc(100vh - 100px);
}

.docs-sidebar {
    width: 280px;
    background: var(--q-dark-page);
    border-right: 1px solid var(--q-separator-color);
    padding: 2rem 1rem;
    
    .body--light & {
        background: #f5f5f5;
    }
    
    h3 {
        margin: 0 0 1rem 0;
        color: var(--q-primary);
        font-weight: 600;
    }
}

.doc-nav-item {
    border-radius: 8px;
    margin-bottom: 0.5rem;
    
    &.q-item--active {
        background: var(--q-primary);
        color: white;
    }
}

.docs-content {
    flex: 1;
    padding: 2rem;
    max-width: 800px;
    
    h1 {
        color: var(--q-primary);
        margin-bottom: 1rem;
    }
    
    h2 {
        margin: 2rem 0 1rem 0;
        color: var(--q-secondary);
    }
    
    p {
        line-height: 1.6;
        margin-bottom: 1rem;
    }
    
    ul, ol {
        margin-bottom: 1rem;
        padding-left: 1.5rem;
        
        li {
            margin-bottom: 0.5rem;
            line-height: 1.5;
        }
    }
    
    code {
        background: var(--q-dark-page);
        padding: 2px 6px;
        border-radius: 4px;
        font-family: 'Courier New', monospace;
        
        .body--light & {
            background: #f0f0f0;
        }
    }
}

@media (max-width: 768px) {
    .docs-container {
        flex-direction: column;
    }
    
    .docs-sidebar {
        width: 100%;
        padding: 1rem;
    }
    
    .docs-content {
        padding: 1rem;
    }
}
</style>