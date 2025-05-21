// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** 获取运行中的工具实例列表 获取所有运行中的工具实例的信息 GET /admin/toolset/instances */
export async function getAdminToolsetInstances(options?: {
  [key: string]: any;
}) {
  return request<API.ToolInstanceDTO[]>('/admin/toolset/instances', {
    method: 'GET',
    ...(options || {}),
  });
}

/** 创建工具实例 根据工具ID和参数创建并运行工具实例 POST /admin/toolset/instances */
export async function postAdminToolsetInstances(
  body: API.ToolCreateInstanceReq,
  options?: { [key: string]: any },
) {
  return request<API.ToolInstanceDTO>('/admin/toolset/instances', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** WebSocket 连接工具实例 通过 WebSocket 方式 attach 到指定工具实例 GET /admin/toolset/instances/${param0}/attach */
export async function getAdminToolsetInstancesIdAttach(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getAdminToolsetInstancesIdAttachParams,
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<any>(`/admin/toolset/instances/${param0}/attach`, {
    method: 'GET',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** 获取工具实例日志 获取指定工具实例的日志 GET /admin/toolset/instances/${param0}/log */
export async function getAdminToolsetInstancesIdLog(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getAdminToolsetInstancesIdLogParams,
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<any>(`/admin/toolset/instances/${param0}/log`, {
    method: 'GET',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** 获取工具列表 获取所有可用工具的信息 GET /admin/toolset/list */
export async function getAdminToolsetList(options?: { [key: string]: any }) {
  return request<API.ToolDTO[]>('/admin/toolset/list', {
    method: 'GET',
    ...(options || {}),
  });
}
