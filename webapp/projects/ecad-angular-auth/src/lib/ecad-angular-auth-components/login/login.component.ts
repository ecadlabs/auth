import { Component, OnInit, Inject, Input } from '@angular/core';
import { Router } from '@angular/router';
import { FormBuilder, Validators, FormGroup } from '@angular/forms';
import { ILoginService } from '../../ecad-angular-auth/interfaces';
import { LoginService } from '../../ecad-angular-auth/tokens';

export interface LoginFormConfig {
  successUrlRedirect: string;
}

@Component({
  selector: 'auth-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.scss']
})
export class LoginComponent {

  @Input()
  config: LoginFormConfig;

  public loginForm: FormGroup;

  constructor(
    @Inject(LoginService)
    private loginService: ILoginService,
    private router: Router,
    fb: FormBuilder,
  ) {
    this.loginForm = fb.group({
      'username': ['', [Validators.required]],
      'password': ['', [Validators.required]]
    });
  }

  async onSubmit() {
    await this.loginService.login(this.loginForm.value).toPromise();
    this.router.navigateByUrl(this.config.successUrlRedirect);
  }
}
