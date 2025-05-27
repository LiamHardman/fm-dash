<template>
    <q-page class="landing-page">
        <!-- Hero Section -->
        <div class="hero-section">
            <div class="hero-container">
                <div class="hero-content">
                    <div class="hero-badge">
                        <q-icon name="sports_soccer" size="1.2rem" />
                        <span>Football Manager 2024</span>
                    </div>
                    <h1 class="hero-title">
                        Unlock Your
                        <span class="gradient-text">Squad's Potential</span>
                    </h1>
                    <p class="hero-subtitle">
                        Advanced player analysis and team optimization for Football Manager 2024. 
                        Upload your data and discover insights that will transform your tactical approach.
                    </p>
                    <div class="hero-actions">
                        <q-btn
                            unelevated
                            color="primary"
                            size="lg"
                            @click="navigateToUpload"
                            class="primary-action-btn"
                        >
                            <q-icon name="upload" class="q-mr-sm" />
                            Start Analysis
                        </q-btn>
                        <q-btn
                            outline
                            color="primary"
                            size="lg"
                            @click="navigateToDemo"
                            class="secondary-action-btn"
                        >
                            <q-icon name="play_arrow" class="q-mr-sm" />
                            View Demo Data
                        </q-btn>
                    </div>
                </div>
                <div class="hero-visual">
                    <div class="visual-card">
                        <q-icon name="analytics" size="4rem" color="primary" />
                        <div class="visual-stats">
                            <div class="stat">
                                <div class="stat-number">10,000+</div>
                                <div class="stat-label">Players Analyzed</div>
                            </div>
                            <div class="stat">
                                <div class="stat-number">500+</div>
                                <div class="stat-label">Teams Optimized</div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Features Section -->
        <div class="features-section">
            <div class="features-container">
                <div class="section-header">
                    <h2 class="section-title">Powerful Analysis Tools</h2>
                    <p class="section-subtitle">
                        Everything you need to understand and optimize your Football Manager squad
                    </p>
                </div>
                <div class="features-grid">
                    <div class="feature-card">
                        <div class="feature-icon">
                            <q-icon name="search" size="2.5rem" color="primary" />
                        </div>
                        <h3 class="feature-title">Player Search</h3>
                        <p class="feature-description">
                            Advanced filtering and search capabilities to find the perfect players for your system
                        </p>
                    </div>
                    <div class="feature-card">
                        <div class="feature-icon">
                            <q-icon name="bar_chart" size="2.5rem" color="secondary" />
                        </div>
                        <h3 class="feature-title">Player Stats</h3>
                        <p class="feature-description">
                            Comprehensive stats breakdown with role-specific ratings and performance metrics
                        </p>
                    </div>
                    <div class="feature-card">
                        <div class="feature-icon">
                            <q-icon name="compare" size="2.5rem" color="accent" />
                        </div>
                        <h3 class="feature-title">Comparisons</h3>
                        <p class="feature-description">
                            Side-by-side comparisons to identify upgrades and tactical alternatives
                        </p>
                    </div>
                    <div class="feature-card">
                        <div class="feature-icon">
                            <q-icon name="group_work" size="2.5rem" color="info" />
                        </div>
                        <h3 class="feature-title">Team Analysis</h3>
                        <p class="feature-description">
                            Squad composition insights and team chemistry analysis across all positions
                        </p>
                    </div>
                </div>
            </div>
        </div>

        <!-- CTA Section -->
        <div class="cta-section">
            <div class="cta-container">
                <div class="cta-content">
                    <h2 class="cta-title">Ready to Optimize Your Squad?</h2>
                    <q-btn
                        unelevated
                        color="primary"
                        size="xl"
                        @click="navigateToUpload"
                        class="cta-btn"
                    >
                        <q-icon name="rocket_launch" class="q-mr-sm" />
                        Get Started Now
                    </q-btn>
                </div>
            </div>
        </div>
    </q-page>
</template>

<script>
import { defineComponent } from "vue";
import { useRouter } from "vue-router";

export default defineComponent({
    name: "LandingPage",
    setup() {
        const router = useRouter();

        const navigateToUpload = () => {
            router.push("/upload");
        };

        const navigateToDemo = () => {
            router.push("/dataset/demo");
        };

        return {
            navigateToUpload,
            navigateToDemo,
        };
    },
});
</script>

