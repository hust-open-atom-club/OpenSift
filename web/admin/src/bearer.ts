let token = "";

export function getToken() {
  if (!token) {
    token = localStorage.getItem("token") || "";
  }
  return token;
}

export function setToken(t: string) {
  token = t;
  localStorage.setItem("token", t);
}