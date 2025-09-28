/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/**/*.{js,jsx,ts,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'terminal': {
          'bg': '#0a0a0a',
          'text': '#00ff00',
          'accent': '#ffff00',
          'error': '#ff4444',
          'success': '#44ff44',
        }
      },
      fontFamily: {
        'mono': ['Monaco', 'Menlo', 'Ubuntu Mono', 'monospace'],
      },
    },
  },
  plugins: [],
}