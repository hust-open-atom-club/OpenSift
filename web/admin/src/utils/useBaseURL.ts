import { useAppData } from "@umijs/max";

/**
 * 
 * @returns The base name of the application, ensuring it does not end with a slash. e.g., "/admin", "", just join it like ${location}${baseURL}/xxxx.
 */
export default function () {
  const appData = useAppData();
  let baseName = (appData.basename || "/");
  if (baseName.endsWith("/")) {
    baseName = baseName.slice(0, -1);
  }
  return baseName;
}