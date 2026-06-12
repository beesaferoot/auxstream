/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // Core citrus surfaces
        paper: "#f7f8ef",
        "paper-tint": "#eef3df",
        ink: "#11140c",
        "ink-text": "#15160f",
        "player-bg": "#0d0f08",
        card: "#ffffff",
        // Lime accent ramp
        lime: {
          DEFAULT: "#b6f03c",
          bright: "#d8ff6e",
          deep: "#5ba20a",
          shadow: "#1c3a00",
          tint: "#eafabf",
        },
        // Citrus text + lines (light surfaces)
        muted: {
          DEFAULT: "#5a5f4c",
          2: "#7e8369",
          3: "#8a8f76",
        },
        faint: {
          DEFAULT: "#9aa085",
          2: "#a6ab92",
        },
        line: {
          DEFAULT: "#e4e6d6",
          2: "#ebeddf",
          sep: "#f1f2e9",
          sep2: "#e2e4d6",
        },
        // Dark surfaces (player / now-playing)
        "surface-dark": "#171a11",
        "border-dark": {
          DEFAULT: "#2a2e20",
          2: "#2c3122",
          3: "#3a3f2d",
        },
        "muted-dark": {
          DEFAULT: "#cdd3b6",
          2: "#b7bca5",
          3: "#9aa085",
        },
        "faint-dark": {
          DEFAULT: "#7e8369",
          2: "#8b9075",
          3: "#5f6450",
        },
        "wave-off": "#363b29",
        danger: "#c43d3d",
      },
      fontFamily: {
        display: ["'Bricolage Grotesque'", "sans-serif"],
        sans: ["'Hanken Grotesk'", "-apple-system", "BlinkMacSystemFont", "'Segoe UI'", "sans-serif"],
        mono: ["'Space Mono'", "monospace"],
      },
      borderRadius: {
        pill: "30px",
        tile: "13px",
        nav: "16px",
        hero: "26px",
        bar: "20px",
      },
      boxShadow: {
        cover: "0 8px 20px rgba(20,30,0,.12)",
        hero: "0 24px 50px rgba(0,0,0,.4)",
        "player-cover": "0 40px 90px rgba(0,0,0,.55)",
        bar: "0 18px 44px rgba(0,0,0,.28)",
        lime: "0 10px 26px rgba(182,240,60,.3)",
        "lime-hover": "0 14px 32px rgba(182,240,60,.45)",
        "lime-big": "0 14px 36px rgba(182,240,60,.4)",
        logo: "0 0 0 1px #11140c, 0 8px 22px rgba(182,240,60,.35)",
        card: "0 12px 30px rgba(20,30,0,.08)",
        row: "0 8px 22px rgba(20,30,0,.07)",
        menu: "0 24px 54px rgba(20,30,0,.22)",
        modal: "0 40px 100px rgba(0,0,0,.4)",
      },
      keyframes: {
        auxUp: {
          from: { opacity: "0", transform: "translateY(16px)" },
          to: { opacity: "1", transform: "none" },
        },
        auxPop: {
          from: { opacity: "0", transform: "scale(.97)" },
          to: { opacity: "1", transform: "none" },
        },
        auxEq: {
          "0%,100%": { transform: "scaleY(.3)" },
          "50%": { transform: "scaleY(1)" },
        },
        auxWave: {
          "0%,100%": { transform: "scaleY(.5)" },
          "50%": { transform: "scaleY(1)" },
        },
        auxGlow: {
          "0%,100%": { opacity: ".45", transform: "scale(1)" },
          "50%": { opacity: ".8", transform: "scale(1.05)" },
        },
        auxSpinIn: {
          from: { opacity: "0", transform: "translateY(40px) scale(.98)" },
          to: { opacity: "1", transform: "none" },
        },
      },
      animation: {
        "aux-up": "auxUp .5s both",
        "aux-pop": "auxPop .35s both",
        "aux-eq": "auxEq .9s ease-in-out infinite",
        "aux-wave": "auxWave 1.1s ease-in-out infinite",
        "aux-glow": "auxGlow 6s ease-in-out infinite",
        "aux-spin-in": "auxSpinIn .42s cubic-bezier(.2,.8,.2,1) both",
      },
    },
  },
  plugins: [],
}
