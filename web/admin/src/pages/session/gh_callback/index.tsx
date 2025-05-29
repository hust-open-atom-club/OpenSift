import { getAdminSessionGithubCallback } from "@/services/csapi/admin";
import { LoadingOutlined } from "@ant-design/icons";
import { useModel, useParams, useRequest, useRoutes, useSearchParams } from "@umijs/max"
import { Button, Spin } from "antd";
import { history } from "@umijs/max";
import { setToken } from "@/bearer";
import useBaseURL from "@/utils/useBaseURL";

export default function Login() {
  const [search] = useSearchParams();
  const code = search.get("code");
  const ret_uri = search.get("ret_uri");
  const { refresh } = useModel("@@initialState");

  const baseURL = useBaseURL();

  const { loading, error } = useRequest(async () => {
    if (!code) {
      throw new Error("code is required");
    }
    const { token } = await getAdminSessionGithubCallback({
      code,
    })
    token && setToken(token);
    await refresh();
    setTimeout(() => {
      let u = ret_uri;
      if (baseURL) u = u?.slice(baseURL.length) || null;
      // redirect to github login page
      history.push(u || "/");
    }, 0);
  });



  return <div className="h-screen w-screen overflow-auto" style={{
    background: "radial-gradient(circle, rgba(208, 227, 255,1) 0%, rgba(255,255,255,1) 100%)",
  }}>
    <div className="w-96 mx-auto">
      <img className="h-20 mt-40 mx-auto mb-8" src="/logo.svg" alt="logo" />
      <div className="flex flex-col border bg-white shadow-md rounded-lg p-10 h-72">
        <div className="grow text-center">
          {!!loading && <><div className="mb-8"> <Spin /> </div> 正在完成登录，请稍等...</>}
          {!!error && <div className="text-red-500">登录失败，请检查该 GitHub 账号是否授权。
            <div className="mt-4">
              <Button onClick={() => {
                history.push("/session?ret_uri=" + encodeURIComponent(ret_uri || "/"));
              }} className="w-full" type="primary" >重新登录</Button>
            </div>
          </div>}
          {!loading && !error && <div className="text-green-500">登录成功！</div>}
        </div>
        <div className="text-gray-500 text-sm mt-4">
          登录相关问题，请在飞书群组中反馈。
        </div>
      </div>

    </div>
  </div>


}