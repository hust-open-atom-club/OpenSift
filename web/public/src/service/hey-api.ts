import type { CreateClientConfig } from './client/client.gen';

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  baseUrl: (!global.document ? (process.env.BACKEND || "http://localhost:5000") : "") + "/api/v1"
});