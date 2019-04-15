import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { RouterModule } from '@angular/router';
import { MatToolbarModule, MatSnackBarModule } from '@angular/material';
import { AppComponent } from './app.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import {
  EcadAngularAuthModule,
  LoggedinGuard,
  PermissionsGuard,
  EcadAngularAuthAdminComponentsModule,
  EcadAngularAuthAdminModule,
  EcadAngularAuthComponentsModule
} from '@ecadlabs/angular-auth';
import { LoginComponent } from './login/login.component';
import { ResetPasswordComponent } from './reset-password/reset-password.component';
import { ResetPasswordEmailComponent } from './reset-password-email/reset-password-email.component';
import { ProtectedComponent } from './protected/protected.component';
import { UserDetailPageComponent } from './user-detail-page/user-detail-page.component';
import { UserLogsComponent } from './user-logs/user-logs.component';
import { RequestEmailChangeComponent } from './request-email-change/request-email-change.component';
import { HTTP_INTERCEPTORS } from '@angular/common/http';
import { ErrorHttpInterceptor } from './error-http-interceptor';

export function tokenGetter() {
  return localStorage.getItem('token');
}
export function tokenSetter(value: string) {
  localStorage.setItem('token', value);
}

@NgModule({
  declarations: [
    AppComponent,
    LoginComponent,
    ResetPasswordComponent,
    ResetPasswordEmailComponent,
    ProtectedComponent,
    UserDetailPageComponent,
    UserLogsComponent,
    RequestEmailChangeComponent
  ],
  imports: [
    MatSnackBarModule,
    EcadAngularAuthModule.forRoot({
      loginUrl: '/api/v1/login',
      whiteListUrl: '/api/v1/checkip',
      tokenGetter,
      tokenSetter,
      passwordResetUrl: '/api/v1/password_reset',
      sendResetEmailUrl: '/api/v1/request_password_reset',
      loginPageUrl: '',
      roleGuardRedirectUrl: '',
      autoRefreshInterval: 5000,
      tokenPropertyPrefix: 'com.ecadlabs.auth',
      rolesPermissionsMapping: {
        admin: ['show.is-admin']
      },
      defaultRole: 'owner',
      emailChangeValidationUrl: '/api/v1/email_update',
      emailUpdateUrl: '/api/v1/request_email_update'
    }),
    EcadAngularAuthComponentsModule,
    EcadAngularAuthAdminComponentsModule,
    EcadAngularAuthAdminModule.forRoot({
      roles: [
        { value: 'regular', displayValue: 'Regular' },
        { value: 'admin', displayValue: 'Admin' }
      ],
      apiEndpoint: '/api/v1/users',
      emailUpdateUrl: '/api/v1/request_email_update'
    }),
    BrowserModule,
    BrowserAnimationsModule,
    RouterModule.forRoot([
      { path: '', pathMatch: 'full', component: LoginComponent },
      { path: 'reset-password', component: ResetPasswordComponent },
      { path: 'reset-password-email', component: ResetPasswordEmailComponent },
      { path: 'request-email-change', component: RequestEmailChangeComponent },
      {
        path: 'protected',
        component: ProtectedComponent,
        data: { permissions: ['show.is-admin'] },
        canActivate: [LoggedinGuard, PermissionsGuard]
      },
      {
        path: 'user/:id',
        component: UserDetailPageComponent,
        data: { permissions: ['show.is-admin'] },
        canActivate: [LoggedinGuard, PermissionsGuard]
      },
      {
        path: 'logs',
        component: UserLogsComponent,
        data: { permissions: ['show.is-admin'] },
        canActivate: [LoggedinGuard, PermissionsGuard]
      }
    ]),
    MatToolbarModule
  ],
  providers: [
    {
      provide: HTTP_INTERCEPTORS,
      useClass: ErrorHttpInterceptor,
      multi: true
    }
  ],
  bootstrap: [AppComponent]
})
export class AppModule {}
