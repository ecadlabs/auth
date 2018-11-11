export interface AuthConfig {
    loginUrl: string;
    whiteListUrl?: string;
    tokenGetter: () => string;
    tokenSetter: (value: string) => void;
    passwordResetUrl: string;
    sendResetEmailUrl: string;
    loginPageUrl: string;
    autoRefreshInterval?: number;
    tokenPropertyPrefix?: string;
    rolesPermissionsMapping: {
        [key: string]: string[]
    };
    emailValidationRegex?: RegExp;
    emailUpdateUrl: string;
    emailChangeValidationUrl: string;
}
