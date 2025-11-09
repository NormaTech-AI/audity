---
title: "Internal TPRM Audit Platform — Product Requirements Document (PRD)"
version: "1.0"
date: "November 3, 2025"
---

# Internal TPRM Audit Platform (PRD)

## 1. Introduction
### 1.1 Problem
Nishaj Infotech currently manages a high volume of client audits during "audit season" using manual processes. Clients, such as Bagadia, must submit audit reports to their regulators (e.g., NSE, BSE, MCX). This manual process of distributing checklists, collecting evidence, tracking progress, and generating final reports is inefficient, time-consuming, and difficult to scale.
### 1.2 Goal
The primary goal is to build an internal, multi-tenant TPRM (Third-Party Risk Management) software. This platform will centralize and streamline the entire audit lifecycle.

This includes:

- Onboarding clients (e.g., Bagadia, Kiwi Capital) and assigning specific compliance checklists
- Providing clients a portal to answer questions and upload evidence
- Enabling Nishaj auditors to efficiently validate evidence, approve or reject submissions, and manage the workflow
- Generating a final, signed audit report that can be delivered to the client
### 1.3 MVP Scope
This document defines the Minimum Viable Product (MVP) required to launch the platform. The focus is on delivering the core workflows for client management, evidence collection, and report generation. Advanced features, such as full OPA integration for RBAC, are deferred post-MVP.
## 2. User Personas
The system will be built to serve six key user roles:
### 2.1 Nishaj Infotech (Internal)

- **Admin** ("NishajAdmin"): System super-user. Manages all users, onboards new clients, and assigns compliance frameworks.
- **Auditor**: The primary reviewer. Validates client evidence against checklist questions, approves/rejects submissions, and generates the final audit report.
- **Team Member** (e.g., Nikhil, Ashish): Supports the Auditor and may be involved in discussions or partial reviews.
- **POC (Internal)**: The main Nishaj point-of-contact managing the client relationship and overall audit project.

### 2.2 Client (External)

- **POC (Client)**: The primary contact at the client company (e.g., Bagadia). Responsible for the audit on their end and can delegate questions.
- **Assigned Stakeholder**: Employee at the client company who is assigned specific questions to answer and provide evidence for.
## 3. Core Workflows & Features (MVP)
### 3.1 Workflow: Client Onboarding

As an Admin:

- Access the Admin Panel to manage clients
- Onboard a new client by filling out their details (Client Name, POC Email)
- Select which compliance checklists (e.g., NSE, BSE, NCDEX) apply to the client
- Set a due date for the audit

On submission:

- The system provisions a new, isolated PostgreSQL database and a MinIO bucket for the client
- The Client POC receives an email notification to log in and begin the audit
### 3.2 Workflow: Questionnaire & Evidence Submission

As a Client POC or Assigned Stakeholder:

- Log in via OIDC and view a dashboard of assigned audits and their progress
- View the full questionnaire
- Delegate specific questions to Assigned Stakeholders (Client POC)
- Assigned Stakeholders see only the questions assigned to them
- Answer each question (Yes/No/N/A or free text) with a mandatory explanation
- Upload evidence files (PDFs, logs, screenshots, etc.) per question

### 3.3 Workflow: Audit Review & Validation

As an Auditor:

- View a list of active clients and their submission progress
- Select a client and review submitted answers and evidence for each question
- Approve a submission, Reject (with mandatory comments), or Refer for internal discussion

Clients are notified of rejections and can resubmit new evidence.

### 3.4 Workflow: Report Generation & Signing

Audit report flow:

- Auditor initiates "Generate Report" once all questions are Approved
- System generates a final, unsigned PDF from an HTML/CSS template
- Auditor downloads the unsigned PDF and signs it locally (DSC), then uploads the signed PDF back to the platform
- Signed report is made available to the Client POC for download

## 4. Technical Architecture & Stack

High-level technologies and rationale:

| Category      | Technology / Approach                         | Rationale / Notes |
|---------------|-----------------------------------------------|-------------------|
| Monorepo      | Turborepo                                     | Monorepo tooling for the frontend/backend packages |
| Backend       | Go (Echo framework)                           | Fast, typed backend with good concurrency support |
| Frontend      | React (Vite)                                  | Modern frontend with fast dev server |
| Database      | PostgreSQL                                    | Separate DB per tenant (MVP) |
| DB Access     | sqlc (Go codegen)                             | Strong typed SQL access |
| UI            | shadcn/ui & Tailwind CSS                      | Component library + utility CSS |
| State Mgmt    | Zustand                                       | Lightweight state management |
| Data Fetching | TanStack Query                                | Caching and server state management |
| File Storage  | MinIO                                         | S3-compatible object storage, per-tenant buckets |
| Auth          | OIDC (Google/Microsoft) & JWTs                | External identity + JWTs for API auth |
| Config        | Viper                                         | Centralized config management |
| Container     | Docker                                         | Containerize services |
| Orchestration | Kubernetes (K8s)                               | Production orchestration |
| CI/CD         | GitHub Actions                                 | Automated pipelines |


## 5. Key Architectural Decisions (MVP)

### Data Tenancy: Separate Database per Client

- Strict data isolation: each client has its own dedicated PostgreSQL database and its own MinIO bucket
- A central `tenant_db` is used by the Go backend to store the mapping of clients to their encrypted DB credentials

### Automatic Tenant Provisioning

When an Admin onboards a new client, the Go backend will automatically:

1. Create a new, isolated database and a unique DB user/password
2. Run DB migrations (sqlc schema) on the new database
3. Store the new client's encrypted credentials in `tenant_db`

An Echo middleware will read the incoming user's JWT, look up their `client_id`, and attach the correct client-specific DB connection to the request context.

### Authorization (RBAC): DB-backed MVP

- Authorization will be handled via a database-backed RBAC model
- The system will use `roles`, `permissions`, and `role_permissions` tables to define what each persona can do
- A Go middleware will check permissions on authenticated API routes; full OPA integration is deferred post-MVP

### Report Generation: HTML-to-PDF

- The final audit report will be generated by populating a dynamic HTML/CSS template with all audit data
- The Go backend will use a headless browser library (e.g., `chromedp`) to convert the HTML into a high-fidelity PDF to preserve styles and layout

### Digital Signatures (DSC): Manual Upload

- The MVP does not integrate directly with DSC hardware or signing APIs
- Workflow: Auditor downloads unsigned PDF, signs locally with existing DSC tools, then uploads the signed PDF back to the platform

