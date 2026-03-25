# Stage 1: Build .NET client tool
FROM mcr.microsoft.com/dotnet/sdk:9.0 AS dotnet-builder
WORKDIR /app
COPY tests/dot-net-client/dot-net-client.csproj ./
RUN dotnet restore
COPY tests/dot-net-client/Program.cs ./
RUN dotnet publish -c Release -o /dotnet-client

# Stage 2: Final image
FROM mcr.microsoft.com/dotnet/sdk:9.0

# Install dependencies
USER root
RUN apt-get update && apt-get install -y \
    curl \
    postgresql-client \
    apt-transport-https \
    ca-certificates \
    gnupg \
    lsb-release \
    && curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian bookworm stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null \
    && apt-get update && apt-get install -y docker-ce-cli \
    && rm -rf /var/lib/apt/lists/*

# Install PowerShell and Pester
RUN curl -L https://github.com/PowerShell/PowerShell/releases/download/v7.4.1/powershell-7.4.1-linux-arm64.tar.gz -o powershell.tar.gz \
    && mkdir -p /opt/microsoft/powershell/7 \
    && tar zxf powershell.tar.gz -C /opt/microsoft/powershell/7 \
    && chmod +x /opt/microsoft/powershell/7/pwsh \
    && ln -sf /opt/microsoft/powershell/7/pwsh /usr/bin/pwsh \
    && rm powershell.tar.gz \
    && pwsh -Command "Set-PSRepository -Name 'PSGallery' -InstallationPolicy Trusted; \
    Install-Module -Name Pester -Force -AllowClobber -Scope AllUsers"

# Copy .NET client tool
COPY --from=dotnet-builder /dotnet-client /usr/local/bin/dot-net-client

# Add the local CA certificate
COPY certs/dev-root-ca.crt /usr/local/share/ca-certificates/mkcert.crt
RUN update-ca-certificates

WORKDIR /test
