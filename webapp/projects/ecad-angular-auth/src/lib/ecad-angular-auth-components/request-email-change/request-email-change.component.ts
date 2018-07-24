import { Component, OnInit, Inject } from '@angular/core';
import { FormGroup, Validators, FormBuilder } from '@angular/forms';
import { LOGIN_SERVICE, AUTH_CONFIG } from '../../ecad-angular-auth/tokens';
import { ILoginService } from '../../ecad-angular-auth/interfaces/login-service.i';
import { Router } from '@angular/router';
import { AuthConfig } from '../../ecad-angular-auth/interfaces/auth-config.i';
import { switchMap, tap, first } from 'rxjs/operators';

@Component({
  selector: 'auth-request-email-change',
  templateUrl: './request-email-change.component.html',
  styleUrls: ['./request-email-change.component.scss']
})
export class RequestEmailChangeComponent {

  public requestEmailChangeForm: FormGroup;

  public success = false;

  constructor(
    @Inject(LOGIN_SERVICE)
    private loginService: ILoginService,
    private router: Router,
    fb: FormBuilder,
    @Inject(AUTH_CONFIG)
    private authConfig: AuthConfig
  ) {
    this.requestEmailChangeForm = fb.group({
      'email': ['', [Validators.required, Validators.pattern(
        this.authConfig.emailValidationRegex || /^.+@.+\..{2,3}$/
      )]],
    });
  }

  async onSubmit() {
    this.loginService.user.pipe(
      first(),
      switchMap((user) => {
        return this.loginService.updateEmail(user.sub, this.requestEmailChangeForm.value.email);
      }),
      tap(() => this.success = true),
    ).subscribe();
  }
}
