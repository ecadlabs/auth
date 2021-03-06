export interface AuthConfig {
  loginUrl: string;
  whiteListUrl?: string;
  tokenGetter: () => string;
  tokenSetter: (value: string) => void;
  passwordResetUrl: string;
  sendResetEmailUrl: string;
  loginPageUrl: string;
  refreshUrl?: string;
  autoRefreshInterval?: number;
  tokenPropertyPrefix?: string;
  rolesPermissionsMapping: {
    [key: string]: string[];
  };
  emailValidationRegex?: RegExp;
  defaultRole: string;
  emailUpdateUrl: string;
  emailChangeValidationUrl: string;
  roleGuardRedirectUrl?: string;
}
