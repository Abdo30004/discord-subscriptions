import type { Config } from 'tailwindcss';

const config: Config = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        background: '#0B0E14',
        foreground: '#F1F5F9',
        card: {
          DEFAULT: '#131822',
          foreground: '#F1F5F9',
          border: '#1E293B',
        },
        blurple: {
          DEFAULT: '#5865F2',
          hover: '#4752C4',
          subtle: 'rgba(88, 101, 242, 0.15)',
        },
        emerald: {
          glow: 'rgba(16, 185, 129, 0.2)',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
      },
      backgroundImage: {
        'radial-gradient': 'radial-gradient(circle at 50% 0%, var(--tw-gradient-stops))',
      },
    },
  },
  plugins: [],
};

export default config;
