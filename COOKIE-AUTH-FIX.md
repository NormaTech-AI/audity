# Cookie Authentication Fix

## Problem Identified

The NGINX gateway was **blocking cookies** from being passed between the frontend and backend services, causing 401 Unauthorized errors even when users were logged in.

## Root Cause

The `nginx.conf` file was missing the critical `proxy_pass_header` directives needed to forward cookies:
- `Set-Cookie` header (backend → frontend)
- `Cookie` header (frontend → backend)

Without these directives, NGINX was stripping cookies from requests and responses.

## Solution Applied

Added cookie forwarding directives to all three service routes in `nginx.conf`:

### Auth Service (`/api/auth/`)
```nginx
proxy_pass_header Set-Cookie;
proxy_pass_header Cookie;
```

### Client Service (`/api/clients`)
```nginx
proxy_pass_header Set-Cookie;
proxy_pass_header Cookie;
```

### Tenant Service (`/api/tenant/`)
```nginx
proxy_pass_header Set-Cookie;
proxy_pass_header Cookie;
```

## Changes Made

**File**: `/nginx.conf`

Added to each location block:
```nginx
# Pass cookies between client and backend
proxy_pass_header Set-Cookie;
proxy_pass_header Cookie;
```

## How It Works Now

### Login Flow:
1. User logs in via `/api/auth/login/google` or `/api/auth/login/microsoft`
2. Auth service sets `auth_token` cookie
3. **NGINX now forwards the `Set-Cookie` header** to the frontend
4. Browser stores the cookie

### Authenticated Request Flow:
1. Frontend makes API call (e.g., `/api/tenant/dashboard`)
2. Browser automatically includes `auth_token` cookie
3. **NGINX now forwards the `Cookie` header** to the backend
4. Backend validates the token from the cookie
5. Request succeeds ✅

## Testing

After restarting the gateway:
```bash
docker-compose restart gateway
```

### Test Authentication:
1. **Login**: Navigate to login page and authenticate
2. **Check Cookie**: Open DevTools → Application → Cookies → `http://localhost:8080`
   - You should see `auth_token` cookie
3. **Test Dashboard**: Navigate to `/dashboard`
   - Should load without 401 errors
4. **Check Network**: Open DevTools → Network
   - Dashboard API call should show `Cookie: auth_token=...` in request headers
   - Response should be 200 OK

### Manual Test:
```bash
# 1. Login and get cookie from browser DevTools
# 2. Test dashboard endpoint
curl -H "Cookie: auth_token=YOUR_TOKEN" \
  http://localhost:8080/api/tenant/dashboard

# Should return dashboard data, not 401
```

## Additional Frontend Fix

Also updated the dashboard component to only fetch data when authenticated:

**File**: `/apps/frontend/app/routes/dashboard.tsx`
```typescript
const { data: dashboardData, isLoading } = useQuery({
  queryKey: ['dashboard'],
  queryFn: async () => {
    const response = await api.dashboard.getTenantDashboardData();
    return response.data;
  },
  enabled: isAuthenticated, // Only fetch when user is authenticated
});
```

This prevents the dashboard from trying to fetch data before authentication is complete.

## Architecture Overview

```
Frontend (localhost:5173)
    ↓ withCredentials: true
NGINX Gateway (localhost:8080)
    ↓ proxy_pass_header Set-Cookie/Cookie
Backend Services
    ├── Auth Service (8082)
    ├── Client Service (8081)
    └── Tenant Service (8081)
```

## Security Notes

✅ **HTTP-only cookies**: More secure than localStorage
✅ **CORS configured**: `Access-Control-Allow-Credentials: true`
✅ **withCredentials**: Enabled in axios client
✅ **Cookie forwarding**: Now properly configured in NGINX

## Verification Checklist

- [x] NGINX configuration updated
- [x] Gateway restarted
- [x] Frontend query enabled only when authenticated
- [x] CORS headers include `Access-Control-Allow-Credentials: true`
- [x] Axios client has `withCredentials: true`
- [x] Cookie forwarding directives added to all routes

## Summary

The 401 Unauthorized errors were caused by NGINX not forwarding cookies. After adding `proxy_pass_header Set-Cookie` and `proxy_pass_header Cookie` to all service routes and restarting the gateway, cookies now flow properly between frontend and backend, enabling cookie-based authentication to work correctly.

The dashboard should now load successfully for authenticated users! 🎉
