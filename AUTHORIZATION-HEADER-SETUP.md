# Authorization Header Setup

## Overview
Added Authorization header support to the frontend API client. Now all requests will include both:
1. **Cookie authentication** (HTTP-only cookie) - More secure
2. **Authorization header** (Bearer token) - Standard REST API practice

## Changes Made

### 1. API Client - Request Interceptor

**File**: `/apps/frontend/app/api/client.ts`

Added a request interceptor to automatically attach the Authorization header:

```typescript
// Request interceptor to add Authorization header
apiClient.interceptors.request.use(
  (config) => {
    // Try to get token from localStorage
    const token = localStorage.getItem('auth_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);
```

**What it does**:
- Intercepts every outgoing request
- Reads token from `localStorage`
- Adds `Authorization: Bearer <token>` header if token exists

---

### 2. Auth Context - Token Storage

**File**: `/apps/frontend/app/contexts/AuthContext.tsx`

Updated login and logout functions to manage localStorage:

```typescript
// Login function
const login = async (token: string) => {
  try {
    // Store token in localStorage for Authorization header
    if (typeof window !== 'undefined') {
      localStorage.setItem('auth_token', token);
    }
    
    // ... rest of login logic
  } catch (error) {
    // Clear token on error
    if (typeof window !== 'undefined') {
      localStorage.removeItem('auth_token');
    }
    throw error;
  }
};

// Logout function
const logout = () => {
  api.auth.logout();
  queryClient.setQueryData(['auth', 'user'], null);
  // Clear token from localStorage
  if (typeof window !== 'undefined') {
    localStorage.removeItem('auth_token');
  }
};
```

---

### 3. OAuth Callback - Token Storage

**File**: `/apps/frontend/app/routes/auth.callback.tsx`

Store token during OAuth callback:

```typescript
const handleCallback = async (token: string) => {
  try {
    // Store token in localStorage for Authorization header
    localStorage.setItem('auth_token', token);
    
    // Send the token to backend to set it as HTTP-only cookie
    const response = await api.auth.setTokenCookie(token);
    
    // ... rest of callback logic
  } catch (error) {
    // Clear token on error
    localStorage.removeItem('auth_token');
    navigate('/login');
  }
};
```

---

### 4. NGINX - Authorization Header Forwarding

**File**: `/nginx.conf`

Already added by you - forwards Authorization header to backend services:

```nginx
proxy_set_header Authorization $http_authorization;
```

This is present in all three service routes:
- `/api/auth/` → Auth Service
- `/api/clients` → Client Service
- `/api/tenant/` → Tenant Service

---

## How It Works

### Authentication Flow:

1. **User logs in via OAuth**:
   ```
   Frontend → OAuth Provider → Callback with token
   ```

2. **Token storage** (OAuth callback):
   ```
   localStorage.setItem('auth_token', token)
   Backend sets HTTP-only cookie
   ```

3. **Subsequent API requests**:
   ```
   Request Interceptor → Reads token from localStorage
                      → Adds Authorization: Bearer <token>
                      → Also sends auth_token cookie
   ```

4. **NGINX Gateway**:
   ```
   Receives request with:
   - Authorization: Bearer <token> header
   - Cookie: auth_token=<token>
   
   Forwards both to backend service
   ```

5. **Backend validates**:
   ```
   Auth Middleware checks:
   1. Authorization header first
   2. Falls back to cookie if no header
   3. Validates JWT token
   4. Allows request if valid
   ```

---

## Request Example

### Before (Cookie only):
```http
GET /api/tenant/dashboard HTTP/1.1
Host: localhost:8080
Cookie: auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

### After (Cookie + Authorization header):
```http
GET /api/tenant/dashboard HTTP/1.1
Host: localhost:8080
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Cookie: auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Content-Type: application/json
```

---

## Backend Support

Your backend auth middleware already supports both methods:

```go
// From packages/go/auth/middleware.go

// Try to get token from Authorization header first
authHeader := c.Request().Header.Get("Authorization")
if authHeader != "" {
    parts := strings.Split(authHeader, " ")
    if len(parts) == 2 && parts[0] == "Bearer" {
        tokenString = parts[1]
    }
}

// If no token in header, try to get from cookie
if tokenString == "" {
    cookie, err := c.Cookie("auth_token")
    if err == nil && cookie.Value != "" {
        tokenString = cookie.Value
    }
}
```

**Priority**:
1. ✅ Authorization header (checked first)
2. ✅ Cookie (fallback)

---

## Security Considerations

### Dual Authentication Approach:

**Authorization Header (localStorage)**:
- ✅ Standard REST API practice
- ✅ Works with mobile apps and external clients
- ⚠️ Vulnerable to XSS attacks
- ✅ Easy to inspect and debug

**HTTP-only Cookie**:
- ✅ More secure (not accessible via JavaScript)
- ✅ Automatic CSRF protection
- ✅ Browser handles it automatically
- ⚠️ Doesn't work with mobile apps

**Best Practice**: Use both for maximum compatibility and security:
- Web app uses cookies (more secure)
- Authorization header as backup/standard
- Mobile apps can use Authorization header only

---

## Testing

### 1. Check Token Storage:
```javascript
// In browser console after login
localStorage.getItem('auth_token')
// Should return: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 2. Check Request Headers:
1. Open DevTools → Network tab
2. Make any API request (e.g., dashboard)
3. Click on the request
4. Check **Request Headers**:
   ```
   Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   Cookie: auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
   ```

### 3. Test Dashboard:
```bash
# Should now work with Authorization header
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/tenant/dashboard
```

---

## Files Modified

1. ✅ `/apps/frontend/app/api/client.ts` - Request interceptor
2. ✅ `/apps/frontend/app/contexts/AuthContext.tsx` - Token storage on login/logout
3. ✅ `/apps/frontend/app/routes/auth.callback.tsx` - Token storage on OAuth callback
4. ✅ `/nginx.conf` - Authorization header forwarding (already done by you)

---

## Summary

Your frontend now sends authentication in **two ways**:

1. **Authorization: Bearer <token>** header (via request interceptor)
2. **Cookie: auth_token=<token>** (via withCredentials)

Both are forwarded by NGINX to backend services, and your backend auth middleware checks both (header first, then cookie).

This provides:
- ✅ Standard REST API authentication
- ✅ Better debugging (can see token in headers)
- ✅ Compatibility with tools like Postman
- ✅ Fallback to secure HTTP-only cookies
- ✅ Works with your existing backend code

**No backend changes needed** - your auth middleware already supports both methods! 🎉
