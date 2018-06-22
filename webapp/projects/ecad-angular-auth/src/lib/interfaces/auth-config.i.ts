export interface AuthConfig {
    loginUrl: string;
    whiteListUrl: string;
    tokenName: string;
    passwordResetUrl: string;
    sendResetEmailUrl: string;
    loginPageUrl: string;
    autoRefreshInterval?: number;
}
