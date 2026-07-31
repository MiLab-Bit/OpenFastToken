import globals from 'globals'
import js from '@eslint/js'
import pluginQuery from '@tanstack/eslint-plugin-query'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import { defineConfig } from 'eslint/config'
import tseslint from 'typescript-eslint'

export default defineConfig(
  {
    ignores: [
      'dist',
      'src/components/ui',
      '*.config.*',
      '*.setup.*',
      // Generated files (e.g. TanStack Router route tree) contain ~90 `as any`
      // casts that must never be linted or edited by hand.
      '**/*.gen.ts',
      'src/routeTree.gen.ts',
    ],
  },
  {
    extends: [
      js.configs.recommended,
      ...tseslint.configs.recommended,
      ...pluginQuery.configs['flat/recommended'],
    ],
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2020,
      globals: globals.browser,
    },
    plugins: {
      'react-hooks': reactHooks,
      'react-refresh': reactRefresh,
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'react-hooks/incompatible-library': 'off',
      // Allow setState in effects for derived state patterns
      // See: https://react.dev/learn/you-might-not-need-an-effect
      'react-hooks/set-state-in-effect': 'warn',
      // Warn on immutability issues but don't block builds
      // Some false positives for local variables in useMemo/useCallback
      'react-hooks/immutability': 'warn',
      // React Compiler rules - warn only for non-compiler projects
      'react-hooks/static-components': 'warn',
      'react-hooks/refs': 'warn',
      'react-hooks/purity': 'warn',
      'react-refresh/only-export-components': [
        'warn',
        { allowConstantExport: true },
      ],
      'no-console': ['error', { allow: ['warn', 'error', 'info'] }],
      // Ban handwritten `any` to keep the codebase type-safe and explicit.
      '@typescript-eslint/no-explicit-any': 'error',
      'no-unused-vars': 'off',
      '@typescript-eslint/no-unused-vars': [
        'error',
        {
          args: 'all',
          argsIgnorePattern: '^_',
          caughtErrors: 'all',
          caughtErrorsIgnorePattern: '^_',
          destructuredArrayIgnorePattern: '^_',
          varsIgnorePattern: '^_',
          ignoreRestSiblings: true,
        },
      ],
      '@typescript-eslint/consistent-type-imports': [
        'error',
        {
          prefer: 'type-imports',
          fixStyle: 'inline-type-imports',
          disallowTypeAnnotations: false,
        },
      ],
      'no-duplicate-imports': 'error',
      // TanStack Query exhaustive-deps is too strict for stable dependencies like `t`
      '@tanstack/query/exhaustive-deps': 'warn',
      // Can have false positives for try-catch control flow
      'no-useless-assignment': 'warn',
      // Performance-related rules
      'no-restricted-syntax': [
        'error',
        {
          selector: 'CallExpression[callee.name="ReactDOM.render"]',
          message: 'Use createRoot instead of ReactDOM.render (React 18+)',
        },
        {
          // Bans the double cast `x as unknown as T` (nested TSAsExpression).
          // Prefer a correct single cast, a typed helper, or (for genuine
          // third-party untyped APIs) an `eslint-disable-next-line` with a
          // reason comment.
          selector: 'TSAsExpression[expression.type="TSAsExpression"]',
          message:
            'Avoid the double cast "as unknown as T". Provide a correct type, use the `unsafeCast` helper, or add an `eslint-disable-next-line no-restricted-syntax` with a reason.',
        },
      ],
      // Prevent common performance issues
      'react-hooks/exhaustive-deps': 'warn',
      // Ensure proper key usage in lists
      'react-hooks/rules-of-hooks': 'error',
      // Additional TypeScript performance rules
      '@typescript-eslint/no-floating-promises': 'error',
      '@typescript-eslint/no-misused-promises': 'error',
      '@typescript-eslint/no-unnecessary-type-constraint': 'warn',
    },
  },
  {
    files: ['src/routes/**/*.{ts,tsx}'],
    plugins: {
      'react-refresh': reactRefresh,
    },
    rules: {
      'react-refresh/only-export-components': 'off',
      // Relax some rules for route files
      '@typescript-eslint/no-unused-vars': 'warn',
    },
  },
  // Additional configuration for test files
  {
    files: ['**/*.test.{ts,tsx}', '**/*.spec.{ts,tsx}', '**/tests/**'],
    rules: {
      '@typescript-eslint/no-unused-vars': 'warn',
      'no-console': 'off',
    },
  }
)
