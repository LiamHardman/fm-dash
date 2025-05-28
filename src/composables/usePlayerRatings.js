import { computed } from 'vue'

export function usePlayerRatings() {
    // Unified rating color system
    const getUnifiedRatingClass = (rating) => {
        if (rating >= 90) return "rating-90";
        if (rating >= 85) return "rating-85";
        if (rating >= 80) return "rating-80";
        if (rating >= 75) return "rating-75";
        if (rating >= 70) return "rating-70";
        if (rating >= 65) return "rating-65";
        if (rating >= 60) return "rating-60";
        if (rating >= 55) return "rating-55";
        if (rating >= 50) return "rating-50";
        return "rating-below-50";
    };

    // Attribute mapping from short names to full names
    const attributeFullNameMap = {
        COR: "Corners",
        CRO: "Crossing",
        DRI: "Dribbling",
        FIN: "Finishing",
        FRE: "Free Kicks",
        HEA: "Heading",
        LON: "Long Shots",
        LTH: "Long Throws",
        MAR: "Marking",
        PAS: "Passing",
        PEN: "Penalty Taking",
        TAC: "Tackling",
        TEC: "Technique",
        AGG: "Aggression",
        ANT: "Anticipation",
        BRA: "Bravery",
        COM: "Composure",
        CON: "Concentration",
        DEC: "Decisions",
        DET: "Determination",
        FLA: "Flair",
        LEA: "Leadership",
        OTB: "Off The Ball",
        POS: "Positioning",
        TEA: "Teamwork",
        VIS: "Vision",
        WOR: "Work Rate",
        ACC: "Acceleration",
        AGI: "Agility",
        BAL: "Balance",
        JUM: "Jumping Reach",
        NAT: "Natural Fitness",
        PAC: "Pace",
        STA: "Stamina",
        STR: "Strength",
        // FIFA stats
        DIV: "Diving",
        HAN: "Handling",
        KIC: "Kicking",
        REF: "Reflexes",
        SPD: "Speed (GK)",
        PHY: "Physical",
        DEF: "Defending",
        SHO: "Shooting"
    };

    // Attribute groupings for better organization
    const attributeGroups = computed(() => ({
        technical: {
            name: "Technical",
            attrs: ["COR", "CRO", "DRI", "FIN", "FRE", "HEA", "LON", "LTH", "MAR", "PAS", "PEN", "TAC", "TEC"]
        },
        mental: {
            name: "Mental", 
            attrs: ["AGG", "ANT", "BRA", "COM", "CON", "DEC", "DET", "FLA", "LEA", "OTB", "POS", "TEA", "VIS", "WOR"]
        },
        physical: {
            name: "Physical",
            attrs: ["ACC", "AGI", "BAL", "JUM", "NAT", "PAC", "STA", "STR"]
        },
        goalkeeper: {
            name: "Goalkeeper",
            attrs: ["DIV", "HAN", "KIC", "REF", "SPD"]
        },
        fifa: {
            name: "FIFA",
            attrs: ["PAC", "SHO", "PAS", "DRI", "DEF", "PHY"]
        }
    }));

    return {
        getUnifiedRatingClass,
        attributeFullNameMap,
        attributeGroups
    }
}