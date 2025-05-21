// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** GitHub OAuth callback Handles the GitHub OAuth callback and returns JWT token if user is authorized GET /admin/session/github/callback */
export async function getAdminSessionGithubCallback(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getAdminSessionGithubCallbackParams,
  options?: { [key: string]: any },
) {
  return request<API.GitHubCallbackResp>('/admin/session/github/callback', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** Get github client id Get github client id GET /admin/session/github/clientid */
export async function getAdminSessionGithubClientid(options?: {
  [key: string]: any;
}) {
  return request<API.GitHubClientIDResp>('/admin/session/github/clientid', {
    method: 'GET',
    ...(options || {}),
  });
}

/** Get user information Returns the authenticated user's username and policy GET /admin/session/userinfo */
export async function getAdminSessionUserinfo(options?: {
  [key: string]: any;
}) {
  return request<API.UserInfoResp>('/admin/session/userinfo', {
    method: 'GET',
    ...(options || {}),
  });
}
