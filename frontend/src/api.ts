import axios from 'axios'
import jwtDecode from 'jwt-decode'
import { ref } from 'vue'

const api = axios.create({
  baseURL: '/api/',
  validateStatus: function (status) {
    return (status >= 200 && status < 300) || status == 401 || status == 403;
  },
  withCredentials: true
})

function getCookie(cname: string) : string {
  const name = cname + "=";
  const decodedCookie = decodeURIComponent(document.cookie);
  const ca = decodedCookie.split(';');
  for(let i = 0; i <ca.length; i++) {
    let c = ca[i];
    while (c.charAt(0) == ' ') {
      c = c.substring(1);
    }
    if (c.indexOf(name) == 0) {
      return c.substring(name.length, c.length);
    }
  }
  return "";
}

function deleteAuthToken() {
    document.cookie = 'ID-Token=; Max-Age=-99999999; Secure=true; SameSite=Strict; Domain=hbank.julianh.de; Path=/';
}

interface AuthToken {
  sub: string
  exp: number
}

function decodeAuthToken(token: string) : AuthToken {
  return jwtDecode(token) as AuthToken
}

function getAuthToken() : AuthToken | null {
  const authTokenStr = getCookie("ID-Token")
  if (!authTokenStr) {
    return null
  }

  return decodeAuthToken(authTokenStr)
}

export const authenticated = ref(false)

auth()

export async function auth() : Promise<string> {
  let authToken = getAuthToken()
  if (authToken) {
    localStorage.setItem("userId", authToken.sub)
    localStorage.setItem("authTokenExpiredAt", authToken.exp.toString())
    authenticated.value = true

    return authToken.sub
  }

  await navigator.locks.request("token_refresh", async () => await api.get("/auth/refresh")) 

  authToken = getAuthToken()
  if (authToken) {
    localStorage.setItem("userId", authToken.sub)
    localStorage.setItem("authTokenExpiredAt", authToken.exp.toString())
    authenticated.value = true

    return authToken.sub
  }

  authenticated.value = false
  return ""
}

export async function logout() : Promise<void> {
  await navigator.locks.request("token_refresh", async () => await api.post("/auth/logout")) 
  localStorage.removeItem("userId")
  localStorage.removeItem("authTokenExpiredAt")
  authenticated.value = false
  deleteAuthToken()
}

export default api

interface Config {
  captchaEnabled: boolean
  emailEnabled: boolean
  minNameLength: number,
  maxNameLength: number,
  minDescriptionLength: number,
  maxDescriptionLength: number,
  minPasswordLength: number,
  maxPasswordLength: number,
  minEmailLength: number,
  maxEmailLength: number,
  maxProfilePictureFileSize: number,
  loginTokenLifetime: number,
  emailCodeLifetime: number,
  authTokenLifetime: number,
  refreshTokenLifetime: number,
  sendEmailTimeout: number,
  maxPageSize: number,
  idProvider: string
}

let loadedConfig: Config | null = null;

export async function config(): Promise<Config> {
  if (!loadedConfig) {
    await loadConfig()
  }
  return loadedConfig ?? {} as Config
}

async function loadConfig() : Promise<void> {
  try {
    const res = await api.get('/status')
    if (!res.data.success) {
      console.error(res.data.message)
    }
    loadedConfig = res.data.config
  } catch (e) {
    console.error(e)
  }
}
