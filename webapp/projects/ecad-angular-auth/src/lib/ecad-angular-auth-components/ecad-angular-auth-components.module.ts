import { NgModule, ModuleWithProviders } from '@angular/core';
import { CommonModule } from '@angular/common';
import { LoginComponent } from './login/login.component';
import { ReactiveFormsModule } from '@angular/forms';
import { MatInputModule, MatCardModule, MatButtonModule } from '@angular/material';
import { ResetPasswordFormComponent } from './reset-password-form/reset-password-form.component';
import { ResetPasswordEmailFormComponent } from './reset-password-email-form/reset-password-email-form.component';
import { AlertComponent } from './alert/alert.component';
import { EcadPermissionsDirective } from './ecad-permissions.directive';
import { RequestEmailChangeComponent } from './request-email-change/request-email-change.component';

@NgModule({
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatInputModule,
    MatCardModule,
    MatButtonModule
  ],
  declarations: [
    LoginComponent,
    ResetPasswordFormComponent,
    ResetPasswordEmailFormComponent,
    EcadPermissionsDirective,
    AlertComponent,
    RequestEmailChangeComponent
  ],
  exports: [
    RequestEmailChangeComponent,
    LoginComponent,
    ResetPasswordFormComponent,
    ResetPasswordEmailFormComponent,
    EcadPermissionsDirective
  ]
})
export class EcadAngularAuthComponentsModule {
}
