import type { InitialState } from "./app";

export default (initialState: InitialState) => {
  const policy = initialState?.user?.policy || [];
  const isSuperAdmin = initialState?.user?.policy?.includes("all") || false;

  const access: {
    [key: string]: boolean;
  } = {
    canViewGitFile: policy.includes('gitfile'),
  };

  if (isSuperAdmin) {
    for (const key in access) {
      access[key] = true;
    }
  }
  return access;
};
