/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ['./internal/views/**/*.{html,js,templ}'],
    theme: {
        extend: {},
        fontFamily: {
            mono: ['JetBrains Mono', 'monospace'],
        },
    },
    plugins: [],
}
