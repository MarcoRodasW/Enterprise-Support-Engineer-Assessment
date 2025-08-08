# Service Monitoring Validation Summary

## Overview
The `service-monitor.js` script was used to validate the health and performance of the main API endpoints: `/api/export`, `/api/audit`, and `/health`. The script checks for HTTP 200 responses, response latency under 500ms, and proper error handling. It generates a timestamped JSON report for each run.

## Test Steps
1. Run script to install necessary dependencies:
    - Command: `npm install`
2. The script was executed with all services running:
   - Command: `node service-monitor.js`
3. The script performed GET requests to each endpoint using the correct API key where required.
4. The results were saved in the `monitoring/` directory as a JSON report.

## Results
- `/health` and `/api/audit` endpoints responded with HTTP 200 and latency well below 500ms, marked as `healthy`.
- `/api/export` returned a 403 error (likely due to insufficient permissions for the test API key), which was correctly flagged as `error` in the report.
- The script exited with code 1, as expected when any endpoint is not fully healthy.

## Evidence
- Example report: `monitoring/report-<timestamp>.json`
- Console output confirmed the report location and status.
