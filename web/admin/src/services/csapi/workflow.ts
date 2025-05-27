// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** 获取 workflow 日志 获取指定轮次和名称的 workflow 日志文件 GET /admin/workflows/${param0}/logs/${param1} */
export async function getAdminWorkflowsIdLogsName(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getAdminWorkflowsIdLogsNameParams,
  options?: { [key: string]: any },
) {
  const { id: param0, name: param1, ...queryParams } = params;
  return request<string>(`/admin/workflows/${param0}/logs/${param1}`, {
    method: 'GET',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** 杀死 workflow 任务 杀死当前运行中的 workflow 任务 POST /admin/workflows/kill */
export async function postAdminWorkflowsKill(
  body: API.KillWorkflowJobReq,
  options?: { [key: string]: any },
) {
  return request<any>('/admin/workflows/kill', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** 获取最大 workflow 轮次 ID 获取当前最大的 workflow 轮次 ID GET /admin/workflows/maxRounds */
export async function getAdminWorkflowsMaxRounds(options?: {
  [key: string]: any;
}) {
  return request<API.RoundResp>('/admin/workflows/maxRounds', {
    method: 'GET',
    ...(options || {}),
  });
}

/** 获取指定 workflow 轮次详情 根据轮次 ID 获取 workflow 详情 GET /admin/workflows/rounds/${param0} */
export async function getAdminWorkflowsRoundsId(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getAdminWorkflowsRoundsIdParams,
  options?: { [key: string]: any },
) {
  const { id: param0, ...queryParams } = params;
  return request<API.RoundDTO>(`/admin/workflows/rounds/${param0}`, {
    method: 'GET',
    params: { ...queryParams },
    ...(options || {}),
  });
}

/** 启动或停止 workflow 启动或停止 workflow 运行状态 POST /admin/workflows/status */
export async function postAdminWorkflowsStatus(
  body: API.UpdateWorkflowStatusReq,
  options?: { [key: string]: any },
) {
  return request<any>('/admin/workflows/status', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}
