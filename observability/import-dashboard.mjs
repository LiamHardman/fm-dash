// Imports a Grafana dashboard JSON file via the HTTP API. Node has no shell-encoding
// pitfalls (unlike passing JSON through bash command substitution), so this is the
// reliable path for importing all 6 app-observability dashboards.
import { readFileSync } from "node:fs";

const file = process.argv[2];
if (!file) {
  console.error("usage: node import-dashboard.mjs <dashboard.json>");
  process.exit(1);
}

const dashboard = JSON.parse(readFileSync(file, "utf8"));
const auth = Buffer.from("admin:admin").toString("base64");

const body = {
  dashboard,
  overwrite: true,
  inputs: [
    { name: "DS_PROMETHEUS", type: "datasource", pluginId: "prometheus", value: "PBFA97CFB590B2093" },
  ],
  folderId: 0,
};

const res = await fetch("http://localhost:3001/api/dashboards/import", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    Authorization: `Basic ${auth}`,
  },
  body: JSON.stringify(body),
});

const json = await res.json();
console.log(res.status, JSON.stringify(json));
if (!res.ok || json.imported !== true) {
  process.exit(1);
}
