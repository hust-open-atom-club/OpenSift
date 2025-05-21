import type { CreateClientConfig } from './client/client.gen';

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  baseUrl: (!global.document ? process.env.NEXT_PUBLIC_BACKEND : "") + "/api/v1"
});