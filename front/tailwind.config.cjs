module.exports = {
	content: [
		'src/**/*.svelte',
		'src/**/*.html'
	],
	theme: {
		placeholderColor: {
      'primary': '#3490dc',
      'secondary': '#ffed4a',
      'danger': '#e3342f',
    },
	},
	variants: {},
	plugins: [
    require('@tailwindcss/typography'),
    require('daisyui')
  ],
  daisyui: {
    themes: ["light", "dark"],
  },
};
