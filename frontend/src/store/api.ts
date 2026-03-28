import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'

// Base RTK Query API — endpoints are injected by the generated client.ts
// The generated file (src/api/client.ts) is created by running:
//   npm run codegen   (from frontend/)
// or:
//   ./commands.sh openapi:codegen   (from backend/)
export const api = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({
    baseUrl: '/api',
    credentials: 'include',
  }),
  tagTypes: [],
  endpoints: () => ({}),
})
