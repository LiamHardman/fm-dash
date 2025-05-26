<template>
    <q-page class="landing-page">
        <div class="landing-container">
            <div class="hero-section">
                <div class="logo-section">
                    <h1 class="app-title">FMDB</h1>
                    <div class="title-underline"></div>
                </div>
                <p class="app-subtitle">
                    Football Manager Database and Player Analysis Tool
                </p>
                <div class="action-buttons">
                    <q-btn class="primary-btn" @click="navigateToUpload" flat>
                        Get Started
                    </q-btn>
                    <q-btn class="secondary-btn" @click="navigateToDemo" flat>
                        Demo
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
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    background: linear-gradient(135deg, #1a237e 0%, #283593 50%, #3949ab 100%);
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
}

.landing-container {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    max-width: 600px;
    margin: 0 auto;
    padding: 2rem;
}

.hero-section {
    color: white;
    z-index: 1;
    position: relative;
}

.logo-section {
    margin-bottom: 2rem;
}

.app-title {
    font-size: 3.5rem;
    font-weight: 300;
    margin: 0;
    letter-spacing: 8px;
    text-transform: uppercase;
    font-family:
        "Inter",
        -apple-system,
        sans-serif;
    line-height: 1;
}

.title-underline {
    width: 60px;
    height: 2px;
    background: rgba(255, 255, 255, 0.8);
    margin: 1rem auto;
    border-radius: 1px;
}

.app-subtitle {
    font-size: 1.1rem;
    margin: 0 0 3rem 0;
    opacity: 0.85;
    font-weight: 300;
    line-height: 1.6;
    max-width: 400px;
    margin-left: auto;
    margin-right: auto;
}

.action-buttons {
    display: flex;
    justify-content: center;
    gap: 1.5rem;
    flex-wrap: wrap;
}

.primary-btn {
    padding: 16px 40px;
    font-size: 0.95rem;
    font-weight: 500;
    letter-spacing: 1px;
    text-transform: uppercase;
    border-radius: 2px;
    background: rgba(255, 255, 255, 0.95);
    color: #1a237e;
    min-width: 140px;

    &:hover {
        background: white;
        transform: translateY(-1px);
        box-shadow: 0 6px 20px rgba(0, 0, 0, 0.15);
    }

    transition: all 0.2s ease;
}

.secondary-btn {
    padding: 16px 40px;
    font-size: 0.95rem;
    font-weight: 500;
    letter-spacing: 1px;
    text-transform: uppercase;
    border-radius: 2px;
    background: transparent;
    color: rgba(255, 255, 255, 0.9);
    border: 1px solid rgba(255, 255, 255, 0.3);
    min-width: 140px;

    &:hover {
        background: rgba(255, 255, 255, 0.1);
        border-color: rgba(255, 255, 255, 0.6);
        transform: translateY(-1px);
    }

    transition: all 0.2s ease;
}

.landing-footer {
    padding: 2rem;
    text-align: center;
    color: rgba(255, 255, 255, 0.6);
    font-size: 0.85rem;
    font-weight: 300;
    z-index: 1;
    position: relative;

    p {
        margin: 0;
    }
}

@media (max-width: 768px) {
    .app-title {
        font-size: 2.5rem;
        letter-spacing: 4px;
    }

    .app-subtitle {
        font-size: 1rem;
        margin-bottom: 2rem;
    }

    .landing-container {
        padding: 1rem;
    }

    .action-buttons {
        flex-direction: column;
        align-items: center;
        gap: 1rem;
    }

    .primary-btn,
    .secondary-btn {
        width: 100%;
        max-width: 200px;
    }

    .landing-footer {
        padding: 1.5rem 1rem;
        font-size: 0.8rem;
    }
}

@media (max-width: 480px) {
    .app-title {
        font-size: 2rem;
        letter-spacing: 2px;
    }
}
</style>