<style lang="scss" scoped>
.landing-page {
    min-height: 100vh;
    background: linear-gradient(135deg, #f8f9fc 0%, #ffffff 100%);
    
    .body--dark & {
        background: linear-gradient(135deg, #1e1e1e 0%, #2a2a2a 100%);
    }
}

// Hero Section
.hero-section {
    padding: 4rem 0;
    background: linear-gradient(135deg, #1a237e 0%, #283593 50%, #3949ab 100%);
    color: white;
    position: relative;
    overflow: hidden;
    
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
    
    .hero-container {
        max-width: 1200px;
        margin: 0 auto;
        padding: 0 2rem;
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 4rem;
        align-items: center;
        position: relative;
        z-index: 1;
    }
    
    .hero-content {
        .hero-badge {
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
            background: rgba(255, 255, 255, 0.1);
            padding: 0.5rem 1rem;
            border-radius: 20px;
            font-size: 0.85rem;
            font-weight: 500;
            margin-bottom: 2rem;
            backdrop-filter: blur(10px);
        }
        
        .hero-title {
            font-size: 3.5rem;
            font-weight: 700;
            line-height: 1.1;
            margin: 0 0 1.5rem 0;
            
            .gradient-text {
                background: linear-gradient(135deg, #64b5f6 0%, #42a5f5 100%);
                -webkit-background-clip: text;
                -webkit-text-fill-color: transparent;
                background-clip: text;
            }
        }
        
        .hero-subtitle {
            font-size: 1.2rem;
            line-height: 1.6;
            margin: 0 0 2.5rem 0;
            opacity: 0.9;
            font-weight: 300;
        }
        
        .hero-actions {
            display: flex;
            gap: 1rem;
            flex-wrap: wrap;
            
            .primary-action-btn {
                padding: 1rem 2rem;
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
                padding: 1rem 2rem;
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
    
    .hero-visual {
        display: flex;
        justify-content: center;
        
        .visual-card {
            background: rgba(255, 255, 255, 0.1);
            border-radius: 20px;
            padding: 3rem 2rem;
            text-align: center;
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.2);
            
            .visual-stats {
                margin-top: 2rem;
                display: flex;
                gap: 2rem;
                justify-content: center;
                
                .stat {
                    .stat-number {
                        font-size: 2rem;
                        font-weight: 700;
                        color: #64b5f6;
                        margin-bottom: 0.5rem;
                    }
                    
                    .stat-label {
                        font-size: 0.9rem;
                        opacity: 0.8;
                    }
                }
            }
        }
    }
}

// Features Section
.features-section {
    padding: 6rem 0;
    background: #f8f9fc;
    
    .body--dark & {
        background: #1e1e1e;
    }
    
    .features-container {
        max-width: 1200px;
        margin: 0 auto;
        padding: 0 2rem;
    }
    
    .section-header {
        text-align: center;
        margin-bottom: 4rem;
        
        .section-title {
            font-size: 3rem;
            font-weight: 700;
            color: #1a237e;
            margin: 0 0 1rem 0;
            
            .body--dark & {
                color: rgba(255, 255, 255, 0.9);
            }
        }
        
        .section-subtitle {
            font-size: 1.2rem;
            color: #666;
            margin: 0;
            max-width: 600px;
            margin: 0 auto;
            
            .body--dark & {
                color: rgba(255, 255, 255, 0.7);
            }
        }
    }
    
    .features-grid {
        display: grid;
        grid-template-columns: repeat(4, 1fr);
        gap: 2rem;
        
        .feature-card {
            background: white;
            padding: 2.5rem 2rem;
            border-radius: 16px;
            text-align: center;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
            transition: all 0.3s ease;
            
            &:hover {
                transform: translateY(-8px);
                box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15);
            }
            
            .body--dark & {
                background: rgba(255, 255, 255, 0.03);
                box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
                
                &:hover {
                    box-shadow: 0 12px 40px rgba(0, 0, 0, 0.5);
                }
            }
            
            .feature-icon {
                margin-bottom: 1.5rem;
                
                .q-icon {
                    padding: 1rem;
                    background: rgba(26, 35, 126, 0.1);
                    border-radius: 12px;
                    
                    .body--dark & {
                        background: rgba(255, 255, 255, 0.1);
                    }
                }
            }
            
            .feature-title {
                font-size: 1.3rem;
                font-weight: 600;
                color: #1a237e;
                margin: 0 0 1rem 0;
                
                .body--dark & {
                    color: rgba(255, 255, 255, 0.9);
                }
            }
            
            .feature-description {
                color: #666;
                line-height: 1.6;
                margin: 0;
                
                .body--dark & {
                    color: rgba(255, 255, 255, 0.7);
                }
            }
        }
    }
}

// CTA Section
.cta-section {
    padding: 6rem 0;
    background: linear-gradient(135deg, #1a237e 0%, #283593 100%);
    color: white;
    text-align: center;
    
    .cta-container {
        max-width: 800px;
        margin: 0 auto;
        padding: 0 2rem;
    }
    
    .cta-content {
        .cta-title {
            font-size: 3rem;
            font-weight: 700;
            margin: 0 0 2.5rem 0;
            line-height: 1.2;
        }
        
        .cta-btn {
            padding: 1.2rem 3rem;
            font-size: 1.1rem;
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
}

// Responsive Design
@media (max-width: 1024px) {
    .hero-section {
        .hero-container {
            grid-template-columns: 1fr;
            gap: 3rem;
            text-align: center;
        }
        
        .hero-content {
            .hero-title {
                font-size: 3rem;
            }
        }
    }
    
    .features-section {
        .section-header {
            .section-title {
                font-size: 2.5rem;
            }
        }
        
        .features-grid {
            grid-template-columns: repeat(2, 1fr);
        }
    }
    
    .cta-section {
        .cta-content {
            .cta-title {
                font-size: 2.5rem;
            }
        }
    }
}

@media (max-width: 768px) {
    .hero-section {
        padding: 3rem 0;
        
        .hero-container {
            padding: 0 1rem;
            gap: 2rem;
        }
        
        .hero-content {
            .hero-title {
                font-size: 2.5rem;
            }
            
            .hero-subtitle {
                font-size: 1.1rem;
            }
            
            .hero-actions {
                flex-direction: column;
                align-items: center;
                
                .primary-action-btn,
                .secondary-action-btn {
                    width: 100%;
                    max-width: 280px;
                }
            }
        }
        
        .hero-visual {
            .visual-card {
                padding: 2rem 1.5rem;
                
                .visual-stats {
                    gap: 1.5rem;
                    
                    .stat {
                        .stat-number {
                            font-size: 1.5rem;
                        }
                        
                        .stat-label {
                            font-size: 0.8rem;
                        }
                    }
                }
            }
        }
    }
    
    .features-section {
        padding: 4rem 0;
        
        .features-container {
            padding: 0 1rem;
        }
        
        .section-header {
            margin-bottom: 3rem;
            
            .section-title {
                font-size: 2rem;
            }
            
            .section-subtitle {
                font-size: 1rem;
            }
        }
        
        .features-grid {
            grid-template-columns: 1fr;
            gap: 1.5rem;
            
            .feature-card {
                padding: 2rem 1.5rem;
            }
        }
    }
    
    .cta-section {
        padding: 4rem 0;
        
        .cta-container {
            padding: 0 1rem;
        }
        
        .cta-content {
            .cta-title {
                font-size: 2rem;
            }
            
            .cta-btn {
                padding: 1rem 2rem;
                font-size: 1rem;
                width: 100%;
                max-width: 280px;
            }
        }
    }
}

@media (max-width: 480px) {
    .hero-section {
        .hero-content {
            .hero-title {
                font-size: 2rem;
            }
            
            .hero-subtitle {
                font-size: 1rem;
                margin-bottom: 2rem;
            }
        }
        
        .hero-visual {
            .visual-card {
                padding: 1.5rem 1rem;
                
                .visual-stats {
                    flex-direction: column;
                    gap: 1rem;
                }
            }
        }
    }
    
    .features-section {
        .section-header {
            .section-title {
                font-size: 1.8rem;
            }
        }
        
        .features-grid {
            .feature-card {
                padding: 1.5rem 1rem;
                
                .feature-title {
                    font-size: 1.1rem;
                }
                
                .feature-description {
                    font-size: 0.9rem;
                }
            }
        }
    }
    
    .cta-section {
        .cta-content {
            .cta-title {
                font-size: 1.8rem;
            }
        }
    }
}
</style>
