/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ['./internal/views/**/*.{html,js,templ}'],
    theme: {
        extend: {},
        fontFamily: {
            sans: ['JetBrains Mono', 'monospace'],
            serif: ['JetBrains Mono', 'monospace'],
            mono: ['JetBrains Mono', 'monospace'],
        },
    },
    plugins: [],
}
