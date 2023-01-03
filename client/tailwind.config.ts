import type { Config } from 'tailwindcss'
import themes from 'daisyui/src/colors/themes'

export default <Partial<Config>>{
  plugins: [require('daisyui')],
  daisyui: {
    themes: [
      {
        light: {
          ...themes['[data-theme=light]'],
          primary: '#4ade80',
          secondary: '#3b82f6'
        }
      }, {
        dark: {
          ...themes['[data-theme=dark]'],
          primary: '#4ade80',
          secondary: '#3b82f6'
        }
      }
    ]
  }
}
