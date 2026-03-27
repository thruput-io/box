# box

## Configuration Overrides

### Service Bus Emulator Config
You can provide your own `Config.json` for the Service Bus Emulator by setting the `SERVICEBUS_CONFIG` environment variable to the absolute path of your file before running `make up`:

```bash
export SERVICEBUS_CONFIG=/path/to/your/backend/project/Config.json
make up
```

By default, it uses `./tests/DefaultServiceBusConfig.json`.

### Identity Service Config
You can provide your own `Config.yaml` for the Mock Entra ID service by setting the `IDENTITY_CONFIG` environment variable:

```bash
export IDENTITY_CONFIG=/path/to/your/backend/project/IdentityConfig.yaml
make up
```

By default, it uses `./entra/DefaultConfig.yaml`.

### Application Image Override
You can provide your own application image (e.g., from a local project's build) by setting the `APP_IMAGE` environment variable:

```bash
# Build your local project first
cd /Users/johan/workspace/MaintenanceMonitorBackend
docker build -t my-app-backend .

# In box root
export APP_IMAGE=my-app-backend
make up-app
```

By default, it uses `mocking-monitor:latest`.

### Custom Environment Variables
You can pass any environment variables to your app by creating a `.env` file in the root of the project:

```bash
# Example .env file
MY_VAR=hello
ANOTHER_VAR=world
```

These variables will be available inside the app container when running `make up-app`.

### Identity Service Overview
You can view the current identity configuration (tenants, app registrations, clients, and users) by visiting:
`https://entra.web.internal`

This page is served by the mock Entra service and provides a clear overview of the credentials and scopes available for testing.
