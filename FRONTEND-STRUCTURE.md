# Frontend Application Structure

**Framework:** React Router v7 + React 19  
**UI Library:** Radix UI + Tailwind CSS  
**State Management:** TanStack Query (React Query)  
**Type Safety:** TypeScript

---

## 📁 Directory Structure

```
apps/frontend/
├── app/
│   ├── api/                      # API client and endpoints
│   │   ├── client.ts            # Axios instance with interceptors
│   │   └── index.ts             # API methods (auth, tenants, clients, users, rbac)
│   │
│   ├── components/              # Reusable components
│   │   ├── layout/              # Layout components
│   │   │   ├── Sidebar.tsx      # Navigation sidebar
│   │   │   └── DashboardLayout.tsx  # Main layout wrapper
│   │   └── ui/                  # shadcn/ui components
│   │       ├── button.tsx
│   │       ├── card.tsx
│   │       ├── input.tsx
│   │       └── label.tsx
│   │
│   ├── contexts/                # React contexts
│   │   └── AuthContext.tsx      # Authentication state & methods
│   │
│   ├── lib/                     # Utilities
│   │   └── utils.ts             # cn() helper for classnames
│   │
│   ├── routes/                  # Page components
│   │   ├── login.tsx            # OAuth login page
│   │   ├── dashboard.tsx        # Main dashboard
│   │   └── tenants._index.tsx   # Tenants list page
│   │
│   ├── types/                   # TypeScript definitions
│   │   └── index.ts             # All type definitions
│   │
│   ├── app.css                  # Global styles
│   ├── root.tsx                 # Root component
│   └── routes.ts                # Route configuration
│
├── public/                      # Static assets
├── package.json                 # Dependencies
├── tsconfig.json               # TypeScript config
├── vite.config.ts              # Vite config
└── react-router.config.ts      # React Router config
```

---

## 🎨 Features Implemented

### 1. Authentication System ✅

**OAuth Integration:**
- Google OAuth login
- Microsoft OAuth login
- Token-based authentication
- Automatic token refresh
- Secure token storage

**Auth Context:**
```typescript
const { user, loading, login, logout, isAuthenticated } = useAuth();
```

**Protected Routes:**
- Automatic redirect to login if unauthenticated
- Token validation on page load
- Refresh token on 401 errors

### 2. Dashboard ✅

**Features:**
- Real-time statistics cards
- Recent tenants list
- Recent clients list
- Quick action buttons
- Responsive grid layout

**Stats Displayed:**
- Total/Active Tenants
- Total/Active Clients
- Total/Active Users
- High Risk Clients

### 3. Tenant Management ✅

**List View:**
- Search functionality
- Grid layout with cards
- Status badges (active/inactive/suspended/pending)
- Quick actions (edit, activate/deactivate)
- Empty state handling

**Features:**
- Subdomain display
- Database name
- Creation date
- Status management

### 4. Navigation & Layout ✅

**Sidebar:**
- Collapsible design
- Active route highlighting
- User profile section
- Logout button
- Permission-based menu items

**Menu Items:**
- Dashboard
- Tenants
- Clients
- Users
- Roles & Permissions
- Assessments (coming soon)
- Settings

### 5. UI Components ✅

**shadcn/ui Components:**
- Button (variants: default, outline, ghost, destructive)
- Card (with header, content, footer)
- Input (with icons support)
- Label (form labels)

**Styling:**
- Tailwind CSS
- Dark mode support (via remix-themes)
- Responsive design
- Smooth transitions

---

## 🔌 API Integration

### API Client Configuration

```typescript
// Base URL from environment
baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080/api'

// Automatic token injection
Authorization: Bearer <token>

// Auto token refresh on 401
```

### Available APIs

#### Authentication
```typescript
api.auth.getGoogleLoginURL()
api.auth.getMicrosoftLoginURL()
api.auth.validateToken()
api.auth.refreshToken()
api.auth.logout()
```

#### Tenants
```typescript
api.tenants.list(params)
api.tenants.getById(id)
api.tenants.create(payload)
api.tenants.update(id, payload)
api.tenants.delete(id)
api.tenants.activate(id)
api.tenants.deactivate(id)
```

#### Clients
```typescript
api.clients.list(params)
api.clients.getById(id)
api.clients.create(payload)
api.clients.update(id, payload)
api.clients.delete(id)
api.clients.getStats(id)
```

#### Users
```typescript
api.users.list(params)
api.users.getById(id)
api.users.create(payload)
api.users.update(id, payload)
api.users.delete(id)
api.users.getRoles(id)
api.users.assignRole(payload)
api.users.removeRole(userId, roleId)
```

#### RBAC
```typescript
api.rbac.listRoles()
api.rbac.getRole(id)
api.rbac.createRole(payload)
api.rbac.updateRole(id, payload)
api.rbac.deleteRole(id)
api.rbac.listPermissions()
api.rbac.assignPermission(roleId, permissionId)
api.rbac.removePermission(roleId, permissionId)
```

#### Dashboard
```typescript
api.dashboard.getData()
api.dashboard.getStats()
```

---

## 📊 Type Definitions

### Core Types

```typescript
// User & Authentication
User, LoginCredentials, AuthResponse, OAuthProvider

// Tenants
Tenant, TenantStatus, TenantSettings, CreateTenantPayload, UpdateTenantPayload

// Clients
Client, ClientStatus, RiskTier, CreateClientPayload, UpdateClientPayload

// RBAC
Role, Permission, UserRole, AssignRolePayload

// Dashboard
DashboardStats, DashboardData, ActivityLog

// Future Scope
Assessment, Question, Document, RiskAssessment

// API Responses
ApiResponse<T>, PaginatedResponse<T>, ApiError

// Utilities
QueryParams, FilterOption, SortOption
```

