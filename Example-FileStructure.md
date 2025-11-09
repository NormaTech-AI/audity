This is an example for how we use to implement file structure for monolithic architechture of both frontend and backend
```
.
|-- README.md
|-- check_db_roles.ps1
|-- check_db_roles.sh
|-- db_test.sh
|-- diagnose_roles.ps1
|-- diagnose_roles.sh
|-- fix_role_names.ps1
|-- manual_db_check.sh
|-- frontend
|   |-- Dockerfile
|   |-- README.md
|   |-- app
|   |   |-- api
|   |   |   `-- index.ts
|   |   |-- app.css
|   |   |-- auth
|   |   |   |-- auth-dev-tools.tsx
|   |   |   |-- hooks
|   |   |   |   |-- useAuth.ts
|   |   |   |   `-- useLogin.ts
|   |   |   |-- index.ts
|   |   |   |-- layouts
|   |   |   |   |-- guest-layout.tsx
|   |   |   |   `-- protected-layout.tsx
|   |   |   `-- providers
|   |   |       |-- auth-provider.tsx
|   |   |       `-- dev-tools-provider.tsx
|   |   |-- components
|   |   |   |-- (auth)
|   |   |   |   `-- login-form.tsx
|   |   |   |-- ActiveCyclesWidget.tsx
|   |   |   |-- bulk-actions.tsx
|   |   |   |-- cycle-action-card.tsx
|   |   |   |-- delegation-modal.tsx
|   |   |   |-- full-page-loader.tsx
|   |   |   |-- instructions-step.tsx
|   |   |   |-- mode-toggle.tsx
|   |   |   |-- notification-bell.tsx
|   |   |   |-- panel-header.tsx
|   |   |   |-- providers
|   |   |   |   `-- tanstack-provider.tsx
|   |   |   |-- questions-step.tsx
|   |   |   |-- rating-templates.tsx
|   |   |   |-- ratings-step.tsx
|   |   |   |-- sidebar
|   |   |   |   |-- app-sidebar.tsx
|   |   |   |   |-- data.json
|   |   |   |   |-- nav-main.tsx
|   |   |   |   `-- nav-user.tsx
|   |   |   |-- stats-card.tsx
|   |   |   |-- summary-step.tsx
|   |   |   |-- ui
|   |   |   |   |-- accordion.tsx
|   |   |   |   |-- alert-dialog.tsx
|   |   |   |   |-- alert.tsx
|   |   |   |   |-- avatar.tsx
|   |   |   |   |-- badge.tsx
|   |   |   |   |-- button.tsx
|   |   |   |   |-- card.tsx
|   |   |   |   |-- checkbox.tsx
|   |   |   |   |-- command.tsx
|   |   |   |   |-- dialog.tsx
|   |   |   |   |-- dropdown-menu.tsx
|   |   |   |   |-- form.tsx
|   |   |   |   |-- input.tsx
|   |   |   |   |-- label.tsx
|   |   |   |   |-- popover.tsx
|   |   |   |   |-- progress.tsx
|   |   |   |   |-- radio-group.tsx
|   |   |   |   |-- scroll-area.tsx
|   |   |   |   |-- select.tsx
|   |   |   |   |-- separator.tsx
|   |   |   |   |-- sheet.tsx
|   |   |   |   |-- sidebar.tsx
|   |   |   |   |-- skeleton.tsx
|   |   |   |   |-- slider.tsx
|   |   |   |   |-- sonner.tsx
|   |   |   |   |-- supervisor-combobox.tsx
|   |   |   |   |-- switch.tsx
|   |   |   |   |-- table.tsx
|   |   |   |   |-- textarea.tsx
|   |   |   |   `-- tooltip.tsx
|   |   |   `-- widgets
|   |   |       `-- manager-widget.tsx
|   |   |-- contexts
|   |   |   `-- NotificationContext.tsx
|   |   |-- data
|   |   |   `-- sidebar.data.ts
|   |   |-- hooks
|   |   |   |-- index.ts
|   |   |   |-- use-mobile.ts
|   |   |   |-- useAvailableManagers.ts
|   |   |   |-- useBulkPublishSubmissions.ts
|   |   |   |-- useCreateSubmission.ts
|   |   |   |-- useDashboardData.ts
|   |   |   |-- useDelegateSubmission.ts
|   |   |   |-- useFinalizeSubmission.ts
|   |   |   |-- useGetManagerSubmission.ts
|   |   |   |-- useGetSubmission.ts
|   |   |   |-- useGetSubmissionForCycle.tsx
|   |   |   |-- useManagerSubmissions.ts
|   |   |   |-- useMyActiveCycles.ts
|   |   |   |-- useProfile.ts
|   |   |   |-- usePublishSubmission.ts
|   |   |   |-- usePublishedSubmissions.ts
|   |   |   |-- useSupervisors.ts
|   |   |   |-- useSyncQuestions.ts
|   |   |   |-- useUpdateProfile.ts
|   |   |   |-- useUpsertManagerRatings.ts
|   |   |   |-- useUpsertRatings.ts
|   |   |   `-- useUserPermissions.ts
|   |   |-- lib
|   |   |   `-- utils.ts
|   |   |-- root.tsx
|   |   |-- routes
|   |   |   |-- (app)
|   |   |   |   |-- appraisal.tsx
|   |   |   |   |-- dashboard.tsx
|   |   |   |   |-- layout.tsx
|   |   |   |   |-- performance-report.$submissionId.tsx
|   |   |   |   `-- redirect.tsx
|   |   |   |-- (auth)
|   |   |   |   `-- login.tsx
|   |   |   |-- (manager)
|   |   |   |   |-- analytics.tsx
|   |   |   |   |-- layout.tsx
|   |   |   |   |-- reports.tsx
|   |   |   |   |-- review.tsx
|   |   |   |   `-- team.tsx
|   |   |   |-- action.set-theme.ts
|   |   |   `-- auth.callback.tsx
|   |   |-- routes.ts
|   |   |-- sessions.server.ts
|   |   |-- types
|   |   |   `-- index.ts
|   |   `-- welcome
|   |       |-- logo-dark.svg
|   |       |-- logo-light.svg
|   |       `-- welcome.tsx
|   |-- components
|   |   `-- widgets
|   |-- components.json
|   |-- package.json
|   |-- pnpm-lock.yaml
|   |-- public
|   |   |-- favicon.ico
|   |   |-- logo-light.png
|   |   |-- logo.png
|   |   `-- microsoft_icon.svg
|   |-- react-router.config.ts
|   |-- tsconfig.json
|   `-- vite.config.ts
|-- backend
|   |-- Dockerfile
|   |-- certs
|   |   |-- saml.crt
|   |   `-- saml.key
|   |-- cmd
|   |   |-- seeder
|   |   |   `-- main.go
|   |   `-- server
|   |       `-- main.go
|   |-- config.yaml
|   |-- db
|   |   |-- migrations
|   |   |   |-- 000001_init_schema.down.sql
|   |   |   |-- 000001_init_schema.up.sql
|   |   |   |-- 000002_seed_default_roles_and_permissions.down.sql
|   |   |   `-- 000002_seed_default_roles_and_permissions.up.sql
|   |   `-- queries
|   |       |-- affiliation.sql
|   |       |-- appraisal.sql
|   |       |-- company.sql
|   |       |-- group.sql
|   |       |-- manager.sql
|   |       |-- permission.sql
|   |       |-- role.sql
|   |       |-- token.sql
|   |       `-- user.sql
|   |-- docs
|   |   |-- docs.go
|   |   |-- swagger.json
|   |   `-- swagger.yaml
|   |-- go.mod
|   |-- go.sum
|   |-- internal
|   |   |-- api
|   |   |   |-- handler
|   |   |   |   |-- affiliation_handler.go
|   |   |   |   |-- auth_handler.go
|   |   |   |   |-- company_handler.go
|   |   |   |   |-- cycle_handler.go
|   |   |   |   |-- group_handler.go
|   |   |   |   |-- health_handler.go
|   |   |   |   |-- manager_handler.go
|   |   |   |   |-- parameter_handler.go
|   |   |   |   |-- permission_handler.go
|   |   |   |   |-- role_handler.go
|   |   |   |   |-- submission_handler.go
|   |   |   |   |-- supervisor_handler.go
|   |   |   |   `-- user_handler.go
|   |   |   |-- middleware
|   |   |   |   |-- auth_middleware.go
|   |   |   |   |-- permission_middleware.go
|   |   |   |   `-- token_injector_middleware.go
|   |   |   |-- renderer.go
|   |   |   |-- request
|   |   |   |   |-- affiliation.go
|   |   |   |   |-- appraisal.go
|   |   |   |   |-- auth.go
|   |   |   |   |-- company.go
|   |   |   |   |-- group.go
|   |   |   |   |-- manager.go
|   |   |   |   |-- permission.go
|   |   |   |   |-- role.go
|   |   |   |   `-- user.go
|   |   |   |-- response
|   |   |   |   |-- affiliation.go
|   |   |   |   |-- appraisal.go
|   |   |   |   |-- comapny.go
|   |   |   |   |-- group.go
|   |   |   |   |-- manager.go
|   |   |   |   |-- role.go
|   |   |   |   `-- user.go
|   |   |   `-- router.go
|   |   |-- auth
|   |   |   |-- paseto.go
|   |   |   |-- saml.go
|   |   |   `-- samlSP.go
|   |   |-- config
|   |   |   `-- config.go
|   |   |-- logger
|   |   |   `-- logger.go
|   |   `-- store
|   |       |-- postgres
|   |       |   |-- affiliation.sql.go
|   |       |   |-- appraisal.sql.go
|   |       |   |-- company.sql.go
|   |       |   |-- db.go
|   |       |   |-- enum_helpers.go
|   |       |   |-- group.sql.go
|   |       |   |-- manager.sql.go
|   |       |   |-- models.go
|   |       |   |-- permission.sql.go
|   |       |   |-- querier.go
|   |       |   |-- role.sql.go
|   |       |   |-- stub_methods.go
|   |       |   |-- token.sql.go
|   |       |   `-- user.sql.go
|   |       |-- postgres.go
|   |       `-- store.go
|   |-- sqlc.yaml
|   |-- tmp
|   |   |-- build-errors.log
|   |   `-- main.exe
|   `-- web
|       `-- templates
|           `-- redoc.html
```