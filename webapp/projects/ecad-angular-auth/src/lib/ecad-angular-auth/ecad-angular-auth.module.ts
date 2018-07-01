import { NgModule, ModuleWithProviders } from '@angular/core';
import { HttpClientModule } from '@angular/common/http';
import { JwtModule } from '@auth0/angular-jwt';
import { AUTH_CONFIG, LOGIN_SERVICE, PASSWORD_RESET } from './tokens';
import { StandardLoginService } from './login/standard-login.service';
import { AuthConfig } from './interfaces/auth-config.i';
import { PasswordResetService } from './password-reset/password-reset.service';
import { IpWhiteListedGuard } from './guards/ip-whitelisted.guard';
import { LoggedinGuard } from './guards/loggedin.guard';
import { PermissionsGuard } from './guards/permissions.guard';

export const blacklistedRoutes = [];
export const whiteListedDomain = [new RegExp('^null$'), new RegExp(`.*${location.hostname}.*`)];
export let tokenName = '';
export function tokenGetter() {
  return window.localStorage.getItem(tokenName);
}

@NgModule({
  imports: [
    HttpClientModule,
  ],
  declarations: [],
  exports: []
})
export class EcadAngularAuthModule {
  public static forRoot(config: AuthConfig): ModuleWithProviders {
    return {
      ngModule: EcadAngularAuthModule,
      providers: [
        ...JwtModule.forRoot({
          config: {
            blacklistedRoutes: [config.loginUrl],
            whitelistedDomains: [new RegExp('^null$'), new RegExp(`.*${location.hostname}.*`)],
            tokenGetter: config.tokenGetter,
          }
        }).providers,
        { provide: AUTH_CONFIG, useValue: config },
        { provide: LOGIN_SERVICE, useClass: StandardLoginService },
        { provide: PASSWORD_RESET, useClass: PasswordResetService },
        IpWhiteListedGuard,
        LoggedinGuard,
        PermissionsGuard
      ]
    };
  }
}
