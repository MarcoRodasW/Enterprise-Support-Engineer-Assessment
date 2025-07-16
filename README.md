# PartnerHero Enterprise Support Engineer Assessment

## Overview
This practical assessment evaluates your skills in troubleshooting, testing, and issue resolution in a production API system. You’ll work with a Go-based service that interacts with MySQL to address critical production issues reported by users.

---

## Table of Contents
1. [Setup Instructions](#setup-instructions)  
2. [Assessment Tasks](#assessment-tasks)  
3. [Expected Deliverables](#expected-deliverables)  
4. [Troubleshooting Tips](#troubleshooting-tips)  
5. [Time Guidance](#time-guidance)  
6. [Submission](#submission)  
7. [Key Files Structure](#key-files-structure)  
8. [Docker Setup (Optional)](#docker-setup-optional)  
9. [Assessment Focus Areas](#assessment-focus-areas)  

---

## Setup Instructions

### Prerequisites
- Go 1.20+
- MySQL 5.7+
- Postman (or equivalent API client)

### 1. Database Setup
```bash
mysql -u root -p < sample_database.sql
````

### 2. Environment Configuration

Create a `.env` file in the project root:

```env
DB_USER=root
DB_PASS=password
DB_HOST=localhost
```

### 3. Start API Server

```bash
go mod download
go run main.go
```

### 4. Verify Setup

```bash
curl http://localhost:8080/health
# Should return: OK
```

---

## Assessment Tasks

### Task 1: Troubleshoot & Fix Critical Issues

1. **Rate Limit Not Resetting (TICKET-001)**

   * **Symptoms:** Users receive permanent `429 Too Many Requests` after hitting their limit.
   * **Evidence:**

     * Application logs
     * Postman “Load Test” reproduction
   * **Expected:** Rate limits reset daily.

2. **GDPR Compliance Violation (TICKET-002)**

   * **Symptoms:** Full API keys exposed in audit logs.
   * **Evidence:**

     * `GET /api/audit` response via Postman
   * **Expected:** Mask API keys to show only first 6 characters followed by `...`.

---

### Task 2: Implement Security Enhancement

* Create a GDPR‑compliant database view: `masked_audit_logs`.
* Update the `/api/audit` endpoint to use this view instead of exposing raw data.
* Verify no sensitive data is exposed.

---

### Task 3: Validation & Testing

* **Postman Tests:**

  * Verify rate limit resets after 24 hours.
  * Confirm audit logs show masked API keys.
  * Ensure no sensitive data is present.

* **Go Tests:**

  * Add at least one unit test covering critical functionality (e.g., rate‑limit reset logic, key masking).

---

### Task 4: Customer Communication

* Draft a client‑facing email in `/docs/email.md` that:

  * Explains the issue resolution in non‑technical terms.
  * Provides steps for the client to verify fixes.
  * Sets expectations for deployment timeline.

---

### Task 5: Service Monitoring (Node.js Validation)

Implement a Node.js script to monitor API health and performance:

1. **Requirements**:
   - Create `service-monitor.js` that checks:
     - `/api/export` (authenticated with API key `key_DEF456`)
     - `/api/audit`
     - `/health`
   - Validate:
     - HTTP 200 status for all endpoints
     - Response latency < 500ms
     - Proper error handling for timeouts/errors
   - Generate timestamped JSON reports
   - Exit with code 0 (success) only if all services healthy

2. **Validation**:
   ```bash
   # Test with healthy services
   node service-monitor.js
   echo $? # Should return 0
   
   # Test with simulated failure (stop API server)
   node service-monitor.js
   echo $? # Should return 1
   ```
   
3. **Deliverables**:
     - Functional monitoring script
     - Sample report in /monitoring/report-<timestamp>.json
     - Brief validation summary in /docs/tests.md

---

## Expected Deliverables

1. **Code Fixes**

   * Rate limit reset logic in Go.
   * API key masking implementation.
   * `masked_audit_logs` SQL view.
   * Service monitoring script (service-monitor.js)

2. **Validation Evidence**

   * Updated Postman collection with tests.
   * Screenshots of passing tests.
   * Sample SQL queries used for verification.
   * Sample monitoring reports
   * Node.js script execution results

3. **Documentation**

   * `/docs/tests.md`: Detailed test plan including monitoring validation
   * `/docs/email.md`: Customer email draft

4. **Pull Request**

   * Branch must be named as your **Full Name**
   * Include all code and documentation.
   * Clear PR description explaining your approach and changes.

---

## Troubleshooting Tips

* Use Postman’s “Load Test” feature to reproduce rate‑limiting issues.
* Inspect application logs for error patterns.
* Test with different API keys:

  * `key_ABC123` (should be rate limited)
  * `key_DEF456` (should have audit access).
* Review the database schema in `sample_database.sql`.
* For Node.js monitoring: Install dependencies with `npm install axios`

---

## Time Guidance

| Task                    | Estimated Time      |
| ----------------------- | ------------------- |
| Setup                   | 15 minutes          |
| Troubleshooting & Fixes | 2 hours             |
| Testing & Validation    | 1 hour 15 minutes   | 
| Documentation           | 30 minutes          |
| Final Review            | 15 minutes          |

## Completion Window
You'll have **48 hours** from receiving this assessment to submit your solution. 
We estimate the active working time is approximately 4 hours.

Please schedule your work during times that fit your personal commitments.

---

## Submission

1. Create branch named with your **full name** (e.g., `christian-middle-rivera`)
2. Commit all updates (code, tests, docs) to this branch.
3. Push to your private fork.
4. Open a Pull Request including:

   * All deliverables.
   * Comments explaining your approach and fixes.

---

## Key Files Structure

```
.
├── main.go
├── service-monitor.js       
├── sample_database.sql
├── collection.json
├── .env
├── docker-compose.yml
├── go.mod
├── README.md
├── docs/
│   ├── email.md
│   └── tests.md             
└── monitoring/
    ├── logs.txt
    └── monitoring-report-*.json
```

---

## Docker Setup (Optional)

For candidates preferring Docker, use the provided `docker-compose.yml`:

```yaml
version: '3.8'
services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      DB_HOST: db
      DB_USER: root
      DB_PASS: password
    depends_on:
      db:
        condition: service_healthy

  db:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: wikimedia_prod
    ports:
      - "3306:3306"
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-p$${MYSQL_ROOT_PASSWORD}"]
      interval: 5s
      timeout: 10s
      retries: 5
    volumes:
      - ./sample_database.sql:/docker-entrypoint-initdb.d/init.sql
```

---

## Assessment Focus Areas

1. **Troubleshooting Skills**

   * Diagnose and fix rate‑limiting logic.
   * Identify and remediate security vulnerabilities.

2. **Technical Solutions**

   * Implement time‑based rate limits.
   * Add API key masking in Go and SQL.
   * Create GDPR‑compliant database views.
   * Implement Node.js service monitoring
   * Implement validation for API health metrics

3. **Validation**

   * Write and automate Postman tests.
   * Perform manual verification steps.
   * Automated health checks with Node.js

4. **Communication**

   * Produce clear, concise technical documentation.
   * Craft a customer‑facing email in plain language.
   * Provide a well‑structured PR description.
