const axios = require('axios');
const fs = require('fs');
const path = require('path');


const MAX_LATENCY_MS = 500;
const MAX_TIMEOUT_MS = 2000;
const API_KEY = 'key_DEF456';
const ENDPOINTS = [
  {
    name: 'export',
    url: 'http://localhost:8080/api/export',
    headers: { 'X-API-Key': API_KEY }
  },
  {
    name: 'audit',
    url: 'http://localhost:8080/api/audit',
    headers: { 'X-API-Key': API_KEY }
  },
  {
    name: 'health',
    url: 'http://localhost:8080/health',
    headers: {}
  }
];

function getStatus(res, latency) {
  if (res.status === 200 && latency < MAX_LATENCY_MS) return 'healthy';
  if (res.status === 200) return 'slow';
  return 'error';
}

async function checkEndpoint({ name, url, headers }) {
  const start = Date.now();
  try {
    const res = await axios.get(url, { headers, timeout: MAX_TIMEOUT_MS });
    const latency = Date.now() - start;
    return {
      endpoint: name,
      url,
      httpCode: res.status,
      status: getStatus(res, latency),
      latency,
      error: null,
      timestamp: new Date().toISOString()
    };
  } catch (e) {
    const latency = Date.now() - start;
    return {
      endpoint: name,
      url,
      httpCode: e.response ? e.response.status : null,
      status: 'error',
      latency,
      error: e.message,
      timestamp: new Date().toISOString()
    };
  }
}

function writeReport(results) {
  const reportDir = path.join(__dirname, 'monitoring');
  if (!fs.existsSync(reportDir)) fs.mkdirSync(reportDir, { recursive: true });
  const reportPath = path.join(reportDir, `report-${Date.now()}.json`);
  fs.writeFileSync(reportPath, JSON.stringify(results, null, 2));
  console.log(`Report written to ${reportPath}`);
}

async function main() {
  const results = await Promise.all(ENDPOINTS.map(checkEndpoint));
  writeReport(results);
  const allHealthy = results.every(r => r.status === 'healthy');
  if (!allHealthy) {
    console.error('Some endpoints are unhealthy or slow.');
    process.exit(1);
  }
  process.exit(0);
}

main();
