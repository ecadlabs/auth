import { NgModule, ModuleWithProviders } from '@angular/core';
import { HttpClientModule, HTTP_INTERCEPTORS } from '@angular/common/http';
import { JwtModule } from '@auth0/angular-jwt';
import { AUTH_CONFIG, LOGIN_SERVICE, PASSWORD_RESET } from './tokens';
import { StandardLoginService } from './login/standard-login.service';
import { AuthConfig } from './interfaces/auth-config.i';
import { PasswordResetService } from './password-reset/password-reset.service';
import { IpWhiteListedGuard } from './guards/ip-whitelisted.guard';
import { LoggedinGuard } from './guards/loggedin.guard';
import { PermissionsGuard } from './guards/permissions.guard';
import { RoleGuard } from './guards/role.guard';
import { JwtHelperService } from './jwt-helper.service';
import { JwtInterceptor } from './jwt.interceptor';

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
        {
          provide: HTTP_INTERCEPTORS,
          useClass: JwtInterceptor,
          multi: true
        },
        { provide: AUTH_CONFIG, useValue: config },
        { provide: LOGIN_SERVICE, useClass: StandardLoginService },
        { provide: PASSWORD_RESET, useClass: PasswordResetService },
        JwtHelperService,
        IpWhiteListedGuard,
        LoggedinGuard,
        PermissionsGuard,
        RoleGuard
      ]
    };
  }
}
