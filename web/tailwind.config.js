/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{vue,js,ts,jsx,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        sans: [
          "-apple-system",
          "BlinkMacSystemFont",
          '"Apple Color Emoji"',
          '"Segoe UI Emoji"',
          '"Segoe UI Symbol"',
          '"Segoe UI"',
          "Roboto",
          "Helvetica",
          "Arial",
          "sans-serif"
        ],
      },
      colors: {
        ink: {
          950: "#13151e",
          900: "#1b2030",
          800: "#262e45",
          700: "#344060",
        },
        sky: {
          300: "#76cbff",
          400: "#4cbaf7",
          500: "#18a2f0",
        },
        coral: {
          300: "#ff8a6c",
          400: "#ff7350",
        },
      },
      boxShadow: {
        soft: "0 12px 30px rgba(9, 14, 31, 0.18)",
      },
      animation: {
        "slide-in": "slideIn 0.35s ease-out",
        "fade-rise": "fadeRise 0.45s ease-out",
      },
      keyframes: {
        slideIn: {
          "0%": { transform: "translateX(-14px)", opacity: "0" },
          "100%": { transform: "translateX(0)", opacity: "1" },
        },
        fadeRise: {
          "0%": { transform: "translateY(12px)", opacity: "0" },
          "100%": { transform: "translateY(0)", opacity: "1" },
        },
      },
    },
  },
  plugins: [],
}
