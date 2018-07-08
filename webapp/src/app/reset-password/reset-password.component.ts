import { Component, OnInit } from '@angular/core';
import {
  ResetPasswordFormConfig
} from 'ecad-angular-auth';

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
