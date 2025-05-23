type IBestAFSRoute = Partial<{
  path: string,
  component: string,
  name: string, // 兼容此写法
  redirect: string,
  icon: string,
  // 更多功能查看
  // https://beta-pro.ant.design/docs/advanced-menu
  // ---
  // 新页面打开
  target: string,
  // 不展示顶栏
  headerRender: boolean,
  // 不展示页脚
  footerRender: boolean,
  layout: boolean,
  // 不展示菜单
  menuRender: boolean,
  // 不展示菜单顶栏
  menuHeaderRender: boolean,
  // 权限配置，需要与 plugin-access 插件配合使用
  access: string,
  // 隐藏子菜单
  hideChildrenInMenu: boolean,
  // 隐藏自己和子菜单
  hideInMenu: boolean,
  // 在面包屑中隐藏
  hideInBreadcrumb: boolean,
  // 子项往上提，仍旧展示,
  flatMenu: boolean,
  routes: IBestAFSRoute[],
}>;

const routes: IBestAFSRoute[] = [
  {
    path: "/",
    redirect: "/gitfile",
  },
  {
    path: "/session",
    component: "session/login",
    name: "登录",
    layout: false,
  },
  {
    path: "/session/gh_callback",
    component: "session/gh_callback",
    name: "登录",
    layout: false,
  },
  {
    name: "仓库存储",
    path: "/gitfile",
    component: "gitfile",
    access: "canViewGitFile"
  },
  {
    name: "工作流",
    path: "/workflow",
    component: "workflow",
  },
  {
    name: "工具集",
    path: "/toolset",
    routes: [
      {
        path: "/toolset",
        redirect: "/toolset/create",
      },
      {
        name: "创建进程",
        path: "create",
        component: "toolset/create",
      },
      {
        name: "附加到进程",
        path: "attach",
        component: "toolset/attach",
      }
    ]

  }
];

export default routes;