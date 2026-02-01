import { defineConfig } from '@kubb/core'
import { pluginOas } from '@kubb/plugin-oas'
import { pluginTs } from '@kubb/plugin-ts'
import { pluginVueQuery } from '@kubb/plugin-vue-query'
import { pluginZod } from '@kubb/plugin-zod'

export default defineConfig({
  root: '.',
  input: {
    // Fallback: HTTP → локальный файл
    path: process.env.SWAGGER_URL || '../backend/docs/swagger.json',
  },
  output: {
    path: './src/api/generated',
    clean: true,
  },
  plugins: [
    pluginOas({ output: false, validate: true }),
    pluginTs({
      output: { path: './types' },
      group: { type: 'tag' },
      enumType: 'asConst',
      dateType: 'string',
      exclude: [{ type: 'tag', pattern: 'admin' }],
    }),
    pluginZod({
      output: { path: './schemas' },
      group: { type: 'tag' },
      typed: true,
      dateType: 'string',
      exclude: [{ type: 'tag', pattern: 'admin' }],
    }),
    pluginVueQuery({
      output: { path: './hooks' },
      group: { type: 'tag' },
      client: { importPath: '@/api/client' },
      dataReturnType: 'data',
      pathParamsType: 'object',
      exclude: [{ type: 'tag', pattern: 'admin' }],
    }),
  ],
})
