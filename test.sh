#!/bin/bash
# heatlth check for all services
echo "Auth Check" && curl -X GET http://localhost:8080/health/auth
echo "Tenant Check" && curl -X GET http://localhost:8080/health/tenant
echo "Client Check" && curl -X GET http://localhost:8080/health/client
echo "Framework Check" && curl -X GET http://localhost:8080/health/framework
echo "Audit Cycle Check" && curl -X GET http://localhost:8080/api/audit-cycle

