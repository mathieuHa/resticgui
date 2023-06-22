import { Theme } from './src/utils/types'; 

module.exports = {
	content: [
		'src/**/*.svelte',
		'src/**/*.html'
	],
	theme: {
	},
	variants: {},
	plugins: [
    require('@tailwindcss/typography'),
    require('daisyui')
  ],
  daisyui: {
    themes: [Theme.light, Theme.dark],
  },
};
