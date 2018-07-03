import { Component, OnInit } from '@angular/core';
import {
  ResetPasswordFormConfig
} from '../../../projects/ecad-angular-auth/src/lib/ecad-angular-auth-components/reset-password-form/reset-password-form.component';

@Component({
  selector: 'app-reset-password',
  templateUrl: './reset-password.component.html',
  styleUrls: ['./reset-password.component.scss']
})
export class ResetPasswordComponent implements OnInit {

  config: ResetPasswordFormConfig = {
    successUrlRedirect: '/',
  };

  constructor() { }

  ngOnInit() {
  }

}
