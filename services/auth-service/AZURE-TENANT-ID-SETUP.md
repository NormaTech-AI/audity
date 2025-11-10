# How to Find Your Azure AD Tenant ID

## Error Fixed
The error `AADSTS50194: Application is not configured as a multi-tenant application` has been resolved by updating the code to use a tenant-specific endpoint instead of the `/common` endpoint.

## Steps to Get Your Tenant ID

### Option 1: Azure Portal (Recommended)
1. Go to [Azure Portal](https://portal.azure.com/)
2. Navigate to **Azure Active Directory** (or **Microsoft Entra ID**)
3. In the Overview page, you'll see **Tenant ID** (also called Directory ID)
4. Copy the GUID (e.g., `12345678-1234-1234-1234-123456789abc`)

### Option 2: App Registration Page
1. Go to [Azure Portal](https://portal.azure.com/)
2. Navigate to **Azure Active Directory** > **App registrations**
3. Click on your app: **HRMS- DEV** (Application ID: `fce22d45-894f-48a2-9529-628bb971e274`)
4. In the Overview page, you'll see **Directory (tenant) ID**
5. Copy the GUID

### Option 3: From the Error URL
If you look at the authentication URL when the error occurs, you might see it in the URL structure.

## Update Configuration

Once you have your Tenant ID, update the `config.yaml` file:

```yaml
auth:
  microsoft_tenant_id: "YOUR_ACTUAL_TENANT_ID_HERE"
```

Replace `YOUR_TENANT_ID_HERE` with your actual tenant ID (it should be a GUID format like `12345678-1234-1234-1234-123456789abc`).

## Restart the Service

After updating the configuration:

```bash
# If running with Docker
docker-compose restart auth-service

# If running locally
# Stop the current process and restart
go run main.go
```

## Alternative: Make App Multi-Tenant (Not Recommended for Your Use Case)

If you want to use the `/common` endpoint instead, you can configure your Azure AD app as multi-tenant:

1. Go to Azure Portal > App registrations > Your app
2. Click on **Authentication**
3. Under **Supported account types**, select:
   - **Accounts in any organizational directory (Any Azure AD directory - Multitenant)**
4. Save the changes

However, using a tenant-specific endpoint is the recommended approach for single-tenant applications.
