import { Component, OnInit, Inject } from '@angular/core';
import { Router } from '@angular/router';
import { LoginService, PasswordReset } from '../../ecad-angular-auth/tokens';
import { ILoginService, IPasswordReset } from '../../ecad-angular-auth/interfaces';
import { FormBuilder, Validators, FormGroup, AbstractControl } from '@angular/forms';
import { getParameterByName } from '../../utils';

export interface ResetPasswordFormConfig {
  successUrlRedirect: string;
}

@Component({
  selector: 'auth-reset-password-form',
  templateUrl: './reset-password-form.component.html',
  styleUrls: ['./reset-password-form.component.scss']
})
export class ResetPasswordFormComponent implements OnInit {

  public config: ResetPasswordFormConfig;

  public resetPasswordForm: FormGroup;

  constructor(
    @Inject(PasswordReset)
    private resetPassword: IPasswordReset,
    private router: Router,
    fb: FormBuilder,
  ) {
    this.resetPasswordForm = fb.group({
      'confirmPassword': ['', [Validators.required]],
      'password': ['', [Validators.required]]
    }, {
      validator: this.passwordConfirming
    });
  }

  async onSubmit() {
    await this.resetPassword.resetPassword(getParameterByName('token'), this.resetPasswordForm.value.password).toPromise();
    await this.router.navigateByUrl(this.config.successUrlRedirect);
  }

  passwordConfirming(c: AbstractControl): { invalid: boolean } {
    if (c.get('password').value !== c.get('confirmPassword').value) {
      return { invalid: true };
    }
  }

  ngOnInit() {
  }

}
