// @ts-ignore
/* eslint-disable */
import { request } from '@umijs/max';

/** Get Git File List GET /admin/gitfiles */
export async function getAdminGitfiles(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getAdminGitfilesParams,
  options?: { [key: string]: any },
) {
  return request<API.PageDTOModelGitFileDTO>('/admin/gitfiles', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** Append to collector manual list POST /admin/gitfiles/manual */
export async function postAdminGitfilesManual(
  body: API.GitFileAppendManualReq,
  options?: { [key: string]: any },
) {
  return request<any>('/admin/gitfiles/manual', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}

/** Start Git File Collector Starts the Git File collection process POST /admin/gitfiles/start */
export async function postAdminGitfilesStart(options?: { [key: string]: any }) {
  return request<any>('/admin/gitfiles/start', {
    method: 'POST',
    ...(options || {}),
  });
}

/** Get git files statistics and collector status GET /admin/gitfiles/status */
export async function getAdminGitfilesStatus(options?: { [key: string]: any }) {
  return request<API.GitFileStatusResp>('/admin/gitfiles/status', {
    method: 'GET',
    ...(options || {}),
  });
}

/** Stop Git File Collector Stops the Git File collection process POST /admin/gitfiles/stop */
export async function postAdminGitfilesStop(options?: { [key: string]: any }) {
  return request<any>('/admin/gitfiles/stop', {
    method: 'POST',
    ...(options || {}),
  });
}

/** Get score histories Get score histories by git link GET /histories */
export async function getHistories(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getHistoriesParams,
  options?: { [key: string]: any },
) {
  return request<API.PageDTOModelResultDTO>('/histories', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** Get ranking results Get ranking results, optionally including all details GET /rankings */
export async function getRankings(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getRankingsParams,
  options?: { [key: string]: any },
) {
  return request<API.PageDTOModelRankingResultDTO>('/rankings', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** Search score results by git link Search score results by git link
NOTE: All details are ignored, should use /results/:scoreid to get details
NOTE: Maxium take count is 1000 GET /results */
export async function getResults(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getResultsParams,
  options?: { [key: string]: any },
) {
  return request<API.PageDTOModelResultDTO>('/results', {
    method: 'GET',
    params: {
      ...params,
    },
    ...(options || {}),
  });
}

/** Get score results Get score results, including all details by scoreid GET /results/${param0} */
export async function getResultsScoreid(
  // 叠加生成的Param类型 (非body参数swagger默认没有生成对象)
  params: API.getResultsScoreidParams,
  options?: { [key: string]: any },
) {
  const { scoreid: param0, ...queryParams } = params;
  return request<API.ResultDTO>(`/results/${param0}`, {
    method: 'GET',
    params: { ...queryParams },
    ...(options || {}),
  });
}
