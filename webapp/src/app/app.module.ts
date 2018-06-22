import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';

import { RouterModule } from '@angular/router';
import { MatInputModule, MatCardModule, MatButtonModule, MatToolbarModule } from '@angular/material';
import { AppComponent } from './app.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ReactiveFormsModule, FormsModule } from '@angular/forms';
import { EcadAngularAuthModule, LoggedinGuard } from 'ecad-angular-auth';
import {
  EcadAngularAuthComponentsModule
} from 'projects/ecad-angular-auth/src/lib/ecad-angular-auth-components/ecad-angular-auth-components.module';
import { LoginComponent } from './login/login.component';
import { ResetPasswordComponent } from './reset-password/reset-password.component';
import { ResetPasswordEmailComponent } from './reset-password-email/reset-password-email.component';
import { ProtectedComponent } from './protected/protected.component';
import { EcadAngularAuthAdminComponentsModule } from 'projects/ecad-angular-auth/src/lib/ecad-angular-auth-admin-components/ecad-angular-auth-admin-components.module';
import { EcadAngularAuthAdminModule } from 'projects/ecad-angular-auth/src/lib/ecad-angular-auth-admin/ecad-angular-auth-admin.module';



@NgModule({
  declarations: [
    AppComponent,
    LoginComponent,
    ResetPasswordComponent,
    ResetPasswordEmailComponent,
    ProtectedComponent,
  ],
  imports: [
    EcadAngularAuthModule.forRoot({
      loginUrl: '/api/v1/login',
      whiteListUrl: 'test',
      tokenName: 'test',
      passwordResetUrl: 'test',
      sendResetEmailUrl: 'test',
      loginPageUrl: '',
    }),
    EcadAngularAuthComponentsModule,
    EcadAngularAuthAdminComponentsModule,
    EcadAngularAuthAdminModule,
    BrowserModule,
    BrowserAnimationsModule,
    RouterModule.forRoot([
      {path: '', pathMatch: 'full', component: LoginComponent},
      {path: 'reset-password', component: ResetPasswordComponent},
      {path: 'reset-password-email', component: ResetPasswordEmailComponent},
      {path: 'protected', component: ProtectedComponent, canActivate: [LoggedinGuard]}
    ]),
    MatToolbarModule
  ],
  providers: [],
  bootstrap: [AppComponent]
})
export class AppModule { }
