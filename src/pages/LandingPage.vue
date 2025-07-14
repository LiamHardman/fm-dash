<template>
    <q-page class="landing-page">
        <div class="main-container">
            <!-- Left Content Panel -->
            <div class="content-panel">
                <div class="hero-content">
                    <h1 class="hero-title">
                        Revolutionize Your
                        <span class="gradient-text">Scouting Approach</span>
                    </h1>
                    <p class="hero-subtitle">
                        Advanced player analysis and team optimization. Upload your data and discover insights that will transform your tactical approach.
                    </p>
                    <div class="hero-actions">
                        <q-btn
                            unelevated
                            color="primary"
                            size="lg"
                            @click="navigateToUpload"
                            class="primary-action-btn"
                            no-caps
                        >
                            <q-icon name="upload" class="q-mr-sm" />
                            Start Analysis
                        </q-btn>
                        <q-btn
                            outline
                            color="primary"
                            size="md"
                            @click="navigateToDemo"
                            class="secondary-action-btn"
                            no-caps
                        >
                            <q-icon name="play_arrow" class="q-mr-xs" />
                            View Demo
                        </q-btn>
                    </div>
                </div>
            </div>

            <!-- Right Features Panel -->
            <div class="features-panel">
                <div class="features-header">
                    <h3 class="features-title">Complete FM24 Analysis Suite</h3>
                </div>

                <div class="features-grid">
                    <div class="feature-card">
                        <div class="feature-icon">
                            <q-icon name="upload" size="1.5rem" color="primary" />
                        </div>
                        <div class="feature-content">
                            <h4 class="feature-title">Data Upload</h4>
                            <p class="feature-description">Upload your FM24 HTML exports for instant analysis</p>
                        </div>
                    </div>
                    
                    <div class="feature-card">
                        <div class="feature-icon">
                            <q-icon name="search" size="1.5rem" color="secondary" />
                        </div>
                        <div class="feature-content">
                            <h4 class="feature-title">Player Search</h4>
                            <p class="feature-description">Advanced filtering by position, age, stats, and more</p>
                        </div>
                    </div>
                    
                    <div class="feature-card">
                        <div class="feature-icon">
                            <q-icon name="stars" size="1.5rem" color="accent" />
                        </div>
                        <div class="feature-content">
                            <h4 class="feature-title">Scouting Tools</h4>
                            <p class="feature-description">Find wonderkids, bargains, and player upgrades</p>
                        </div>
                    </div>
                    
                    <div class="feature-card">
                        <div class="feature-icon">
                            <q-icon name="flag" size="1.5rem" color="info" />
                        </div>
                        <div class="feature-content">
                            <h4 class="feature-title">Nation Analysis</h4>
                            <p class="feature-description">Scout out other national teams.</p>
                        </div>
                    </div>

                    <div class="feature-card">
                        <div class="feature-icon">
                            <q-icon name="emoji_events" size="1.5rem" color="positive" />
                        </div>
                        <div class="feature-content">
                            <h4 class="feature-title">League Analysis</h4>
                            <p class="feature-description">Browse teams and players by competition</p>
                        </div>
                    </div>
                    
                    <div class="feature-card">
                        <div class="feature-icon">
                            <q-icon name="favorite" size="1.5rem" color="warning" />
                        </div>
                        <div class="feature-content">
                            <h4 class="feature-title">Wishlist System</h4>
                            <p class="feature-description">Save and track players across datasets</p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </q-page>
</template>

<script>
import { defineComponent } from 'vue'
import { useRouter } from 'vue-router'

export default defineComponent({
  name: 'LandingPage',
  setup() {
    const router = useRouter()

    const navigateToUpload = () => {
      router.push('/upload')
    }

    const navigateToDemo = () => {
      router.push('/dataset/1e0c8dcd-f6b8-4874-a72e-a2a3bdf20038')
    }

    return {
      navigateToUpload,
      navigateToDemo
    }
  }
})
</script>