---

## 🚀 Pages to Build Next

### 1. Tenant Pages
- ✅ List page (completed)
- ⏳ Create page
- ⏳ Edit page
- ⏳ Detail/View page

### 2. Client Pages
- ⏳ List page
- ⏳ Create page
- ⏳ Edit page
- ⏳ Detail/View page
- ⏳ Risk assessment view

### 3. User Pages
- ⏳ List page
- ⏳ Create page
- ⏳ Edit page
- ⏳ Role assignment

### 4. RBAC Pages
- ⏳ Roles list
- ⏳ Create/Edit role
- ⏳ Permissions list
- ⏳ Permission assignment

### 5. Settings Pages
- ⏳ Profile settings
- ⏳ Tenant settings
- ⏳ Security settings
- ⏳ Notification preferences

### 6. Assessment Pages (Future)
- ⏳ Assessment list
- ⏳ Create assessment
- ⏳ Assessment builder
- ⏳ Results view

---

## 🎯 Future Features

### Phase 1: Core CRUD Operations
- Complete all list/create/edit/delete pages
- Form validation with Zod
- Error handling and toast notifications
- Loading states and skeletons

### Phase 2: Advanced Features
- **Search & Filters:**
  - Advanced search
  - Multi-field filters
  - Saved filters
  
- **Data Tables:**
  - Sortable columns
  - Pagination
  - Bulk actions
  - Export to CSV/Excel

- **File Upload:**
  - Document management
  - Drag & drop
  - Progress indicators
  - File preview

### Phase 3: Assessment Module
- **Assessment Builder:**
  - Question templates
  - Custom questions
  - Scoring logic
  - Conditional questions

- **Assessment Execution:**
  - Client portal
  - Progress tracking
  - Auto-save
  - File attachments

- **Results & Reporting:**
  - Score calculation
  - Risk visualization
  - PDF reports
  - Historical comparison

### Phase 4: Risk Management
- **Risk Dashboard:**
  - Risk heatmap
  - Trend analysis
  - Alerts & notifications
  
- **Risk Scoring:**
  - Automated scoring
  - Custom weights
  - Risk categories
  - Remediation tracking

### Phase 5: Collaboration
- **Comments & Notes:**
  - Threaded discussions
  - @mentions
  - Activity feed

- **Workflows:**
  - Approval workflows
  - Task assignment
  - Due dates & reminders
  - Email notifications

### Phase 6: Analytics & Reporting
- **Dashboards:**
  - Custom widgets
  - Real-time updates
  - Drill-down capabilities

- **Reports:**
  - Scheduled reports
  - Custom templates
  - Multi-format export
  - Email delivery

---

## 🛠️ Development Commands

```bash
# Install dependencies
pnpm install

# Start development server
pnpm dev

# Build for production
pnpm build

# Start production server
pnpm start

# Type checking
pnpm typecheck
```

---

## 🔐 Environment Variables

```env
# API Configuration
VITE_API_URL=http://localhost:8080/api

# OAuth (if needed for client-side)
VITE_GOOGLE_CLIENT_ID=your_google_client_id
VITE_MICROSOFT_CLIENT_ID=your_microsoft_client_id

# Feature Flags
VITE_ENABLE_ASSESSMENTS=false
VITE_ENABLE_RISK_MODULE=false
```

---

## 📱 Responsive Design

**Breakpoints:**
- Mobile: < 640px
- Tablet: 640px - 1024px
- Desktop: > 1024px

**Features:**
- Mobile-first approach
- Collapsible sidebar on mobile
- Responsive grids
- Touch-friendly buttons
- Optimized for all screen sizes

---

## 🎨 Design System

**Colors:**
- Primary: Brand color
- Secondary: Accent color
- Destructive: Error/danger
- Muted: Subtle backgrounds
- Accent: Hover states

**Typography:**
- Headings: Bold, tracking-tight
- Body: Regular, readable
- Captions: Small, muted

**Spacing:**
- Consistent padding/margins
- Grid gaps
- Component spacing

---

## ✅ Best Practices Implemented

1. **Type Safety:**
   - Full TypeScript coverage
   - Strict type checking
   - API response typing

2. **Code Organization:**
   - Feature-based structure
   - Reusable components
   - Separation of concerns

3. **Performance:**
   - React Query caching
   - Lazy loading
   - Code splitting

4. **Security:**
   - Token-based auth
   - Secure storage
   - CSRF protection

5. **UX:**
   - Loading states
   - Error handling
   - Empty states
   - Responsive design

---

## 📚 Next Steps

1. **Complete CRUD Pages:**
   - Tenant create/edit
   - Client management
   - User management
   - Role management

2. **Add Form Validation:**
   - Zod schemas
   - Error messages
   - Field validation

3. **Implement Notifications:**
   - Toast messages
   - Success/error feedback
   - Action confirmations

4. **Add Data Tables:**
   - Sortable columns
   - Pagination
   - Filters
   - Bulk actions

5. **Build Assessment Module:**
   - Question builder
   - Assessment execution
   - Results dashboard

---

**Status:** Foundation Complete ✅  
**Next Phase:** CRUD Operations & Forms  
**Target:** Production-Ready TPRM Platform
