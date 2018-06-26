import { NgModule, ModuleWithProviders } from '@angular/core';
import { HttpClientModule } from '@angular/common/http';
import { JwtModule } from '@auth0/angular-jwt';
import { authConfig, LoginService, PasswordReset } from './tokens';
import { StandardLoginService } from './login/standard-login.service';
import { AuthConfig } from './interfaces';
import { PasswordResetService } from './password-reset/password-reset.service';
import { IpWhiteListedGuard } from './guards/ip-whitelisted.guard';
import { LoggedinGuard } from './guards/loggedin.guard';

export const blacklistedRoutes = [];
export const whiteListedDomain = [new RegExp('^null$'), new RegExp(`.*${location.hostname}.*`)];
export let tokenName = '';
export function tokenGetter() {
  return window.localStorage.getItem(tokenName);
}

@NgModule({
  imports: [
    HttpClientModule,
    JwtModule.forRoot({
      config: {
        blacklistedRoutes,
        whitelistedDomains: whiteListedDomain,
        tokenGetter: tokenGetter,
      }
    }),
  ],
  declarations: [],
  exports: []
})
export class EcadAngularAuthModule {
  public static forRoot(config: AuthConfig): ModuleWithProviders {
    tokenName = config.tokenName;
    blacklistedRoutes.push(config.loginUrl);
    return {
      ngModule: EcadAngularAuthModule,
      providers: [
        { provide: authConfig, useValue: config },
        { provide: LoginService, useClass: StandardLoginService },
        { provide: PasswordReset, useClass: PasswordResetService },
        IpWhiteListedGuard,
        LoggedinGuard
      ]
    };
  }
}
