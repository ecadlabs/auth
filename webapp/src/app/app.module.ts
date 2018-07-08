import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { RouterModule } from '@angular/router';
import { MatToolbarModule } from '@angular/material';
import { AppComponent } from './app.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import {
  EcadAngularAuthModule,
  LoggedinGuard,
  IpWhiteListedGuard,
  PermissionsGuard,
  EcadAngularAuthAdminComponentsModule,
  EcadAngularAuthAdminModule,
  EcadAngularAuthComponentsModule
} from 'ecad-angular-auth';
import { LoginComponent } from './login/login.component';
import { ResetPasswordComponent } from './reset-password/reset-password.component';
import { ResetPasswordEmailComponent } from './reset-password-email/reset-password-email.component';
import { ProtectedComponent } from './protected/protected.component';
import { UserDetailPageComponent } from './user-detail-page/user-detail-page.component';
import { UserLogsComponent } from './user-logs/user-logs.component';

export function tokenGetter() { return localStorage.getItem('token'); }
export function tokenSetter(value: string) { localStorage.setItem('token', value); }

@NgModule({
  declarations: [
    AppComponent,
    LoginComponent,
    ResetPasswordComponent,
    ResetPasswordEmailComponent,
    ProtectedComponent,
    UserDetailPageComponent,
    UserLogsComponent,
  ],
  imports: [
    EcadAngularAuthModule.forRoot({
      loginUrl: '/api/v1/login',
      whiteListUrl: '/api/v1/checkip',
      tokenGetter,
      tokenSetter,
      passwordResetUrl: '/api/v1/password_reset',
      sendResetEmailUrl: '/api/v1/request_password_reset',
      loginPageUrl: '',
      autoRefreshInterval: 5000,
      tokenPropertyPrefix: 'com.ecadlabs.auth',
      rolesPermissionsMapping: {
        'com.ecadlabs.auth.admin': ['show.is-admin']
      },
    }),
    EcadAngularAuthComponentsModule,
    EcadAngularAuthAdminComponentsModule,
    EcadAngularAuthAdminModule.forRoot({
      roles: [
        { value: 'com.ecadlabs.auth.regular', displayValue: 'Regular' },
        { value: 'com.ecadlabs.auth.admin', displayValue: 'Admin' }
      ],
      apiEndpoint: '/api/v1/users'
    }),
    BrowserModule,
    BrowserAnimationsModule,
    RouterModule.forRoot([
      { path: '', pathMatch: 'full', component: LoginComponent },
      { path: 'reset-password', component: ResetPasswordComponent },
      { path: 'reset-password-email', component: ResetPasswordEmailComponent },
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
    MatToolbarModule,
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
