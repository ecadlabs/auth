import { Component, OnInit, Inject } from '@angular/core';
import { FormGroup, FormBuilder, Validators } from '@angular/forms';
import { PASSWORD_RESET, AUTH_CONFIG } from '../../ecad-angular-auth/tokens';
import { IPasswordReset } from '../../ecad-angular-auth/interfaces/password-reset.i';
import { Router } from '@angular/router';
import { AuthConfig } from '../../ecad-angular-auth/interfaces/auth-config.i';

@Component({
  selector: 'auth-reset-password-email-form',
  templateUrl: './reset-password-email-form.component.html',
  styleUrls: ['./reset-password-email-form.component.scss']
})
export class ResetPasswordEmailFormComponent {

  public resetPasswordEmailForm: FormGroup;

  public success = false;

  constructor(
    @Inject(PASSWORD_RESET)
    private resetPassword: IPasswordReset,
    private router: Router,
    fb: FormBuilder,
    @Inject(AUTH_CONFIG)
    private authConfig: AuthConfig
  ) {
    this.resetPasswordEmailForm = fb.group({
      'email': ['', [Validators.required, Validators.pattern(
        this.authConfig.emailValidationRegex || /^.+@.+\..{2,3}$/
      )]],
    });
  }

  async onSubmit() {
    await this.resetPassword.sendResetEmail(this.resetPasswordEmailForm.value.email).toPromise();
    this.success = true;
  }

}
