<template>
    <q-page class="landing-page">
        <div class="landing-container">
            <div class="hero-section">
                <h1 class="app-title">FMDB</h1>
                <p class="app-subtitle">
                    Football Manager Database and Player Analysis Tool
                </p>
                <div class="action-buttons">
                    <q-btn
                        color="primary"
                        size="lg"
                        class="get-started-btn"
                        @click="navigateToUpload"
                        unelevated
                    >
                        Get Started
                    </q-btn>
                    <q-btn
                        color="secondary"
                        size="lg"
                        class="demo-btn"
                        @click="navigateToDemo"
                        unelevated
                    >
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
    align-items: center;
    justify-content: center;
    background: linear-gradient(
        135deg,
        var(--q-primary) 0%,
        var(--q-secondary) 100%
    );
}

.landing-container {
    text-align: center;
    max-width: 800px;
    padding: 2rem;
}

.hero-section {
    color: white;
}

.app-title {
    font-size: 4rem;
    font-weight: 700;
    margin: 0 0 1rem 0;
    text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
    letter-spacing: 2px;
}

.app-subtitle {
    font-size: 1.5rem;
    margin: 0 0 3rem 0;
    opacity: 0.9;
    font-weight: 300;
}

.action-buttons {
    display: flex;
    justify-content: center;
    gap: 1rem;
}

.get-started-btn {
    padding: 12px 32px;
    font-size: 1.1rem;
    font-weight: 600;
    border-radius: 8px;
    background: white;
    color: var(--q-primary);
    
    &:hover {
        transform: translateY(-2px);
        box-shadow: 0 8px 25px rgba(0, 0, 0, 0.2);
    }
    
    transition: all 0.3s ease;
}

.demo-btn {
    padding: 12px 32px;
    font-size: 1.1rem;
    font-weight: 600;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.1);
    color: white;
    border: 2px solid white;
    
    &:hover {
        transform: translateY(-2px);
        box-shadow: 0 8px 25px rgba(0, 0, 0, 0.2);
        background: white;
        color: var(--q-secondary);
    }
    
    transition: all 0.3s ease;
}

@media (max-width: 768px) {
    .app-title {
        font-size: 3rem;
    }
    
    .app-subtitle {
        font-size: 1.2rem;
    }
    
    .landing-container {
        padding: 1rem;
    }
}
</style>