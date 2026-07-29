export default [
  { ignores: ['node_modules/**','dist/**','**/*.gen.ts','src/routeTree.gen.ts','*.config.ts','scripts/**','.tanstack/**'] },
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: { parserOptions: { ecmaFeatures: { jsx: true }, sourceType: 'module' } },
    rules: {
      'no-unused-vars': ['warn', { argsIgnorePattern: '^_', varsIgnorePattern: '^_', ignoreRestSiblings: true }],
    },
  },
];
