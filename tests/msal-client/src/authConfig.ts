import { Configuration, PublicClientApplication, LogLevel } from "@azure/msal-browser";

export const msalConfig: Configuration = {
    auth: {
        clientId: "e697b97c-9b4b-487f-9f7a-248386f78864",
        authority: "https://login.microsoftonline.com/10000000-0000-4000-a000-000000000000/v2.0",
        redirectUri: "https://msal-client.web.internal/",
        postLogoutRedirectUri: "https://msal-client.web.internal/",
    },
    cache: {
        cacheLocation: "sessionStorage",
    },
    system: {
        loggerOptions: {
            loggerCallback: (level, message, containsPii) => {
                if (containsPii) {
                    return;
                }
                switch (level) {
                    case LogLevel.Error:
                        console.error(message);
                        return;
                    case LogLevel.Info:
                        console.info(message);
                        return;
                    case LogLevel.Verbose:
                        console.debug(message);
                        return;
                    case LogLevel.Warning:
                        console.warn(message);
                        return;
                }
            }
        }
    }
};

export const loginRequest = {
    scopes: ["User.Read", "api://aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa/read"],
    state: "custom-state-value-123"
};

export const msalInstance = new PublicClientApplication(msalConfig);