<style lang="scss" scoped>
.landing-page {
    min-height: 100vh;
    height: 100vh;
    overflow: hidden;
    background: linear-gradient(135deg, #1a237e 0%, #283593 50%, #3949ab 100%);
    color: white;
    position: relative;
    
    &::before {
        content: "";
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: radial-gradient(
            circle at 30% 20%,
            rgba(255, 255, 255, 0.05) 0%,
            transparent 50%
        );
        pointer-events: none;
    }
    
    .body--dark & {
        background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
    }
}

.main-container {
    display: flex;
    height: 100vh;
    max-width: 1400px;
    margin: 0 auto;
    padding: 3rem;
    gap: 4rem;
    align-items: center;
    position: relative;
    z-index: 1;
}

.content-panel {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2rem 0;
}

.hero-content {
    max-width: 500px;
    
    .hero-badge {
        display: inline-flex;
        align-items: center;
        gap: 0.5rem;
        background: rgba(255, 255, 255, 0.1);
        padding: 0.5rem 1rem;
        border-radius: 20px;
        font-size: 0.85rem;
        font-weight: 500;
        margin-bottom: 1.5rem;
        backdrop-filter: blur(10px);
    }
    
    .hero-title {
        font-size: 3rem;
        font-weight: 700;
        line-height: 1.1;
        margin: 0 0 1rem 0;
        
        .gradient-text {
            background: linear-gradient(135deg, #64b5f6 0%, #42a5f5 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
    }
    
    .hero-subtitle {
        font-size: 1.1rem;
        line-height: 1.5;
        margin: 0 0 2rem 0;
        opacity: 0.9;
        font-weight: 300;
    }
    
    .hero-actions {
        display: flex;
        gap: 1rem;
        flex-wrap: wrap;
        
        .primary-action-btn {
            padding: 0.8rem 1.8rem;
            font-weight: 600;
            letter-spacing: 0.5px;
            border-radius: 12px;
            transition: all 0.3s ease;
            
            &:hover {
                transform: translateY(-2px);
                box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
            }
        }
        
        .secondary-action-btn {
            padding: 0.8rem 1.5rem;
            font-weight: 600;
            letter-spacing: 0.5px;
            border-radius: 12px;
            border-width: 2px;
            transition: all 0.3s ease;
            
            &:hover {
                transform: translateY(-2px);
                background: rgba(255, 255, 255, 0.1);
            }
        }
    }
}

.features-panel {
    flex: 1;
    padding: 2rem 0;
    height: 100vh;
    display: flex;
    flex-direction: column;
    justify-content: center;
}

.features-header {
    text-align: center;
    margin-bottom: 2rem;
    
    .features-title {
        font-size: 2rem;
        font-weight: 700;
        color: rgba(255, 255, 255, 0.95);
        margin: 0 0 1rem 0;
        
        .body--dark & {
            color: rgba(255, 255, 255, 0.9);
        }
    }
}

.features-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 1.2rem;
    margin-bottom: 2rem;
    align-items: stretch;
    
    .feature-card {
        background: rgba(255, 255, 255, 0.1);
        padding: 1.5rem;
        border-radius: 12px;
        backdrop-filter: blur(10px);
        border: 1px solid rgba(255, 255, 255, 0.2);
        transition: all 0.3s ease;
        display: flex;
        align-items: flex-start;
        gap: 1rem;
        min-height: 100px;
        
        &:hover {
            transform: translateY(-4px);
            background: rgba(255, 255, 255, 0.15);
        }
        
        .body--dark & {
            background: rgba(255, 255, 255, 0.05);
            border: 1px solid rgba(255, 255, 255, 0.1);
            
            &:hover {
                background: rgba(255, 255, 255, 0.1);
            }
        }
        
        .feature-icon {
            flex-shrink: 0;
            margin-top: 0.2rem;
            
            .q-icon {
                padding: 0.8rem;
                background: rgba(255, 255, 255, 0.1);
                border-radius: 8px;
                
                .body--dark & {
                    background: rgba(255, 255, 255, 0.1);
                }
            }
        }
        
        .feature-content {
            flex: 1;
            min-width: 0;
            display: flex;
            flex-direction: column;
            
            .feature-title {
                font-size: 1.1rem;
                font-weight: 600;
                color: rgba(255, 255, 255, 0.95);
                margin: 0 0 0.5rem 0;
                line-height: 1.2;
                white-space: nowrap;
                overflow: hidden;
                text-overflow: ellipsis;
                
                .body--dark & {
                    color: rgba(255, 255, 255, 0.9);
                }
            }
            
            .feature-description {
                color: rgba(255, 255, 255, 0.8);
                line-height: 1.4;
                margin: 0;
                font-size: 0.9rem;
                flex: 1;
                
                .body--dark & {
                    color: rgba(255, 255, 255, 0.7);
                }
            }
        }
    }
}

.quick-start {
    text-align: center;
    
    .quick-start-btn {
        padding: 1rem 2.5rem;
        font-size: 1rem;
        font-weight: 600;
        letter-spacing: 0.5px;
        border-radius: 12px;
        transition: all 0.3s ease;
        
        &:hover {
            transform: translateY(-3px);
            box-shadow: 0 12px 30px rgba(0, 0, 0, 0.2);
        }
    }
}

// Responsive Design
@media (max-width: 1200px) {
    .main-container {
        padding: 2rem;
        gap: 2rem;
    }
    
    .hero-content {
        .hero-title {
            font-size: 2.5rem;
        }
    }
    
    .features-header {
        .features-title {
            font-size: 1.8rem;
        }
    }
}

@media (max-width: 1024px) {
    .main-container {
        flex-direction: column;
        padding: 2rem;
        gap: 3rem;
        justify-content: center;
    }
    
    .content-panel,
    .features-panel {
        flex: none;
        height: auto;
        padding: 0;
    }
    
    .features-header {
        margin-bottom: 1.5rem;
    }
    
    .features-grid {
        margin-bottom: 1.5rem;
        grid-template-columns: repeat(2, 1fr);
    }
}

@media (max-width: 768px) {
    .landing-page {
        overflow-y: auto;
        height: auto;
        min-height: 100vh;
    }
    
    .main-container {
        height: auto;
        padding: 2rem 1rem;
        gap: 2rem;
    }
    
    .hero-content {
        text-align: center;
        
        .hero-title {
            font-size: 2rem;
        }
        
        .hero-subtitle {
            font-size: 1rem;
        }
        
        .hero-actions {
            justify-content: center;
            
            .primary-action-btn,
            .secondary-action-btn {
                width: 100%;
                max-width: 250px;
            }
        }
    }
    
    .features-header {
        .features-title {
            font-size: 1.5rem;
        }
    }
    
    .features-grid {
        grid-template-columns: 1fr;
        gap: 1rem;
        
        .feature-card {
            padding: 1.2rem;
            
            .feature-content {
                .feature-title {
                    font-size: 1rem;
                    white-space: normal; /* Allow wrapping on mobile */
                }
                
                .feature-description {
                    font-size: 0.85rem;
                }
            }
        }
    }
    
    .quick-start {
        .quick-start-btn {
            width: 100%;
            max-width: 250px;
        }
    }
}

@media (max-width: 480px) {
    .main-container {
        padding: 1.5rem;
    }
    
    .hero-content {
        .hero-title {
            font-size: 1.8rem;
        }
        
        .hero-subtitle {
            font-size: 0.9rem;
            margin-bottom: 1.5rem;
        }
    }
    
    .features-header {
        .features-title {
            font-size: 1.3rem;
        }
    }
    
    .features-grid {
        .feature-card {
            padding: 1rem;
            flex-direction: column;
            text-align: center;
            gap: 0.8rem;
            
            .feature-icon {
                margin-top: 0;
            }
            
            .feature-content {
                .feature-title {
                    white-space: normal;
                }
            }
        }
    }
}
</style>
