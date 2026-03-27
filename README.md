# box

A local development and testing environment that provides a suite of simulated cloud services and infrastructure. `box` is designed to streamline development for Azure-based applications by offering high-fidelity mocks for common services, integrated DNS, and automatic TLS certificate management.

## Valid Use Cases

- **Local Azure Development:** Develop and test applications that depend on Azure services like Entra ID (formerly Azure AD), Azure Service Bus, and PostgreSQL without needing an active Azure subscription.
- **Integration Testing:** Run automated integration tests against realistic service mocks. It includes support for Pester, Playwright, and MSAL testing.
- **Microservices Orchestration:** Develop multi-service architectures using a shared local network (`infra-shared-net`) and internal DNS (`*.web.internal`).
- **OAuth/OIDC Flow Testing:** Test authentication and authorization logic using the built-in Entra ID emulator (`entra`) that supports authorization code and client credentials flows with customizable roles.
- **Infrastructure Simulation:** Simulate complex network topologies with Traefik as an entry point, handling TLS termination and routing for all services.

## Key Components

- **Traefik (`entry`):** Acts as the primary ingress, handling routing and TLS termination.
- **CoreDNS (`dns`):** Provides internal DNS for the `*.internal` and `*.microsoftonline.com` domains.
- **Entra ID Emulator (`entra`):** A mock for Azure Entra ID, pre-configured with users, clients, and roles (see `entra/config.md`).
- **Service Bus Emulator:** An AMQP-compatible emulator for Azure Service Bus.
- **PostgreSQL:** A standard PostgreSQL instance for relational data storage.
- **Azure SQL Edge (`sqledge`):** A lightweight SQL Server instance.
- **Browser Service:** A containerized browser (accessible at `https://browser.web.internal`) to test UI and authentication flows in a clean environment.

## Prerequisites

- **Docker Desktop** with Compose support.
- **macOS** is required for the full localhost integration (DNS and certificate trust).
- **Homebrew** (optional, used for some localhost setup scripts).

## Getting Started

### 1. Initial Setup

First, install the `box` command-line helper:

```bash
make install-box
source ~/.zshrc  # or ~/.bashrc
```

This installs a `box` alias that you can use instead of `make -C /path/to/box`.

### 2. Configure Localhost (macOS only)

To enable custom DNS (`*.internal`) and trust the generated TLS certificates on your host machine:

```bash
box setup-localhost
```

This will:
- Add a custom resolver to `/etc/resolver/internal`.
- Create a dedicated keychain and import the Root CA.
- Configure Firefox and Chrome to trust the certificates.

### 3. Start the Environment

```bash
box up
```

This generates certificates, builds images, and starts the core infrastructure and services.

### 4. Access the Portal

Once started, you can access the management portal at:
[https://portal.web.internal](https://portal.web.internal)

## Common Commands

- `box up`: Start the entire environment.
- `box down`: Stop the environment.
- `box clean`: Full cleanup of containers, networks, and volumes.
- `box logs`: View logs for all services.
- `box browser`: Start the environment and open the containerized browser.
- `box rotate-cert`: Re-generate all certificates and refresh host trust.

## Testing

`box` includes several test suites to verify both the environment and your application:

- `box self-test`: Runs all internal integration tests.
- `box test-pester`: Runs PowerShell Pester tests.
- `box test-msal`: Validates authentication flows.
- `box test-browser-playwright`: Runs UI tests using Playwright.

## Documentation

- **Entra ID Configuration:** See [entra/config.md](entra/config.md) for details on users, clients, and roles available in the emulator.
- **Localhost Setup:** See the scripts in `localhost/` for details on how host integration works.
