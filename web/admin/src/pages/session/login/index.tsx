import { getAdminSessionGithubClientid } from "@/services/csapi/admin";
import { GithubFilled } from "@ant-design/icons";
import { useAppData, useModel, useRequest, useSearchParams } from "@umijs/max";
import { App, Button } from "antd";
import { history } from "@umijs/max";
import { getToken } from "@/bearer";
import useBaseURL from "@/utils/useBaseURL";

export default function Login() {
  const { message } = App.useApp();
  const [search] = useSearchParams();
  const retUri = search.get("ret_uri") || "/";
  const { initialState } = useModel('@@initialState');
  if (initialState?.user && getToken()) {
    // already logged in
    // redirect to github login page
    history.push(retUri);
  }

  const baseURL = useBaseURL();

  const {
    run: gotoGitHub,
    loading,
  } = useRequest(async () => {
    let clientid: string | undefined = "";
    try {
      clientid = (await getAdminSessionGithubClientid()).clientId;
    } catch (e) {
      console.error(e);
      return;
    }
    if (!clientid) {
      message.error("获取配置信息失败，请稍后再试。");
      return;
    }
    const redirectURL = `${window.location.origin}${baseURL}/session/gh_callback?ret_uri=${encodeURIComponent(retUri)}`;
    const githubURL = `https://github.com/login/oauth/authorize?client_id=${clientid}&redirect_uri=${encodeURIComponent(redirectURL)}&scope=read:user`;

    setTimeout(() => {
      // redirect to github login page
      window.location.href = githubURL;
    }, 0);

    await message.loading({
      content: "正在跳转到 GitHub 登录页面...",
      key: "login",
      duration: 10,
    });
  }, {
    manual: true,
  })

  return <div className="h-screen w-screen overflow-auto" style={{
    background: "radial-gradient(circle, rgba(208, 227, 255,1) 0%, rgba(255,255,255,1) 100%)",
  }}>
    <div className="w-96 mx-auto">
      <img className="h-20 mt-40 mx-auto mb-8" src="/logo.svg" alt="logo" />
      <div className="flex flex-col border bg-white shadow-md rounded-lg p-10 h-72">
        <div>您正在使用管理平台，需要认证后继续。</div>
        <div className="flex flex-col mt-10 grow">
          <Button loading={loading} icon={<GithubFilled />} size="large" color="default" variant="solid" onClick={gotoGitHub}>使用 GitHub 继续</Button>
        </div>
        <div className="text-gray-500 text-sm mt-4">
          登录相关问题，请在飞书群组中反馈。
        </div>
      </div>

    </div>
  </div>


}