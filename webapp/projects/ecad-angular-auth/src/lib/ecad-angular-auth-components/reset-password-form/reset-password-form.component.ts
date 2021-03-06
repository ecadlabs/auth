import { Component, Inject, Input, OnInit } from '@angular/core';
import {
  AbstractControl,
  FormBuilder,
  FormGroup,
  Validators
} from '@angular/forms';
import { Router } from '@angular/router';
import { IPasswordReset } from '../../ecad-angular-auth/interfaces/password-reset.i';
import { PASSWORD_RESET } from '../../ecad-angular-auth/tokens';
import { getParameterByName } from '../../utils';

export interface ResetPasswordFormConfig {
  successUrlRedirect: string;
  tokenAccessor?: () => string;
}

@Component({
  selector: 'auth-reset-password-form',
  templateUrl: './reset-password-form.component.html',
  styleUrls: ['./reset-password-form.component.scss']
})
export class ResetPasswordFormComponent implements OnInit {
  public token_expired = false;
  public error_occured = false;

  @Input()
  public config: ResetPasswordFormConfig;

  public resetPasswordForm: FormGroup;

  constructor(
    @Inject(PASSWORD_RESET)
    private resetPassword: IPasswordReset,
    private router: Router,
    fb: FormBuilder
  ) {
    this.resetPasswordForm = fb.group(
      {
        confirmPassword: ['', [Validators.required]],
        password: ['', [Validators.required]]
      },
      {
        validator: this.passwordConfirming
      }
    );
  }

  async onSubmit() {
    const token = this.config.tokenAccessor
      ? this.config.tokenAccessor
      : () => getParameterByName('token');
    try {
      await this.resetPassword
        .resetPassword(token(), this.resetPasswordForm.value.password)
        .toPromise();
      await this.router.navigateByUrl(this.config.successUrlRedirect);
    } catch (err) {
      if (err && String(err.status) === '400') {
        this.token_expired = true;
      } else {
        this.error_occured = true;
      }
    }
  }

  passwordConfirming(c: AbstractControl): { invalid: boolean } {
    if (c.get('password').value !== c.get('confirmPassword').value) {
      return { invalid: true };
    }
  }

  ngOnInit() {}
}
