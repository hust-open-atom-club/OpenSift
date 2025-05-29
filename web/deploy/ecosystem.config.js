// prepare for nginx.conf
const fs = require('node:fs');

const backend = process.env.BACKEND || "http://apiserver:5000";
const nginxConf = fs.readFileSync("/app/deploy/nginx.conf.tmpl", "utf8")
const updatedNginxConf = nginxConf.replace(/{{API_ENDPOINT_URL}}/g, backend);
fs.writeFileSync("/app/deploy/nginx.conf", updatedNginxConf);


module.exports = {
  apps: [{
    name: "public_frontend",
    script: "/app/public/server.js",
    env: {
      HOSTNAME: "0.0.0.0",
      PORT: 3000,
      NODE_ENV: "production",
    }
  }, {
    name: "nginx",
    script: "nginx -c /app/deploy/nginx.conf",
  }]
}
