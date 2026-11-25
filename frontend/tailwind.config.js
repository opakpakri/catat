/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{svelte,js,ts}'],
  darkMode: 'class',
  theme: {
    extend: {
      fontFamily: {
        sans: ["'Plus Jakarta Sans'", 'Inter', 'system-ui', 'sans-serif']
      },
      colors: {
        catat: {
          bg: '#121212',
          surface: '#1E1E22',
          border: '#2E2E34',
          accent: '#F05A4F',
          'accent-hover': '#D9483E',
          income: '#3B82F6',
          expense: '#EF4444',
          total: '#F8FAFC',
          emerald: '#10B981',
          text: '#F4F4F5',
          muted: '#A1A1AA'
        }
      }
    }
  },
  plugins: []
};
