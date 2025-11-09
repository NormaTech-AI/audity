# Frontend Quick Start Guide 🚀

## Prerequisites

- Node.js 20+ installed
- pnpm installed (`npm install -g pnpm`)
- Backend services running on `http://localhost:8080`

---

## Installation

```bash
cd apps/frontend
pnpm install
```

---

## Development

```bash
pnpm dev
```

Frontend will start on: **http://localhost:5173**

---

## Environment Setup

Create `.env` file in `apps/frontend/`:

```env
VITE_API_URL=http://localhost:8080/api
```

---

## Project Structure

```
apps/frontend/
├── app/
│   ├── api/              # API client & endpoints
│   ├── components/       # Reusable components
│   ├── contexts/         # React contexts
│   ├── layouts/          # Layout wrappers
│   ├── lib/              # Utilities
│   ├── routes/           # Page components
│   ├── types/            # TypeScript types
│   ├── app.css           # Global styles
│   ├── root.tsx          # Root component
│   └── routes.ts         # Route configuration
├── public/               # Static assets
└── package.json          # Dependencies
```

---

## Available Routes

### Public Routes
- `/login` - OAuth login page

### Protected Routes (Require Authentication)
- `/` or `/dashboard` - Main dashboard
- `/tenants` - Tenants list
- `/clients` - Clients list (to be built)
- `/users` - Users list (to be built)
- `/rbac` - Roles & permissions (to be built)
- `/settings` - Settings (to be built)

---

## Features Implemented ✅

### Authentication
- Google OAuth login
- Microsoft OAuth login
- Token-based auth
- Automatic token refresh
- Protected routes

### Dashboard
- Statistics cards
- Recent activity
- Quick actions
- Responsive design

### Tenant Management
- List view with search
- Status badges
- Quick actions
- Empty states

### UI Components
- Button (multiple variants)
- Card
- Input
- Label
- Sidebar navigation
- Layouts

---

## API Integration

All API calls go through `app/api/index.ts`:

```typescript
import { api } from '~/api';

// Authentication
await api.auth.validateToken();
await api.auth.logout();

// Tenants
const tenants = await api.tenants.list();
const tenant = await api.tenants.getById(id);
await api.tenants.create(payload);

// Dashboard
const data = await api.dashboard.getData();
```

---

## State Management

### React Query (TanStack Query)

```typescript
import { useQuery } from '@tanstack/react-query';

const { data, isLoading, error } = useQuery({
  queryKey: ['tenants'],
  queryFn: async () => {
    const response = await api.tenants.list();
    return response.data;
  },
});
```

### Auth Context

```typescript
import { useAuth } from '~/contexts/AuthContext';

const { user, login, logout, isAuthenticated } = useAuth();
```

---

## Styling

### Tailwind CSS

```tsx
<div className="flex items-center gap-4 p-6">
  <Button variant="outline" size="lg">
    Click me
  </Button>
</div>
```

### Dark Mode Ready

All components support dark mode via Tailwind's `dark:` prefix.

---

## Building for Production

```bash
pnpm build
```

Output: `build/` directory

### Start Production Server

```bash
pnpm start
```

---

## Type Checking

```bash
pnpm typecheck
```

---

## Common Tasks

### Add a New Page

1. Create file in `app/routes/`:
```tsx
// app/routes/my-page.tsx
export default function MyPage() {
  return <div>My Page</div>;
}
```

2. Add route in `app/routes.ts`:
```typescript
route("my-page", "routes/my-page.tsx"),
```

### Add a New API Endpoint

1. Add to `app/api/index.ts`:
```typescript
export const myApi = {
  getData: (): Promise<AxiosResponse<MyData>> =>
    apiClient.get<MyData>('/my-endpoint'),
};
```

2. Add to exports:
```typescript
export const api = {
  // ...
  my: myApi,
};
```

### Add a New Component

1. Create in `app/components/`:
```tsx
// app/components/MyComponent.tsx
export function MyComponent() {
  return <div>Component</div>;
}
```

2. Import and use:
```tsx
import { MyComponent } from '~/components/MyComponent';
```

---

## Troubleshooting

### Port Already in Use

```bash
# Kill process on port 5173
lsof -ti:5173 | xargs kill -9
```

### Module Not Found

```bash
pnpm install
```

### TypeScript Errors

```bash
pnpm typecheck
```

### API Connection Issues

Check:
1. Backend is running on `http://localhost:8080`
2. `.env` file has correct `VITE_API_URL`
3. CORS is enabled on backend

---

## Next Steps

1. **Test Login Flow**
   - Start backend services
   - Visit http://localhost:5173/login
   - Click OAuth button
   - Verify redirect to dashboard

2. **Build CRUD Pages**
   - Tenant create/edit forms
   - Client management
   - User management
   - RBAC configuration

3. **Add Features**
   - Toast notifications
   - Form validation
   - Data tables
   - File uploads

---

## Useful Links

- [React Router Docs](https://reactrouter.com/)
- [TanStack Query Docs](https://tanstack.com/query/latest)
- [Tailwind CSS Docs](https://tailwindcss.com/)
- [Radix UI Docs](https://www.radix-ui.com/)
- [Lucide Icons](https://lucide.dev/)

---

## Support

For issues or questions:
1. Check `FRONTEND-STRUCTURE.md` for architecture details
2. Check `FRONTEND-VERIFICATION.md` for complete checklist
3. Review example code in `example-frontend/`

---

**Happy Coding! 🎉**
