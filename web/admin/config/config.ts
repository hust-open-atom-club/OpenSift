import { defineConfig } from "@umijs/max";
import { join } from "path";
import routes from "./route";

export default defineConfig({
  presets: ['umi-presets-pro'],
  antd: {},
  access: {},
  model: {},
  initialState: {},
  request: {},
  layout: {
    title: "@umijs/max",
  },
  routes,
  npmClient: "yarn",
  base: "/admin",
  publicPath: "/admin/",
  tailwindcss: {},
  proxy: {
    '/api': {
      target: 'http://localhost:5000/',
      changeOrigin: true,
      ws: true,
    }
  },
  openAPI: {
    requestLibPath: "import { request } from '@umijs/max';",
    schemaPath: join(__dirname, 'csapi.json'),
    projectName: 'csapi',
    mock: false,
  },
  esbuildMinifyIIFE: true
});
