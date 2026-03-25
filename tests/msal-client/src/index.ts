import { msalInstance, loginRequest } from "./authConfig";
import { AuthenticationResult } from "@azure/msal-browser";

const loginButton = document.getElementById("login") as HTMLButtonElement;
const logoutButton = document.getElementById("logout") as HTMLButtonElement;
const silentTokenButton = document.getElementById("silent-token") as HTMLButtonElement;
const accountInfo = document.getElementById("account-info") as HTMLPreElement;
const tokenInfo = document.getElementById("token-info") as HTMLPreElement;
const stateInfo = document.getElementById("state-info") as HTMLPreElement;

async function initialize() {
    await msalInstance.initialize();

    msalInstance.handleRedirectPromise()
        .then((response: AuthenticationResult | null) => {
            if (response) {
                console.log("Login successful");
                handleResponse(response);
            } else {
                const currentAccounts = msalInstance.getAllAccounts();
                if (currentAccounts.length > 0) {
                    showAccount(currentAccounts[0]);
                }
            }
        })
        .catch(err => {
            console.error("Redirect Error:", err);
            accountInfo.innerText = "Error: " + err.message;
        });
}

function handleResponse(response: AuthenticationResult) {
    showAccount(response.account);
    tokenInfo.innerText = JSON.stringify(response, null, 2);
    
    if ((response as any).state) {
        stateInfo.innerText = "State received: " + (response as any).state;
    } else {
        stateInfo.innerText = "State not found in response object";
    }
}

function showAccount(account: any) {
    accountInfo.innerText = JSON.stringify(account, null, 2);
    loginButton.style.display = "none";
    logoutButton.style.display = "block";
    silentTokenButton.style.display = "block";
}

loginButton.addEventListener("click", () => {
    msalInstance.loginRedirect(loginRequest);
});

logoutButton.addEventListener("click", () => {
    msalInstance.logoutRedirect();
});

silentTokenButton.addEventListener("click", () => {
    const currentAccounts = msalInstance.getAllAccounts();
    if (currentAccounts.length === 0) {
        console.error("No account found for silent token acquisition");
        return;
    }

    const request = {
        ...loginRequest,
        account: currentAccounts[0]
    };

    msalInstance.acquireTokenSilent(request)
        .then(response => {
            console.log("Silent token acquisition successful");
            handleResponse(response);
        })
        .catch(error => {
            console.error("Silent token acquisition failed:", error);
            tokenInfo.innerText = "Error: " + error.message;
        });
});

initialize();
