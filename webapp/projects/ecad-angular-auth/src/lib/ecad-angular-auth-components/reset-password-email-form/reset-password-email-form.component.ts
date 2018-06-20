import { Component, OnInit, Inject } from '@angular/core';
import { FormGroup, FormBuilder, Validators } from '@angular/forms';
import { PasswordReset } from '../../tokens';
import { IPasswordReset } from '../../interfaces';
import { Router } from '@angular/router';

@Component({
  selector: 'auth-reset-password-email-form',
  templateUrl: './reset-password-email-form.component.html',
  styleUrls: ['./reset-password-email-form.component.scss']
})
export class ResetPasswordEmailFormComponent {

  public resetPasswordEmailForm: FormGroup;

  public success = false;

  constructor(
    @Inject(PasswordReset)
    private resetPassword: IPasswordReset,
    private router: Router,
    fb: FormBuilder,
  ) {
    this.resetPasswordEmailForm = fb.group({
      'email': ['', [Validators.required]],
    });
  }

  async onSubmit() {
    await this.resetPassword.sendResetEmail(this.resetPasswordEmailForm.value.email).toPromise();
    this.success = true;
  }

}