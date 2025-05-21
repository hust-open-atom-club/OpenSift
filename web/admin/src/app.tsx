// 运行时配置

import { history, RequestConfig, RuntimeConfig } from "@umijs/max";
import { App } from "antd";
import { getAdminSessionUserinfo } from "./services/csapi/admin";
import { getToken } from "./bearer";
import RightContent from "./components/RightContent";

export type InitialState = {
  user?: API.UserInfoResp,
}

// 全局初始化数据配置，用于 Layout 用户信息和权限初始化
// 更多信息见文档：https://umijs.org/docs/api/runtime-config#getinitialstate
export async function getInitialState(): Promise<InitialState> {
  let user: API.UserInfoResp | undefined = undefined;
  try {
    user = await getAdminSessionUserinfo();
  } catch (e) {
    if (!history.location.pathname.startsWith("/session")) {
      history.push("/session?ret_uri=" + encodeURIComponent(history.location.pathname + history.location.search));
    }
  }
  return {
    user
  }
}

export const request: RequestConfig = {
  baseURL: '/api/v1',
  // other axios options you want
  errorConfig: {
    errorHandler() {
    },
    errorThrower() {
    }
  },
  requestInterceptors: [(url, options) => {
    const bearerToken = getToken();
    if (bearerToken) {
      if (options.headers) {
        options.headers['Authorization'] = `Bearer ${bearerToken}`;
      } else {
        options.headers = { Authorization: `Bearer ${bearerToken}` };
      }
    }
    if (typeof options.data == 'string') {
      options.data = `\"${options.data}\"`;
    } else if (typeof options.data == 'number') {
      options.data = options.data.toString()
    }
    return { url, options };
  }],
  responseInterceptors: []
};

export const layout: RuntimeConfig['layout'] = () => {
  return {
    logo: '/logo.svg',
    title: "关键开源软件筛选管理平台",
    headerTitleRender(logo, title, props) {
      return <div>{logo}</div>
    },
    rightContentRender(headerProps, dom, props) {
      return <RightContent />;
    },
    menu: {
      locale: false,
    },
    layout: 'top'

  };
};


export const rootContainer: RuntimeConfig['rootContainer'] = (lastRoot) => {
  return <App >
    {lastRoot}
  </App>
}