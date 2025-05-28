// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** 查询发行版包列表 根据发行版、链接、置信度等条件分页查询包列表 GET /admin/label/distributions */
export async function getAdminLabelDistributions(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getAdminLabelDistributionsParams,
  options?: { [key: string]: any },
) {
  return request<API.PageDTOModelDistributionPackageDTO>(
    '/admin/label/distributions',
    {
      method: 'GET',
      params: {
        ...params,
      },
      ...(options || {}),
    },
  );
}

/** 获取所有发行版包的前缀 获取所有支持的发行版包前缀列表 GET /admin/label/distributions/all */
export async function getAdminLabelDistributionsAll(options?: {
  [key: string]: any;
}) {
  return request<string[]>('/admin/label/distributions/all', {
    method: 'GET',
    ...(options || {}),
  });
}

/** 更新发行版包的 Git 链接 更新指定发行版包的 Git 仓库链接和置信度 PUT /admin/label/distributions/gitlink */
export async function putAdminLabelDistributionsGitlink(
  body: API.UpdateDistributionGitLinkReq,
  options?: { [key: string]: any },
) {
  return request<any>('/admin/label/distributions/gitlink', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}
