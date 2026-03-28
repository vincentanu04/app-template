import { api } from "../store/api";
export const addTagTypes = [] as const;
const injectedRtkApi = api
  .enhanceEndpoints({
    addTagTypes,
  })
  .injectEndpoints({
    endpoints: (build) => ({
      postAuthRegister: build.mutation<
        PostAuthRegisterApiResponse,
        PostAuthRegisterApiArg
      >({
        query: (queryArg) => ({
          url: `/auth/register`,
          method: "POST",
          body: queryArg.registerRequest,
        }),
        invalidatesTags: [],
      }),
      postAuthLogin: build.mutation<
        PostAuthLoginApiResponse,
        PostAuthLoginApiArg
      >({
        query: (queryArg) => ({
          url: `/auth/login`,
          method: "POST",
          body: queryArg.loginRequest,
        }),
        invalidatesTags: [],
      }),
      postAuthLogout: build.mutation<
        PostAuthLogoutApiResponse,
        PostAuthLogoutApiArg
      >({
        query: () => ({ url: `/auth/logout`, method: "POST" }),
        invalidatesTags: [],
      }),
      getAuthMe: build.query<GetAuthMeApiResponse, GetAuthMeApiArg>({
        query: () => ({ url: `/auth/me` }),
        providesTags: [],
      }),
      getHealth: build.query<GetHealthApiResponse, GetHealthApiArg>({
        query: () => ({ url: `/health` }),
        providesTags: [],
      }),
    }),
    overrideExisting: false,
  });
export { injectedRtkApi as enhancedApi };
export type PostAuthRegisterApiResponse = /** status 201 Created */ AuthUser;
export type PostAuthRegisterApiArg = {
  registerRequest: RegisterRequest;
};
export type PostAuthLoginApiResponse = /** status 200 Logged in */ AuthUser;
export type PostAuthLoginApiArg = {
  loginRequest: LoginRequest;
};
export type PostAuthLogoutApiResponse = unknown;
export type PostAuthLogoutApiArg = void;
export type GetAuthMeApiResponse = /** status 200 Current user */ AuthUser;
export type GetAuthMeApiArg = void;
export type GetHealthApiResponse = /** status 200 OK */ {
  status: string;
};
export type GetHealthApiArg = void;
export type AuthUser = {
  id: string;
  email: string;
};
export type ErrorResponse = {
  message: string;
};
export type RegisterRequest = {
  email: string;
  password: string;
};
export type LoginRequest = {
  email: string;
  password: string;
};
export const {
  usePostAuthRegisterMutation,
  usePostAuthLoginMutation,
  usePostAuthLogoutMutation,
  useGetAuthMeQuery,
  useGetHealthQuery,
} = injectedRtkApi;
